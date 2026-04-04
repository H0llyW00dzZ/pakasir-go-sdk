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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPool(t *testing.T) {
	require.NotNil(t, Default)
}

func TestPoolGetPut(t *testing.T) {
	buf := Default.Get()
	require.NotNil(t, buf)

	buf.WriteString("hello")
	assert.Equal(t, "hello", buf.String())
	assert.Equal(t, 5, buf.Len())
	assert.Equal(t, []byte("hello"), buf.Bytes())

	buf.Reset()
	Default.Put(buf)

	buf2 := Default.Get()
	require.NotNil(t, buf2)
	buf2.Reset()
	Default.Put(buf2)
}

func TestBufferWrite(t *testing.T) {
	buf := Default.Get()
	defer func() {
		buf.Reset()
		Default.Put(buf)
	}()

	n, err := buf.Write([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "test", buf.String())
}

func TestBufferWriteString(t *testing.T) {
	buf := Default.Get()
	defer func() {
		buf.Reset()
		Default.Put(buf)
	}()

	n, err := buf.WriteString("hello world")
	require.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "hello world", buf.String())
}

func TestBufferWriteByte(t *testing.T) {
	buf := Default.Get()
	defer func() {
		buf.Reset()
		Default.Put(buf)
	}()

	err := buf.WriteByte('X')
	require.NoError(t, err)
	assert.Equal(t, "X", buf.String())
}

func TestBufferSet(t *testing.T) {
	buf := Default.Get()
	defer func() {
		buf.Reset()
		Default.Put(buf)
	}()

	buf.Set([]byte("replaced"))
	assert.Equal(t, "replaced", buf.String())
}

func TestBufferSetString(t *testing.T) {
	buf := Default.Get()
	defer func() {
		buf.Reset()
		Default.Put(buf)
	}()

	buf.SetString("new content")
	assert.Equal(t, "new content", buf.String())
}

func TestBufferReset(t *testing.T) {
	buf := Default.Get()
	defer func() {
		buf.Reset()
		Default.Put(buf)
	}()

	buf.WriteString("data")
	assert.Equal(t, 4, buf.Len())

	buf.Reset()
	assert.Equal(t, 0, buf.Len())
}

func TestPoolPutNonByteBuffer(t *testing.T) {
	// Passing a non-ByteBuffer type should be safely ignored.
	Default.Put(&mockBuffer{})
}

func TestPoolConcurrency(t *testing.T) {
	done := make(chan struct{})
	for range 100 {
		go func() {
			buf := Default.Get()
			buf.WriteString("concurrent")
			buf.Reset()
			Default.Put(buf)
			done <- struct{}{}
		}()
	}
	for range 100 {
		<-done
	}
}

// mockBuffer is a minimal Buffer implementation for testing Put type-assertion.
type mockBuffer struct{}

func (m *mockBuffer) Write([]byte) (int, error)           { return 0, nil }
func (m *mockBuffer) WriteString(string) (int, error)     { return 0, nil }
func (m *mockBuffer) WriteByte(byte) error                { return nil }
func (m *mockBuffer) WriteTo(w io.Writer) (int64, error)  { return 0, nil }
func (m *mockBuffer) ReadFrom(r io.Reader) (int64, error) { return 0, nil }
func (m *mockBuffer) Bytes() []byte                       { return nil }
func (m *mockBuffer) String() string                      { return "" }
func (m *mockBuffer) Len() int                            { return 0 }
func (m *mockBuffer) Set([]byte)                          {}
func (m *mockBuffer) SetString(string)                    {}
func (m *mockBuffer) Reset()                              {}
