package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/pkg/errors"
)

type queue struct {
	Pkt *packetData
	Err error
}

const (
	defaultReceiverQueueSize = 1024
	defaultVxlanPort         = 4789

	vxlanHeaderLength = 8
)

func parseVXLAN(raw []byte, length int) (*packetData, error) {
	if length < vxlanHeaderLength {
		return nil, fmt.Errorf("Too short data for VXLAN header: %d", length)
	}

	pkt := newPacketData(raw[vxlanHeaderLength:length])

	buffer := bytes.NewBuffer(raw)
	if err := binary.Read(buffer, binary.BigEndian, &pkt.Header); err != nil {
		return nil, errors.Wrap(err, "Fail to parse VXLAN header")
	}

	return pkt, nil
}

func listenVXLAN(port, queueSize int) chan *queue {
	ch := make(chan *queue, queueSize)

	go func() {
		defer close(ch)

		sock, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
		if err != nil {
			ch <- &queue{Err: errors.Wrap(err, "Fail to create UDP socket")}
			return
		}
		defer sock.Close()

		buf := make([]byte, 32768)

		for {
			n, _, err := sock.ReadFrom(buf)
			if err != nil {
				ch <- &queue{Err: errors.Wrap(err, "Fail to read UDP data")}
				return
			}

			pkt, err := parseVXLAN(buf, n)
			if err != nil {
				logger.WithError(err).Warn("Fail to parse VXLAN data")
				continue
			}

			q := new(queue)
			q.Pkt = pkt
			ch <- q
		}
	}()

	return ch
}
