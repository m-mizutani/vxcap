//nolint
package vxcap

import "io"

var (
	ParseVXLAN    = parseVXLAN
	ListenVXLAN   = listenVXLAN
	NewPacketData = newPacketData
	NewEmitter    = newEmitter
	NewDumper     = newDumper

	NewJSONPacketDumper = newJSONPacketDumper
)

type PacketData packetData
type JSONRecord jsonRecord

func ToPacketDataSlice(pkt *packetData) []*packetData {
	return []*packetData{pkt}
}

func JSONPacketDumperDump(d dumper, packets []*packetData, w io.Writer) error {
	return d.(*jsonPacketDumper).dump(packets, w)
}

func PcapDumperDump(d dumper, packets []*packetData, w io.Writer) error {
	return d.(*pcapDumper).dump(packets, w)
}
