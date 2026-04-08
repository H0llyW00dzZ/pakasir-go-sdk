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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

func TestAPIErrorError(t *testing.T) {
	e := &APIError{StatusCode: http.StatusBadRequest, Body: "bad request"}
	assert.Equal(t, "pakasir api error: status 400: bad request", e.Error())
}

func TestNewBasic(t *testing.T) {
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidProject)
	assert.Contains(t, err.Error(), "project slug is required")
}

func TestNewIndonesian(t *testing.T) {
	err := New(i18n.Indonesian, ErrInvalidAmount, i18n.MsgInvalidAmount)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidAmount)
	assert.Contains(t, err.Error(), "jumlah harus lebih dari 0")
}

func TestNewWithContextString(t *testing.T) {
	err := New(i18n.English, ErrInvalidPaymentMethod, i18n.MsgInvalidPaymentMethod, "bitcoin")
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidPaymentMethod)
	assert.Contains(t, err.Error(), "unsupported payment method: bitcoin")
}

func TestNewWithCauseError(t *testing.T) {
	cause := fmt.Errorf("network timeout")
	err := New(i18n.English, ErrRequestFailed, i18n.MsgRequestFailedPermanent, cause)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrRequestFailed)
	assert.ErrorIs(t, err, cause)
}

func TestNewWithCauseAndContext(t *testing.T) {
	cause := fmt.Errorf("connection refused")
	err := New(i18n.English, ErrInvalidPaymentMethod, i18n.MsgInvalidPaymentMethod, "dana", cause)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidPaymentMethod)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "unsupported payment method: dana")
}

func TestNewWithEmptyStringIgnored(t *testing.T) {
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject, "")
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidProject)
	assert.Contains(t, err.Error(), "project slug is required")
}

func TestNewWithUnsupportedArgType(t *testing.T) {
	err := New(i18n.English, ErrInvalidProject, i18n.MsgInvalidProject, 42)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidProject)
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrInvalidProject, ErrInvalidAPIKey, ErrInvalidAmount,
		ErrInvalidOrderID, ErrInvalidPaymentMethod,
		ErrNilRequest, ErrEncodeJSON, ErrDecodeJSON,
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
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidProject)
	assert.Contains(t, err.Error(), "project slug is required: extra-context")
	assert.NotContains(t, err.Error(), "EXTRA")
}

func TestNewFormatMessageEmptyContextTrimsSeparator(t *testing.T) {
	// MsgInvalidPaymentMethod = "unsupported payment method: %s" (has %s verb).
	// Empty contextStr should replace %s and trim trailing ": " to prevent
	// dangling separators like "unsupported payment method: : invalid payment method".
	err := New(i18n.English, ErrInvalidPaymentMethod, i18n.MsgInvalidPaymentMethod, "")
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrInvalidPaymentMethod)
	assert.Contains(t, err.Error(), "unsupported payment method: invalid payment method")
	assert.NotContains(t, err.Error(), ": :")
	assert.NotContains(t, err.Error(), "%s")
}

func TestNewMultipleCauses(t *testing.T) {
	cause1 := errors.New("first")
	cause2 := errors.New("second")
	// Only the first error cause should be wrapped.
	err := New(i18n.English, ErrRequestFailed, i18n.MsgRequestFailedPermanent, cause1, cause2)
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, ErrRequestFailed)
	assert.ErrorIs(t, err, cause1)
	assert.NotErrorIs(t, err, cause2)
}

// --- AsType ---

func TestAsTypeMatch(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &APIError{StatusCode: http.StatusNotFound, Body: "not found"})
	apiErr, ok := AsType[*APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "not found", apiErr.Body)
}

func TestAsTypeNoMatch(t *testing.T) {
	err := errors.New("plain error")
	apiErr, ok := AsType[*APIError](err)
	assert.False(t, ok)
	assert.Nil(t, apiErr)
}

func TestAsTypeNilError(t *testing.T) {
	apiErr, ok := AsType[*APIError](nil)
	assert.False(t, ok)
	assert.Nil(t, apiErr)
}

func TestAsTypeDirect(t *testing.T) {
	err := &APIError{StatusCode: http.StatusInternalServerError, Body: "internal"}
	apiErr, ok := AsType[*APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}

func TestAsTypeDeeplyNested(t *testing.T) {
	err := fmt.Errorf("a: %w", fmt.Errorf("b: %w", &APIError{StatusCode: http.StatusBadGateway, Body: "bad gateway"}))
	apiErr, ok := AsType[*APIError](err)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Equal(t, "bad gateway", apiErr.Body)
}

// --- HasType ---

func TestHasTypeMatch(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &APIError{StatusCode: http.StatusForbidden, Body: "forbidden"})
	assert.True(t, HasType[*APIError](err))
}

func TestHasTypeNoMatch(t *testing.T) {
	err := errors.New("plain error")
	assert.False(t, HasType[*APIError](err))
}

func TestHasTypeNilError(t *testing.T) {
	assert.False(t, HasType[*APIError](nil))
}

func TestHasTypeDeeplyNested(t *testing.T) {
	err := fmt.Errorf("a: %w", fmt.Errorf("b: %w", &APIError{StatusCode: http.StatusInternalServerError, Body: "internal"}))
	assert.True(t, HasType[*APIError](err))
}
