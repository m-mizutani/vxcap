package main

import (
	"fmt"

	"github.com/pkg/errors"
)

type vxcap struct {
	RecvPort  int
	QueueSize int

	emitters []recordEmitter
}

func newVxcap() *vxcap {
	cap := vxcap{
		RecvPort:  defaultVxlanPort,
		QueueSize: defaultReceiverQueueSize,
	}
	return &cap
}

func (x *vxcap) addEmitter(e recordEmitter) {
	x.emitters = append(x.emitters, e)
}

func (x *vxcap) selfcheck() error {
	if len(x.emitters) == 0 {
		return fmt.Errorf("No emitter is found")
	}

	for _, emitter := range x.emitters {
		if emitter.getDumper() == nil {
			return fmt.Errorf("Found emitter having no dumper")
		}
	}

	return nil
}

func (x *vxcap) start() error {
	if err := x.selfcheck(); err != nil {
		return err
	}

	for q := range listenVXLAN(x.RecvPort, x.QueueSize) {
		if q.Err != nil {
			return errors.Wrap(q.Err, "Fail to receive UDP")
		}

		for _, emitter := range x.emitters {
			if err := emitter.emit(q.Pkt); err != nil {
				return errors.Wrap(err, "Fail to emit packet")
			}
		}
	}

	return nil
}
