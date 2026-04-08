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

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/gc"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/qr"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

func newTestClient(t *testing.T, serverURL string, opts ...Option) *Client {
	t.Helper()
	allOpts := []Option{WithBaseURL(serverURL), WithRetries(0)}
	allOpts = append(allOpts, opts...)
	return New("test-project", "test-api-key", allOpts...)
}

// --- New ---

func TestNewSuccess(t *testing.T) {
	c := New("my-project", "my-key")
	assert.Equal(t, "my-project", c.project)
	assert.Equal(t, "my-key", c.apiKey)
	assert.Equal(t, DefaultBaseURL, c.baseURL)
	assert.Equal(t, i18n.English, c.language)
	assert.Equal(t, DefaultRetries, c.retries)
}

func TestNewWithAllOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 5 * time.Second}
	c := New("proj", "key",
		WithBaseURL("https://custom.api.com"),
		WithHTTPClient(customHTTP),
		WithLanguage(i18n.Indonesian),
		WithRetries(5),
		WithRetryWait(100*time.Millisecond, 2*time.Second),
	)
	assert.Equal(t, "https://custom.api.com", c.baseURL)
	assert.NotSame(t, customHTTP, c.httpClient)
	assert.Equal(t, customHTTP.Timeout, c.httpClient.Timeout)
	assert.Equal(t, i18n.Indonesian, c.language)
	assert.Equal(t, 5, c.retries)
	assert.Equal(t, 100*time.Millisecond, c.retryWaitMin)
	assert.Equal(t, 2*time.Second, c.retryWaitMax)
}

func TestWithTimeout(t *testing.T) {
	c := New("proj", "key", WithTimeout(5*time.Second))
	assert.Equal(t, 5*time.Second, c.httpClient.Timeout)
}

// --- Do ---

func TestDoSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"status":"ok"}`, string(data))
}

func TestDoEmptyProject(t *testing.T) {
	c := New("", "my-key")
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidProject)
}

func TestDoEmptyAPIKey(t *testing.T) {
	c := New("my-project", "")
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAPIKey)
}

func TestDoEmptyProjectIndonesian(t *testing.T) {
	c := New("", "my-key", WithLanguage(i18n.Indonesian))
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.Contains(t, err.Error(), "slug proyek wajib diisi")
}

func TestDoSetsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		assert.Contains(t, ua, constants.SDKName)
		assert.Contains(t, ua, constants.SDKVersion)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	// Even with a custom http.Client, User-Agent must be set.
	c := newTestClient(t, srv.URL, WithHTTPClient(&http.Client{}))
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
}

func TestDoPostSetsContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Do(context.Background(), http.MethodPost, "/test", []byte(`{}`))
	require.NoError(t, err)
}

func TestDo4xxReturnsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)

	var apiErr *sdkerrors.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestDoRetryOn5xxThenSuccess(t *testing.T) {
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`unavailable`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(3),
		WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
	assert.Equal(t, int32(3), attempt.Load())
}

func TestDoRetryOn429ThenSuccess(t *testing.T) {
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`rate limited`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(3),
		WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
	assert.Equal(t, int32(3), attempt.Load())
}

func TestDoRetryAfterHeaderRespected(t *testing.T) {
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`rate limited`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	start := time.Now()
	c := newTestClient(t, srv.URL,
		WithRetries(2),
		// Set a very small backoff min but a high max so the Retry-After
		// of 1s is not clamped but clearly overrides the tiny min backoff.
		WithRetryWait(1*time.Millisecond, 5*time.Second),
	)
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
	assert.Equal(t, int32(2), attempt.Load())
	// The Retry-After of 1s should have caused a delay of ~1s,
	// far exceeding the 1-2ms backoff.
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond, "Retry-After delay must be respected")
}

func TestDoRetryAfterHeaderClampedToMax(t *testing.T) {
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		if n == 1 {
			// Server requests a 60s wait, but retryWaitMax is 50ms.
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`rate limited`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	start := time.Now()
	c := newTestClient(t, srv.URL,
		WithRetries(2),
		WithRetryWait(1*time.Millisecond, 50*time.Millisecond),
	)
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
	// Should complete quickly because 60s was clamped to 50ms.
	assert.Less(t, elapsed, 1*time.Second, "Retry-After must be clamped to retryWaitMax")
}

func TestDoRetryAfterInvalidHeaderIgnored(t *testing.T) {
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "not-a-number")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`rate limited`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(2),
		WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
	assert.Equal(t, int32(2), attempt.Load())
}

func TestDoPostRetryWithBody(t *testing.T) {
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.JSONEq(t, `{"key":"value"}`, string(body), "body must be complete on attempt %d", n)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`unavailable`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(3),
		WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)
	data, err := c.Do(context.Background(), http.MethodPost, "/test", []byte(`{"key":"value"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
	assert.Equal(t, int32(3), attempt.Load())
}

func TestDoRetriesExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`unavailable`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(2),
		WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	assert.Contains(t, err.Error(), "request failed after")
}

func TestDoContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := c.Do(ctx, http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
}

func TestDoContextDeadlineExceeded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Let the deadline expire before making the request.
	time.Sleep(5 * time.Millisecond)

	_, err := c.Do(ctx, http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
}

func TestDoContextCancelledDuringBackoff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`unavailable`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(5),
		WithRetryWait(5*time.Second, 10*time.Second),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Do(ctx, http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
}

func TestDoContextCancelledDuringRequest(t *testing.T) {
	// The server blocks indefinitely; context cancellation aborts the HTTP call.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, WithRetries(2))
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay so the HTTP request is in-flight.
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := c.Do(ctx, http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
}

func TestDoContextDeadlineDuringRetryRequest(t *testing.T) {
	// Server always returns 503, but a tight deadline ensures the context
	// expires during the backoff or next attempt, not after all retries.
	var attempt atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`unavailable`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(10),
		WithRetryWait(200*time.Millisecond, 1*time.Second),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := c.Do(ctx, http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	// Should have done far fewer attempts than the max 10.
	assert.Less(t, attempt.Load(), int32(10), "context deadline should short-circuit retries")
}

func TestDoNetworkError(t *testing.T) {
	c := newTestClient(t, "http://127.0.0.1:1")
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
}

func TestDoInvalidURL(t *testing.T) {
	c := New("proj", "key", WithBaseURL("://invalid"), WithRetries(0))
	_, doErr := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, doErr)
	t.Log(doErr)
	assert.Contains(t, doErr.Error(), "client: failed to create request")
}

func TestDoNetworkErrorRetryExhausted(t *testing.T) {
	c := New("proj", "key",
		WithBaseURL("http://127.0.0.1:1"),
		WithRetries(1),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
	)

	_, doErr := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, doErr)
	t.Log(doErr)
	assert.ErrorIs(t, doErr, sdkerrors.ErrRequestFailedAfterRetries)
	assert.Contains(t, doErr.Error(), "request failed after")
}

// errorBody is an io.ReadCloser that always returns a configurable error on Read.
type errorBody struct {
	err error
}

func (e *errorBody) Read([]byte) (int, error) { return 0, e.err }
func (e *errorBody) Close() error             { return nil }

// customPool is a minimal [gc.Pool] for testing [WithBufferPool].
type customPool struct{}

func (customPool) Get() gc.Buffer { return nil }
func (customPool) Put(gc.Buffer)  {}

// roundTripFunc adapts a function into an [http.RoundTripper].
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDoReadBodyError(t *testing.T) {
	c := New("proj", "key",
		WithRetries(0),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       &errorBody{err: io.ErrUnexpectedEOF},
				}, nil
			}),
		}),
	)

	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	// With 0 retries the read error becomes the last error and retries are exhausted.
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
}

func TestDoReadBodyErrorRetried(t *testing.T) {
	var attempt atomic.Int32
	c := New("proj", "key",
		WithRetries(2),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				n := attempt.Add(1)
				if n <= 2 {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       &errorBody{err: io.ErrUnexpectedEOF},
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(io.LimitReader(errlessReader("ok"), 2)),
				}, nil
			}),
		}),
	)

	// First two attempts fail on body read; third succeeds.
	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	assert.Equal(t, "ok", string(data))
	assert.Equal(t, int32(3), attempt.Load())
}

// errlessReader is a string-backed reader that avoids importing strings.
type errlessReader string

func (r errlessReader) Read(p []byte) (int, error) {
	n := copy(p, string(r))
	return n, io.EOF
}

func TestDoReadBodyNonRetryableError(t *testing.T) {
	var attempt atomic.Int32
	c := New("proj", "key",
		WithRetries(3),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				attempt.Add(1)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       &errorBody{err: &x509.UnknownAuthorityError{}},
				}, nil
			}),
		}),
	)

	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	// Must fail immediately with ErrRequestFailed, not exhaust retries.
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailed)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	assert.Contains(t, err.Error(), "permanent error")
	// Only one attempt — no retries wasted.
	assert.Equal(t, int32(1), attempt.Load())
}

func TestDoResponseBodyExceedsLimit(t *testing.T) {
	// Build a body that is one byte over DefaultMaxResponseSize (1 MB)
	// so the size check rejects it.
	oversized := bytes.Repeat([]byte("x"), int(DefaultMaxResponseSize)+1)

	c := New("proj", "key",
		WithRetries(0),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(oversized)),
				}, nil
			}),
		}),
	)

	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, sdkerrors.ErrResponseTooLarge)
}

func TestDoResponseBodyCustomLimit(t *testing.T) {
	const customLimit int64 = 512

	c := New("proj", "key",
		WithRetries(0),
		WithMaxResponseSize(customLimit),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				body := bytes.Repeat([]byte("x"), int(customLimit)+1)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
				}, nil
			}),
		}),
	)

	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, sdkerrors.ErrResponseTooLarge)
}

func TestDoResponseBodyExactlyAtLimit(t *testing.T) {
	// A body of exactly DefaultMaxResponseSize bytes must be accepted.
	exactBody := bytes.Repeat([]byte("x"), int(DefaultMaxResponseSize))

	c := New("proj", "key",
		WithRetries(0),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(exactBody)),
				}, nil
			}),
		}),
	)

	data, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	assert.Len(t, data, int(DefaultMaxResponseSize))
}

func TestDoResponseBodyExceedsLimitNotRetried(t *testing.T) {
	var attempt atomic.Int32
	oversized := bytes.Repeat([]byte("x"), int(DefaultMaxResponseSize)+1)

	c := New("proj", "key",
		WithRetries(3),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
		WithHTTPClient(&http.Client{
			Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				attempt.Add(1)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(oversized)),
				}, nil
			}),
		}),
	)

	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	// Must fail immediately with ErrRequestFailed, not exhaust retries.
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailed)
	assert.ErrorIs(t, err, sdkerrors.ErrResponseTooLarge)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	// Only one attempt — no retries wasted on deterministic failure.
	assert.Equal(t, int32(1), attempt.Load())
}

// --- GetBufferPool ---

func TestGetBufferPool(t *testing.T) {
	c := New("proj", "key")
	assert.NotNil(t, c.GetBufferPool())
}

// --- Getters ---

func TestGetters(t *testing.T) {
	c := New("my-project", "my-key", WithLanguage(i18n.Indonesian))
	assert.Equal(t, "my-project", c.Project())
	assert.Equal(t, "my-key", c.APIKey())
	assert.Equal(t, i18n.Indonesian, c.Lang())
}

// --- calculateBackoff ---

func TestCalculateBackoff(t *testing.T) {
	c := New("proj", "key", WithRetryWait(100*time.Millisecond, 1*time.Second))

	// Attempt 1: base = 100ms * 2^0 = 100ms, jitter range is [100ms, 100ms] → always 100ms.
	assert.Equal(t, 100*time.Millisecond, c.calculateBackoff(1))

	// Attempt 2: base = 100ms * 2^1 = 200ms, jitter in [100ms, 200ms].
	for range 20 {
		b := c.calculateBackoff(2)
		assert.GreaterOrEqual(t, b, 100*time.Millisecond)
		assert.LessOrEqual(t, b, 200*time.Millisecond)
	}

	// Attempt 5: clamped to max 1s, jitter in [100ms, 1s].
	for range 20 {
		b := c.calculateBackoff(5)
		assert.GreaterOrEqual(t, b, 100*time.Millisecond)
		assert.LessOrEqual(t, b, 1*time.Second)
	}
}

// --- Options ---

func TestWithHTTPClientNilIgnored(t *testing.T) {
	c := New("proj", "key", WithHTTPClient(nil))
	assert.NotNil(t, c.httpClient, "nil http client must be ignored")
}

func TestWithBufferPoolNilIgnored(t *testing.T) {
	c := New("proj", "key", WithBufferPool(nil))
	assert.NotNil(t, c.GetBufferPool(), "nil buffer pool must be ignored")
}

func TestWithBufferPoolCustom(t *testing.T) {
	custom := &customPool{}
	c := New("proj", "key", WithBufferPool(custom))
	assert.Same(t, custom, c.GetBufferPool(), "custom pool must be set")
}

func TestWithRetriesNegativeClamped(t *testing.T) {
	c := New("proj", "key", WithRetries(-5))
	assert.Equal(t, 0, c.retries, "negative retries must be clamped to 0")
}

func TestWithTimeoutZeroIgnored(t *testing.T) {
	c := New("proj", "key", WithTimeout(0))
	assert.Equal(t, DefaultTimeout, c.httpClient.Timeout, "zero timeout must be ignored")
}

func TestWithTimeoutNegativeIgnored(t *testing.T) {
	c := New("proj", "key", WithTimeout(-1*time.Second))
	assert.Equal(t, DefaultTimeout, c.httpClient.Timeout, "negative timeout must be ignored")
}

func TestWithRetryWaitSwapped(t *testing.T) {
	c := New("proj", "key", WithRetryWait(5*time.Second, 1*time.Second))
	assert.Equal(t, 1*time.Second, c.retryWaitMin, "min > max must be swapped")
	assert.Equal(t, 5*time.Second, c.retryWaitMax, "min > max must be swapped")
}

func TestWithRetryWaitZeroClamped(t *testing.T) {
	c := New("proj", "key", WithRetryWait(0, 0))
	assert.Equal(t, 1*time.Millisecond, c.retryWaitMin, "zero min must be clamped to 1ms")
	assert.Equal(t, 1*time.Millisecond, c.retryWaitMax, "zero max must be clamped to 1ms")
}

func TestWithRetryWaitNegativeClamped(t *testing.T) {
	c := New("proj", "key", WithRetryWait(-5*time.Second, -1*time.Second))
	assert.Equal(t, 1*time.Millisecond, c.retryWaitMin, "negative min must be clamped to 1ms")
	assert.Equal(t, 1*time.Millisecond, c.retryWaitMax, "negative max must be clamped to 1ms")
}

func TestWithMaxResponseSize(t *testing.T) {
	c := New("proj", "key", WithMaxResponseSize(5<<20))
	assert.Equal(t, int64(5<<20), c.maxResponseSize)
}

func TestWithMaxResponseSizeZeroIgnored(t *testing.T) {
	c := New("proj", "key", WithMaxResponseSize(0))
	assert.Equal(t, DefaultMaxResponseSize, c.maxResponseSize, "zero must be ignored")
}

func TestWithMaxResponseSizeNegativeIgnored(t *testing.T) {
	c := New("proj", "key", WithMaxResponseSize(-1))
	assert.Equal(t, DefaultMaxResponseSize, c.maxResponseSize, "negative must be ignored")
}

func TestWithBaseURLTrailingSlashStripped(t *testing.T) {
	c := New("proj", "key", WithBaseURL("https://app.pakasir.com/"))
	assert.Equal(t, "https://app.pakasir.com", c.baseURL, "trailing slash must be stripped")
}

func TestWithBaseURLMultipleTrailingSlashesStripped(t *testing.T) {
	c := New("proj", "key", WithBaseURL("https://app.pakasir.com///"))
	assert.Equal(t, "https://app.pakasir.com", c.baseURL, "all trailing slashes must be stripped")
}

func TestWithBaseURLNoTrailingSlashUnchanged(t *testing.T) {
	c := New("proj", "key", WithBaseURL("https://app.pakasir.com"))
	assert.Equal(t, "https://app.pakasir.com", c.baseURL, "URL without trailing slash must be unchanged")
}

// --- isRetryableStatus ---

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"429", http.StatusTooManyRequests, true},
		{"502", http.StatusBadGateway, true},
		{"503", http.StatusServiceUnavailable, true},
		{"504", http.StatusGatewayTimeout, true},
		{"500", http.StatusInternalServerError, false},
		{"400", http.StatusBadRequest, false},
		{"401", http.StatusUnauthorized, false},
		{"403", http.StatusForbidden, false},
		{"404", http.StatusNotFound, false},
		{"409", http.StatusConflict, false},
		{"200", http.StatusOK, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isRetryableStatus(tt.status))
		})
	}
}

// --- isRetryable ---

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"generic network error", errors.New("timeout"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"response too large", sdkerrors.ErrResponseTooLarge, false},
		{"wrapped response too large", fmt.Errorf("%w: exceeds 1048576 bytes", sdkerrors.ErrResponseTooLarge), false},
		{"tls unknown authority", &x509.UnknownAuthorityError{}, false},
		{"tls hostname error", &x509.HostnameError{}, false},
		{"tls cert invalid", &x509.CertificateInvalidError{}, false},
		{"tls cert verification", &tls.CertificateVerificationError{}, false},
		{"url error wrapping transient", &neturl.Error{Op: "Get", Err: errors.New("timeout")}, true},
		{"url error wrapping tls", &neturl.Error{Op: "Get", Err: &x509.UnknownAuthorityError{}}, false},
		{"dns not found", &net.DNSError{Err: "no such host", Name: "bad.example.com", IsNotFound: true}, false},
		{"dns timeout", &net.DNSError{Err: "i/o timeout", Name: "slow.example.com", IsTimeout: true}, true},
		{"url error wrapping dns not found", &neturl.Error{Op: "Get", Err: &net.DNSError{Err: "no such host", Name: "bad.example.com", IsNotFound: true}}, false},
		{"system roots unavailable", &x509.SystemRootsError{}, false},
		{"tls alert bad certificate", tls.AlertError(42), false},
		{"tls alert handshake failure", tls.AlertError(40), false},
		{"tls record header error", tls.RecordHeaderError{Msg: "not TLS"}, false},
		{"url error wrapping tls alert", &neturl.Error{Op: "Get", Err: tls.AlertError(70)}, false},
		{"ech rejection", &tls.ECHRejectionError{}, false},
		{"url error wrapping ech rejection", &neturl.Error{Op: "Get", Err: &tls.ECHRejectionError{}}, false},
		{"addr error", &net.AddrError{Err: "mismatched address", Addr: "bad"}, false},
		{"unknown network error", net.UnknownNetworkError("tcp7"), false},
		{"invalid addr error", net.InvalidAddrError("wrong type"), false},
		{"url error wrapping addr error", &neturl.Error{Op: "Get", Err: &net.AddrError{Err: "bad", Addr: "x"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isRetryable(tt.err))
		})
	}
}

// --- Do non-retryable network error ---

func TestDoNonRetryableNetworkError(t *testing.T) {
	// Use a TLS server with a self-signed cert so the client gets
	// an x509.UnknownAuthorityError, which is non-retryable.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Use a plain HTTP client (no custom TLS config) so the cert fails.
	c := New("proj", "key",
		WithBaseURL(srv.URL),
		WithHTTPClient(&http.Client{}),
		WithRetries(2),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
	)

	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)
	// Must fail immediately with ErrRequestFailed, not ErrRequestFailedAfterRetries.
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailed)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	assert.Contains(t, err.Error(), "permanent error")
}

// --- QR ---

func TestQRDefaultNotNil(t *testing.T) {
	c := New("proj", "key")
	require.NotNil(t, c.QR(), "default QR generator must not be nil")

	png, err := c.QR().Encode("hello")
	require.NoError(t, err)
	assert.NotEmpty(t, png)
}

func TestWithQRCodeOptionsCustom(t *testing.T) {
	c := New("proj", "key", WithQRCodeOptions(qr.WithSize(512), qr.WithRecoveryLevel(qr.RecoveryHigh)))
	require.NotNil(t, c.QR())

	// Verify the custom options are applied by encoding and checking output.
	png, err := c.QR().Encode("test")
	require.NoError(t, err)
	assert.NotEmpty(t, png)
}

// --- parseRetryAfter ---

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   time.Duration
	}{
		{"empty", "", 0},
		{"seconds", "2", 2 * time.Second},
		{"seconds with spaces", "  5  ", 5 * time.Second},
		{"zero seconds", "0", 0},
		{"negative seconds", "-1", 0},
		{"invalid string", "not-a-number", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parseRetryAfter(tt.header))
		})
	}
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	// Use a future date so the result is positive.
	future := time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)
	d := parseRetryAfter(future)
	// Should be roughly 10s (allow some slack for test execution).
	assert.Greater(t, d, 5*time.Second, "HTTP-date in the future should produce a positive duration")
	assert.Less(t, d, 15*time.Second, "HTTP-date duration should be reasonable")
}

func TestParseRetryAfterHTTPDatePast(t *testing.T) {
	// A past date should return 0 (negative durations are not useful).
	past := time.Now().Add(-10 * time.Second).UTC().Format(http.TimeFormat)
	assert.Equal(t, time.Duration(0), parseRetryAfter(past))
}

// --- AsType end-to-end ---

func TestAsTypeAPIErrorOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)

	apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err)
	require.True(t, ok, "4xx error must be extractable via AsType")
	assert.Equal(t, 403, apiErr.StatusCode)
	assert.Contains(t, apiErr.Body, "forbidden")
}

func TestAsTypeAPIErrorOnRetriesExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"unavailable"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(1),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
	)
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)

	// The last error in the chain is an APIError wrapped inside the
	// retries-exhausted error. AsType must traverse the chain to find it.
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err)
	require.True(t, ok, "APIError must be reachable inside retries-exhausted chain")
	assert.Equal(t, 503, apiErr.StatusCode)
	assert.Contains(t, apiErr.Body, "unavailable")
}

func TestAsTypeNoMatchOnValidationError(t *testing.T) {
	c := New("", "my-key")
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)

	// Validation errors are not APIErrors — AsType must return false.
	apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err)
	assert.False(t, ok, "validation error must not match APIError")
	assert.Nil(t, apiErr)
}

func TestAsTypeAPIErrorOnPermanentFailure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New("proj", "key",
		WithBaseURL(srv.URL),
		WithHTTPClient(&http.Client{}),
		WithRetries(0),
	)
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	t.Log(err)

	// Permanent TLS errors are not APIErrors — AsType must return false.
	apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err)
	assert.False(t, ok, "TLS permanent error must not match APIError")
	assert.Nil(t, apiErr)
}

// --- stopRetry ---

func TestStopRetryErrorAndUnwrap(t *testing.T) {
	inner := errors.New("permanent failure")
	sr := &stopRetry{err: inner}
	assert.Equal(t, "permanent failure", sr.Error())
	assert.Equal(t, inner, sr.Unwrap())
}
