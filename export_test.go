//nolint
package main

import "io"

var (
	ParseVXLAN    = parseVXLAN
	ListenVXLAN   = listenVXLAN
	NewPacketData = newPacketData
)

type EmitterArgument emitterArgument
type PacketData packetData
type JSONRecord jsonRecord

func NewEmitter(args EmitterArgument) (recordEmitter, error) {
	return newEmitter(emitterArgument(args))
}

func ToPacketDataSlice(pkt *packetData) []*packetData {
	return []*packetData{pkt}
}

type JsonPacketDumper jsonPacketDumper
type PcapDumper pcapDumper

func (x *JsonPacketDumper) Dump(packets []*packetData, w io.Writer) error {
	return (*jsonPacketDumper)(x).dump(packets, w)
}

func (x *PcapDumper) Open(w io.Writer) error  { return (*pcapDumper)(x).open(w) }
func (x *PcapDumper) Close(w io.Writer) error { return (*pcapDumper)(x).close(w) }
func (x *PcapDumper) Dump(packets []*packetData, w io.Writer) error {
	return (*pcapDumper)(x).dump(packets, w)
}
