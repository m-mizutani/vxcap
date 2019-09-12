package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"honnef.co/go/pcap"
)

type dumpRecord func([]*packetData, io.Writer) error

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

func dumpJSON(packets []*packetData, w io.Writer) error {
	for _, pkt := range packets {
		var record jsonRecord
		if netLayer := (*pkt.Packet).NetworkLayer(); netLayer != nil {
			netFlow := netLayer.NetworkFlow()
			src, dst := netFlow.Endpoints()
			record.SrcAddr = src.String()
			record.DstAddr = dst.String()

			if ipv4, ok := netLayer.(*layers.IPv4); ok {
				record.Protocol = ipv4.Protocol.String()
			} else if ipv6, ok := netLayer.(*layers.IPv6); ok {
				record.Protocol = ipv6.NextHeader.String()
			}
		}

		if tpLayer := (*pkt.Packet).TransportLayer(); tpLayer != nil {
			tpFlow := tpLayer.TransportFlow()
			src, dst := tpFlow.Endpoints()
			if n, err := strconv.Atoi(src.String()); err == nil {
				record.SrcPort = n
			}
			if n, err := strconv.Atoi(dst.String()); err == nil {
				record.DstPort = n
			}
		}

		if app := (*pkt.Packet).ApplicationLayer(); app != nil {
			record.RawData = app.Payload()
			record.TextData = string(app.Payload())
		}

		data, err := json.Marshal(&record)
		if err != nil {
			return errors.Wrap(err, "Fail to marshal jsonRecord")
		}

		if _, err := w.Write(data); err != nil {
			return errors.Wrap(err, "Fail to write JSON data")
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return errors.Wrap(err, "Fail to write JSON data (LF)")
		}
	}

	return nil
}

type pcapPayload []byte

func (x pcapPayload) Payload() []byte {
	return x
}

func dumpPcap(packets []*packetData, writer io.Writer) error {
	w := pcap.NewWriter(writer)
	w.Header.Network = pcap.DLT_EN10MB
	if err := w.WriteHeader(); err != nil {
		return errors.Wrap(err, "Fail to write header of pcap")
	}

	for _, pkt := range packets {
		p := pcapPayload(pkt.Data)
		pcapPkt := pcap.Packet{
			// Specify a timestamp
			Header: pcap.PacketHeader{Timestamp: pkt.Timestamp},
			Data:   p,
		}

		if err := w.WritePacket(pcapPkt); err != nil {
			return errors.Wrap(err, "Fail to write pcap data")
		}
	}

	return nil
}

/*
func dumpGzipJSON([]*packetData, io.writer) error {}
func dumpParquet([]*packetData, io.writer) error {}
*/
