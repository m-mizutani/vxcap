package main

type packetProcessor struct {
	Argument packetProcessorArgument
	emitter  recordEmitter
}

type packetProcessorArgument struct {
	Format      string // pcap, json
	Destination string // fs, s3, firehose
	DumpBase    string // packet or session
	EmitterArgs emitterArgument
}

func newPacketProcessor(args packetProcessorArgument) (*packetProcessor, error) {

	dumper, err := getDumper(args.DumpBase, args.Format)
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
	if err := x.emitter.emit(pkt); err != nil {
		return err
	}

	return nil
}
