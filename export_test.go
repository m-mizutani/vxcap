//nolint
package main

var (
	ParseVXLAN    = parseVXLAN
	ListenVXLAN   = listenVXLAN
	NewPacketData = newPacketData
	DumpPcap      = dumpPcap
	DumpJSON      = dumpJSON
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
