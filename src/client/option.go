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
	"net/http"
	"time"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/gc"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

// Option is a functional option for configuring the [Client].
type Option func(*Client)

// WithBaseURL sets a custom base URL for the Pakasir API.
// This is useful for testing or pointing to a staging environment.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets a custom [http.Client] for the Pakasir client.
// Use this to configure custom transports, proxies, or TLS settings.
// A nil value is ignored and the default client is retained.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithTimeout sets the timeout on the client's [http.Client].
//
// When combined with [WithHTTPClient], the result depends on ordering:
// if WithTimeout is applied after WithHTTPClient, the custom client's
// timeout is overridden. Apply WithHTTPClient last if the custom client
// should control its own timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.httpClient.Timeout = d
		}
	}
}

// WithLanguage sets the language for SDK error messages.
// Supported values are [i18n.English] and [i18n.Indonesian].
func WithLanguage(lang i18n.Language) Option {
	return func(c *Client) {
		c.language = lang
	}
}

// WithRetries sets the maximum number of retry attempts for transient failures.
// Set to 0 to disable retries. Negative values are clamped to 0.
func WithRetries(n int) Option {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.retries = n
	}
}

// WithBufferPool overrides the default buffer pool used for request
// serialization. A nil value is ignored.
func WithBufferPool(pool gc.Pool) Option {
	return func(c *Client) {
		if pool != nil {
			c.bufferPool = pool
		}
	}
}

// WithRetryWait sets the minimum and maximum wait durations for
// exponential backoff between retries. If min > max, the values are swapped.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Client) {
		if min > max {
			min, max = max, min
		}
		c.retryWaitMin = min
		c.retryWaitMax = max
	}
}
