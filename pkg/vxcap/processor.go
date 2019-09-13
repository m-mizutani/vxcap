package vxcap

import "fmt"

// PacketProcessor controls both of dumper (log enconder) and emitter (log forwarder).
// And it works as interface of log processing by Put() function.
type PacketProcessor struct {
	argument PacketProcessorArgument
	emitter  recordEmitter
}

// PacketProcessorArgument is argument to construct new PacketProcessor
type PacketProcessorArgument struct {
	DumperKey   dumperKey
	EmitterArgs emitterArgument
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
	{Emitter: "fs", Format: "pcap", Target: "packet"}: {"stream", "pcap"},
	{Emitter: "fs", Format: "json", Target: "packet"}: {"stream", "json"},
	{Emitter: "s3", Format: "pcap", Target: "packet"}: {"stream", "pcap"},
	{Emitter: "s3", Format: "json", Target: "packet"}: {"stream", "json"},
}

// NewPacketProcessor is constructor of PacketProcessor. Not only creating instance
// but also setting up emitter and dumper.
func NewPacketProcessor(args PacketProcessorArgument) (*PacketProcessor, error) {
	// Choose emitter mode
	modeKey := emitterModeKey{
		Emitter: args.EmitterArgs.Key.Name,
		Format:  args.DumperKey.Format,
		Target:  args.DumperKey.Target,
	}

	params, ok := emitterModeMap[modeKey]
	if !ok {
		return nil, fmt.Errorf("The settings for emitter and dumper are not allowed: %v", modeKey)
	}
	args.EmitterArgs.Key.Mode = params.Mode
	args.EmitterArgs.Extension = params.Extension

	// construct dumper and emitter
	dumper, err := getDumper(args.DumperKey)
	if err != nil {
		return nil, err
	}

	args.EmitterArgs.Dumper = dumper
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

// Put method input a packet to emitter.
func (x *PacketProcessor) Put(pkt *packetData) error {
	if err := x.emitter.emit([]*packetData{pkt}); err != nil {
		return err
	}

	return nil
}
