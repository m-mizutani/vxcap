package main

//nolint
var (
	ParseVXLAN      = parseVXLAN
	ListenVXLAN     = listenVXLAN
	NewPacketRecord = newPacketRecord
	DumpPcap        = dumpPcap
	DumpJSON        = dumpJSON
)

// nolint
type EmitterArgument emitterArgument
type PacketRecord packetRecord
type JSONRecord jsonRecord

// nolint
func NewEmitter(args EmitterArgument) (recordEmitter, error) {
	return newEmitter(emitterArgument(args))
}

// nolint
func ToPacketRecordSlice(pkt *packetRecord) []*packetRecord {
	return []*packetRecord{pkt}
}
