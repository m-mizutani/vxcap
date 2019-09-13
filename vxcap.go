package main

import (
	"github.com/pkg/errors"
)

type vxcap struct {
	RecvPort  int
	QueueSize int
}

func newVxcap() *vxcap {
	cap := vxcap{
		RecvPort:  defaultVxlanPort,
		QueueSize: defaultReceiverQueueSize,
	}
	return &cap
}

func (x *vxcap) start(proc *packetProcessor) error {
	for q := range listenVXLAN(x.RecvPort, x.QueueSize) {
		if q.Err != nil {
			return errors.Wrap(q.Err, "Fail to receive UDP")
		}

		if err := proc.put(q.Pkt); err != nil {
			return errors.Wrap(err, "Fail to handle packet")
		}
	}

	return nil
}
