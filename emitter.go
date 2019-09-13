package main

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
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type recordEmitter interface {
	emit([]*packetData) error
	close() error
	setDumper(dumper)
	getDumper() dumper
}

type emitterKey struct {
	Name string
	Mode string // batch or stream
}
type emitterConstructor func(emitterArgument) recordEmitter

type emitterArgument struct {
	Key       emitterKey
	Dumper    dumper
	Extension string

	// For fsEmitter
	FsFileName   string
	FsDirPath    string
	FsRotateSize int

	// FOr s3Emitter
	AwsRegion       string
	AwsS3Bucket     string
	AwsS3Prefix     string
	AwsAddTimeKey   bool
	AwsS3FlushCount int
}

type baseEmitter struct {
	Dumper dumper
}

func (x *baseEmitter) setDumper(f dumper) {
	x.Dumper = f
}

func (x *baseEmitter) getDumper() dumper {
	return x.Dumper
}

func newEmitter(args emitterArgument) (recordEmitter, error) {
	emitterMap := map[emitterKey]emitterConstructor{
		{Name: "s3", Mode: "stream"}: newS3StreamEmitter,
		{Name: "fs", Mode: "batch"}:  newFsBatchEmitter,
		{Name: "fs", Mode: "stream"}: newFsStreamEmitter,
	}

	constructor, ok := emitterMap[args.Key]
	if !ok {
		return nil, fmt.Errorf("The pair is not supported: %v", args.Key)
	}

	emitter := constructor(args)

	if args.Dumper == nil {
		return nil, fmt.Errorf("No Dumper. Dumper is required for new emitter")
	}

	emitter.setDumper(args.Dumper)
	return emitter, nil
}

type fsBatchEmitter struct {
	baseEmitter
	Argument emitterArgument
}

func newFsBatchEmitter(args emitterArgument) recordEmitter {
	e := fsBatchEmitter{Argument: args}
	return &e
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

func (x *fsBatchEmitter) close() error {
	return nil
}

type fsStreamEmitter struct {
	baseEmitter
	Argument    emitterArgument
	RotateLimit int
	fd          *os.File
}

func newFsStreamEmitter(args emitterArgument) recordEmitter {
	e := fsStreamEmitter{Argument: args}
	return &e
}

func (x *fsStreamEmitter) emit(packets []*packetData) error {
	if x.fd == nil {
		fd, err := os.Create(filepath.Join(x.Argument.FsDirPath, x.Argument.FsFileName))
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

func (x *fsStreamEmitter) close() error {
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
	Argument  emitterArgument
	pktBuffer []*packetData
}

func newS3StreamEmitter(args emitterArgument) recordEmitter {
	e := s3StreamEmitter{Argument: args}
	return &e
}

func (x *s3StreamEmitter) flush() error {
	var buf bytes.Buffer

	reader, writer := io.Pipe()
	errCh := make(chan error)

	go func() {
		defer writer.Close()
		defer close(errCh)

		if err := x.Dumper.open(&buf); err != nil {
			errCh <- errors.Wrap(err, "Fail to open dumper for S3 object")
		}
		if err := x.Dumper.dump(x.pktBuffer, &buf); err != nil {
			errCh <- errors.Wrap(err, "Fail to dump packets for S3 object")
		}
		if err := x.Dumper.close(&buf); err != nil {
			errCh <- errors.Wrap(err, "Fail to close dumper for S3 object")
		}
	}()

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(x.Argument.AwsRegion),
	}))

	s3Key := x.Argument.AwsS3Prefix
	now := time.Now().UTC()
	if x.Argument.AwsAddTimeKey {
		s3Key += now.Format("2006/01/02/15/")
	}
	s3Key += now.Format("20160102_150405_") +
		strings.Replace(uuid.New().String(), "-", "_", -1) + "." +
		x.Argument.Extension

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

	logger.WithField("resp", resp).Debug("Uploaded S3 object")

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

func (x *s3StreamEmitter) close() error {
	if err := x.flush(); err != nil {
		return errors.Wrap(err, "Fail to upload object to S3 in closing")
	}

	return nil
}
