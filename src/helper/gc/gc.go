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

	"github.com/valyala/bytebufferpool"
)

// Buffer defines the interface for a reusable byte buffer.
//
// It abstracts the [bytebufferpool.ByteBuffer] type to avoid direct dependencies
// and provides a consistent API for buffer manipulation throughout the application.
//
// The interface supports standard I/O operations (ReadFrom, WriteTo) as well as
// efficient string and byte manipulation methods. Implementations must ensure
// that the underlying storage can be reused after Reset() is called.
type Buffer interface {
	// Write appends the contents of p to the buffer.
	//
	// It implements the [io.Writer] interface, allowing the buffer to be used
	// as a destination for standard library I/O operations.
	//
	// Returns the number of bytes written (always len(p)) and nil error.
	Write(p []byte) (int, error)

	// WriteString appends the string s to the buffer.
	//
	// This method is optimized for string appending without unnecessary allocations.
	//
	// Returns the number of bytes written (len(s)) and nil error.
	WriteString(s string) (int, error)

	// WriteByte appends the byte c to the buffer.
	//
	// Returns nil error.
	WriteByte(c byte) error

	// WriteTo writes data to w until the buffer is drained or an error occurs.
	//
	// It implements the [io.WriterTo] interface, allowing efficient data transfer
	// from the buffer to another writer.
	//
	// Returns the number of bytes written and any error from w.Write.
	WriteTo(w io.Writer) (int64, error)

	// ReadFrom reads data from r until EOF and appends it to the buffer.
	//
	// It implements the [io.ReaderFrom] interface, allowing the buffer to efficiently
	// consume data from a reader.
	//
	// Returns the number of bytes read and any error from r.Read.
	ReadFrom(r io.Reader) (int64, error)

	// Bytes returns the accumulated bytes in the buffer.
	//
	// The returned slice is valid only until the next buffer modification.
	Bytes() []byte

	// String returns the accumulated string in the buffer.
	String() string

	// Len returns the number of bytes in the buffer.
	Len() int

	// Set replaces the buffer contents with p.
	//
	// This is equivalent to Reset() followed by Write(p), but more efficient.
	Set(p []byte)

	// SetString replaces the buffer contents with s.
	//
	// This is equivalent to Reset() followed by WriteString(s), but more efficient.
	SetString(s string)

	// Reset clears the buffer, retaining the underlying storage for reuse.
	//
	// This must be called before returning the buffer to the pool to ensure
	// no data leaks between uses.
	Reset()
}

// Pool defines the interface for buffer pooling.
//
// It abstracts the [bytebufferpool.Pool] type to avoid direct dependencies
// and enable efficient memory reuse.
//
// Implementations must be safe for concurrent use by multiple goroutines.
type Pool interface {
	// Get returns a buffer from the pool.
	//
	// The returned buffer may contain garbage data and should be Reset()
	// before use if not using Set/SetString.
	Get() Buffer

	// Put returns a buffer to the pool.
	//
	// The buffer should be Reset() before calling Put() to prevent data leaks.
	Put(b Buffer)
}

// pool wraps [bytebufferpool.Pool] to implement the [Pool] interface.
type pool struct{ p *bytebufferpool.Pool }

// Get returns a buffer from the pool.
func (p *pool) Get() Buffer { return p.p.Get() }

// Put returns a buffer to the pool.
func (p *pool) Put(b Buffer) {
	if buf, ok := b.(*bytebufferpool.ByteBuffer); ok {
		p.p.Put(buf)
	}
}

// Default is the default buffer pool used for efficient memory reuse.
//
// Buffer pooling provides efficient memory reuse for I/O operations,
// especially beneficial in high-concurrency environments processing
// multiple API requests. Memory usage remains low even under high load
// by reusing buffer allocations instead of constant allocation/deallocation.
//
// Example usage for reading HTTP response body:
//
//	buf := gc.Default.Get()
//	defer func() {
//	    buf.Reset()
//	    gc.Default.Put(buf)
//	}()
//
//	if _, err := buf.ReadFrom(resp.Body); err != nil {
//	    return fmt.Errorf("failed to read response: %w", err)
//	}
//
//	// Parse JSON from buffer
//	var result Response
//	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
//	    return fmt.Errorf("failed to parse response: %w", err)
//	}
//
// Example usage for building request body:
//
//	buf := gc.Default.Get()
//	defer func() {
//	    buf.Reset()
//	    gc.Default.Put(buf)
//	}()
//
//	// Encode request to JSON
//	if err := json.NewEncoder(buf).Encode(request); err != nil {
//	    return fmt.Errorf("failed to encode request: %w", err)
//	}
//
//	// Create HTTP request with buffer as body
//	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
var Default Pool = &pool{p: &bytebufferpool.Pool{}}
