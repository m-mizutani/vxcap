package main

import "io"

type recordWriter interface {
	write([]*packetRecord, io.Writer) error
}
