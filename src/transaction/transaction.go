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

package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/internal/request"
)

// Service provides methods for creating, cancelling, and querying
// Pakasir transactions.
type Service struct {
	client *client.Client
}

// NewService creates a new transaction [Service] backed by the given [client.Client].
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// Create initiates a new payment transaction with the specified payment method.
//
// It sends a POST request to /api/transactioncreate/{method} and returns
// the payment details including the QR string or Virtual Account number.
func (s *Service) Create(ctx context.Context, method constants.PaymentMethod, req *CreateRequest) (*CreateResponse, error) {
	if req == nil {
		return nil, sdkerrors.New(s.client.Lang(), sdkerrors.ErrNilRequest, i18n.MsgNilRequest)
	}
	if !method.Valid() {
		return nil, sdkerrors.New(s.client.Lang(), sdkerrors.ErrInvalidPaymentMethod, i18n.MsgInvalidPaymentMethod, method.String())
	}
	if err := s.validateRequest(req.OrderID, req.Amount); err != nil {
		return nil, err
	}

	body := request.Body{
		Project: s.client.Project(),
		OrderID: req.OrderID,
		Amount:  req.Amount,
		APIKey:  s.client.APIKey(),
	}

	data, err := request.EncodeJSON(s.client.GetBufferPool(), s.client.Lang(), body)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/transactioncreate/%s", method)
	data, err = s.client.Do(ctx, http.MethodPost, path, data)
	if err != nil {
		return nil, err
	}

	var resp CreateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, sdkerrors.New(s.client.Lang(), sdkerrors.ErrDecodeJSON, i18n.MsgFailedToDecode, err)
	}

	return &resp, nil
}

// Cancel cancels an existing transaction.
//
// It sends a POST request to /api/transactioncancel.
func (s *Service) Cancel(ctx context.Context, req *CancelRequest) error {
	if req == nil {
		return sdkerrors.New(s.client.Lang(), sdkerrors.ErrNilRequest, i18n.MsgNilRequest)
	}
	if err := s.validateRequest(req.OrderID, req.Amount); err != nil {
		return err
	}

	body := request.Body{
		Project: s.client.Project(),
		OrderID: req.OrderID,
		Amount:  req.Amount,
		APIKey:  s.client.APIKey(),
	}

	data, err := request.EncodeJSON(s.client.GetBufferPool(), s.client.Lang(), body)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, http.MethodPost, "/api/transactioncancel", data)
	return err
}

// Detail retrieves the details and status of a transaction.
//
// It sends a GET request to /api/transactiondetail with query parameters.
// Per the [Pakasir API specification], all parameters including the API key
// are passed as query string values for this endpoint. This means the API
// key will appear in server access logs, reverse proxy logs, and any
// network-level monitoring. All other SDK endpoints transmit the key in
// the POST request body, which is not logged by default.
//
// Callers should ensure their infrastructure redacts or excludes query
// strings from access logs, and should never invoke this method from
// client-side or browser code.
//
// [Pakasir API specification]: https://pakasir.com/p/docs
func (s *Service) Detail(ctx context.Context, req *DetailRequest) (*DetailResponse, error) {
	if req == nil {
		return nil, sdkerrors.New(s.client.Lang(), sdkerrors.ErrNilRequest, i18n.MsgNilRequest)
	}
	if err := s.validateRequest(req.OrderID, req.Amount); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("project", s.client.Project())
	params.Set("amount", strconv.FormatInt(req.Amount, 10))
	params.Set("order_id", req.OrderID)
	params.Set("api_key", s.client.APIKey())

	path := "/api/transactiondetail?" + params.Encode()

	data, err := s.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp DetailResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, sdkerrors.New(s.client.Lang(), sdkerrors.ErrDecodeJSON, i18n.MsgFailedToDecode, err)
	}

	return &resp, nil
}

// validateRequest performs common validation for transaction requests.
func (s *Service) validateRequest(orderID string, amount int64) error {
	return request.ValidateOrderAndAmount(s.client.Lang(), orderID, amount)
}
