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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
)

const testPayload = `{"amount":22000,"order_id":"240910HDE7C9","project":"depodomain","status":"completed","payment_method":"qris","completed_at":"2024-09-10T08:07:02.819+07:00","is_sandbox":false}`

const testSandboxPayload = `{"amount":169950,"order_id":"topup-50-1775420177111-wDQscYsR","project":"oneapi-btz","status":"completed","payment_method":"qris","completed_at":"2026-04-05T20:17:03.889362797Z","is_sandbox":true}`

func assertValidEvent(t *testing.T, event *Event) {
	t.Helper()
	assert.Equal(t, int64(22000), event.Amount)
	assert.Equal(t, "240910HDE7C9", event.OrderID)
	assert.Equal(t, "depodomain", event.Project)
	assert.Equal(t, constants.StatusCompleted, event.Status)
	assert.Equal(t, constants.MethodQRIS, event.PaymentMethod)
	assert.False(t, event.IsSandbox)
}

// --- Parse (io.Reader) ---

func TestParseSuccess(t *testing.T) {
	event, err := Parse(strings.NewReader(testPayload))
	require.NoError(t, err)
	assertValidEvent(t, event)
}

func TestParseNilReader(t *testing.T) {
	_, err := Parse(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilReader)
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse(strings.NewReader(`not json`))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDecodeBody)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestParseEmptyReader(t *testing.T) {
	// Parse must return ErrEmptyBody directly for an empty reader
	// without delegating solely to ParseBytes.
	_, err := Parse(strings.NewReader(""))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyBody)
}

func TestParseReadError(t *testing.T) {
	_, err := Parse(errorReader{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrReadBody)
	assert.True(t, errors.Is(err, io.ErrUnexpectedEOF), "should wrap the underlying cause")
}

func TestParseBodyExceedsLimit(t *testing.T) {
	// DefaultMaxBodySize (1 MB) is the limit. Provide one byte over
	// so the size check fires.
	oversized := strings.NewReader(strings.Repeat("x", int(DefaultMaxBodySize)+1))
	_, err := Parse(oversized)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBodyTooLarge)
	assert.Contains(t, err.Error(), "exceeds")
}

func TestParseBodyExactlyAtLimit(t *testing.T) {
	// A body of exactly DefaultMaxBodySize bytes must be accepted
	// (though it will fail JSON decoding since it's not valid JSON).
	exact := strings.NewReader(strings.Repeat("x", int(DefaultMaxBodySize)))
	_, err := Parse(exact)
	require.Error(t, err)
	// Accepted by the size check, rejected by JSON decoding.
	assert.ErrorIs(t, err, ErrDecodeBody)
	assert.NotErrorIs(t, err, ErrBodyTooLarge)
}

func TestParseCustomMaxBodySize(t *testing.T) {
	// Set a tiny limit and verify it rejects a normal-sized payload.
	_, err := Parse(strings.NewReader(testPayload), WithMaxBodySize(16))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBodyTooLarge)
	assert.Contains(t, err.Error(), "exceeds")
}

func TestParseCustomMaxBodySizeLargeEnough(t *testing.T) {
	event, err := Parse(strings.NewReader(testPayload), WithMaxBodySize(4<<20))
	require.NoError(t, err)
	assertValidEvent(t, event)
}

func TestWithMaxBodySizeZeroIgnored(t *testing.T) {
	// Zero is ignored; default applies. Payload fits within 1 MB.
	event, err := Parse(strings.NewReader(testPayload), WithMaxBodySize(0))
	require.NoError(t, err)
	assertValidEvent(t, event)
}

func TestWithMaxBodySizeNegativeIgnored(t *testing.T) {
	event, err := Parse(strings.NewReader(testPayload), WithMaxBodySize(-1))
	require.NoError(t, err)
	assertValidEvent(t, event)
}

func TestParseRequestCustomMaxBodySize(t *testing.T) {
	r := &http.Request{Body: io.NopCloser(strings.NewReader(testPayload))}
	_, err := ParseRequest(r, WithMaxBodySize(16))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBodyTooLarge)
}

// --- ParseRequest (*http.Request) ---

func TestParseRequestSuccess(t *testing.T) {
	r := &http.Request{Body: io.NopCloser(strings.NewReader(testPayload))}
	event, err := ParseRequest(r)
	require.NoError(t, err)
	assertValidEvent(t, event)
}

func TestParseRequestNilRequest(t *testing.T) {
	_, err := ParseRequest(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilRequest)
}

func TestParseRequestNilBody(t *testing.T) {
	_, err := ParseRequest(&http.Request{Body: nil})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilRequest)
}

func TestParseRequestInvalidJSON(t *testing.T) {
	r := &http.Request{Body: io.NopCloser(strings.NewReader(`not json`))}
	_, err := ParseRequest(r)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDecodeBody)
}

type errorReadCloser struct{}

func (errorReadCloser) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (errorReadCloser) Close() error             { return nil }

func TestParseRequestReadError(t *testing.T) {
	_, err := ParseRequest(&http.Request{Body: errorReadCloser{}})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrReadBody)
}

// --- ParseBytes ---

func TestParseBytesSuccess(t *testing.T) {
	event, err := ParseBytes([]byte(testPayload))
	require.NoError(t, err)
	assertValidEvent(t, event)
}

func TestParseBytesEmpty(t *testing.T) {
	_, err := ParseBytes(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyBody)

	_, err = ParseBytes([]byte{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrEmptyBody)
}

func TestParseBytesInvalidJSON(t *testing.T) {
	_, err := ParseBytes([]byte(`not json`))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDecodeBody)
}

// --- Sandbox payload ---

func TestParseSandboxPayload(t *testing.T) {
	event, err := ParseBytes([]byte(testSandboxPayload))
	require.NoError(t, err)
	assert.Equal(t, int64(169950), event.Amount)
	assert.Equal(t, "topup-50-1775420177111-wDQscYsR", event.OrderID)
	assert.Equal(t, "oneapi-btz", event.Project)
	assert.Equal(t, constants.StatusCompleted, event.Status)
	assert.Equal(t, constants.MethodQRIS, event.PaymentMethod)
	assert.True(t, event.IsSandbox)
}

// --- Event.ParseTime ---

func TestEventParseTimeRFC3339(t *testing.T) {
	e := &Event{CompletedAt: "2024-09-10T08:07:02+07:00"}
	ts, err := e.ParseTime()
	require.NoError(t, err)
	assert.Equal(t, 2024, ts.Year())
}

func TestEventParseTimeNano(t *testing.T) {
	e := &Event{CompletedAt: "2024-09-10T08:07:02.819000000+07:00"}
	ts, err := e.ParseTime()
	require.NoError(t, err)
	assert.Equal(t, 2024, ts.Year())
}

func TestEventParseTimeInvalid(t *testing.T) {
	_, err := (&Event{CompletedAt: "bad"}).ParseTime()
	require.Error(t, err)
}
