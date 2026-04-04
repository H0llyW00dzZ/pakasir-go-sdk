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

// Package errors provides structured, localized error types for the Pakasir SDK.
//
// # Sentinel Errors
//
// The package defines sentinel errors for common validation and transport
// failures. These errors support programmatic handling via [errors.Is]:
//
//   - [ErrInvalidProject]: project slug is empty
//   - [ErrInvalidAPIKey]: API key is empty
//   - [ErrInvalidAmount]: amount is not positive
//   - [ErrInvalidOrderID]: order ID is empty
//   - [ErrInvalidPaymentMethod]: unsupported payment method
//   - [ErrRequestFailed]: HTTP non-2xx response
//   - [ErrRequestFailedAfterRetries]: all retry attempts exhausted
//
// # API Errors
//
// The [APIError] type captures non-2xx HTTP responses from the Pakasir API,
// including the status code and raw response body for diagnostics.
//
// # Localized Error Creation
//
// Use [New] to create errors with localized messages:
//
//	err := errors.New(i18n.Indonesian, errors.ErrInvalidAmount, i18n.MsgInvalidAmount)
//	// err.Error() => "jumlah harus lebih dari 0: invalid amount"
//	// errors.Is(err, errors.ErrInvalidAmount) => true
package errors
