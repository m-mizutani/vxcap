package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var logger = logrus.New()

const vxcapVersion = "0.0.1"

func main() {
	cap := newVxcap()
	var args packetProcessorArgument
	var dumperName string

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
			Destination: &args.EmitterArgs.Name,
		},
		cli.StringFlag{
			Name: "dumper, d", Value: "pcap",
			Destination: &dumperName,
		},
		cli.IntFlag{
			Name: "port, p", Value: defaultVxlanPort,
			Destination: &cap.RecvPort,
		},

		// Options for fsEmitter
		cli.StringFlag{
			Name: "fs-filename", Value: "dump",
			Destination: &args.EmitterArgs.FsFileName,
		},
		cli.StringFlag{
			Name: "fs-dirpath", Value: ".",
			Destination: &args.EmitterArgs.FsDirPath,
		},
		cli.IntFlag{
			Name: "fs-rotate-size", Value: 0, // Not rotate
			Destination: &args.EmitterArgs.FsRotateSize,
		},
	}

	app.Action = func(c *cli.Context) error {
		proc, err := newPacketProcessor(args)
		if err != nil {
			return err
		}

		if err := cap.start(proc); err != nil {
			return err
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logger.WithError(err).Fatal("Fatal Error")
	}
}
