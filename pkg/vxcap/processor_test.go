package vxcap_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/caarlos0/env/v6"
	"github.com/google/uuid"
	"github.com/m-mizutani/vxcap/pkg/vxcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessorPcapFsOutput(t *testing.T) {
	payload := genSamplePacketData()
	pkt := vxcap.NewPacketData(payload)

	tmpFile, err := ioutil.TempFile(".", "test")
	require.NoError(t, err)
	tmpFileName := tmpFile.Name()
	tmpFile.Close()

	proc, err := vxcap.NewPacketProcessor(vxcap.PacketProcessorArgument{
		DumperArgs: vxcap.DumperArguments{
			Format: "pcap",
			Target: "packet",
		},
		EmitterArgs: vxcap.EmitterArguments{
			Name:       "fs",
			FsFileName: tmpFileName,
		},
	})
	require.NoError(t, err)

	require.NoError(t, proc.Put(pkt))
	require.NoError(t, proc.Shutdown())

	cmd := exec.Command("tcpdump", "-nr", tmpFileName)
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "167.71.184.66.53472 > 172.30.2.104.8088")
	require.NoError(t, os.Remove(tmpFileName))
}

type awsConfig struct {
	AwsRegion   string `env:"VXCAP_AWS_REGION"`
	AwsS3Bucket string `env:"VXCAP_AWS_S3_BUCKET"`
	AwsS3Prefix string `env:"VXCAP_AWS_S3_PREFIX"`
}

func loadAwsConfig(t *testing.T) *awsConfig {
	var config awsConfig
	require.NoError(t, env.Parse(&config))
	if config.AwsRegion == "" || config.AwsS3Bucket == "" || config.AwsS3Prefix == "" {
		t.Skip("VXCAP_AWS_REGION, VXCAP_AWS_S3_BUCKET and VXCAP_AWS_S3_PREFIX are required for S3 test")
		return nil
	}
	return &config
}

func setupObjectsForAwsS3(t *testing.T) (*s3.S3, *awsConfig, string) {
	config := loadAwsConfig(t)
	prefix := config.AwsS3Prefix + uuid.New().String() + "/"

	ssn := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.AwsRegion),
	}))
	s3client := s3.New(ssn)

	return s3client, config, prefix
}

func TestProcessorJsonS3Output(t *testing.T) {
	pkt := vxcap.NewPacketData(genSamplePacketData())

	s3client, config, prefix := setupObjectsForAwsS3(t)
	proc, err := vxcap.NewPacketProcessor(vxcap.PacketProcessorArgument{
		DumperArgs: vxcap.DumperArguments{
			Format: "json",
			Target: "packet",
		},
		EmitterArgs: vxcap.EmitterArguments{
			Name:        "s3",
			AwsRegion:   config.AwsRegion,
			AwsS3Bucket: config.AwsS3Bucket,
			AwsS3Prefix: prefix,
		},
	})
	require.NoError(t, err)

	require.NoError(t, proc.Put(pkt))
	require.NoError(t, proc.Shutdown())

	resp, err := s3client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(config.AwsS3Bucket),
		Prefix: aws.String(prefix),
	})
	require.NoError(t, err)

	assert.Equal(t, 1, len(resp.Contents))

	obj, err := s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(config.AwsS3Bucket),
		Key:    resp.Contents[0].Key,
	})
	require.NoError(t, err)
	raw, err := ioutil.ReadAll(obj.Body)
	require.NoError(t, err)
	var jdata vxcap.JSONRecord
	err = json.Unmarshal(raw, &jdata)
	require.NoError(t, err)
	assert.Equal(t, "167.71.184.66", jdata.SrcAddr)
}

func TestProcessorJsonS3FlushCount(t *testing.T) {
	pkt := vxcap.NewPacketData(genSamplePacketData())

	s3client, config, prefix := setupObjectsForAwsS3(t)
	proc, err := vxcap.NewPacketProcessor(vxcap.PacketProcessorArgument{
		DumperArgs: vxcap.DumperArguments{
			Format: "json",
			Target: "packet",
		},
		EmitterArgs: vxcap.EmitterArguments{
			Name:            "s3",
			AwsRegion:       config.AwsRegion,
			AwsS3Bucket:     config.AwsS3Bucket,
			AwsS3Prefix:     prefix,
			AwsS3FlushCount: 3,
		},
	})
	require.NoError(t, err)

	for i := 0; i < 7; i++ {
		require.NoError(t, proc.Put(pkt))
	}
	require.NoError(t, proc.Shutdown())

	resp, err := s3client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(config.AwsS3Bucket),
		Prefix: aws.String(prefix),
	})
	require.NoError(t, err)

	assert.Equal(t, 3, len(resp.Contents))
	bucket := aws.String(config.AwsS3Bucket)

	count := 0

	obj, err := s3client.GetObject(&s3.GetObjectInput{Bucket: bucket, Key: resp.Contents[0].Key})
	require.NoError(t, err)
	raw, err := ioutil.ReadAll(obj.Body)
	require.NoError(t, err)
	count += strings.Count(string(raw), "\n")

	obj, err = s3client.GetObject(&s3.GetObjectInput{Bucket: bucket, Key: resp.Contents[1].Key})
	require.NoError(t, err)
	raw, err = ioutil.ReadAll(obj.Body)
	require.NoError(t, err)
	count += strings.Count(string(raw), "\n")

	obj, err = s3client.GetObject(&s3.GetObjectInput{Bucket: bucket, Key: resp.Contents[2].Key})
	require.NoError(t, err)
	raw, err = ioutil.ReadAll(obj.Body)
	require.NoError(t, err)
	count += strings.Count(string(raw), "\n")

	assert.Equal(t, 7, count)
}

func TestProcessorS3ConfigError(t *testing.T) {
	var err error
	_, err = vxcap.NewPacketProcessor(vxcap.PacketProcessorArgument{
		DumperArgs: vxcap.DumperArguments{
			Format: "json",
			Target: "packet",
		},
		EmitterArgs: vxcap.EmitterArguments{
			Name: "s3",
			// Missing AwsRegion:   "test",
			AwsS3Bucket: "test",
			AwsS3Prefix: "test",
		},
	})
	assert.Error(t, err)

	_, err = vxcap.NewPacketProcessor(vxcap.PacketProcessorArgument{
		DumperArgs: vxcap.DumperArguments{
			Format: "json",
			Target: "packet",
		},
		EmitterArgs: vxcap.EmitterArguments{
			Name:      "s3",
			AwsRegion: "test",
			// Missing AwsS3Bucket: "test"
			AwsS3Prefix: "test",
		},
	})
	assert.Error(t, err)

	_, err = vxcap.NewPacketProcessor(vxcap.PacketProcessorArgument{
		DumperArgs: vxcap.DumperArguments{
			Format: "json",
			Target: "packet",
		},
		EmitterArgs: vxcap.EmitterArguments{
			Name:        "s3",
			AwsRegion:   "test",
			AwsS3Bucket: "test",
			// Missing AwsS3Prefix: "test",
		},
	})
	assert.NoError(t, err) // AwsS3Prefix is optional
}
