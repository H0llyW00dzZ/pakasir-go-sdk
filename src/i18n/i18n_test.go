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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnglish(t *testing.T) {
	tests := []struct {
		key  MessageKey
		want string
	}{
		{MsgInvalidProject, "project slug is required"},
		{MsgInvalidAPIKey, "API key is required"},
		{MsgInvalidAmount, "amount must be greater than 0"},
		{MsgInvalidOrderID, "order ID is required"},
		{MsgInvalidPaymentMethod, "unsupported payment method: %s"},
		{MsgRequestFailedPermanent, "request failed due to permanent error"},
		{MsgRequestFailedAfterRetries, "request failed after %d retries"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			assert.Equal(t, tt.want, Get(English, tt.key))
		})
	}
}

func TestGetIndonesian(t *testing.T) {
	tests := []struct {
		key  MessageKey
		want string
	}{
		{MsgInvalidProject, "slug proyek wajib diisi"},
		{MsgInvalidAPIKey, "API key wajib diisi"},
		{MsgInvalidAmount, "jumlah harus lebih dari 0"},
		{MsgInvalidOrderID, "ID pesanan wajib diisi"},
		{MsgInvalidPaymentMethod, "metode pembayaran tidak didukung: %s"},
		{MsgRequestFailedPermanent, "permintaan gagal karena kesalahan permanen"},
		{MsgRequestFailedAfterRetries, "permintaan gagal setelah %d percobaan ulang"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			assert.Equal(t, tt.want, Get(Indonesian, tt.key))
		})
	}
}

func TestGetFallbackToEnglish(t *testing.T) {
	got := Get("fr", MsgInvalidProject)
	assert.Equal(t, "project slug is required", got, "unsupported language should fall back to English")
}

func TestGetUnknownKeyFallbackToKeyString(t *testing.T) {
	unknownKey := MessageKey("totally_unknown_key")
	assert.Equal(t, string(unknownKey), Get(English, unknownKey))
}

func TestGetUnknownLanguageUnknownKey(t *testing.T) {
	unknownKey := MessageKey("nonexistent_key")
	assert.Equal(t, string(unknownKey), Get("zz", unknownKey))
}
