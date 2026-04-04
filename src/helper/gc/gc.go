// Copyright 2026 H0llyW00dzZ
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gc

import (
	"bytes"
	"sync"
)

// Pool is a memory-efficient buffer pool backed by [sync.Pool].
// It reduces allocation pressure during JSON serialization
// and HTTP request/response processing.
type Pool struct {
	pool sync.Pool
}

// Default is the package-level default buffer pool,
// ready for use without initialization.
var Default = NewPool()

// NewPool creates a new [Pool] instance.
func NewPool() *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

// Get retrieves a [bytes.Buffer] from the pool.
// The buffer is reset before being returned.
func (p *Pool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// Put returns a [bytes.Buffer] to the pool for reuse.
func (p *Pool) Put(buf *bytes.Buffer) {
	p.pool.Put(buf)
}
