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
	NewPcapDumper       = newPcapDumper
)

type PacketData packetData
type JSONRecord jsonRecord

func ToPacketDataSlice(pkt *packetData) []*packetData {
	return []*packetData{pkt}
}

func JSONPacketDumperDump(d dumper, packets []*packetData, w io.Writer) error {
	if err := d.(*jsonPacketDumper).open(w); err != nil {
		return err
	}
	if err := d.(*jsonPacketDumper).dump(packets, w); err != nil {
		return err
	}
	if err := d.(*jsonPacketDumper).close(w); err != nil {
		return err
	}
	return nil
}

func PcapDumperDump(d dumper, packets []*packetData, w io.Writer) error {
	if err := d.(*pcapDumper).open(w); err != nil {
		return err
	}
	if err := d.(*pcapDumper).dump(packets, w); err != nil {
		return err
	}
	if err := d.(*pcapDumper).close(w); err != nil {
		return err
	}
	return nil
}
