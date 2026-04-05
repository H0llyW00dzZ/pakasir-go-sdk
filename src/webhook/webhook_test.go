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
	assert.Equal(t, "qris", event.PaymentMethod)
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

func TestParseReadError(t *testing.T) {
	_, err := Parse(errorReader{})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrReadBody)
	assert.True(t, errors.Is(err, io.ErrUnexpectedEOF), "should wrap the underlying cause")
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
	assert.Equal(t, "qris", event.PaymentMethod)
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
