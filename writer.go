package main

import "io"

type writer interface {
	write(w io.Writer) error
}
