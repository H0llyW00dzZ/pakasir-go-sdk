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

package transaction

import (
	"time"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/internal/timefmt"
)

// CreateRequest contains the parameters for creating a new transaction.
type CreateRequest struct {
	// OrderID is the unique identifier for this transaction in your system.
	// Example: "INV20240910-123456" or "1298".
	OrderID string `json:"order_id"`

	// Amount is the transaction amount in the smallest currency unit (no decimals).
	// Example: 99000 for Rp 99.000.
	Amount int64 `json:"amount"`
}

// CreateResponse is the API response for a transaction creation.
type CreateResponse struct {
	Payment PaymentInfo `json:"payment"`
}

// PaymentInfo contains the payment details returned by the Pakasir API
// after creating a transaction.
type PaymentInfo struct {
	// Project is the Pakasir project slug.
	Project string `json:"project"`

	// OrderID is the transaction identifier.
	OrderID string `json:"order_id"`

	// Amount is the original requested amount.
	Amount int64 `json:"amount"`

	// Fee is the transaction fee charged by the payment provider.
	Fee int64 `json:"fee"`

	// TotalPayment is the total amount the customer must pay (Amount + Fee).
	TotalPayment int64 `json:"total_payment"`

	// PaymentMethod is the selected payment channel (e.g., "qris", "bni_va").
	PaymentMethod string `json:"payment_method"`

	// PaymentNumber is the QR string or Virtual Account number
	// that the customer uses to complete payment.
	PaymentNumber string `json:"payment_number"`

	// ExpiredAt is the payment expiration timestamp as returned by the API.
	// The format is typically RFC3339 with nanoseconds.
	ExpiredAt string `json:"expired_at"`
}

// ParseTime parses the [PaymentInfo.ExpiredAt] field into a [time.Time].
// It attempts RFC3339Nano first, then falls back to RFC3339.
func (p *PaymentInfo) ParseTime() (time.Time, error) {
	return timefmt.Parse(p.ExpiredAt)
}

// CancelRequest contains the parameters for cancelling a transaction.
type CancelRequest struct {
	// OrderID is the transaction identifier to cancel.
	OrderID string `json:"order_id"`

	// Amount is the transaction amount.
	Amount int64 `json:"amount"`
}

// DetailRequest contains the parameters for querying transaction details.
type DetailRequest struct {
	// OrderID is the transaction identifier to look up.
	OrderID string

	// Amount is the transaction amount.
	Amount int64
}

// DetailResponse is the API response for a transaction detail query.
type DetailResponse struct {
	Transaction TransactionInfo `json:"transaction"`
}

// TransactionInfo contains the transaction details returned by the Pakasir API.
type TransactionInfo struct {
	// Amount is the transaction amount.
	Amount int64 `json:"amount"`

	// OrderID is the transaction identifier.
	OrderID string `json:"order_id"`

	// Project is the Pakasir project slug.
	Project string `json:"project"`

	// Status is the current transaction status (e.g., "completed", "pending").
	Status constants.TransactionStatus `json:"status"`

	// PaymentMethod is the payment channel used.
	PaymentMethod string `json:"payment_method"`

	// CompletedAt is the completion timestamp as returned by the API.
	// The format may vary (RFC3339 with timezone offset).
	CompletedAt string `json:"completed_at"`
}

// ParseTime parses the [TransactionInfo.CompletedAt] field into a [time.Time].
// It attempts RFC3339Nano first, then falls back to RFC3339.
func (t *TransactionInfo) ParseTime() (time.Time, error) {
	return timefmt.Parse(t.CompletedAt)
}
