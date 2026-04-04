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

package request

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

func TestValidateOrderAndAmountSuccess(t *testing.T) {
	err := ValidateOrderAndAmount(i18n.English, "INV123", 99000)
	assert.NoError(t, err)
}

func TestValidateOrderAndAmountEmptyOrderID(t *testing.T) {
	err := ValidateOrderAndAmount(i18n.English, "", 99000)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
}

func TestValidateOrderAndAmountZeroAmount(t *testing.T) {
	err := ValidateOrderAndAmount(i18n.English, "INV123", 0)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAmount)
}

func TestValidateOrderAndAmountNegativeAmount(t *testing.T) {
	err := ValidateOrderAndAmount(i18n.English, "INV123", -100)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAmount)
}

func TestValidateOrderAndAmountIndonesian(t *testing.T) {
	err := ValidateOrderAndAmount(i18n.Indonesian, "", 99000)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
	assert.Contains(t, err.Error(), "ID pesanan wajib diisi")
}
