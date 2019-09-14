package vxcap_test

import (
	"testing"

	"github.com/m-mizutani/vxcap/pkg/vxcap"
	"github.com/stretchr/testify/assert"
)

func TestEmitterNoName(t *testing.T) {
	var args vxcap.EmitterArguments
	emitter, err := vxcap.NewEmitter(args)
	assert.Error(t, err)
	assert.Nil(t, emitter)
}
