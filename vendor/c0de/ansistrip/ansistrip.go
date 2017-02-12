package stripansi

import (
	"io"
	"regexp"
)

var (
	ansiSequence = regexp.MustCompile("\u001b\\[.*?m")
)

// AnsiStripper strips ansi
type AnsiStripper struct {
	w io.Writer
}

// New creates a new ansistripper
func New(w io.Writer) *AnsiStripper {
	return &AnsiStripper{
		w: w,
	}
}

func (as *AnsiStripper) Write(b []byte) (n int, err error) {
	return as.w.Write(StripAnsi(b))
}

// StripAnsi removes all ANSI Escape Sequences from the byteslice
func StripAnsi(b []byte) []byte {
	return ansiSequence.ReplaceAll(b, []byte(""))
}
