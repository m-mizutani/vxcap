package vxcap

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Logger is logging interface of the pacakge. Basically It's disabled by default.
// But it can be enabled by changing log level by SetLevel for debugging.
var Logger = logrus.New()

func init() {
	Logger.SetLevel(logrus.FatalLevel)
}

// VXCap is one of main components of the package
type VXCap struct {
	RecvPort  int
	QueueSize int
}

// New is constructor of VXCap
func New() *VXCap {
	cap := VXCap{
		RecvPort:  DefaultVxlanPort,
		QueueSize: DefaultReceiverQueueSize,
	}
	return &cap
}

func init() {

}

// Start invokes UDP listener for VXLAN and forward captured packets to processor.
func (x *VXCap) Start(proc *PacketProcessor) error {
	for q := range listenVXLAN(x.RecvPort, x.QueueSize) {
		if q.Err != nil {
			return errors.Wrap(q.Err, "Fail to receive UDP")
		}

		if err := proc.Put(q.Pkt); err != nil {
			return errors.Wrap(err, "Fail to handle packet")
		}
	}

	return nil
}
