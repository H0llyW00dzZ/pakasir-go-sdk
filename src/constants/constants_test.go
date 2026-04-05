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

package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaymentMethodValid(t *testing.T) {
	validMethods := []PaymentMethod{
		MethodCIMBNiagaVA, MethodBNIVA, MethodQRIS, MethodSampoernaVA,
		MethodBNCVA, MethodMaybankVA, MethodPermataVA, MethodATMBersamaVA,
		MethodArthaGrahaVA, MethodBRIVA, MethodPaypal,
	}

	for _, m := range validMethods {
		t.Run(string(m), func(t *testing.T) {
			assert.True(t, m.Valid(), "expected %q to be valid", m)
		})
	}
}

func TestPaymentMethodInvalid(t *testing.T) {
	invalid := []PaymentMethod{"", "unknown", "bitcoin", "dana", "gopay"}
	for _, m := range invalid {
		t.Run(string(m), func(t *testing.T) {
			assert.False(t, m.Valid(), "expected %q to be invalid", m)
		})
	}
}

func TestPaymentMethodString(t *testing.T) {
	assert.Equal(t, "qris", MethodQRIS.String())
	assert.Equal(t, "bni_va", MethodBNIVA.String())
}

func TestTransactionStatusValid(t *testing.T) {
	valid := []TransactionStatus{
		StatusCompleted, StatusPending, StatusExpired, StatusCancelled, StatusCanceled,
	}
	for _, s := range valid {
		t.Run(string(s), func(t *testing.T) {
			assert.True(t, s.Valid(), "expected %q to be valid", s)
		})
	}
}

func TestTransactionStatusInvalid(t *testing.T) {
	invalid := []TransactionStatus{"", "unknown", "failed", "refunded"}
	for _, s := range invalid {
		t.Run(string(s), func(t *testing.T) {
			assert.False(t, s.Valid(), "expected %q to be invalid", s)
		})
	}
}

func TestTransactionStatusString(t *testing.T) {
	assert.Equal(t, "completed", StatusCompleted.String())
	assert.Equal(t, "pending", StatusPending.String())
}

func TestUserAgent(t *testing.T) {
	ua := UserAgent()
	assert.Contains(t, ua, SDKName)
	assert.Contains(t, ua, SDKVersion)
	assert.Contains(t, ua, SDKRepository)
	assert.Equal(t, SDKName+"/"+SDKVersion+" (+"+SDKRepository+")", ua)
}
