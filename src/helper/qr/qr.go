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

package qr

import (
	"errors"
	"fmt"
	"io"
	"os"

	goqrcode "github.com/skip2/go-qrcode"
)

// Sentinel errors for programmatic handling via [errors.Is].
var (
	// ErrEmptyContent is returned when an empty string is passed to [QR.Encode].
	ErrEmptyContent = errors.New("qr: content must not be empty")

	// ErrEncodeFailed is returned when the underlying QR encoding library
	// fails to generate a QR code (e.g., content exceeds maximum capacity).
	ErrEncodeFailed = errors.New("qr: encode failed")
)

// QR provides QR code generation with configurable options.
//
// Create a new instance using [New]. A QR instance is safe for concurrent
// use after creation.
type QR struct{ cfg *config }

// New creates a new QR code generator with the given options.
//
// If no options are provided, the generator uses default settings:
// 256×256 pixels, medium recovery level, black foreground, white background.
//
// Example:
//
//	q := qr.New(qr.WithSize(512), qr.WithRecoveryLevel(qr.RecoveryHigh))
//	png, err := q.Encode(paymentNumber)
func New(opts ...Option) *QR {
	cfg := defaults()
	for _, opt := range opts {
		opt(cfg)
	}
	return &QR{cfg: cfg}
}

// Encode encodes content into a QR code and returns the PNG image as bytes.
//
// The content is typically a QRIS payment string from the PaymentNumber field
// of a transaction create response (e.g., [transaction.PaymentInfo.PaymentNumber]).
//
// Example:
//
//	png, err := q.Encode(paymentInfo.PaymentNumber)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (q *QR) Encode(content string) ([]byte, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}

	code, err := goqrcode.New(content, q.cfg.level)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEncodeFailed, err)
	}

	code.ForegroundColor = q.cfg.foregroundColor
	code.BackgroundColor = q.cfg.backgroundColor

	return code.PNG(q.cfg.size)
}

// Write encodes content into a QR code and writes the PNG image to w.
//
// This is ideal for serving QR codes via HTTP. Set the Content-Type header
// to "image/png" before calling Write.
//
// Example:
//
//	w.Header().Set("Content-Type", "image/png")
//	err := q.Write(w, paymentInfo.PaymentNumber)
func (q *QR) Write(w io.Writer, content string) error {
	png, err := q.Encode(content)
	if err != nil {
		return err
	}

	_, err = w.Write(png)
	return err
}

// WriteFile encodes content into a QR code and saves the PNG image to a file.
//
// The file is created with permissions 0644. If the file already exists,
// it is overwritten.
//
// Example:
//
//	err := q.WriteFile("payment_qr.png", paymentInfo.PaymentNumber)
func (q *QR) WriteFile(filename, content string) error {
	png, err := q.Encode(content)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, png, 0644)
}
