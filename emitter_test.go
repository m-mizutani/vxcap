package main_test

import (
	"testing"

	vxcap "github.com/m-mizutani/vxcap"
	"github.com/stretchr/testify/assert"
)

func TestEmitterNoName(t *testing.T) {
	var args vxcap.EmitterArgument
	emitter, err := vxcap.NewEmitter(args)
	assert.Error(t, err)
	assert.Nil(t, emitter)
}
