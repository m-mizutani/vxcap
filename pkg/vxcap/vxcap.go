package vxcap

import (
	"os"
	"os/signal"
	"syscall"
	"time"

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

// Start invokes UDP listener for VXLAN and forward captured packets to processor.
func (x *VXCap) Start(proc Processor) error {
	// Setup channels
	queueCh := listenVXLAN(x.RecvPort, x.QueueSize)
	tickerCh := time.Tick(time.Second)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM)
	defer signal.Stop(signalCh)

MainLoop:
	for {
		select {
		case q := <-queueCh:
			if q.Err != nil {
				return errors.Wrap(q.Err, "Fail to receive UDP")
			}

			if err := proc.Put(q.Pkt); err != nil {
				return errors.Wrap(err, "Fail to handle packet")
			}

		case t := <-tickerCh:
			if err := proc.Tick(t); err != nil {
				return errors.Wrap(err, "Fail in tick process")
			}

		case s := <-signalCh:
			Logger.WithField("signal", s).Warn("Caught signal (should be SIGTERM")
			if err := proc.Shutdown(); err != nil {
				return errors.Wrap(err, "Fail in shutdown process")
			}
			break MainLoop
		}
	}

	return nil
}
