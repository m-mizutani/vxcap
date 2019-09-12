package main

type emitter interface {
	emit(*packetRecord) error
	setWriter(*recordWriter)
}

type fsEmitter struct {
	RotateLimit int
	writer      *recordWriter
}

func newFsEmitter() *fsEmitter {
	e := fsEmitter{}
	return &e
}

func (x *fsEmitter) emit(pkt *packetRecord) error {
	return nil
}
