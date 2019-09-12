package main

import "time"

type vxlanHeader struct {
	Flag               uint16
	GroupPolicyID      uint16
	NetworkIndentifier [3]byte
	Reserved           [1]byte
}

type packetRecord struct {
	Data      []byte
	Header    vxlanHeader
	Timestamp time.Time
}
