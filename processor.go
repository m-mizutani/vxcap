package main

type packetProcessor struct {
	Argument packetProcessorArgument
	emitter  recordEmitter
}

type packetProcessorArgument struct {
	DumperKey   dumperKey
	EmitterArgs emitterArgument
}

func newPacketProcessor(args packetProcessorArgument) (*packetProcessor, error) {
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
