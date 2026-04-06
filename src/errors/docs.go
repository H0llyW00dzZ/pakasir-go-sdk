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
//   - [ErrNilRequest]: nil request pointer passed to a service method
//   - [ErrEncodeJSON]: JSON marshaling of a request body failed
//   - [ErrRequestFailed]: HTTP non-2xx response
//   - [ErrRequestFailedAfterRetries]: all retry attempts exhausted
//
// # API Errors
//
// The [APIError] type captures non-2xx HTTP responses from the Pakasir API,
// including the status code and raw response body for diagnostics.
//
// # Type-Safe Error Extraction
//
// [AsType] is a generic convenience wrapper around the standard library's
// [errors.AsType]. It extracts the first error in the chain matching a
// concrete type, eliminating the need for a separate variable declaration:
//
//	if apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err); ok {
//	    fmt.Printf("status %d: %s\n", apiErr.StatusCode, apiErr.Body)
//	}
//
// This is re-exported so that callers who import this package as sdkerrors
// do not need an additional import of the standard errors package.
//
// # Localized Error Creation
//
// Use [New] to create errors with localized messages:
//
//	err := errors.New(lang, errors.ErrEncodeJSON, i18n.MsgFailedToEncode)
//
//	err := errors.New(i18n.Indonesian, errors.ErrInvalidAmount, i18n.MsgInvalidAmount)
//	// err.Error() => "jumlah harus lebih dari 0: invalid amount"
//	// errors.Is(err, errors.ErrInvalidAmount) => true
//
// [New] inspects variadic arguments for an error cause and/or a string
// context. An error is wrapped with %w so the original cause is preserved
// in the chain; a string is either substituted into a %s verb or appended
// as a suffix.
package errors
