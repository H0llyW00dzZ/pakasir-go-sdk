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

package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

// Sentinel errors for programmatic handling via [errors.Is].
var (
	// ErrInvalidProject is returned when the project slug is empty.
	ErrInvalidProject = errors.New("invalid project")

	// ErrInvalidAPIKey is returned when the API key is empty.
	ErrInvalidAPIKey = errors.New("invalid api key")

	// ErrInvalidAmount is returned when the amount is not greater than zero.
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrInvalidOrderID is returned when the order ID is empty.
	ErrInvalidOrderID = errors.New("invalid order id")

	// ErrInvalidPaymentMethod is returned when an unsupported payment method is used.
	ErrInvalidPaymentMethod = errors.New("invalid payment method")

	// ErrRequestFailed is returned when the API returns a non-2xx status code.
	ErrRequestFailed = errors.New("request failed")

	// ErrRequestFailedAfterRetries is returned when all retry attempts are exhausted.
	ErrRequestFailedAfterRetries = errors.New("request failed after retries")
)

// APIError represents an error response from the Pakasir API.
// It captures the HTTP status code and the raw response body
// for diagnostic purposes.
type APIError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface for [APIError].
func (e *APIError) Error() string {
	return fmt.Sprintf("pakasir api error: status %d: %s", e.StatusCode, e.Body)
}

// New creates a localized error wrapping the provided sentinel error.
// The localized message is resolved using the given language and message key.
//
// Additional variadic arguments are inspected for:
//   - error: used as an additional wrapped cause
//   - string: used as contextual detail (e.g., an invalid method name)
//
// The returned error supports [errors.Is] against the sentinel.
func New(lang i18n.Language, sentinel error, key i18n.MessageKey, args ...any) error {
	msg := formatMessage(i18n.Get(lang, key), parseContextStr(args))

	if cause := parseCause(args); cause != nil {
		return fmt.Errorf("%s: %w: %w", msg, sentinel, cause)
	}
	return fmt.Errorf("%s: %w", msg, sentinel)
}

// parseCause returns the first error value found in args, or nil.
func parseCause(args []any) error {
	for _, arg := range args {
		if e, ok := arg.(error); ok {
			return e
		}
	}
	return nil
}

// parseContextStr returns the first non-empty string value found in args.
func parseContextStr(args []any) string {
	for _, arg := range args {
		if s, ok := arg.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// formatMessage applies contextStr to msg. If msg contains a %s verb the
// context is substituted in; otherwise it is appended as a suffix.
func formatMessage(msg, contextStr string) string {
	if contextStr == "" {
		return msg
	}
	if strings.Contains(msg, "%s") {
		return fmt.Sprintf(msg, contextStr)
	}
	return msg + ": " + contextStr
}

// NewWithFormat creates a localized error wrapping the provided sentinel error,
// formatting the message with the given format arguments (e.g., status codes, retry counts).
func NewWithFormat(lang i18n.Language, sentinel error, key i18n.MessageKey, formatArgs ...any) error {
	msg := fmt.Sprintf(i18n.Get(lang, key), formatArgs...)
	return fmt.Errorf("%s: %w", msg, sentinel)
}
