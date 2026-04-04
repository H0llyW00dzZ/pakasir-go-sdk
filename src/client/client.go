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
	"math"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/gc"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

const (
	// DefaultBaseURL is the production Pakasir API base URL.
	DefaultBaseURL = "https://app.pakasir.com"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultRetries is the default number of retry attempts.
	DefaultRetries = 3

	// DefaultRetryWaitMin is the minimum wait time between retries.
	DefaultRetryWaitMin = 1 * time.Second

	// DefaultRetryWaitMax is the maximum wait time between retries.
	DefaultRetryWaitMax = 30 * time.Second
)

// Client is the core Pakasir API client.
// It manages HTTP communication, authentication, retry logic,
// and buffer pooling for efficient request processing.
//
// Create a new client using [New] with functional options:
//
//	c := client.New("my-project", "api-key-xxx",
//	    client.WithTimeout(10 * time.Second),
//	    client.WithLanguage(i18n.Indonesian),
//	)
type Client struct {
	// Project is the Pakasir project slug.
	Project string

	// APIKey is the API key for authenticating requests.
	APIKey string

	// BaseURL is the Pakasir API base URL.
	// Defaults to [DefaultBaseURL].
	BaseURL string

	// HTTPClient is the underlying HTTP client used for requests.
	HTTPClient *http.Client

	// Language determines the locale for SDK error messages.
	Language i18n.Language

	// Retries is the maximum number of retry attempts for transient failures.
	Retries int

	// RetryWaitMin is the minimum backoff duration between retries.
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum backoff duration between retries.
	RetryWaitMax time.Duration

	// bufferPool is the internal buffer pool for JSON serialization.
	bufferPool gc.Pool
}

// New creates a new Pakasir API [Client] with the given project slug, API key,
// and optional configuration via functional options.
//
// Credential validation (project and API key) is deferred to [Client.Do],
// so callers do not need to handle an error at initialization time.
func New(project, apiKey string, opts ...Option) *Client {
	c := &Client{
		Project:      project,
		APIKey:       apiKey,
		BaseURL:      DefaultBaseURL,
		HTTPClient:   &http.Client{Timeout: DefaultTimeout},
		Language:     i18n.English,
		Retries:      DefaultRetries,
		RetryWaitMin: DefaultRetryWaitMin,
		RetryWaitMax: DefaultRetryWaitMax,
		bufferPool:   gc.Default,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Do executes an HTTP request with the configured retry logic.
// It is exported for use by service packages within the SDK.
//
// The body parameter is the pre-encoded request payload. A fresh
// [bytes.Reader] is created for each attempt, so retries always send the
// complete body. For GET requests, body may be nil.
func (c *Client) Do(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	if c.Project == "" {
		return nil, sdkerrors.New(c.Language, sdkerrors.ErrInvalidProject, i18n.MsgInvalidProject)
	}
	if c.APIKey == "" {
		return nil, sdkerrors.New(c.Language, sdkerrors.ErrInvalidAPIKey, i18n.MsgInvalidAPIKey)
	}

	var lastErr error

	for attempt := 0; attempt <= c.Retries; attempt++ {
		if err := c.waitForRetry(ctx, attempt); err != nil {
			return nil, err
		}

		req, err := c.buildRequest(ctx, method, path, body)
		if err != nil {
			return nil, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryable(err) {
				return nil, fmt.Errorf("%s: %w: %w",
					fmt.Sprintf(i18n.Get(c.Language, i18n.MsgRequestFailedAfterRetries), c.Retries),
					sdkerrors.ErrRequestFailed,
					lastErr,
				)
			}
			continue
		}

		data, readErr := c.readResponseBody(resp)
		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return data, nil
		}

		apiErr := &sdkerrors.APIError{
			StatusCode: resp.StatusCode,
			Body:       string(data),
		}

		if !isRetryableStatus(resp.StatusCode) {
			return nil, apiErr
		}
		lastErr = apiErr
	}

	// All retries exhausted.
	return nil, fmt.Errorf("%s: %w: %w",
		fmt.Sprintf(i18n.Get(c.Language, i18n.MsgRequestFailedAfterRetries), c.Retries),
		sdkerrors.ErrRequestFailedAfterRetries,
		lastErr,
	)
}

// GetBufferPool returns the client's buffer pool for use by services.
func (c *Client) GetBufferPool() gc.Pool {
	return c.bufferPool
}

// waitForRetry blocks until the backoff timer fires or the context is
// cancelled. On the first attempt (0) it returns immediately.
func (c *Client) waitForRetry(ctx context.Context, attempt int) error {
	if attempt == 0 {
		return nil
	}

	timer := time.NewTimer(c.calculateBackoff(attempt))
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// buildRequest creates a fully-formed [http.Request] with the SDK
// headers set. A fresh [bytes.Reader] wraps body on each call so
// retries always send the complete payload.
func (c *Client) buildRequest(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", constants.UserAgent())
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// readResponseBody drains the response into a pooled buffer, copies the
// data out, and returns the buffer to the pool.
func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	buf := c.bufferPool.Get()
	_, readErr := buf.ReadFrom(resp.Body)
	resp.Body.Close()

	if readErr != nil {
		buf.Reset()
		c.bufferPool.Put(buf)
		return nil, readErr
	}

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	buf.Reset()
	c.bufferPool.Put(buf)
	return data, nil
}

// calculateBackoff returns the wait duration for the given attempt
// using exponential backoff clamped to [RetryWaitMin, RetryWaitMax].
func (c *Client) calculateBackoff(attempt int) time.Duration {
	mult := math.Pow(2, float64(attempt-1))
	wait := min(time.Duration(float64(c.RetryWaitMin)*mult), c.RetryWaitMax)
	return wait
}

// isRetryableStatus determines whether a request should be retried
// based on the HTTP status code.
func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// isRetryable determines whether a network-level error is transient
// and worth retrying. TLS certificate errors and other permanent
// failures return false to avoid wasting retry attempts.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Unwrap *url.Error to inspect the underlying cause.
	if urlErr, ok := errors.AsType[*neturl.Error](err); ok {
		err = urlErr.Err
	}

	// TLS certificate errors are permanent — do not retry.
	if _, ok := errors.AsType[*tls.CertificateVerificationError](err); ok {
		return false
	}
	if _, ok := errors.AsType[*x509.UnknownAuthorityError](err); ok {
		return false
	}
	if _, ok := errors.AsType[*x509.HostnameError](err); ok {
		return false
	}
	if _, ok := errors.AsType[*x509.CertificateInvalidError](err); ok {
		return false
	}

	// All other network errors (timeouts, connection refused, etc.)
	// are considered transient.
	return true
}
