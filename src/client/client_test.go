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
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
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
	assert.Same(t, customHTTP, c.httpClient)
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
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidProject)
}

func TestDoEmptyAPIKey(t *testing.T) {
	c := New("my-project", "")
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAPIKey)
}

func TestDoEmptyProjectIndonesian(t *testing.T) {
	c := New("", "my-key", WithLanguage(i18n.Indonesian))
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slug proyek wajib diisi")
}

func TestDoSetsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		assert.Contains(t, ua, "pakasir-go-sdk")
		assert.Contains(t, ua, constants.SDKVersion)
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`error`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL,
		WithRetries(2),
		WithRetryWait(1*time.Millisecond, 5*time.Millisecond),
	)
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
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
}

func TestDoNetworkError(t *testing.T) {
	c := newTestClient(t, "http://127.0.0.1:1")
	_, err := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, err)
}

func TestDoInvalidURL(t *testing.T) {
	c := New("proj", "key", WithBaseURL("://invalid"), WithRetries(0))
	_, doErr := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, doErr)
	assert.Contains(t, doErr.Error(), "failed to create request")
}

func TestDoNetworkErrorRetryExhausted(t *testing.T) {
	c := New("proj", "key",
		WithBaseURL("http://127.0.0.1:1"),
		WithRetries(1),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
	)

	_, doErr := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, doErr)
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
	// Must fail immediately with ErrRequestFailed, not exhaust retries.
	assert.ErrorIs(t, err, sdkerrors.ErrRequestFailed)
	assert.NotErrorIs(t, err, sdkerrors.ErrRequestFailedAfterRetries)
	assert.Contains(t, err.Error(), "permanent error")
	// Only one attempt — no retries wasted.
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

// --- isRetryableStatus ---

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"500", http.StatusInternalServerError, true},
		{"502", http.StatusBadGateway, true},
		{"503", http.StatusServiceUnavailable, true},
		{"504", http.StatusGatewayTimeout, true},
		{"400", http.StatusBadRequest, false},
		{"401", http.StatusUnauthorized, false},
		{"403", http.StatusForbidden, false},
		{"404", http.StatusNotFound, false},
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
		{"tls unknown authority", &x509.UnknownAuthorityError{}, false},
		{"tls hostname error", &x509.HostnameError{}, false},
		{"tls cert invalid", &x509.CertificateInvalidError{}, false},
		{"tls cert verification", &tls.CertificateVerificationError{}, false},
		{"url error wrapping transient", &neturl.Error{Op: "Get", Err: errors.New("timeout")}, true},
		{"url error wrapping tls", &neturl.Error{Op: "Get", Err: &x509.UnknownAuthorityError{}}, false},
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
