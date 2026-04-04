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
	"fmt"
	"io"
	"net/http"
	"time"
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
	Status string `json:"status"`

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

// Parse reads and decodes a Pakasir webhook POST body from an [http.Request].
//
// It reads the full request body, closes it, and unmarshals the JSON
// into an [Event] struct.
//
// Important: Callers must validate the returned Event's Amount and OrderID
// against their own system records, as recommended by the Pakasir documentation.
func Parse(r *http.Request) (*Event, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("webhook: request body is nil")
	}
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("webhook: failed to read body: %w", err)
	}

	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("webhook: failed to decode body: %w", err)
	}

	return &event, nil
}
