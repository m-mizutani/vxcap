package main_test

import (
	"reflect"
	"testing"

	vxcap "github.com/m-mizutani/vxcap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmitterNoName(t *testing.T) {
	var args vxcap.EmitterArgument
	emitter, err := vxcap.NewEmitter(args)
	assert.Error(t, err)
	assert.Nil(t, emitter)
}

func TestEmitterGetFsEmitter(t *testing.T) {
	var args vxcap.EmitterArgument
	args.Name = "fs"
	emitter, err := vxcap.NewEmitter(args)
	require.NoError(t, err)
	assert.NotNil(t, emitter)

	tp := reflect.TypeOf(emitter)
	require.Equal(t, reflect.Ptr, tp.Kind())
	assert.Equal(t, "fsEmitter", tp.Elem().Name())
}
