package main

//nolint
var (
	ParseVXLAN  = parseVXLAN
	ListenVXLAN = listenVXLAN
)

// nolint
type EmitterArgument emitterArgument

func NewEmitter(args EmitterArgument) (recordEmitter, error) {
	return newEmitter(emitterArgument(args))
}
