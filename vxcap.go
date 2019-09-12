package main

import "github.com/pkg/errors"

type vxcap struct {
	RecvPort  int
	QueueSize int

	emitters []emitter
}

func newVxcap() *vxcap {
	cap := vxcap{}
	return &cap
}

func (x *vxcap) addEmitter(e emitter) {
	x.emitters = append(x.emitters, e)
}

func (x *vxcap) start() error {
	for q := range listenVXLAN(x.RecvPort, x.QueueSize) {
		if q.Err != nil {
			return errors.Wrap(q.Err, "Fail to receive UDP")
		}
	}

	return nil
}
