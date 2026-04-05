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
	"bytes"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleQRIS is a fake QRIS-format payment string for testing.
// This follows the EMV QR Code TLV structure but uses entirely fictitious values.
const sampleQRIS = "00020101021226610016ID.CO.SHOPEE.WWW01189360091800216005230208216005230303UME51440014ID.CO.QRIS.WWW0215ID10243228429300303UME5204792953033605409100003.005802ID5907Pakasir6012KAB. KEBUMEN61055439262230519SP25RZRATEQI2HQ65Q46304A079"

// pngHeader is the magic bytes at the start of every valid PNG file.
var pngHeader = []byte{0x89, 0x50, 0x4E, 0x47}

// --- New ---

func TestNew(t *testing.T) {
	t.Run("returns non-nil with defaults", func(t *testing.T) {
		q := New()
		require.NotNil(t, q)
		assert.Equal(t, DefaultSize, q.cfg.size)
		assert.Equal(t, RecoveryMedium, q.cfg.level)
		assert.Equal(t, color.Black, q.cfg.foregroundColor)
		assert.Equal(t, color.White, q.cfg.backgroundColor)
	})

	t.Run("applies options", func(t *testing.T) {
		q := New(
			WithSize(512),
			WithRecoveryLevel(RecoveryHigh),
			WithForegroundColor(color.RGBA{R: 0, G: 0, B: 128, A: 255}),
			WithBackgroundColor(color.RGBA{R: 240, G: 240, B: 240, A: 255}),
		)
		assert.Equal(t, 512, q.cfg.size)
		assert.Equal(t, RecoveryHigh, q.cfg.level)
	})
}

// --- Encode ---

func TestQREncode(t *testing.T) {
	q := New()

	t.Run("encodes content to PNG bytes", func(t *testing.T) {
		png, err := q.Encode("hello world")
		require.NoError(t, err)
		assert.NotEmpty(t, png)
		assert.Equal(t, pngHeader, png[:4])
	})

	t.Run("encodes QRIS payload", func(t *testing.T) {
		png, err := q.Encode(sampleQRIS)
		require.NoError(t, err)
		assert.NotEmpty(t, png)
		assert.Equal(t, pngHeader, png[:4])
	})

	t.Run("returns ErrEmptyContent for empty string", func(t *testing.T) {
		_, err := q.Encode("")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyContent)
	})

	t.Run("returns ErrEncodeFailed for content too long", func(t *testing.T) {
		// QR codes have a maximum capacity of ~2,953 bytes.
		// Generate content that exceeds this limit.
		longContent := string(make([]byte, 5000))
		_, err := q.Encode(longContent)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEncodeFailed)
		assert.NotErrorIs(t, err, ErrEmptyContent)
		t.Logf("Full error: %v", err)
	})

	t.Run("respects custom size option", func(t *testing.T) {
		small := New(WithSize(64))
		large := New(WithSize(1024))

		smallPNG, err := small.Encode("test")
		require.NoError(t, err)

		largePNG, err := large.Encode("test")
		require.NoError(t, err)

		assert.Greater(t, len(largePNG), len(smallPNG))
	})

	t.Run("respects recovery level option", func(t *testing.T) {
		low := New(WithRecoveryLevel(RecoveryLow))
		high := New(WithRecoveryLevel(RecoveryHighest))

		lowPNG, err := low.Encode("test")
		require.NoError(t, err)

		highPNG, err := high.Encode("test")
		require.NoError(t, err)

		assert.Equal(t, pngHeader, lowPNG[:4])
		assert.Equal(t, pngHeader, highPNG[:4])
	})

	t.Run("respects color options", func(t *testing.T) {
		q := New(
			WithForegroundColor(color.RGBA{R: 0, G: 0, B: 128, A: 255}),
			WithBackgroundColor(color.RGBA{R: 240, G: 240, B: 240, A: 255}),
		)

		png, err := q.Encode("test")
		require.NoError(t, err)
		assert.NotEmpty(t, png)
	})

	t.Run("ignores nil color options", func(t *testing.T) {
		q := New(
			WithForegroundColor(nil),
			WithBackgroundColor(nil),
		)
		png, err := q.Encode("test")
		require.NoError(t, err)
		assert.NotEmpty(t, png)
	})

	t.Run("ignores non-positive size", func(t *testing.T) {
		q := New(WithSize(0))
		png, err := q.Encode("test")
		require.NoError(t, err)
		assert.NotEmpty(t, png)

		q2 := New(WithSize(-1))
		png2, err := q2.Encode("test")
		require.NoError(t, err)
		assert.NotEmpty(t, png2)
	})
}

// --- Write ---

func TestQRWrite(t *testing.T) {
	q := New()

	t.Run("writes PNG to buffer", func(t *testing.T) {
		var buf bytes.Buffer
		err := q.Write(&buf, "hello world")
		require.NoError(t, err)
		assert.NotEmpty(t, buf.Bytes())
		assert.Equal(t, pngHeader, buf.Bytes()[:4])
	})

	t.Run("writes QRIS payload to buffer", func(t *testing.T) {
		large := New(WithSize(512))
		var buf bytes.Buffer
		err := large.Write(&buf, sampleQRIS)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.Bytes())
	})

	t.Run("returns error for empty content", func(t *testing.T) {
		var buf bytes.Buffer
		err := q.Write(&buf, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyContent)
	})
}

// --- WriteFile ---

func TestQRWriteFile(t *testing.T) {
	q := New()

	t.Run("writes PNG to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "test_qr.png")

		err := q.WriteFile(filename, "hello world")
		require.NoError(t, err)

		data, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Equal(t, pngHeader, data[:4])
	})

	t.Run("writes QRIS payload to file with options", func(t *testing.T) {
		high := New(WithSize(512), WithRecoveryLevel(RecoveryHigh))
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "qris_payment.png")

		err := high.WriteFile(filename, sampleQRIS)
		require.NoError(t, err)

		data, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("returns error for empty content", func(t *testing.T) {
		tmpDir := t.TempDir()
		filename := filepath.Join(tmpDir, "empty.png")

		err := q.WriteFile(filename, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrEmptyContent)

		_, err = os.Stat(filename)
		assert.True(t, os.IsNotExist(err))
	})
}

// --- Defaults ---

func TestDefaults(t *testing.T) {
	cfg := defaults()

	assert.Equal(t, DefaultSize, cfg.size)
	assert.Equal(t, RecoveryMedium, cfg.level)
	assert.Equal(t, color.Black, cfg.foregroundColor)
	assert.Equal(t, color.White, cfg.backgroundColor)
}
