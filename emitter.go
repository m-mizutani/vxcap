package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type recordEmitter interface {
	emit(*packetRecord) error
	setDumper(dumpRecord)
	getDumper() dumpRecord
}

type emitterArgument struct {
	Name string

	// For fsEmitter
	FsFileName   string
	FsDirPath    string
	FsRotateSize int
}

type baseEmitter struct {
	Dump dumpRecord
}

func (x *baseEmitter) setDumper(f dumpRecord) {
	x.Dump = f
}

func (x *baseEmitter) getDumper() dumpRecord {
	return x.Dump
}

func newEmitter(args emitterArgument) (recordEmitter, error) {
	switch args.Name {
	case "fs":
		return newFsEmitter(args), nil
	default:
		return nil, fmt.Errorf("Invalid emitter name: %s", args.Name)
	}
}

type fsEmitter struct {
	baseEmitter
	dump        dumpRecord
	Argument    emitterArgument
	RotateLimit int
	FlushSize   int
	PktBuffer   []*packetRecord
}

func newFsEmitter(args emitterArgument) *fsEmitter {
	e := fsEmitter{}
	return &e
}

func (x *fsEmitter) emit(pkt *packetRecord) error {
	x.PktBuffer = append(x.PktBuffer, pkt)

	if len(x.PktBuffer) > x.FlushSize {
		fd, err := os.Create(filepath.Join(x.Argument.FsDirPath, x.Argument.FsFileName))
		if err != nil {
			return errors.Wrap(err, "Fail to create a dump file for emitter")
		}
		defer fd.Close()

		if err := x.dump(x.PktBuffer, fd); err != nil {
			return err
		}

		x.PktBuffer = []*packetRecord{}
	}

	return nil
}
