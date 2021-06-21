/*
Copyright 2020-2021 The UnDistro authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package template

import (
	"bytes"
)

// bufPool represents a reusable buffer pool for executing templates into.
var bufPool *BufferPool

// BufferPool implements a pool of bytes.Buffers in the form of a bounded channel.
// Pulled from the github.com/oxtoacart/bpool package (Apache licensed).
type BufferPool struct {
	c chan *bytes.Buffer
}

// NewBufferPool creates a new BufferPool bounded to the given size.
func NewBufferPool(size int) (bp *BufferPool) {
	return &BufferPool{
		c: make(chan *bytes.Buffer, size),
	}
}

// Get gets a Buffer from the BufferPool, or creates a new one if none are
// available in the pool.
func (bp *BufferPool) Get() (b *bytes.Buffer) {
	select {
	case b = <-bp.c:
	// reuse existing buffer
	default:
		// create new buffer
		b = bytes.NewBuffer([]byte{})
	}
	return
}

// Put returns the given Buffer to the BufferPool.
func (bp *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	select {
	case bp.c <- b:
	default: // Discard the buffer if the pool is full.
	}
}

// Initialize buffer pool for writing templates into.
func init() {
	bufPool = NewBufferPool(64)
}
