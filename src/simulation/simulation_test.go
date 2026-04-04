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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
)

func newTestService(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := client.New("test-project", "test-key",
		client.WithBaseURL(srv.URL), client.WithRetries(0),
	)
	return NewService(c), srv
}

func TestPaySuccess(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/paymentsimulation", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	defer srv.Close()

	err := svc.Pay(context.Background(), &PayRequest{OrderID: "INV123", Amount: 99000})
	require.NoError(t, err)
}

func TestPayNilRequest(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	err := svc.Pay(context.Background(), nil)
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
}

func TestPayEmptyOrderID(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	err := svc.Pay(context.Background(), &PayRequest{OrderID: "", Amount: 99000})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
}

func TestPayInvalidAmount(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	assert.ErrorIs(t, svc.Pay(context.Background(), &PayRequest{OrderID: "INV123", Amount: 0}), sdkerrors.ErrInvalidAmount)
	assert.ErrorIs(t, svc.Pay(context.Background(), &PayRequest{OrderID: "INV123", Amount: -500}), sdkerrors.ErrInvalidAmount)
}

func TestPayAPIError(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"not sandbox"}`))
	})
	defer srv.Close()

	err := svc.Pay(context.Background(), &PayRequest{OrderID: "INV123", Amount: 99000})
	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	assert.ErrorAs(t, err, &apiErr)
}
