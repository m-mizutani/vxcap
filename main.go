package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var logger = logrus.New()

const vxcapVersion = "0.0.1"

func main() {
	cap := newVxcap()
	var emitterArgs emitterArgument
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
			Destination: &emitterArgs.Name,
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
			Destination: &emitterArgs.FsFileName,
		},
		cli.StringFlag{
			Name: "fs-dirpath", Value: ".",
			Destination: &emitterArgs.FsDirPath,
		},
		cli.IntFlag{
			Name: "fs-rotate-size", Value: 0, // Not rotate
			Destination: &emitterArgs.FsRotateSize,
		},
	}

	app.Action = func(c *cli.Context) error {
		dumper, err := getDumper(dumperName)
		if err != nil {
			return err
		}
		emitter, err := newEmitter(emitterArgs)
		if err != nil {
			return err
		}

		emitter.setDumper(dumper)
		cap.addEmitter(emitter)

		if err := cap.start(); err != nil {
			return err
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
