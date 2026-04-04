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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

func newTestClient(t *testing.T, serverURL string, opts ...Option) *Client {
	t.Helper()
	allOpts := []Option{WithBaseURL(serverURL), WithRetries(0)}
	allOpts = append(allOpts, opts...)
	c, err := New("test-project", "test-api-key", allOpts...)
	require.NoError(t, err)
	return c
}

// --- New ---

func TestNewSuccess(t *testing.T) {
	c, err := New("my-project", "my-key")
	require.NoError(t, err)
	assert.Equal(t, "my-project", c.Project)
	assert.Equal(t, "my-key", c.APIKey)
	assert.Equal(t, DefaultBaseURL, c.BaseURL)
	assert.Equal(t, i18n.English, c.Language)
	assert.Equal(t, DefaultRetries, c.Retries)
}

func TestNewEmptyProject(t *testing.T) {
	_, err := New("", "my-key")
	require.Error(t, err)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidProject)
}

func TestNewEmptyAPIKey(t *testing.T) {
	_, err := New("my-project", "")
	require.Error(t, err)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAPIKey)
}

func TestNewEmptyProjectIndonesian(t *testing.T) {
	_, err := New("", "my-key", WithLanguage(i18n.Indonesian))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slug proyek wajib diisi")
}

func TestNewWithAllOptions(t *testing.T) {
	customHTTP := &http.Client{Timeout: 5 * time.Second}
	c, err := New("proj", "key",
		WithBaseURL("https://custom.api.com"),
		WithHTTPClient(customHTTP),
		WithLanguage(i18n.Indonesian),
		WithRetries(5),
		WithRetryWait(100*time.Millisecond, 2*time.Second),
	)
	require.NoError(t, err)
	assert.Equal(t, "https://custom.api.com", c.BaseURL)
	assert.Same(t, customHTTP, c.HTTPClient)
	assert.Equal(t, i18n.Indonesian, c.Language)
	assert.Equal(t, 5, c.Retries)
	assert.Equal(t, 100*time.Millisecond, c.RetryWaitMin)
	assert.Equal(t, 2*time.Second, c.RetryWaitMax)
}

func TestWithTimeout(t *testing.T) {
	c, err := New("proj", "key", WithTimeout(5*time.Second))
	require.NoError(t, err)
	assert.Equal(t, 5*time.Second, c.HTTPClient.Timeout)
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
	_, err := c.Do(context.Background(), http.MethodPost, "/test", strings.NewReader(`{}`))
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
	attempt := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt <= 2 {
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
	assert.Equal(t, 3, attempt)
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
	c, err := New("proj", "key", WithBaseURL("://invalid"), WithRetries(0))
	require.NoError(t, err)

	_, doErr := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, doErr)
	assert.Contains(t, doErr.Error(), "failed to create request")
}

func TestDoNetworkErrorRetryExhausted(t *testing.T) {
	c, err := New("proj", "key",
		WithBaseURL("http://127.0.0.1:1"),
		WithRetries(1),
		WithRetryWait(1*time.Millisecond, 2*time.Millisecond),
	)
	require.NoError(t, err)

	_, doErr := c.Do(context.Background(), http.MethodGet, "/test", nil)
	require.Error(t, doErr)
	assert.Contains(t, doErr.Error(), "request failed after")
}

func TestDoReadBodyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100000")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial"))
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, WithRetries(0))
	// Exercises the readErr path — outcome depends on timing.
	c.Do(context.Background(), http.MethodGet, "/test", nil)
}

// --- GetBufferPool ---

func TestGetBufferPool(t *testing.T) {
	c, _ := New("proj", "key")
	assert.NotNil(t, c.GetBufferPool())
}

// --- calculateBackoff ---

func TestCalculateBackoff(t *testing.T) {
	c, _ := New("proj", "key", WithRetryWait(100*time.Millisecond, 1*time.Second))

	assert.Equal(t, 100*time.Millisecond, c.calculateBackoff(1))  // 100ms * 2^0
	assert.Equal(t, 200*time.Millisecond, c.calculateBackoff(2))  // 100ms * 2^1
	assert.Equal(t, 1*time.Second, c.calculateBackoff(5))          // clamped to max
}

// --- isRetryable ---

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name   string
		status int
		err    error
		want   bool
	}{
		{"network error", 0, errors.New("timeout"), true},
		{"500", http.StatusInternalServerError, nil, true},
		{"502", http.StatusBadGateway, nil, true},
		{"503", http.StatusServiceUnavailable, nil, true},
		{"504", http.StatusGatewayTimeout, nil, true},
		{"400", http.StatusBadRequest, nil, false},
		{"401", http.StatusUnauthorized, nil, false},
		{"403", http.StatusForbidden, nil, false},
		{"404", http.StatusNotFound, nil, false},
		{"200", http.StatusOK, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isRetryable(tt.status, tt.err))
		})
	}
}
