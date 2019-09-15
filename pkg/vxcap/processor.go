package vxcap

import (
	"fmt"
	"time"
)

type Processor interface {
	Setup() error
	Put(pkt *packetData) error
	Tick(now time.Time) error
	Shutdown() error
}

// PacketProcessor controls both of dumper (log enconder) and emitter (log forwarder).
// And it works as interface of log processing by Put() function.
type PacketProcessor struct {
	argument PacketProcessorArgument
	emitter  recordEmitter
	ready    bool
}

// PacketProcessorArgument is argument to construct new PacketProcessor
type PacketProcessorArgument struct {
	DumperArgs  DumperArguments
	EmitterArgs EmitterArguments
}

type emitterModeKey struct {
	Emitter string
	Format  string
	Target  string
}
type emitterParams struct {
	Mode      string
	Extension string
}

var emitterModeMap = map[emitterModeKey]emitterParams{
	{Emitter: "fs", Format: "pcap", Target: "packet"}:       {"stream", "pcap"},
	{Emitter: "fs", Format: "json", Target: "packet"}:       {"stream", "json"},
	{Emitter: "s3", Format: "pcap", Target: "packet"}:       {"stream", "pcap"},
	{Emitter: "s3", Format: "json", Target: "packet"}:       {"stream", "json"},
	{Emitter: "firehose", Format: "json", Target: "packet"}: {"stream", "json"},
}

// NewPacketProcessor is constructor of PacketProcessor. Not only creating instance
// but also setting up emitter and dumper.
func NewPacketProcessor(args PacketProcessorArgument) (*PacketProcessor, error) {
	// Choose emitter mode
	modeKey := emitterModeKey{
		Emitter: args.EmitterArgs.Name,
		Format:  args.DumperArgs.Format,
		Target:  args.DumperArgs.Target,
	}

	params, ok := emitterModeMap[modeKey]
	if !ok {
		return nil, fmt.Errorf("The settings for emitter and dumper are not allowed: %v", modeKey)
	}
	args.EmitterArgs.mode = params.Mode
	args.EmitterArgs.extension = params.Extension

	// construct dumper and emitter
	dumper, err := newDumper(args.DumperArgs)
	if err != nil {
		return nil, err
	}

	args.EmitterArgs.dumper = dumper
	emitter, err := newEmitter(args.EmitterArgs)
	if err != nil {
		return nil, err
	}

	proc := PacketProcessor{
		argument: args,
		emitter:  emitter,
	}

	return &proc, nil
}

// Setup must be invoked before calling Put()
func (x *PacketProcessor) Setup() error {
	if err := x.emitter.setup(); err != nil {
		return err
	}

	x.ready = true
	return nil
}

// Put method input a packet to emitter.
func (x *PacketProcessor) Put(pkt *packetData) error {
	if !x.ready {
		return fmt.Errorf("PacketProcessor is not ready, run Setup() at first")
	}

	if err := x.emitter.emit([]*packetData{pkt}); err != nil {
		return err
	}

	return nil
}

// Tick involves timer handler to manage timeout process.
func (x *PacketProcessor) Tick(now time.Time) error {
	return nil
}

// Shutdown starts closing process of emitter.
func (x *PacketProcessor) Shutdown() error {
	if err := x.emitter.teardown(); err != nil {
		return err
	}

	return nil
}
