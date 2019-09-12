package main

type emitter interface {
	emit(packetRecord) error
}
