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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSuccess(t *testing.T) {
	body := `{"amount":22000,"order_id":"240910HDE7C9","project":"depodomain","status":"completed","payment_method":"qris","completed_at":"2024-09-10T08:07:02.819+07:00"}`
	r := &http.Request{Body: io.NopCloser(strings.NewReader(body))}

	event, err := Parse(r)
	require.NoError(t, err)
	assert.Equal(t, int64(22000), event.Amount)
	assert.Equal(t, "240910HDE7C9", event.OrderID)
	assert.Equal(t, "depodomain", event.Project)
	assert.Equal(t, "completed", event.Status)
	assert.Equal(t, "qris", event.PaymentMethod)
}

func TestParseNilBody(t *testing.T) {
	_, err := Parse(&http.Request{Body: nil})
	require.Error(t, err)
}

func TestParseInvalidJSON(t *testing.T) {
	r := &http.Request{Body: io.NopCloser(strings.NewReader(`not json`))}
	_, err := Parse(r)
	require.Error(t, err)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (errorReader) Close() error             { return nil }

func TestParseReadError(t *testing.T) {
	_, err := Parse(&http.Request{Body: errorReader{}})
	require.Error(t, err)
}

func TestEventParseCompletedAtRFC3339(t *testing.T) {
	e := &Event{CompletedAt: "2024-09-10T08:07:02+07:00"}
	ts, err := e.ParseCompletedAt()
	require.NoError(t, err)
	assert.Equal(t, 2024, ts.Year())
}

func TestEventParseCompletedAtNano(t *testing.T) {
	e := &Event{CompletedAt: "2024-09-10T08:07:02.819000000+07:00"}
	ts, err := e.ParseCompletedAt()
	require.NoError(t, err)
	assert.Equal(t, 2024, ts.Year())
}

func TestEventParseCompletedAtInvalid(t *testing.T) {
	_, err := (&Event{CompletedAt: "bad"}).ParseCompletedAt()
	require.Error(t, err)
}
