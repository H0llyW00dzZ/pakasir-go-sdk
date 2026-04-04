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

package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
)

// Sentinel errors for programmatic handling via [errors.Is].
var (
	// ErrNilReader is returned when a nil [io.Reader] is passed to [Parse].
	ErrNilReader = errors.New("webhook: reader is nil")

	// ErrNilRequest is returned when a nil [http.Request] or nil body
	// is passed to [ParseRequest].
	ErrNilRequest = errors.New("webhook: request body is nil")

	// ErrEmptyBody is returned when the webhook payload is empty.
	ErrEmptyBody = errors.New("webhook: body is empty")

	// ErrReadBody is returned when reading the webhook body fails.
	ErrReadBody = errors.New("webhook: failed to read body")

	// ErrDecodeBody is returned when JSON decoding of the webhook body fails.
	ErrDecodeBody = errors.New("webhook: failed to decode body")
)

// Event represents a payment notification received from the Pakasir webhook.
//
// When a customer successfully completes a payment, Pakasir sends an HTTP POST
// with this structure to your configured webhook URL.
//
// Important: Always verify that the Amount and OrderID match a pending
// transaction in your system before processing the event.
type Event struct {
	// Amount is the transaction amount.
	Amount int64 `json:"amount"`

	// OrderID is the transaction identifier from your system.
	OrderID string `json:"order_id"`

	// Project is the Pakasir project slug.
	Project string `json:"project"`

	// Status is the transaction status (typically "completed").
	Status constants.TransactionStatus `json:"status"`

	// PaymentMethod is the payment channel used (e.g., "qris", "bni_va").
	PaymentMethod string `json:"payment_method"`

	// CompletedAt is the payment completion timestamp as returned by Pakasir.
	CompletedAt string `json:"completed_at"`
}

// ParseCompletedAt parses the [Event.CompletedAt] field into a [time.Time].
// It attempts RFC3339 and RFC3339Nano formats.
func (e *Event) ParseCompletedAt() (time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, e.CompletedAt)
	if err != nil {
		t, err = time.Parse(time.RFC3339, e.CompletedAt)
	}
	return t, err
}

// Parse decodes a Pakasir webhook payload from an [io.Reader].
//
// This is the framework-agnostic entry point. It works with any Go HTTP
// framework by accepting the request body reader directly:
//
//   - net/http: webhook.Parse(r.Body)
//   - Gin:      webhook.Parse(c.Request.Body)
//   - Echo:     webhook.Parse(c.Request().Body)
//   - Chi:      webhook.Parse(r.Body)
//
// The caller is responsible for closing the reader if required.
//
// Important: Callers must validate the returned Event's Amount and OrderID
// against their own system records, as recommended by the Pakasir documentation.
func Parse(r io.Reader) (*Event, error) {
	if r == nil {
		return nil, ErrNilReader
	}

	// Limit body reads to 1 MB to guard against oversized payloads.
	const maxBodySize = 1 << 20 // 1 MB
	data, err := io.ReadAll(io.LimitReader(r, maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadBody, err)
	}

	return ParseBytes(data)
}

// ParseRequest is a convenience wrapper that decodes a Pakasir webhook
// payload from a standard [http.Request].
//
// It reads the full request body, closes it, and unmarshals the JSON
// into an [Event] struct.
func ParseRequest(r *http.Request) (*Event, error) {
	if r == nil || r.Body == nil {
		return nil, ErrNilRequest
	}
	defer r.Body.Close()

	return Parse(r.Body)
}

// ParseBytes decodes a Pakasir webhook payload from raw bytes.
//
// This is useful for frameworks that provide the body as []byte
// (e.g., Fiber's c.Body()) or when the body has already been read.
func ParseBytes(data []byte) (*Event, error) {
	if len(data) == 0 {
		return nil, ErrEmptyBody
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeBody, err)
	}

	return &event, nil
}
