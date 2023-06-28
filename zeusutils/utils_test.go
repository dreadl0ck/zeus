package zeusutils_test

import (
	"github.com/dreadl0ck/zeus/zeusutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadArg(t *testing.T) {
	os.Args = []string{"zeus", "test=value"}
	val := zeusutils.LoadArg("test")
	assert.Equal(t, "value", val)

	os.Args = []string{"zeus", "test=\tvalue\t"}
	val = zeusutils.LoadArg("test")
	assert.Equal(t, "value", val)

	os.Args = []string{"zeus", "test= value "}
	val = zeusutils.LoadArg("test")
	assert.Equal(t, "value", val)
}

func TestLoadArgs(t *testing.T) {
	os.Args = []string{"zeus", "arg1=value", "arg2=value ", "arg3= value"}
	for k, v := range zeusutils.LoadArgs() {
		switch k {
		case "arg1":
			assert.Equal(t, "value", v)
		case "arg2":
			assert.Equal(t, "value", v)
		case "arg3":
			assert.Equal(t, "value", v)
		}
	}
}

func TestRequireEnv(t *testing.T) {
	// $PATH should always be set
	_ = zeusutils.RequireEnv("PATH")
}

func TestTrimStringLiterals(t *testing.T) {
	val := zeusutils.TrimStringLiterals("\"value\"")
	assert.Equal(t, "value", val)

	val = zeusutils.TrimStringLiterals("'value'")
	assert.Equal(t, "value", val)

	val = zeusutils.TrimStringLiterals("\"value")
	assert.Equal(t, "\"value", val)
}
