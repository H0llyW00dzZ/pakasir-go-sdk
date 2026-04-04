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

package simulation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/internal/request"
)

// PayRequest contains the parameters for simulating a payment.
type PayRequest struct {
	// OrderID is the transaction identifier to simulate payment for.
	OrderID string `json:"order_id"`

	// Amount is the transaction amount.
	Amount int64 `json:"amount"`
}

// Service provides methods for simulating payments in Pakasir sandbox mode.
type Service struct {
	client *client.Client
}

// NewService creates a new simulation [Service] backed by the given [client.Client].
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// Pay simulates a payment for a pending transaction.
//
// This endpoint is only available when the project is in Sandbox mode.
// It triggers the webhook callback as if a real payment was completed.
//
// It sends a POST request to /api/paymentsimulation.
func (s *Service) Pay(ctx context.Context, req *PayRequest) error {
	if req.OrderID == "" {
		return sdkerrors.New(s.client.Language, sdkerrors.ErrInvalidOrderID, i18n.MsgInvalidOrderID)
	}
	if req.Amount <= 0 {
		return sdkerrors.New(s.client.Language, sdkerrors.ErrInvalidAmount, i18n.MsgInvalidAmount)
	}

	body := request.Body{
		Project: s.client.Project,
		OrderID: req.OrderID,
		Amount:  req.Amount,
		APIKey:  s.client.APIKey,
	}

	buf := s.client.GetBufferPool().Get()
	defer func() {
		buf.Reset()
		s.client.GetBufferPool().Put(buf)
	}()

	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	_, err := s.client.Do(ctx, http.MethodPost, "/api/paymentsimulation", bytes.NewReader(buf.Bytes()))
	return err
}
