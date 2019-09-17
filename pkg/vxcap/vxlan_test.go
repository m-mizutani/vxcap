package vxcap_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/m-mizutani/vxcap/pkg/vxcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sampleHeader    = []byte{0x08, 0x00, 0x00, 0x01, 0xa8, 0xee, 0xd6, 0x00}
	sampleEther     = []byte{0x0a, 0x66, 0x53, 0x0c, 0x59, 0xc4, 0x0a, 0x40, 0x8d, 0x4d, 0x24, 0x0e, 0x08, 0x00}
	sampleIPHeader  = []byte{0x45, 0x00, 0x01, 0x21, 0x9c, 0xe7, 0x40, 0x00, 0x26, 0x06, 0xa8, 0xdf, 0xa7, 0x47, 0xb8, 0x42, 0xac, 0x1e, 0x02, 0x68}
	sampleTCPHeader = []byte{0xd0, 0xe0, 0x1f, 0x98, 0x57, 0xd9, 0xc0, 0x71, 0x34, 0x04, 0x0e, 0x1f, 0x50, 0x18, 0x39, 0x08, 0x54, 0x10, 0x00, 0x00}
	samplePayload   = []byte("POST /ws/v1/cluster/apps/new-application HTTP/1.1\r\nHost: 54.65.xxx.xxx:8088\r\nContent-Length: 0\r\nUser-Agent: python-requests/2.6.0 CPython/2.6.6 Linux/2.6.32-754.17.1.el6.x86_64\r\nConnection: keep-alive\r\nAccept: */*\r\nAccept-Encoding: gzip, deflate\r\n\r\n")
)

func TestParseVxlanNormal(t *testing.T) {
	hdr := sampleHeader
	ether := sampleEther

	var data []byte
	data = append(data, hdr...)
	data = append(data, ether...)

	pkt, err := vxcap.ParseVXLAN(data, len(data))
	require.NoError(t, err)
	assert.Equal(t, len(ether), len(pkt.Data))
	assert.Equal(t, uint16(1), pkt.Header.GroupPolicyID)
	assert.Equal(t, [3]byte{0xa8, 0xee, 0xd6}, pkt.Header.NetworkIndentifier)
}

func TestParseVxlanLength(t *testing.T) {
	hdr := sampleHeader

	pkt, err := vxcap.ParseVXLAN(hdr, len(hdr))
	require.NoError(t, err)
	assert.Equal(t, 0, len(pkt.Data))

	tooShortHdr := hdr[0:7]
	_, err = vxcap.ParseVXLAN(tooShortHdr, len(tooShortHdr))
	require.Error(t, err)
}

func TestVxlanListener(t *testing.T) {
	hdr := sampleHeader
	ether := sampleEther

	var data []byte
	data = append(data, hdr...)
	data = append(data, ether...)

	port := 30000 + rand.Int()%10000
	ch := vxcap.ListenVXLAN(port, 10)

	time.Sleep(time.Second) // Wait for UDP server listening

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%d", port))
	require.NoError(t, err)
	sock, err := net.DialUDP("udp", nil, addr)
	require.NoError(t, err)

	n, err := sock.Write(data)
	require.NoError(t, err)
	require.Equal(t, len(data), n)

	q := <-ch
	assert.NoError(t, q.Err)
	assert.Equal(t, len(ether), len(q.Pkt.Data))
	assert.Equal(t, uint16(1), q.Pkt.Header.GroupPolicyID)
}
