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

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

// Option is a functional option for configuring the [Client].
type Option func(*Client)

// WithBaseURL sets a custom base URL for the Pakasir API.
// This is useful for testing or pointing to a staging environment.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithHTTPClient sets a custom [http.Client] for the Pakasir client.
// Use this to configure custom transports, proxies, or TLS settings.
// A nil value is ignored and the default client is retained.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.HTTPClient = httpClient
		}
	}
}

// WithTimeout sets the timeout for the default HTTP client.
// This option is ignored if [WithHTTPClient] is also provided,
// since the custom client manages its own timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.HTTPClient.Timeout = d
	}
}

// WithLanguage sets the language for SDK error messages.
// Supported values are [i18n.English] and [i18n.Indonesian].
func WithLanguage(lang i18n.Language) Option {
	return func(c *Client) {
		c.Language = lang
	}
}

// WithRetries sets the maximum number of retry attempts for transient failures.
// Set to 0 to disable retries. Negative values are clamped to 0.
func WithRetries(n int) Option {
	return func(c *Client) {
		if n < 0 {
			n = 0
		}
		c.Retries = n
	}
}

// WithRetryWait sets the minimum and maximum wait durations for
// exponential backoff between retries.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Client) {
		c.RetryWaitMin = min
		c.RetryWaitMax = max
	}
}
