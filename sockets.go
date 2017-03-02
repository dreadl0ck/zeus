/*
 *  ZEUS - An Electrifying Build System
 *  Copyright (c) 2017 Philipp Mieden <dreadl0ck@protonmail.ch>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"errors"
	"sync"

	"github.com/desertbit/glue"
)

var (
	// ErrSocketNotFound means the socketID does not exist
	ErrSocketNotFound = errors.New("no socket found for id")
)

// SocketStore contains the socket connections and provides an interface for safe concurrent access
type SocketStore struct {
	mutex   *sync.Mutex
	sockets []*glue.Socket
}

// RemoveSocket removes an unauthorized socket
func (s *SocketStore) RemoveSocket(socket *glue.Socket) {

	s.Lock()
	defer s.Unlock()

	for i, sock := range s.sockets {
		if sock == socket {
			s.sockets = append(s.sockets[:i], s.sockets[i+1:]...)
		}
	}

	return
}

// AddSocket adds an unauthorized socket
func (s *SocketStore) AddSocket(socket *glue.Socket) {
	s.Lock()
	defer s.Unlock()

	s.sockets = append(s.sockets, socket)
}

// Clear closes all socket connections
func (s *SocketStore) Clear() {

	s.Lock()
	defer s.Unlock()

	if len(s.sockets) > 0 {
		for _, socket := range s.sockets {
			socket.Close()
		}
	}

	return
}

// NumSockets returns the current amount of socket connections managed by the SocketStore
func (s *SocketStore) NumSockets() int {

	s.Lock()
	defer s.Unlock()

	return len(s.sockets)
}

// NewSocketStore constructs a new SocketStore
func NewSocketStore() *SocketStore {
	return &SocketStore{
		mutex:   &sync.Mutex{},
		sockets: make([]*glue.Socket, 0),
	}
}

// Lock is an alias for locking the Sockt Store mutex
func (s *SocketStore) Lock() {
	s.mutex.Lock()
}

// Unlock is an alias for unlocking the Socket Store mutex
func (s *SocketStore) Unlock() {
	s.mutex.Unlock()
}

// OnNewSocket is executed when a new glue client connects
func (s *SocketStore) OnNewSocket(socket *glue.Socket) {

	socket.OnRead(func(sessionID string) {
		socket.Write("access granted")
	})

	// Set a function which is triggered as soon as the socket is closed.
	socket.OnClose(func() {
		s.RemoveSocket(socket)
	})

	s.AddSocket(socket)
}
