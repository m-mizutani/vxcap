package main

import (
	"fmt"
	"os"

	"github.com/m-mizutani/vxcap/pkg/vxcap"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const vxcapVersion = "0.0.1"

var logLevelMap = map[string]logrus.Level{
	"trace": logrus.TraceLevel,
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
}

func main() {
	cap := vxcap.New()
	var args vxcap.PacketProcessorArgument
	var logLevel string

	app := cli.NewApp()
	app.Name = "vxcap"
	app.Usage = "Dump tool encapslated packet on VXNET"
	app.Version = vxcapVersion
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Masayoshi Mizutani",
			Email: "mizutani@sfc.wide.ad.jp",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "emitter, e", Value: "fs",
			Usage:       "Destination to save data [fs,s3,firehose]",
			Destination: &args.EmitterArgs.Name,
		},
		cli.StringFlag{
			Name: "dumper, d", Value: "pcap",
			Usage:       "Write format [pcap,json]",
			Destination: &args.DumperArgs.Format,
		},
		cli.StringFlag{
			Name: "log-level", Value: "info",
			Usage:       "Log level [trace,debug,info,warn,error]",
			Destination: &logLevel,
		},
		/*
			TODO: this option is not available for now
			cli.StringFlag{
				Name: "target, t", Value: "packet",
				Destination: &args.DumperArgs.Target,
			},
		*/
		cli.IntFlag{
			Name: "port, p", Value: vxcap.DefaultVxlanPort,
			Usage:       "UDP port of VXLAN receiver",
			Destination: &cap.RecvPort,
		},
		cli.IntFlag{
			Name: "receiver-queue-size", Value: vxcap.DefaultReceiverQueueSize,
			Usage:       "Queue size between UDP server and packet processor",
			Destination: &cap.QueueSize,
		},
		// Options for fsEmitter
		cli.StringFlag{
			Name: "fs-filename", Value: "dump",
			Usage:       "Base file name for FS emitter",
			Destination: &args.EmitterArgs.FsFileName,
		},
		cli.StringFlag{
			Name: "fs-dirpath", Value: ".",
			Usage:       "Output directory for FS emitter",
			Destination: &args.EmitterArgs.FsDirPath,
		},
		/*
			TODO: Implement rotation mechanism
			cli.IntFlag{
				Name: "fs-rotate-size", Value: 0, // Not rotate
				Usage:       "Threshold size of file rotation for FS emitter",
				Destination: &args.EmitterArgs.FsRotateSize,
			},
		*/

		// Options for AWS emitter
		cli.StringFlag{
			Name:        "aws-region",
			Usage:       "AWS region for emitter to AWS",
			Destination: &args.EmitterArgs.AwsRegion,
		},
		// == s3Emitter
		cli.StringFlag{
			Name:        "aws-s3-bucket",
			Usage:       "AWS S3 bucket name for S3 emitter",
			Destination: &args.EmitterArgs.AwsS3Bucket,
		},
		cli.StringFlag{
			Name:        "aws-s3-prefix",
			Usage:       "Prefix of AWS S3 object key for S3 emitter",
			Destination: &args.EmitterArgs.AwsS3Prefix,
		},
		cli.BoolFlag{
			Name:        "aws-s3-add-time-key",
			Usage:       "Enable to add time key to S3 object key for S3 emitter",
			Destination: &args.EmitterArgs.AwsS3AddTimeKey,
		},
		cli.IntFlag{
			Name:        "aws-s3-flush-count",
			Usage:       "Threshold of record number to flush object to AWS S3 bucket",
			Destination: &args.EmitterArgs.AwsS3FlushCount,
		},
		// == firehoseEmitter
		cli.StringFlag{
			Name:        "aws-firehose-name",
			Usage:       "Name of AWS Firehose for Firehose emitter",
			Destination: &args.EmitterArgs.AwsFirehoseName,
		},
		cli.IntFlag{
			Name:        "aws-firehose-flush-size",
			Usage:       "Threshold of record size to flush object to AWS Firehose",
			Destination: &args.EmitterArgs.AwsFirehoseFlushSize,
		},
		// Options for Dumper
		cli.BoolFlag{
			Name:        "enable-json-text",
			Usage:       "Enable human readable application layer payload in json format",
			Destination: &args.DumperArgs.EnableJSONTextPayload,
		},
		cli.BoolFlag{
			Name:        "enable-json-raw",
			Usage:       "Enable raw application layer payload (base64 encoded) in json format",
			Destination: &args.DumperArgs.EnableJSONRawPayload,
		},
	}

	// Target option can use only "packet" for now
	args.DumperArgs.Target = "packet"

	app.Action = func(c *cli.Context) error {
		level, ok := logLevelMap[logLevel]
		if !ok {
			return fmt.Errorf("Invalid log level: %s", logLevel)
		}
		vxcap.Logger.SetLevel(level)

		proc, err := vxcap.NewPacketProcessor(args)
		if err != nil {
			return err
		}

		if err := cap.Start(proc); err != nil {
			return err
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		vxcap.Logger.WithError(err).Fatal("Fatal Error")
	}
}
