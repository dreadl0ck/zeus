// ZEUS - A Powerful Build System
// Copyright (c) 2017 Philipp Mieden <dreadl0ck@protonmail.ch>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package ansistrip

import (
	"io"
	"regexp"
	"sync"
)

var (
	ansiSequence = regexp.MustCompile("\u001b\\[.*?m")
	version      = 0.1
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

// AtomicAnsiStripper is thread safe
type AtomicAnsiStripper struct {
	w     io.Writer
	mutex *sync.Mutex
}

// NewAtomic creates a new AtomicAnsiStripper
func NewAtomic(w io.Writer) *AtomicAnsiStripper {
	return &AtomicAnsiStripper{
		w:     w,
		mutex: &sync.Mutex{},
	}
}

func (as *AtomicAnsiStripper) Write(b []byte) (n int, err error) {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	return as.w.Write(StripAnsi(b))
}

// StripAnsi removes all ANSI Escape Sequences from the byteslice
func StripAnsi(b []byte) []byte {
	return ansiSequence.ReplaceAll(b, []byte(""))
}
