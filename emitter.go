package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	Key    emitterKey
	Dumper dumper

	// For fsEmitter
	FsFileName   string
	FsDirPath    string
	FsRotateSize int
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
