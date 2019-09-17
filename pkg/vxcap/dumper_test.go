package vxcap_test

import (
	"bytes"
	"encoding/json"

	"github.com/m-mizutani/vxcap/pkg/vxcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"os"
	"testing"
)

// var vxcapTestFS = os.Getenv("VXCAP_TEST_FS")

var vxcapTestFS = "on"

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

	dumper := vxcap.NewPcapDumper(vxcap.DumperArguments{
		Format: "pcap",
		Target: "packet",
	})
	err = vxcap.PcapDumperDump(dumper, vxcap.ToPacketDataSlice(pkt), w)
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

	dumper := vxcap.NewJSONPacketDumper(vxcap.DumperArguments{
		Format: "pcap",
		Target: "packet",
	})
	err = vxcap.JSONPacketDumperDump(dumper, vxcap.ToPacketDataSlice(pkt), w)
	require.NoError(t, err)
}

func TestJsonDumpBuffer(t *testing.T) {
	payload := genSamplePacketData()
	pkt := vxcap.NewPacketData(payload)

	buf := new(bytes.Buffer)

	dumper := vxcap.NewJSONPacketDumper(vxcap.DumperArguments{
		Format:                "json",
		Target:                "packet",
		EnableJSONTextPayload: true,
	})
	err := vxcap.JSONPacketDumperDump(dumper, vxcap.ToPacketDataSlice(pkt), buf)
	require.NoError(t, err)

	var d vxcap.JSONRecord
	err = json.Unmarshal(buf.Bytes(), &d)
	require.NoError(t, err)

	assert.Equal(t, "167.71.184.66", d.SrcAddr)
	assert.Equal(t, "172.30.2.104", d.DstAddr)
	assert.Equal(t, "TCP", d.Protocol)
	assert.Equal(t, 53472, d.SrcPort)
	assert.Equal(t, 8088, d.DstPort)
	assert.Contains(t, d.TextPayload, "POST /ws/v1/cluster/apps/new-application")
	assert.Contains(t, d.TextPayload, "\r\n\r\n") // tail LF of HTTP request
	assert.Equal(t, 0, len(d.RawPayload))
}

func TestJsonDumpNoText(t *testing.T) {
	payload := genSamplePacketData()
	pkt := vxcap.NewPacketData(payload)
	buf := new(bytes.Buffer)

	dumper := vxcap.NewJSONPacketDumper(vxcap.DumperArguments{
		Format:               "json",
		Target:               "packet",
		EnableJSONRawPayload: true,
	})
	err := vxcap.JSONPacketDumperDump(dumper, vxcap.ToPacketDataSlice(pkt), buf)
	require.NoError(t, err)

	var d vxcap.JSONRecord
	err = json.Unmarshal(buf.Bytes(), &d)
	require.NoError(t, err)

	assert.NotContains(t, d.TextPayload, "POST /ws/v1/cluster/apps/new-application")
	assert.NotEqual(t, 0, len(d.RawPayload))
}
