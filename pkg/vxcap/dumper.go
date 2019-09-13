package vxcap

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
	"honnef.co/go/pcap"
)

type dumper interface {
	open(io.Writer) error
	dump([]*packetData, io.Writer) error
	close(io.Writer) error
}

type dumperKey struct {
	Format string
	Target string // packet or session
}

func getDumper(key dumperKey) (dumper, error) {
	dumperMap := map[dumperKey]dumper{
		{Format: "json", Target: "packet"}: &jsonPacketDumper{},
		{Format: "pcap", Target: "packet"}: &pcapDumper{},
	}

	d, ok := dumperMap[key]
	if !ok {
		return nil, fmt.Errorf("The pair is not supported: %v", key)
	}

	return d, nil
}

type baseDumper struct{}

func (x *baseDumper) open(io.Writer) error  { return nil }
func (x *baseDumper) close(io.Writer) error { return nil }

type jsonPacketDumper struct {
	baseDumper
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

func (x *jsonPacketDumper) dump(packets []*packetData, w io.Writer) error {
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

// pcapDumper is not concurrency safe for now
type pcapDumper struct {
	writer     *pcap.Writer
	baseDumper //nolint
}

type pcapPayload []byte

func (x pcapPayload) Payload() []byte {
	return x
}

func (x *pcapDumper) open(writer io.Writer) error {
	w := pcap.NewWriter(writer)
	w.Header.Network = pcap.DLT_EN10MB
	if err := w.WriteHeader(); err != nil {
		return errors.Wrap(err, "Fail to write header of pcap")
	}
	x.writer = w
	return nil
}

func (x *pcapDumper) close(writer io.Writer) error {
	x.writer = nil
	return nil
}

func (x *pcapDumper) dump(packets []*packetData, writer io.Writer) error {
	if x.writer == nil {
		return fmt.Errorf("pcapDumper.writer is not set, assertion error")
	}

	for _, pkt := range packets {
		p := pcapPayload(pkt.Data)
		pcapPkt := pcap.Packet{
			// Specify a timestamp
			Header: pcap.PacketHeader{Timestamp: pkt.Timestamp},
			Data:   p,
		}

		if err := x.writer.WritePacket(pcapPkt); err != nil {
			return errors.Wrap(err, "Fail to write pcap data")
		}
	}

	return nil
}

/*
func dumpGzipJSON([]*packetData, io.writer) error {}
func dumpParquet([]*packetData, io.writer) error {}
*/
