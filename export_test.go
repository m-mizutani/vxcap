//nolint
package main

var (
	ParseVXLAN      = parseVXLAN
	ListenVXLAN     = listenVXLAN
	NewPacketRecord = newPacketRecord
	DumpPcap        = dumpPcap
	DumpJSON        = dumpJSON
)

type EmitterArgument emitterArgument
type PacketRecord packetRecord
type JSONRecord jsonRecord

func NewEmitter(args EmitterArgument) (recordEmitter, error) {
	return newEmitter(emitterArgument(args))
}

func ToPacketRecordSlice(pkt *packetRecord) []*packetRecord {
	return []*packetRecord{pkt}
}
