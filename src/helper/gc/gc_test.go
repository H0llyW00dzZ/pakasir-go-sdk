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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	p := NewPool()
	require.NotNil(t, p)
}

func TestDefaultPool(t *testing.T) {
	require.NotNil(t, Default)
}

func TestPoolGetPut(t *testing.T) {
	p := NewPool()

	buf := p.Get()
	require.NotNil(t, buf)
	assert.Equal(t, 0, buf.Len(), "Get should return an empty buffer")

	buf.WriteString("hello")
	assert.Equal(t, "hello", buf.String())

	p.Put(buf)

	buf2 := p.Get()
	assert.Equal(t, 0, buf2.Len(), "Get after Put should return a reset buffer")
	p.Put(buf2)
}

func TestPoolConcurrency(t *testing.T) {
	p := NewPool()
	done := make(chan struct{})

	for range 100 {
		go func() {
			buf := p.Get()
			buf.WriteString("concurrent")
			p.Put(buf)
			done <- struct{}{}
		}()
	}

	for range 100 {
		<-done
	}
}
