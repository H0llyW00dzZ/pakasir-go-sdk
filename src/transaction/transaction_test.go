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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/gc"
)

func newTestService(t *testing.T, handler http.HandlerFunc) (*Service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := client.New("test-project", "test-key",
		client.WithBaseURL(srv.URL),
		client.WithRetries(0),
	)
	return NewService(c), srv
}

// --- Create ---

func TestCreateSuccess(t *testing.T) {
	mockResp := CreateResponse{
		Payment: PaymentInfo{
			Project: "test-project", OrderID: "INV123", Amount: 99000,
			Fee: 1003, TotalPayment: 100003, PaymentMethod: "qris",
			PaymentNumber: "00020101021226...",
			ExpiredAt:     "2025-09-19T01:18:49.678622564Z",
		},
	}

	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/transactioncreate/qris", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResp)
	})
	defer srv.Close()

	resp, err := svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{
		OrderID: "INV123", Amount: 99000,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(100003), resp.Payment.TotalPayment)
	assert.Equal(t, "00020101021226...", resp.Payment.PaymentNumber)
}

func TestCreateInvalidMethod(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Create(context.Background(), "bitcoin", &CreateRequest{OrderID: "INV123", Amount: 99000})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidPaymentMethod)
}

func TestCreateEmptyOrderID(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{OrderID: "", Amount: 99000})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
}

func TestCreateInvalidAmount(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{OrderID: "INV123", Amount: 0})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAmount)

	_, err = svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{OrderID: "INV123", Amount: -100})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAmount)
}

func TestCreateAPIError(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad"}`))
	})
	defer srv.Close()

	_, err := svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{OrderID: "INV123", Amount: 99000})
	require.Error(t, err)
	var apiErr *sdkerrors.APIError
	assert.ErrorAs(t, err, &apiErr)
}

func TestCreateInvalidJSONResponse(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid}`))
	})
	defer srv.Close()

	_, err := svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{OrderID: "INV123", Amount: 99000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestCreateNilRequest(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Create(context.Background(), constants.MethodQRIS, nil)
	assert.ErrorIs(t, err, sdkerrors.ErrNilRequest)
}

// --- Cancel ---

func TestCancelSuccess(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/transactioncancel", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	defer srv.Close()

	err := svc.Cancel(context.Background(), &CancelRequest{OrderID: "INV123", Amount: 99000})
	require.NoError(t, err)
}

func TestCancelEmptyOrderID(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	err := svc.Cancel(context.Background(), &CancelRequest{OrderID: "", Amount: 99000})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
}

func TestCancelInvalidAmount(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	err := svc.Cancel(context.Background(), &CancelRequest{OrderID: "INV123", Amount: 0})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAmount)
}

func TestCancelNilRequest(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	err := svc.Cancel(context.Background(), nil)
	assert.ErrorIs(t, err, sdkerrors.ErrNilRequest)
}

// --- Detail ---

func TestDetailSuccess(t *testing.T) {
	mockResp := DetailResponse{
		Transaction: TransactionInfo{
			Amount: 22000, OrderID: "240910HDE7C9", Project: "test-project",
			Status: constants.StatusCompleted, PaymentMethod: "qris",
			CompletedAt: "2024-09-10T08:07:02.819+07:00",
		},
	}

	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "test-project", r.URL.Query().Get("project"))
		assert.Equal(t, "240910HDE7C9", r.URL.Query().Get("order_id"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockResp)
	})
	defer srv.Close()

	resp, err := svc.Detail(context.Background(), &DetailRequest{OrderID: "240910HDE7C9", Amount: 22000})
	require.NoError(t, err)
	assert.Equal(t, constants.StatusCompleted, resp.Transaction.Status)
}

func TestDetailEmptyOrderID(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Detail(context.Background(), &DetailRequest{OrderID: "", Amount: 22000})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidOrderID)
}

func TestDetailInvalidAmount(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Detail(context.Background(), &DetailRequest{OrderID: "INV123", Amount: -1})
	assert.ErrorIs(t, err, sdkerrors.ErrInvalidAmount)
}

func TestDetailInvalidJSONResponse(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	})
	defer srv.Close()

	_, err := svc.Detail(context.Background(), &DetailRequest{OrderID: "INV123", Amount: 22000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestDetailNilRequest(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {})
	defer srv.Close()

	_, err := svc.Detail(context.Background(), nil)
	assert.ErrorIs(t, err, sdkerrors.ErrNilRequest)
}

// --- Time Parsers ---

func TestPaymentInfoParseExpiredAtNano(t *testing.T) {
	p := &PaymentInfo{ExpiredAt: "2025-09-19T01:18:49.678622564Z"}
	ts, err := p.ParseExpiredAt()
	require.NoError(t, err)
	assert.Equal(t, 2025, ts.Year())
}

func TestPaymentInfoParseExpiredAtRFC3339(t *testing.T) {
	p := &PaymentInfo{ExpiredAt: "2025-09-19T01:18:49Z"}
	ts, err := p.ParseExpiredAt()
	require.NoError(t, err)
	assert.Equal(t, 2025, ts.Year())
}

func TestPaymentInfoParseExpiredAtInvalid(t *testing.T) {
	p := &PaymentInfo{ExpiredAt: "not-a-date"}
	_, err := p.ParseExpiredAt()
	require.Error(t, err)
}

func TestTransactionInfoParseCompletedAtRFC3339(t *testing.T) {
	info := &TransactionInfo{CompletedAt: "2024-09-10T08:07:02+07:00"}
	ts, err := info.ParseCompletedAt()
	require.NoError(t, err)
	assert.Equal(t, 2024, ts.Year())
}

func TestTransactionInfoParseCompletedAtNano(t *testing.T) {
	info := &TransactionInfo{CompletedAt: "2024-09-10T08:07:02.819000000+07:00"}
	ts, err := info.ParseCompletedAt()
	require.NoError(t, err)
	assert.Equal(t, 2024, ts.Year())
}

func TestTransactionInfoParseCompletedAtInvalid(t *testing.T) {
	info := &TransactionInfo{CompletedAt: "invalid"}
	_, err := info.ParseCompletedAt()
	require.Error(t, err)
}

// --- Encode errors ---

// errorBuffer is a [gc.Buffer] whose Write always fails.
type errorBuffer struct{}

func (errorBuffer) Write([]byte) (int, error)         { return 0, assert.AnError }
func (errorBuffer) WriteString(string) (int, error)   { return 0, assert.AnError }
func (errorBuffer) WriteByte(byte) error              { return assert.AnError }
func (errorBuffer) WriteTo(io.Writer) (int64, error)  { return 0, nil }
func (errorBuffer) ReadFrom(io.Reader) (int64, error) { return 0, nil }
func (errorBuffer) Bytes() []byte                     { return nil }
func (errorBuffer) String() string                    { return "" }
func (errorBuffer) Len() int                          { return 0 }
func (errorBuffer) Set([]byte)                        {}
func (errorBuffer) SetString(string)                  {}
func (errorBuffer) Reset()                            {}

// errorPool returns an [errorBuffer] from Get.
type errorPool struct{}

func (errorPool) Get() gc.Buffer { return errorBuffer{} }
func (errorPool) Put(gc.Buffer)  {}

func newEncodeErrorClient() *client.Client {
	return client.New("test-project", "test-key",
		client.WithBaseURL("http://localhost"),
		client.WithRetries(0),
		client.WithBufferPool(errorPool{}),
	)
}

func TestCreateEncodeError(t *testing.T) {
	svc := NewService(newEncodeErrorClient())
	_, err := svc.Create(context.Background(), constants.MethodQRIS, &CreateRequest{OrderID: "INV123", Amount: 99000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encode request")
}

func TestCancelEncodeError(t *testing.T) {
	svc := NewService(newEncodeErrorClient())
	err := svc.Cancel(context.Background(), &CancelRequest{OrderID: "INV123", Amount: 99000})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encode request")
}

func TestDetailDoError(t *testing.T) {
	// Point at a closed server to trigger a Do error.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := client.New("test-project", "test-key",
		client.WithBaseURL(srv.URL),
		client.WithRetries(0),
	)
	svc := NewService(c)
	_, err := svc.Detail(context.Background(), &DetailRequest{OrderID: "INV123", Amount: 22000})
	require.Error(t, err)
}
