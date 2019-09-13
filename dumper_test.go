package main_test

import (
	"bytes"
	"encoding/json"

	vxcap "github.com/m-mizutani/vxcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"os"
	"testing"
)

var vxcapTestFS = os.Getenv("VXCAP_TEST_FS")

// var vxcapTestFS = "on"

func genSamplePacketData() []byte {
	var payload []byte
	payload = append(payload, sampleEther...)
	payload = append(payload, sampleIPHeader...)
	payload = append(payload, sampleTCPHeader...)
	payload = append(payload, samplePayload...)
	return payload
}

func TestPcapDumpFileSystem(t *testing.T) {
	if vxcapTestFS == "" {
		t.Skip("VXCAP_TEST_FS is not set")
	}

	payload := genSamplePacketData()
	pkt := vxcap.NewPacketData(payload)

	w, err := os.Create("test_dumppcap_fs.pcap")
	require.NoError(t, err)

	d := vxcap.PcapDumper{}
	err = d.Open(w)
	require.NoError(t, err)
	err = d.Dump(vxcap.ToPacketDataSlice(pkt), w)
	require.NoError(t, err)
	err = d.Close(w)
	require.NoError(t, err)
}

func TestJsonDumpFileSystem(t *testing.T) {
	if vxcapTestFS == "" {
		t.Skip("VXCAP_TEST_FS is not set")
	}

	payload := genSamplePacketData()
	pkt := vxcap.NewPacketData(payload)

	w, err := os.Create("test_dumpjson_fs.json")
	require.NoError(t, err)

	d := vxcap.JsonPacketDumper{}
	err = d.Dump(vxcap.ToPacketDataSlice(pkt), w)
	require.NoError(t, err)
}

func TestJsonDumpBuffer(t *testing.T) {
	payload := genSamplePacketData()
	pkt := vxcap.NewPacketData(payload)

	buf := new(bytes.Buffer)

	dumper := vxcap.JsonPacketDumper{}
	err := dumper.Dump(vxcap.ToPacketDataSlice(pkt), buf)
	require.NoError(t, err)

	var d vxcap.JSONRecord
	err = json.Unmarshal(buf.Bytes(), &d)
	require.NoError(t, err)

	assert.Equal(t, "167.71.184.66", d.SrcAddr)
	assert.Equal(t, "172.30.2.104", d.DstAddr)
	assert.Equal(t, "TCP", d.Protocol)
	assert.Equal(t, 53472, d.SrcPort)
	assert.Equal(t, 8088, d.DstPort)
	assert.Contains(t, d.TextData, "POST /ws/v1/cluster/apps/new-application")
	assert.Contains(t, d.TextData, "\n\n") // tail LF of HTTP request
}
