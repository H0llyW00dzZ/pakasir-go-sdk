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
	"image/color"

	goqrcode "github.com/skip2/go-qrcode"
)

// RecoveryLevel represents the QR code error correction level.
//
// Higher levels allow more of the QR code to be damaged while remaining
// readable, but produce physically larger codes.
type RecoveryLevel = goqrcode.RecoveryLevel

// Error correction levels for QR code generation.
const (
	// RecoveryLow provides approximately 7% error recovery.
	RecoveryLow RecoveryLevel = goqrcode.Low

	// RecoveryMedium provides approximately 15% error recovery (default).
	RecoveryMedium RecoveryLevel = goqrcode.Medium

	// RecoveryHigh provides approximately 25% error recovery.
	RecoveryHigh RecoveryLevel = goqrcode.High

	// RecoveryHighest provides approximately 30% error recovery.
	RecoveryHighest RecoveryLevel = goqrcode.Highest
)

// Default values for QR code generation.
const (
	// DefaultSize is the default QR code image width and height in pixels.
	DefaultSize = 256
)

// config holds the QR code generation settings.
type config struct {
	size            int
	level           RecoveryLevel
	foregroundColor color.Color
	backgroundColor color.Color
}

// defaults returns a config with default settings.
func defaults() *config {
	return &config{
		size:            DefaultSize,
		level:           RecoveryMedium,
		foregroundColor: color.Black,
		backgroundColor: color.White,
	}
}

// Option is a functional option for configuring QR code generation.
//
// Pass these to [New] when creating a [QR] instance.
type Option func(*config)

// WithSize sets the QR code image size in pixels (width and height).
//
// A positive value sets a fixed image size. Non-positive values are ignored,
// and the default size is retained.
// Default is 256 pixels.
//
// Example:
//
//	qr.New(qr.WithSize(512))
func WithSize(pixels int) Option {
	return func(c *config) {
		if pixels > 0 {
			c.size = pixels
		}
	}
}

// WithRecoveryLevel sets the QR code error correction level.
//
// Higher levels make the QR code more resilient to damage but increase size.
// Default is [RecoveryMedium].
//
// Example:
//
//	qr.New(qr.WithRecoveryLevel(qr.RecoveryHigh))
func WithRecoveryLevel(level RecoveryLevel) Option {
	return func(c *config) {
		c.level = level
	}
}

// WithForegroundColor sets the foreground (module) color of the QR code.
//
// A nil value is ignored, and the default color is retained.
// Default is [color.Black].
//
// Example:
//
//	qr.New(qr.WithForegroundColor(color.RGBA{R: 0, G: 0, B: 128, A: 255}))
func WithForegroundColor(fg color.Color) Option {
	return func(c *config) {
		if fg != nil {
			c.foregroundColor = fg
		}
	}
}

// WithBackgroundColor sets the background color of the QR code.
//
// A nil value is ignored, and the default color is retained.
// Default is [color.White].
//
// Example:
//
//	qr.New(qr.WithBackgroundColor(color.RGBA{R: 240, G: 240, B: 240, A: 255}))
func WithBackgroundColor(bg color.Color) Option {
	return func(c *config) {
		if bg != nil {
			c.backgroundColor = bg
		}
	}
}
