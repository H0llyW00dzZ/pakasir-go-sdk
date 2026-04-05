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
	"encoding/json"

	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/gc"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
)

// Body is the standard request body structure used by the Pakasir API.
// It is used internally by all services to build consistent API payloads.
type Body struct {
	Project string `json:"project"`
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
	APIKey  string `json:"api_key"`
}

// ValidateOrderAndAmount performs common validation for service requests.
// It returns a localized error if the order ID is empty or the amount is
// not positive.
func ValidateOrderAndAmount(lang i18n.Language, orderID string, amount int64) error {
	if orderID == "" {
		return sdkerrors.New(lang, sdkerrors.ErrInvalidOrderID, i18n.MsgInvalidOrderID)
	}
	if amount <= 0 {
		return sdkerrors.New(lang, sdkerrors.ErrInvalidAmount, i18n.MsgInvalidAmount)
	}
	return nil
}

// EncodeJSON encodes v as JSON using a buffer from the provided pool,
// copies the result into an independent []byte, and returns the buffer
// to the pool. The caller does not need to manage the pool lifecycle.
func EncodeJSON(pool gc.Pool, lang i18n.Language, v any) ([]byte, error) {
	buf := pool.Get()
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil, sdkerrors.New(lang, sdkerrors.ErrEncodeJSON, i18n.MsgFailedToEncode)
	}

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	return data, nil
}
