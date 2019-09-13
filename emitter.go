package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type recordEmitter interface {
	emit(*packetData) error
	close() error
	setDumper(dumper)
	getDumper() dumper
}

type emitterOutputMode int

const (
	emitterOutputStream emitterOutputMode = iota
	emitterOutputBatch  emitterOutputMode = iota
)

type emitterArgument struct {
	Name       string
	OutputMode emitterOutputMode
	Dumper     dumper

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
	var emitter recordEmitter
	switch args.Name {
	case "fs":
		emitter = newFsBatchEmitter(args)
	default:
		return nil, fmt.Errorf("Invalid emitter name: %s", args.Name)
	}

	if args.Dumper == nil {
		return nil, fmt.Errorf("No Dumper. Dumper is required for new emitter")
	}

	emitter.setDumper(args.Dumper)
	return emitter, nil
}

type fsBatchEmitter struct {
	baseEmitter
	Argument    emitterArgument
	RotateLimit int
	FlushSize   int
	PktBuffer   []*packetData
}

func newFsBatchEmitter(args emitterArgument) *fsBatchEmitter {
	e := fsBatchEmitter{Argument: args}
	return &e
}

func (x *fsBatchEmitter) emit(pkt *packetData) error {
	x.PktBuffer = append(x.PktBuffer, pkt)

	if len(x.PktBuffer) > x.FlushSize {
		fd, err := os.Create(filepath.Join(x.Argument.FsDirPath, x.Argument.FsFileName))
		if err != nil {
			return errors.Wrap(err, "Fail to create a dump file for emitter")
		}
		defer fd.Close()

		if err := x.Dumper.open(fd); err != nil {
			return err
		}
		if err := x.Dumper.dump(x.PktBuffer, fd); err != nil {
			return err
		}
		if err := x.Dumper.close(fd); err != nil {
			return err
		}

		x.PktBuffer = []*packetData{}
	}

	return nil
}

func (x *fsBatchEmitter) close() error {
	return nil
}
