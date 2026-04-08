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
//   - [WithBaseURL]: Override the API base URL (e.g., for staging); trailing slashes are stripped
//   - [WithHTTPClient]: Provide a custom [http.Client]
//   - [WithTimeout]: Set the HTTP request timeout (zero/negative ignored)
//   - [WithLanguage]: Set the locale for SDK error messages
//   - [WithRetries]: Configure the number of retry attempts
//   - [WithRetryWait]: Configure backoff min/max durations (auto-swapped if inverted; non-positive values clamped to 1ms)
//   - [WithBufferPool]: Provide a custom buffer pool
//   - [WithMaxResponseSize]: Set the maximum response body size (default 1 MB)
//   - [WithQRCodeOptions]: Configure QR code generation settings
//
// # Encapsulation
//
// All [Client] struct fields are unexported. Read-only access is provided
// via getter methods: [Client.Project], [Client.APIKey], [Client.Lang],
// [Client.GetBufferPool], and [Client.QR]. Configuration must be done
// through [New] and functional options.
//
// # Thread Safety
//
// A [Client] must not be modified after first use. Concurrent calls to
// [Client.Do] are safe.
//
// # Retry Logic
//
// The client automatically retries requests that encounter transient
// failures (429 Too Many Requests, gateway errors 502/503/504, and
// network errors) using exponential backoff with full jitter. A 500
// Internal Server Error is not retried because it indicates a server-side
// bug rather than a transient condition. When a 429 response includes
// a Retry-After header (delay-seconds or HTTP-date per RFC 9110), the
// indicated delay is used instead of the calculated backoff, clamped to the
// configured maximum wait ([WithRetryWait]). Delay-seconds values are
// capped at 24 hours before conversion to [time.Duration] to prevent
// integer overflow from malicious headers.
//
// The [Client.Do] method delegates each attempt to unexported helpers
// (validateCredentials, executeAttempt, handleResponse) to keep cyclomatic
// complexity low. Non-retryable errors are signalled internally via an
// unexported stopRetry wrapper so the retry loop breaks immediately
// without wrapping the error in a retries-exhausted message.
//
// Permanent failures are never retried:
//
//   - HTTP 500 Internal Server Error (server bug, not transient)
//   - Client errors (4xx other than 429)
//   - TLS certificate/handshake errors ([tls.CertificateVerificationError], [x509.UnknownAuthorityError], [x509.SystemRootsError], [tls.AlertError], [tls.RecordHeaderError], [tls.ECHRejectionError], etc.)
//   - Permanent DNS failures ([net.DNSError] with IsNotFound: true, i.e., NXDOMAIN)
//   - Address misconfiguration ([net.AddrError], [net.UnknownNetworkError], [net.InvalidAddrError])
//   - Oversized responses ([errors.ErrResponseTooLarge])
//
// DNS timeouts remain retryable. Response body reads are limited to
// [DefaultMaxResponseSize] (1 MB) by default; use [WithMaxResponseSize]
// to adjust. Oversized responses are rejected early and the pooled buffer
// is returned immediately without allocating a copy.
//
// All requests include an Accept: application/json header.
//
// # QR Code Generation
//
// The client exposes a pre-configured QR code generator via [Client.QR].
// This is used to render QRIS payment_number strings (from transaction
// create responses) as PNG images for display to customers.
//
//	c := client.New("my-project", "api-key",
//	    client.WithQRCodeOptions(qr.WithSize(512)),
//	)
//	png, err := c.QR().Encode(paymentInfo.PaymentNumber)
//
// The QR code generator can also be used standalone via the [qr] package:
//
//	q := qr.New(qr.WithSize(512))
//	png, err := q.Encode(paymentNumber)
//
// [errors.ErrResponseTooLarge]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#ErrResponseTooLarge
// [qr]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/qr
//
// [codes.Internal]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
package client
