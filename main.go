package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var logger = logrus.New()

func main() {
	app := cli.NewApp()
	app.Name = "vxcap"
	app.Usage = "Dump encapslated packet on VXNET"

	app.Action = func(c *cli.Context) error {
		cap := newVxcap()
		if err := cap.start(); err != nil {
			return err
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
