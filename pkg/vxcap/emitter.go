package vxcap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type recordEmitter interface {
	setup() error
	emit([]*packetData) error
	teardown() error
	tick(t time.Time) error
	setDumper(dumper)
	getDumper() dumper
}

type emitterKey struct {
	Name string
	Mode string // batch or stream
}
type emitterConstructor func(EmitterArguments) (recordEmitter, error)

// EmitterArguments is for construction of emitter
type EmitterArguments struct {
	Name string
	mode string // batch or stream, the field should be set by PacketProcessor

	dumper    dumper
	extension string

	// For fsEmitter
	FsFileName   string
	FsDirPath    string
	FsRotateSize int

	// For aws service
	AwsRegion string

	// For s3Emitter
	AwsS3Bucket        string
	AwsS3Prefix        string
	AwsS3AddTimeKey    bool
	AwsS3FlushCount    int
	AwsS3FlushInterval int

	// For firehoseEmitter
	AwsFirehoseName          string
	AwsFirehoseFlushSize     int
	AwsFirehoseFlushInterval int
}

const (
	// DefaultAwsS3FlushCount is limit of data buffer for S3 emitter.
	DefaultAwsS3FlushCount = 4096
	// DefaultAwsS3FlushInterval is seconds of interval to flush data for S3 emitter
	DefaultAwsS3FlushInterval = 300

	// DefaultAwsFirehoseFlushSize is threshold of flush to firehose.
	DefaultAwsFirehoseFlushSize = 2 * 1024 * 1024 // 2MB
	// DefaultAwsFIrehoseFlushInterval is seconds of interval to flush data for Firehose emitter
	DefaultAwsFIrehoseFlushInterval = 300
)

type baseEmitter struct {
	Dumper dumper
}

func (x *baseEmitter) setDumper(f dumper)       { x.Dumper = f }
func (x *baseEmitter) getDumper() dumper        { return x.Dumper }
func (x *baseEmitter) setup() error             { return nil }
func (x *baseEmitter) teardown() error          { return nil }
func (x *baseEmitter) tick(now time.Time) error { return nil }

func newEmitter(args EmitterArguments) (recordEmitter, error) {
	emitterMap := map[emitterKey]emitterConstructor{
		{Name: "s3", Mode: "stream"}:       newS3StreamEmitter,
		{Name: "fs", Mode: "batch"}:        newFsBatchEmitter,
		{Name: "fs", Mode: "stream"}:       newFsStreamEmitter,
		{Name: "firehose", Mode: "stream"}: newFirehoseEmitter,
	}

	key := emitterKey{
		Name: args.Name,
		Mode: args.mode,
	}
	constructor, ok := emitterMap[key]
	if !ok {
		return nil, fmt.Errorf("The pair is not supported: %v", key)
	}

	emitter, err := constructor(args)
	if err != nil {
		return nil, err
	}

	if args.dumper == nil {
		return nil, fmt.Errorf("No Dumper. Dumper is required for new emitter")
	}

	emitter.setDumper(args.dumper)
	return emitter, nil
}

type fsBatchEmitter struct {
	baseEmitter
	Argument EmitterArguments
}

func newFsBatchEmitter(args EmitterArguments) (recordEmitter, error) {
	e := fsBatchEmitter{Argument: args}
	return &e, nil
}

func (x *fsBatchEmitter) emit(pkt []*packetData) error {

	fd, err := os.Create(filepath.Join(x.Argument.FsDirPath, x.Argument.FsFileName))
	if err != nil {
		return errors.Wrap(err, "Fail to create a dump file for emitter")
	}
	defer fd.Close()

	if err := x.Dumper.open(fd); err != nil {
		return err
	}
	if err := x.Dumper.dump(pkt, fd); err != nil {
		return err
	}
	if err := x.Dumper.close(fd); err != nil {
		return err
	}

	return nil
}

type fsStreamEmitter struct {
	baseEmitter
	Argument    EmitterArguments
	DirPath     string
	FileName    string
	RotateLimit int
	fd          *os.File
}

func newFsStreamEmitter(args EmitterArguments) (recordEmitter, error) {
	emitter := fsStreamEmitter{
		Argument: args,
		DirPath:  ".",
		FileName: "dump." + args.extension,
	}

	if args.FsDirPath != "" {
		emitter.DirPath = args.FsDirPath
	}
	if args.FsFileName != "" {
		emitter.FileName = args.FsFileName
	}

	return &emitter, nil
}

func (x *fsStreamEmitter) emit(packets []*packetData) error {
	if x.fd == nil {
		fd, err := os.Create(filepath.Join(x.DirPath, x.FileName))
		if err != nil {
			return errors.Wrap(err, "Fail to create a dump file for emitter")
		}
		x.fd = fd

		if err := x.Dumper.open(x.fd); err != nil {
			return err
		}
	}

	if err := x.Dumper.dump(packets, x.fd); err != nil {
		return err
	}
	return nil
}

func (x *fsStreamEmitter) teardown() error {
	defer x.fd.Close()

	if x.fd != nil {
		if err := x.Dumper.close(x.fd); err != nil {
			return err
		}
	}
	return nil
}

type s3StreamEmitter struct {
	baseEmitter
	Argument      EmitterArguments
	pktBuffer     []*packetData
	flushCount    int
	flushInterval int
	lastFlush     time.Time
}

func newS3StreamEmitter(args EmitterArguments) (recordEmitter, error) {
	if args.AwsRegion == "" {
		return nil, fmt.Errorf("AwsRegion is not set for S3 emitter")
	}
	if args.AwsS3Bucket == "" {
		return nil, fmt.Errorf("AwsS3Bucket is not set for S3 emitter")
	}

	emitter := s3StreamEmitter{
		Argument:      args,
		flushCount:    DefaultAwsS3FlushCount,
		flushInterval: DefaultAwsS3FlushInterval,
		lastFlush:     time.Now(),
	}

	if args.AwsS3FlushCount > 0 {
		emitter.flushCount = args.AwsS3FlushCount
	}
	if args.AwsS3FlushInterval > 0 {
		emitter.flushInterval = args.AwsS3FlushInterval
	}
	return &emitter, nil
}

func (x *s3StreamEmitter) flush() error {
	x.lastFlush = time.Now()
	if len(x.pktBuffer) == 0 {
		return nil
	}

	reader, writer := io.Pipe()
	errCh := make(chan error)

	go func() {
		defer writer.Close()
		defer close(errCh)

		if err := x.Dumper.open(writer); err != nil {
			errCh <- errors.Wrap(err, "Fail to open dumper for S3 object")
		}
		if err := x.Dumper.dump(x.pktBuffer, writer); err != nil {
			errCh <- errors.Wrap(err, "Fail to dump packets for S3 object")
		}
		if err := x.Dumper.close(writer); err != nil {
			errCh <- errors.Wrap(err, "Fail to close dumper for S3 object")
		}

		x.pktBuffer = []*packetData{}
	}()

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.Argument.AwsRegion),
	}))

	s3Key := x.Argument.AwsS3Prefix
	now := time.Now().UTC()
	if x.Argument.AwsS3AddTimeKey {
		s3Key += now.Format("2006/01/02/15/")
	}
	s3Key += now.Format("20160102_150405_") +
		strings.Replace(uuid.New().String(), "-", "", -1) + "." +
		x.Argument.extension

	uploader := s3manager.NewUploader(ssn)
	resp, err := uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: &x.Argument.AwsS3Bucket,
		Key:    &s3Key,
	})

	if err != nil {
		return errors.Wrap(err, "Fail to PutObject in Emitter")
	}

	if err := <-errCh; err != nil {
		return err
	}

	Logger.WithField("resp", resp).Debug("Uploaded S3 object")

	return nil
}

func (x *s3StreamEmitter) emit(packets []*packetData) error {
	x.pktBuffer = append(x.pktBuffer, packets...)

	if len(x.pktBuffer) >= x.Argument.AwsS3FlushCount {
		if err := x.flush(); err != nil {
			return errors.Wrap(err, "Fail to upload object to S3")
		}
	}

	return nil
}

func (x *s3StreamEmitter) teardown() error {
	if err := x.flush(); err != nil {
		return errors.Wrap(err, "Fail to upload object to S3 in closing")
	}

	return nil
}

func (x *s3StreamEmitter) tick(now time.Time) error {
	if now.Sub(x.lastFlush) > time.Second*time.Duration(x.flushInterval) {
		if err := x.flush(); err != nil {
			return err
		}
	}
	return nil
}

type vxcapFirehoseClient interface {
	PutRecordBatch(*firehose.PutRecordBatchInput) (*firehose.PutRecordBatchOutput, error)
}

var newFirehoseClient = func(awsRegion string) vxcapFirehoseClient {
	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	}))

	client := firehose.New(ssn)

	return client
}

type firehoseEmitter struct {
	baseEmitter
	Argument       EmitterArguments
	firehoseClient vxcapFirehoseClient
	pktBuffer      [][]byte
	pktBufferSize  int
	flushSize      int
	flushInterval  int
	lastFlush      time.Time
}

func newFirehoseEmitter(args EmitterArguments) (recordEmitter, error) {
	e := firehoseEmitter{
		Argument:      args,
		flushSize:     DefaultAwsFirehoseFlushSize,
		flushInterval: DefaultAwsFIrehoseFlushInterval,
		lastFlush:     time.Now(),
	}

	if args.AwsFirehoseFlushSize > 0 {
		e.flushSize = args.AwsFirehoseFlushSize
	}
	if args.AwsFirehoseFlushInterval > 0 {
		e.flushInterval = args.AwsFirehoseFlushInterval
	}

	return &e, nil
}

func (x *firehoseEmitter) flush() error {
	x.lastFlush = time.Now()
	if len(x.pktBuffer) == 0 {
		return nil
	}

	recordsBatchInput := &firehose.PutRecordBatchInput{}
	recordsBatchInput = recordsBatchInput.SetDeliveryStreamName(x.Argument.AwsFirehoseName)

	records := []*firehose.Record{}

	for _, buf := range x.pktBuffer {
		record := &firehose.Record{Data: buf}
		records = append(records, record)
	}

	recordsBatchInput = recordsBatchInput.SetRecords(records)

	resp, err := x.firehoseClient.PutRecordBatch(recordsBatchInput)
	if err != nil {
		return errors.Wrap(err, "Fail to put firehose records")
	}

	Logger.WithField("resp", resp).Debug("Done Firehose PutRecordBatch")

	x.pktBuffer = [][]byte{}
	x.pktBufferSize = 0

	return nil
}

func (x *firehoseEmitter) setup() error {
	x.firehoseClient = newFirehoseClient(x.Argument.AwsRegion)
	return nil
}

func (x *firehoseEmitter) emit(pkt []*packetData) error {
	for _, p := range pkt {
		buf := new(bytes.Buffer)
		if err := x.Dumper.dump([]*packetData{p}, buf); err != nil {
			return errors.Wrap(err, "Fail to encode data for firehose record")
		}

		raw := buf.Bytes()
		x.pktBuffer = append(x.pktBuffer, raw)
		x.pktBufferSize += len(raw)

		if x.pktBufferSize >= x.flushSize {
			if err := x.flush(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (x *firehoseEmitter) teardown() error {
	if err := x.flush(); err != nil {
		return err
	}

	return nil
}

func (x *firehoseEmitter) tick(now time.Time) error {
	fmt.Println("tick")
	if now.Sub(x.lastFlush) > time.Second*time.Duration(x.flushInterval) {
		if err := x.flush(); err != nil {
			return err
		}
	}
	return nil
}
