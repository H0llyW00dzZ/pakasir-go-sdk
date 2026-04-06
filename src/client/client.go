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
	"math/rand/v2"
	"net/http"
	neturl "net/url"
	"time"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/gc"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/qr"
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

	// maxResponseSize is the maximum number of bytes the client will read
	// from a response body. This guards against unbounded memory consumption
	// from misbehaving servers.
	maxResponseSize = 10 << 20 // 10 MB
)

// Client is the core Pakasir API client.
// It manages HTTP communication, authentication, retry logic,
// and buffer pooling for efficient request processing.
//
// A Client must not be modified after first use. Concurrent calls
// to [Client.Do] are safe.
//
// Create a new client using [New] with functional options:
//
//	c := client.New("my-project", "api-key-xxx",
//	    client.WithTimeout(10 * time.Second),
//	    client.WithLanguage(i18n.Indonesian),
//	)
type Client struct {
	// project is the Pakasir project slug.
	project string
	// apiKey is the Pakasir API key.
	apiKey string
	// baseURL is the Pakasir API base URL.
	baseURL string
	// httpClient is the HTTP client used to make requests.
	httpClient *http.Client
	// language is the language used for SDK error messages.
	language i18n.Language
	// retries is the number of retry attempts.
	retries int
	// retryWaitMin is the minimum wait time between retries.
	retryWaitMin time.Duration
	// retryWaitMax is the maximum wait time between retries.
	retryWaitMax time.Duration
	// bufferPool is the buffer pool used to allocate buffers for request and response bodies.
	bufferPool gc.Pool
	// qrGen is the pre-configured QR code generator.
	qrGen *qr.QR
}

// New creates a new Pakasir API [Client] with the given project slug, API key,
// and optional configuration via functional options.
//
// Credential validation (project and API key) is deferred to [Client.Do],
// so callers do not need to handle an error at initialization time.
func New(project, apiKey string, opts ...Option) *Client {
	c := &Client{
		project:      project,
		apiKey:       apiKey,
		baseURL:      DefaultBaseURL,
		httpClient:   &http.Client{Timeout: DefaultTimeout},
		language:     i18n.English,
		retries:      DefaultRetries,
		retryWaitMin: DefaultRetryWaitMin,
		retryWaitMax: DefaultRetryWaitMax,
		bufferPool:   gc.Default,
		qrGen:        qr.New(),
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
	if c.project == "" {
		return nil, sdkerrors.New(c.language, sdkerrors.ErrInvalidProject, i18n.MsgInvalidProject)
	}
	if c.apiKey == "" {
		return nil, sdkerrors.New(c.language, sdkerrors.ErrInvalidAPIKey, i18n.MsgInvalidAPIKey)
	}

	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		if err := c.waitForRetry(ctx, attempt); err != nil {
			return nil, err
		}

		req, err := c.buildRequest(ctx, method, path, body)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryable(err) {
				return nil, fmt.Errorf("%s: %w: %w",
					i18n.Get(c.language, i18n.MsgRequestFailedPermanent),
					sdkerrors.ErrRequestFailed,
					lastErr,
				)
			}
			continue
		}

		data, readErr := c.readResponseBody(resp)
		if readErr != nil {
			lastErr = readErr
			if !isRetryable(readErr) {
				return nil, fmt.Errorf("%s: %w: %w",
					i18n.Get(c.language, i18n.MsgRequestFailedPermanent),
					sdkerrors.ErrRequestFailed,
					lastErr,
				)
			}
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
		fmt.Sprintf(i18n.Get(c.language, i18n.MsgRequestFailedAfterRetries), c.retries),
		sdkerrors.ErrRequestFailedAfterRetries,
		lastErr,
	)
}

// GetBufferPool returns the client's buffer pool for use by services.
func (c *Client) GetBufferPool() gc.Pool { return c.bufferPool }

// Project returns the Pakasir project slug.
func (c *Client) Project() string { return c.project }

// APIKey returns the API key for authenticating requests.
func (c *Client) APIKey() string { return c.apiKey }

// Lang returns the language used for SDK error messages.
func (c *Client) Lang() i18n.Language { return c.language }

// QR returns the client's pre-configured QR code generator.
//
// The returned [qr.QR] instance can be used to encode QRIS payment strings
// into PNG images. Configure QR options via [WithQRCodeOptions] when
// creating the client.
//
// Example:
//
//	png, err := c.QR().Encode(paymentInfo.PaymentNumber)
//
//	// Or serve directly via HTTP:
//	w.Header().Set("Content-Type", "image/png")
//	err := c.QR().Write(w, paymentInfo.PaymentNumber)
func (c *Client) QR() *qr.QR { return c.qrGen }

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

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", constants.UserAgent())
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// readResponseBody drains the response into a pooled buffer, copies the
// data out, and returns the buffer to the pool. The read is limited to
// [maxResponseSize] bytes to guard against unbounded memory consumption.
func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	buf := c.bufferPool.Get()
	_, readErr := buf.ReadFrom(io.LimitReader(resp.Body, maxResponseSize))
	resp.Body.Close()

	if readErr != nil {
		buf.Reset()
		c.bufferPool.Put(buf)
		return nil, fmt.Errorf("reading response body: %w", readErr)
	}

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	buf.Reset()
	c.bufferPool.Put(buf)
	return data, nil
}

// calculateBackoff returns the wait duration for the given attempt
// using exponential backoff with full jitter, clamped to [RetryWaitMin, RetryWaitMax].
// The jitter randomizes the wait in [RetryWaitMin, computed] to avoid thundering herd.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	mult := math.Pow(2, float64(attempt-1))
	wait := min(time.Duration(float64(c.retryWaitMin)*mult), c.retryWaitMax)
	// Full jitter: uniform random in [retryWaitMin, wait].
	if wait > c.retryWaitMin {
		wait = c.retryWaitMin + time.Duration(rand.Int64N(int64(wait-c.retryWaitMin+1)))
	}
	return wait
}

// isRetryableStatus determines whether a request should be retried
// based on the HTTP status code.
func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
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
