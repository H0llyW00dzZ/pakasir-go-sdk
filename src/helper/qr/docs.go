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

// Package qr provides QR code generation for payment strings.
//
// It is designed for rendering QRIS payment_number values (returned by
// the transaction create API) into PNG images that can be displayed to
// customers in any Go HTTP framework.
//
// # Basic Usage
//
//	q := qr.New()
//	png, err := q.Encode(paymentInfo.PaymentNumber)
//
// # Serving via HTTP
//
// The [QR.Write] method writes a PNG image directly to any [io.Writer],
// making it easy to integrate with any HTTP framework:
//
//	// net/http
//	w.Header().Set("Content-Type", "image/png")
//	err := q.Write(w, paymentInfo.PaymentNumber)
//
//	// Gin
//	c.Header("Content-Type", "image/png")
//	err := q.Write(c.Writer, paymentInfo.PaymentNumber)
//
//	// Echo
//	c.Response().Header().Set("Content-Type", "image/png")
//	err := q.Write(c.Response(), paymentInfo.PaymentNumber)
//
// # Saving to File
//
// The [QR.WriteFile] method encodes and saves a QR code directly to a
// PNG file:
//
//	err := q.WriteFile("payment_qr.png", paymentInfo.PaymentNumber)
//
// # Customization
//
// QR code size, error correction level, and colors are configurable
// via functional options:
//
//	q := qr.New(
//	    qr.WithSize(512),
//	    qr.WithRecoveryLevel(qr.RecoveryHigh),
//	    qr.WithForegroundColor(color.Black),
//	    qr.WithBackgroundColor(color.White),
//	)
//
// # Error Handling
//
// All methods return sentinel errors for programmatic handling via
// [errors.Is]:
//
//   - [ErrEmptyContent]: returned when an empty string is passed
//   - [ErrEncodeFailed]: returned when QR encoding fails (wraps cause)
//
// # Thread Safety
//
// A [QR] instance is safe for concurrent use by multiple goroutines
// after creation.
package qr
