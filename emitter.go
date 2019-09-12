package main

import "fmt"

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
	RotateLimit int
}

func newFsEmitter(args emitterArgument) *fsEmitter {
	e := fsEmitter{}
	return &e
}

func (x *fsEmitter) emit(pkt *packetRecord) error {
	return nil
}
