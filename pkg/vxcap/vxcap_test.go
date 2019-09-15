package vxcap_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/m-mizutani/vxcap/pkg/vxcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DummyProcessor struct {
	tickCount      int
	calledShutdown bool
	vxcap.PacketProcessor
}

func (x *DummyProcessor) Tick(now time.Time) error {
	x.tickCount += 1
	return nil
}

func (x *DummyProcessor) Shutdown() error {
	x.calledShutdown = true
	return nil
}

func TestVxcapTimerAndSignal(t *testing.T) {
	dummy := DummyProcessor{}
	cap := vxcap.New()

	go func() {
		time.Sleep(2 * time.Second)

		proc, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)
		err = proc.Signal(syscall.SIGTERM)
		require.NoError(t, err)
	}()

	err := cap.Start(&dummy)
	require.NoError(t, err)
	assert.True(t, dummy.calledShutdown)
	assert.NotEqual(t, 0, dummy.tickCount)
}
