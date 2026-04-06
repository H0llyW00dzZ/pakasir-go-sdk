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

package i18n

// MessageKey is a typed key for looking up localized messages.
type MessageKey string

// Message keys used throughout the SDK.
//
// Keys whose translations contain format verbs (e.g., %d) are intended
// for use with [fmt.Sprintf] directly, not through [errors.New].
const (
	MsgInvalidProject            MessageKey = "invalid_project"
	MsgInvalidAPIKey             MessageKey = "invalid_api_key"
	MsgInvalidAmount             MessageKey = "invalid_amount"
	MsgInvalidOrderID            MessageKey = "invalid_order_id"
	MsgInvalidPaymentMethod      MessageKey = "invalid_payment_method"
	MsgNilRequest                MessageKey = "nil_request"
	MsgFailedToEncode            MessageKey = "failed_to_encode"
	MsgRequestFailedPermanent    MessageKey = "request_failed_permanent"
	MsgRequestFailedAfterRetries MessageKey = "request_failed_after_retries"
)

// translations holds all localized messages keyed by [Language] and [MessageKey].
var translations = map[Language]map[MessageKey]string{
	English: {
		MsgInvalidProject:            "project slug is required",
		MsgInvalidAPIKey:             "API key is required",
		MsgInvalidAmount:             "amount must be greater than 0",
		MsgInvalidOrderID:            "order ID is required",
		MsgInvalidPaymentMethod:      "unsupported payment method: %s",
		MsgNilRequest:                "request must not be nil",
		MsgFailedToEncode:            "failed to encode request as JSON",
		MsgRequestFailedPermanent:    "request failed due to permanent error",
		MsgRequestFailedAfterRetries: "request failed after %d retries",
	},
	Indonesian: {
		MsgInvalidProject:            "slug proyek wajib diisi",
		MsgInvalidAPIKey:             "API key wajib diisi",
		MsgInvalidAmount:             "jumlah harus lebih dari 0",
		MsgInvalidOrderID:            "ID pesanan wajib diisi",
		MsgInvalidPaymentMethod:      "metode pembayaran tidak didukung: %s",
		MsgNilRequest:                "permintaan tidak boleh nil",
		MsgFailedToEncode:            "gagal mengenkode permintaan sebagai JSON",
		MsgRequestFailedPermanent:    "permintaan gagal karena kesalahan permanen",
		MsgRequestFailedAfterRetries: "permintaan gagal setelah %d percobaan ulang",
	},
}
