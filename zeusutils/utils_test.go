package zeusutils_test

import (
	"github.com/dreadl0ck/zeus/zeusutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadArg(t *testing.T) {
	os.Args = []string{"zeus", "test=asdf"}
	val := zeusutils.LoadArg("test")
	assert.Equal(t, val, "asdf")

	os.Args = []string{"zeus", "test=\"asdf\""}
	val = zeusutils.LoadArg("test")
	assert.Equal(t, val, "asdf")

	os.Args = []string{"zeus", "test='asdf'"}
	val = zeusutils.LoadArg("test")
	assert.Equal(t, val, "asdf")
}

func TestRequireEnv(t *testing.T) {
	_ = zeusutils.RequireEnv("PATH")
}
