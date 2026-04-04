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
// automatic retry with exponential backoff for transient failures.
//
// # Basic Usage
//
//	c, err := client.New("my-project", "api-key-xxx",
//	    client.WithTimeout(10 * time.Second),
//	    client.WithLanguage(i18n.Indonesian),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
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
//   - [WithTimeout]: Set the HTTP request timeout
//   - [WithLanguage]: Set the locale for SDK error messages
//   - [WithRetries]: Configure the number of retry attempts
//   - [WithRetryWait]: Configure backoff min/max durations
//
// # Retry Logic
//
// The client automatically retries requests that encounter transient
// failures (5xx server errors and network errors) using exponential
// backoff. Client errors (4xx) are never retried.
package client
