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
	"net"
	"net/http"
	"strconv"
	"strings"
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

	// DefaultMaxResponseSize is the default maximum number of bytes the
	// client will read from a response body. This guards against unbounded
	// memory consumption from misbehaving servers.
	DefaultMaxResponseSize int64 = 1 << 20 // 1 MB
)

// stopRetry wraps an error that must not be retried. The [Client.Do]
// loop checks for this type to break immediately and return the inner
// error to the caller without wrapping it in a retries-exhausted message.
type stopRetry struct{ err error }

func (s *stopRetry) Error() string { return s.err.Error() }
func (s *stopRetry) Unwrap() error { return s.err }

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
	// maxResponseSize is the maximum bytes to read from a response body.
	maxResponseSize int64
	// qrGen is the pre-configured QR code generator.
	qrGen *qr.QR
}

// New creates a new Pakasir API [Client] with the given project slug, API key,
// and optional configuration via functional options.
//
// Options are applied in order. When combining [WithHTTPClient] and
// [WithTimeout], pass [WithHTTPClient] first so that [WithTimeout]
// overrides the custom client's timeout. If [WithHTTPClient] is applied
// last, it replaces the entire HTTP client and discards earlier timeout
// changes.
//
// Credential validation (project and API key) is deferred to [Client.Do],
// so callers do not need to handle an error at initialization time.
func New(project, apiKey string, opts ...Option) *Client {
	c := &Client{
		project:         project,
		apiKey:          apiKey,
		baseURL:         DefaultBaseURL,
		httpClient:      &http.Client{Timeout: DefaultTimeout},
		language:        i18n.English,
		retries:         DefaultRetries,
		retryWaitMin:    DefaultRetryWaitMin,
		retryWaitMax:    DefaultRetryWaitMax,
		bufferPool:      gc.Default,
		qrGen:           qr.New(),
		maxResponseSize: DefaultMaxResponseSize,
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
//
// When the server responds with 429 Too Many Requests and includes a
// Retry-After header, the client respects the indicated delay (clamped
// to [Client.retryWaitMax]) instead of using the calculated backoff.
func (c *Client) Do(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	if err := c.validateCredentials(); err != nil {
		return nil, err
	}

	var lastErr error
	var retryAfterHint time.Duration

	for attempt := 0; attempt <= c.retries; attempt++ {
		if err := c.waitForRetry(ctx, attempt, retryAfterHint); err != nil {
			return nil, err
		}

		data, hint, err := c.executeAttempt(ctx, method, path, body)
		if err == nil {
			return data, nil
		}
		if stop := (*stopRetry)(nil); errors.As(err, &stop) {
			return nil, stop.err
		}
		lastErr = err
		retryAfterHint = hint
	}

	return nil, c.retriesExhaustedError(lastErr)
}

// validateCredentials returns an error if the project slug or API key is empty.
func (c *Client) validateCredentials() error {
	if c.project == "" {
		return sdkerrors.New(c.language, sdkerrors.ErrInvalidProject, i18n.MsgInvalidProject)
	}
	if c.apiKey == "" {
		return sdkerrors.New(c.language, sdkerrors.ErrInvalidAPIKey, i18n.MsgInvalidAPIKey)
	}
	return nil
}

// executeAttempt performs a single HTTP round-trip and interprets the result.
// It returns the response data on success, or an error to be retried.
// A non-zero retryAfterHint is set when the server responds with 429 and
// includes a Retry-After header.
//
// Returning a non-nil error does not necessarily mean the caller should
// retry; [permanentError] results are returned directly to the caller
// of [Client.Do] via a wrapped panic-free sentinel.
func (c *Client) executeAttempt(ctx context.Context, method, path string, body []byte) ([]byte, time.Duration, error) {
	req, err := c.buildRequest(ctx, method, path, body)
	if err != nil {
		return nil, 0, &stopRetry{err}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if !isRetryable(err) {
			return nil, 0, &stopRetry{c.permanentError(err)}
		}
		if ctx.Err() != nil {
			return nil, 0, &stopRetry{ctx.Err()}
		}
		return nil, 0, err
	}

	return c.handleResponse(resp)
}

// handleResponse reads the response body and classifies the HTTP status.
// It returns the body bytes on 2xx, a non-retryable [sdkerrors.APIError]
// on 4xx (except 429), or a retryable error for 5xx/429 statuses.
// A non-zero duration is returned when a 429 response includes a
// Retry-After header.
func (c *Client) handleResponse(resp *http.Response) ([]byte, time.Duration, error) {
	data, readErr := c.readResponseBody(resp)
	if readErr != nil {
		if !isRetryable(readErr) {
			return nil, 0, &stopRetry{c.permanentError(readErr)}
		}
		return nil, 0, readErr
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return data, 0, nil
	}

	apiErr := &sdkerrors.APIError{
		StatusCode: resp.StatusCode,
		Body:       string(data),
	}

	if !isRetryableStatus(resp.StatusCode) {
		return nil, 0, &stopRetry{apiErr}
	}

	var hint time.Duration
	if resp.StatusCode == http.StatusTooManyRequests {
		hint = parseRetryAfter(resp.Header.Get("Retry-After"))
	}
	return nil, hint, apiErr
}

// retriesExhaustedError wraps lastErr with a localized retries-exhausted message.
func (c *Client) retriesExhaustedError(lastErr error) error {
	msg := fmt.Sprintf(i18n.Get(c.language, i18n.MsgRequestFailedAfterRetries), c.retries)
	return fmt.Errorf("%s: %w: %w", msg, sdkerrors.ErrRequestFailedAfterRetries, lastErr)
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

// permanentError wraps err as a non-retryable failure with a localized message.
func (c *Client) permanentError(err error) error {
	return fmt.Errorf("%s: %w: %w",
		i18n.Get(c.language, i18n.MsgRequestFailedPermanent),
		sdkerrors.ErrRequestFailed,
		err,
	)
}

// waitForRetry blocks until the backoff timer fires or the context is
// cancelled. On the first attempt (0) it returns immediately.
//
// If retryAfterHint is positive (parsed from a Retry-After header), it
// is used instead of the calculated exponential backoff, clamped to
// [Client.retryWaitMax].
func (c *Client) waitForRetry(ctx context.Context, attempt int, retryAfterHint time.Duration) error {
	if attempt == 0 {
		return nil
	}

	wait := c.calculateBackoff(attempt)
	if retryAfterHint > 0 {
		wait = min(retryAfterHint, c.retryWaitMax)
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
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
		return nil, fmt.Errorf("client: failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", constants.UserAgent())
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// readResponseBody drains the response into a pooled buffer, copies the
// data out, and returns the buffer to the pool. The read is limited to
// [maxResponseSize]+1 bytes so that a body of exactly [maxResponseSize]
// is accepted while anything larger is rejected with [sdkerrors.ErrResponseTooLarge].
//
// To avoid holding a full buffer when the body exceeds the limit, the
// size check is performed before copying the data out.
func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	buf := c.bufferPool.Get()

	// Read one extra byte beyond the limit so we can distinguish
	// "exactly at limit" (valid) from "over limit" (rejected).
	_, readErr := buf.ReadFrom(io.LimitReader(resp.Body, c.maxResponseSize+1))

	// Check the size before anything else so that on oversize
	// rejection the buffer is returned to the pool immediately
	// without allocating a copy.
	if int64(buf.Len()) > c.maxResponseSize {
		buf.Reset()
		c.bufferPool.Put(buf)
		return nil, fmt.Errorf("%w: exceeds %d bytes", sdkerrors.ErrResponseTooLarge, c.maxResponseSize)
	}

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
// and worth retrying. TLS certificate/handshake errors, system root pool
// errors, permanent DNS failures, oversized responses, and other permanent
// failures return false to avoid wasting retry attempts.
//
// The [errors.AsType] checks traverse the entire error chain, including
// errors nested inside [net/url.Error] via its Unwrap method, so no
// manual unwrapping is required.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Permanent DNS failures (NXDOMAIN) are not transient — do not retry.
	// DNS timeouts remain retryable because IsTimeout is true and
	// IsNotFound is false. Checked before the switch because we need
	// the extracted value to inspect IsNotFound.
	if dnsErr, ok := sdkerrors.AsType[*net.DNSError](err); ok {
		return !dnsErr.IsNotFound
	}

	switch {
	// Oversized responses are deterministic — do not retry.
	case errors.Is(err, sdkerrors.ErrResponseTooLarge):
		return false
	// TLS/x509 certificate and handshake errors are permanent — do not retry.
	case sdkerrors.HasType[*tls.CertificateVerificationError](err),
		sdkerrors.HasType[*x509.UnknownAuthorityError](err),
		sdkerrors.HasType[*x509.HostnameError](err),
		sdkerrors.HasType[*x509.CertificateInvalidError](err),
		sdkerrors.HasType[*x509.SystemRootsError](err),
		sdkerrors.HasType[tls.AlertError](err),
		sdkerrors.HasType[tls.RecordHeaderError](err):
		return false
	// All other network errors (timeouts, connection refused, etc.)
	// are considered transient.
	default:
		return true
	}
}

// parseRetryAfter parses the value of a Retry-After HTTP header.
// It supports both delay-seconds ("120") and HTTP-date
// ("Mon, 07 Apr 2026 02:30:00 GMT") formats per RFC 9110 Section 10.2.3.
// If the header is empty or unparseable, zero is returned, causing the
// caller to fall back to exponential backoff.
func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}

	// Try delay-seconds first (most common for 429 responses).
	// Cap at 24 hours to prevent overflow when converting to time.Duration.
	if seconds, err := strconv.ParseInt(strings.TrimSpace(header), 10, 64); err == nil && seconds > 0 {
		const maxSeconds int64 = 86400 // 24 hours
		return time.Duration(min(seconds, maxSeconds)) * time.Second
	}

	// Try HTTP-date format.
	if t, err := http.ParseTime(header); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}

	return 0
}
