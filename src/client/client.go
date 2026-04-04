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
	"fmt"
	"io"
	"math"
	"net/http"
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
//	c, err := client.New("my-project", "api-key-xxx",
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
// It returns an error if the project or API key is empty.
func New(project, apiKey string, opts ...Option) (*Client, error) {
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

	if c.Project == "" {
		return nil, sdkerrors.New(c.Language, sdkerrors.ErrInvalidProject, i18n.MsgInvalidProject)
	}
	if c.APIKey == "" {
		return nil, sdkerrors.New(c.Language, sdkerrors.ErrInvalidAPIKey, i18n.MsgInvalidAPIKey)
	}

	return c, nil
}

// Do executes an HTTP request with the configured retry logic.
// It is exported for use by service packages within the SDK.
//
// The body parameter should be a pre-encoded [io.Reader] (e.g., a [bytes.Buffer]).
// For GET requests, body may be nil.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			waitTime := c.calculateBackoff(attempt)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
			}
		}

		// Check context before making the request.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		url := c.BaseURL + path

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", constants.UserAgent())
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		buf := c.bufferPool.Get()
		_, readErr := buf.ReadFrom(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			buf.Reset()
			c.bufferPool.Put(buf)
			lastErr = readErr
			continue
		}

		data := make([]byte, buf.Len())
		copy(data, buf.Bytes())
		buf.Reset()
		c.bufferPool.Put(buf)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return data, nil
		}

		apiErr := &sdkerrors.APIError{
			StatusCode: resp.StatusCode,
			Body:       string(data),
		}

		if !isRetryable(resp.StatusCode, nil) {
			return nil, apiErr
		}
		lastErr = apiErr
	}

	// All retries exhausted.
	return nil, fmt.Errorf("%s: %w",
		fmt.Sprintf(i18n.Get(c.Language, i18n.MsgRequestFailedAfterRetries), c.Retries),
		lastErr,
	)
}

// GetBufferPool returns the client's buffer pool for use by services.
func (c *Client) GetBufferPool() gc.Pool {
	return c.bufferPool
}

// calculateBackoff returns the wait duration for the given attempt
// using exponential backoff clamped to [RetryWaitMin, RetryWaitMax].
func (c *Client) calculateBackoff(attempt int) time.Duration {
	mult := math.Pow(2, float64(attempt-1))
	wait := min(time.Duration(float64(c.RetryWaitMin)*mult), c.RetryWaitMax)
	return wait
}

// isRetryable determines whether a request should be retried
// based on the HTTP status code and/or error.
func isRetryable(statusCode int, err error) bool {
	if err != nil {
		// Network-level errors are generally retryable.
		return true
	}
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
