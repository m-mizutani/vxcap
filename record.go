package main

import (
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type vxlanHeader struct {
	Flag               uint16
	GroupPolicyID      uint16
	NetworkIndentifier [3]byte
	Reserved           [1]byte
}

type packetRecord struct {
	Data      []byte
	Packet    *gopacket.Packet
	Header    vxlanHeader
	Timestamp time.Time
}

func newPacketRecord(buf []byte, length int) *packetRecord {
	pkt := new(packetRecord)
	pkt.Timestamp = time.Now()

	pkt.Data = make([]byte, length)
	copy(pkt.Data, buf)

	gopkt := gopacket.NewPacket(pkt.Data, layers.LayerTypeEthernet, gopacket.Lazy)
	pkt.Packet = &gopkt

	return pkt
}
