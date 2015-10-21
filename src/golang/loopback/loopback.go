/*
 *  Ants - Let the ants handle your serial communication.
 *  Copyright (C) 2015  Roland Singer <roland.singer[at]desertbit.com>
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

// Package loopback is a small wrapper to provide an io.ReadWriteCloser interface
// which writes read data back.
package loopback

import (
	"errors"
	"io"
	"sync"
)

var (
	ErrIsClosed = errors.New("loopback is closed")
)

type loopback struct {
	buffer   []byte
	mutex    sync.Mutex
	isClosed bool
}

func New() io.ReadWriteCloser {
	return &loopback{}
}

func (l *loopback) Read(p []byte) (n int, err error) {
	// Lock the mutex.
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if closed.
	if l.isClosed {
		return 0, ErrIsClosed
	}

	// Check if buffer is empty.
	// Note: if l.buffer == nil then len(l.buffer) == 0.
	if len(l.buffer) == 0 {
		return 0, nil
	}

	// Determind how many bytes to read.
	n = len(p)
	if n > len(l.buffer) {
		n = len(l.buffer)
	}

	// Read the bytes and add them to the passed slice.
	for i := 0; i < n; i++ {
		p[i] = l.buffer[i]
	}

	// Remove the read bytes from the buffer.
	l.buffer = l.buffer[n:]

	return n, nil
}

func (l *loopback) Write(p []byte) (n int, err error) {
	// Lock the mutex.
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if closed.
	if l.isClosed {
		return 0, ErrIsClosed
	}

	// Add the bytes to the buffer.
	l.buffer = append(l.buffer, p...)

	return len(p), nil
}

func (l *loopback) Close() error {
	// Lock the mutex.
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Check if closed.
	if l.isClosed {
		return ErrIsClosed
	}

	// Update the flag.
	l.isClosed = true

	return nil
}
