package main

import "fmt"

type packetProcessor struct {
	Argument packetProcessorArgument
	emitter  recordEmitter
}

type packetProcessorArgument struct {
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

func newPacketProcessor(args packetProcessorArgument) (*packetProcessor, error) {
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

	proc := packetProcessor{
		Argument: args,
		emitter:  emitter,
	}

	return &proc, nil
}

func (x *packetProcessor) put(pkt *packetData) error {
	if err := x.emitter.emit([]*packetData{pkt}); err != nil {
		return err
	}

	return nil
}
