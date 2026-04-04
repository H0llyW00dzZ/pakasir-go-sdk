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

// Package client provides the core HTTP client for the Pakasir payment gateway SDK.
//
// It handles authentication, request execution, buffer pooling, and
// automatic retry with exponential backoff and jitter for transient failures.
//
// # Basic Usage
//
//	c := client.New("my-project", "api-key-xxx",
//	    client.WithTimeout(10 * time.Second),
//	    client.WithLanguage(i18n.Indonesian),
//	)
//
//	// Use c with service packages:
//	txnService := transaction.NewService(c)
//
// # Configuration
//
// The client supports functional options for customization:
//
//   - [WithBaseURL]: Override the API base URL (e.g., for staging)
//   - [WithHTTPClient]: Provide a custom [http.Client]
//   - [WithTimeout]: Set the HTTP request timeout (zero/negative ignored)
//   - [WithLanguage]: Set the locale for SDK error messages
//   - [WithRetries]: Configure the number of retry attempts
//   - [WithRetryWait]: Configure backoff min/max durations (auto-swapped if inverted)
//   - [WithBufferPool]: Provide a custom buffer pool
//
// # Encapsulation
//
// All [Client] struct fields are unexported. Read-only access is provided
// via getter methods: [Client.Project], [Client.APIKey], [Client.Lang],
// and [Client.GetBufferPool]. Configuration must be done through [New]
// and functional options.
//
// # Thread Safety
//
// A [Client] must not be modified after first use. Concurrent calls to
// [Client.Do] are safe.
//
// # Retry Logic
//
// The client automatically retries requests that encounter transient
// failures (5xx server errors and network errors) using exponential
// backoff with full jitter. Client errors (4xx) and permanent TLS
// certificate errors are never retried.
package client
