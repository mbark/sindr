package sindrtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.starlark.net/starlark"
)

func AssertValue(t *testing.T, value starlark.Value, canHash bool) {
	assert.NotEmpty(t, value.String())
	assert.NotEmpty(t, value.Type())
	hash, err := value.Hash()
	if canHash {
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)
	} else {
		assert.Error(t, err)
	}
	assert.True(t, bool(value.Truth()))

	// just freeze to verify it works
	value.Freeze()

	a, ok := value.(starlark.HasAttrs)
	if ok {
		assert.NotEmpty(t, a.AttrNames())
		for _, n := range a.AttrNames() {
			_, err := a.Attr(n)
			assert.NoError(t, err)
		}
	}
}
