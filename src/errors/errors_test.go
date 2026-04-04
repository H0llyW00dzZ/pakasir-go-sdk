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

package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

func TestAPIErrorError(t *testing.T) {
	e := &APIError{StatusCode: 400, Body: "bad request"}
	assert.Equal(t, "pakasir api error: status 400: bad request", e.Error())
}

func TestNewBasic(t *testing.T) {
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidProject)
	assert.Contains(t, err.Error(), "project slug is required")
}

func TestNewIndonesian(t *testing.T) {
	err := New(i18n.Indonesian, ErrInvalidAmount, i18n.MsgInvalidAmount)
	assert.ErrorIs(t, err, ErrInvalidAmount)
	assert.Contains(t, err.Error(), "jumlah harus lebih dari 0")
}

func TestNewWithContextString(t *testing.T) {
	err := New(i18n.English, ErrInvalidPaymentMethod, i18n.MsgInvalidPaymentMethod, "bitcoin")
	assert.ErrorIs(t, err, ErrInvalidPaymentMethod)
	assert.Contains(t, err.Error(), "unsupported payment method: bitcoin")
}

func TestNewWithCauseError(t *testing.T) {
	cause := fmt.Errorf("network timeout")
	err := New(i18n.English, ErrRequestFailed, i18n.MsgRequestFailed, cause)
	assert.ErrorIs(t, err, ErrRequestFailed)
	assert.ErrorIs(t, err, cause)
}

func TestNewWithCauseAndContext(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := New(i18n.English, ErrInvalidPaymentMethod, i18n.MsgInvalidPaymentMethod, "dana", cause)
	assert.ErrorIs(t, err, ErrInvalidPaymentMethod)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "unsupported payment method: dana")
}

func TestNewWithEmptyStringIgnored(t *testing.T) {
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject, "")
	assert.ErrorIs(t, err, ErrInvalidProject)
	assert.Contains(t, err.Error(), "project slug is required")
}

func TestNewWithUnsupportedArgType(t *testing.T) {
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject, 42)
	assert.ErrorIs(t, err, ErrInvalidProject)
}

func TestNewWithFormat(t *testing.T) {
	err := NewWithFormat(i18n.English, ErrRequestFailedAfterRetries, i18n.MsgRequestFailedAfterRetries, 3)
	assert.ErrorIs(t, err, ErrRequestFailedAfterRetries)
	assert.Contains(t, err.Error(), "request failed after 3 retries")
}

func TestNewWithFormatIndonesian(t *testing.T) {
	err := NewWithFormat(i18n.Indonesian, ErrRequestFailed, i18n.MsgRequestFailed, 500)
	assert.ErrorIs(t, err, ErrRequestFailed)
	assert.Contains(t, err.Error(), "permintaan gagal dengan status 500")
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrInvalidProject, ErrInvalidAPIKey, ErrInvalidAmount,
		ErrInvalidOrderID, ErrInvalidPaymentMethod,
		ErrRequestFailed, ErrRequestFailedAfterRetries,
	}
	for _, s := range sentinels {
		t.Run(s.Error(), func(t *testing.T) {
			require.NotNil(t, s)
			assert.NotEmpty(t, s.Error())
		})
	}
}

func TestNewContextStrOnNonFormatMessage(t *testing.T) {
	// MsgInvalidProject = "project slug is required" (no %s verb).
	// Passing a contextStr should append it as a suffix, not garble the message.
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject, "extra-context")
	assert.ErrorIs(t, err, ErrInvalidProject)
	assert.Contains(t, err.Error(), "project slug is required: extra-context")
	assert.NotContains(t, err.Error(), "EXTRA")
}

func TestNewMultipleCauses(t *testing.T) {
	cause1 := errors.New("first")
	cause2 := errors.New("second")
	// Only the first error cause should be wrapped.
	err := New(i18n.English, ErrRequestFailed, i18n.MsgRequestFailed, cause1, cause2)
	assert.ErrorIs(t, err, ErrRequestFailed)
	assert.ErrorIs(t, err, cause1)
}
