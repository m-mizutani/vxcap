package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"honnef.co/go/pcap"
)

type dumpRecord func([]*packetRecord, io.Writer) error

func getDumper(name string) (dumpRecord, error) {
	switch name {
	case "json":
		return dumpJSON, nil
	case "pcap":
		return dumpPcap, nil
	default:
		return nil, fmt.Errorf("Invalid dumpMethod name: %s", name)
	}
}

type jsonRecord struct {
	// Five tuple
	Protocol string `json:"proto"`
	SrcAddr  string `json:"src_addr"`
	DstAddr  string `json:"dst_addr"`
	SrcPort  int    `json:"src_port,omitempty"`
	DstPort  int    `json:"dst_port,omitempty"`

	// TCP
	TCPFlag string `json:"tcp_flag,omitempty"`
	TCPSeq  uint32 `json:"tcp_seq,omitempty"`

	// Data part
	TextData string `json:"text_data,omitempty"`
	RawData  []byte `json:"raw_data,omitempty"`
}

func dumpJSON(packets []*packetRecord, w io.Writer) error {
	for _, pkt := range packets {
		var record jsonRecord
		netLayer := (*pkt.Packet).NetworkLayer()

		netFlow := netLayer.NetworkFlow()
		src, dst := netFlow.Endpoints()

		record.SrcAddr = src.String()
		record.DstAddr = dst.String()

		data, err := json.Marshal(&record)
		if err != nil {
			return errors.Wrap(err, "Fail to marshal jsonRecord")
		}

		if _, err := w.Write(data); err != nil {
			return errors.Wrap(err, "Fail to write JSON data")
		}
	}

	return nil
}

type pcapPayload []byte

func (x pcapPayload) Payload() []byte {
	return x
}

func dumpPcap(packets []*packetRecord, writer io.Writer) error {
	w := pcap.NewWriter(writer)
	w.Header.Network = pcap.DLT_EN10MB
	w.WriteHeader()

	for _, pkt := range packets {
		p := pcapPayload(pkt.Data)
		w.WritePacket(pcap.Packet{
			// Specify a timestamp
			Header: pcap.PacketHeader{Timestamp: pkt.Timestamp},
			Data:   p,
		})
	}

	return nil
}

/*
func dumpGzipJSON([]*packetRecord, io.writer) error {}
func dumpParquet([]*packetRecord, io.writer) error {}
*/
