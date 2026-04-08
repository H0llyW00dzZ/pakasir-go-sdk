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
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/internal/grpctest"
	pakasirv1 "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/pakasir/v1"
	sdktxn "github.com/H0llyW00dzZ/pakasir-go-sdk/src/transaction"
)

// --- Unit tests ---

func TestNewService(t *testing.T) {
	svc := NewService(nil)
	require.NotNil(t, svc)
}

func TestPaymentInfoToProto(t *testing.T) {
	p := &sdktxn.PaymentInfo{
		Project:       "myproject",
		OrderID:       "INV-001",
		Amount:        50000,
		Fee:           1000,
		TotalPayment:  51000,
		PaymentMethod: constants.MethodQRIS,
		PaymentNumber: "00020101...",
		ExpiredAt:     "2026-04-07T12:00:00Z",
	}

	pb := paymentInfoToProto(p)
	require.NotNil(t, pb)
	assert.Equal(t, "myproject", pb.GetProject())
	assert.Equal(t, "INV-001", pb.GetOrderId())
	assert.Equal(t, int64(50000), pb.GetAmount())
	assert.Equal(t, int64(1000), pb.GetFee())
	assert.Equal(t, int64(51000), pb.GetTotalPayment())
	assert.Equal(t, pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS, pb.GetPaymentMethod())
	assert.Equal(t, "00020101...", pb.GetPaymentNumber())
	assert.NotNil(t, pb.GetExpiredAt())
}

func TestPaymentInfoToProtoNil(t *testing.T) {
	assert.Nil(t, paymentInfoToProto(nil))
}

func TestTransactionInfoToProto(t *testing.T) {
	ti := &sdktxn.TransactionInfo{
		Amount:        99000,
		OrderID:       "INV-002",
		Project:       "shop",
		Status:        constants.StatusCompleted,
		PaymentMethod: constants.MethodPermataVA,
		CompletedAt:   "2026-04-07T15:30:00Z",
	}

	pb := transactionInfoToProto(ti)
	require.NotNil(t, pb)
	assert.Equal(t, int64(99000), pb.GetAmount())
	assert.Equal(t, "INV-002", pb.GetOrderId())
	assert.Equal(t, "shop", pb.GetProject())
	assert.Equal(t, pakasirv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED, pb.GetStatus())
	assert.Equal(t, pakasirv1.PaymentMethod_PAYMENT_METHOD_PERMATA_VA, pb.GetPaymentMethod())
	assert.NotNil(t, pb.GetCompletedAt())
}

func TestTransactionInfoToProtoNil(t *testing.T) {
	assert.Nil(t, transactionInfoToProto(nil))
}

// --- Test helpers ---

// mockPakasirServer returns an httptest.Server that simulates the Pakasir API.
func mockPakasirServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, constants.PathTransactionCreate+"/"):
			json.NewEncoder(w).Encode(map[string]any{
				"payment": map[string]any{
					"project":        "testproject",
					"order_id":       "E2E-001",
					"amount":         50000,
					"fee":            1000,
					"total_payment":  51000,
					"payment_method": "qris",
					"payment_number": "00020101...",
					"expired_at":     "2026-12-25T23:59:59Z",
				},
			})
		case r.URL.Path == constants.PathTransactionCancel:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		case r.URL.Path == constants.PathTransactionDetail:
			json.NewEncoder(w).Encode(map[string]any{
				"transaction": map[string]any{
					"amount":         50000,
					"order_id":       "E2E-001",
					"project":        "testproject",
					"status":         "completed",
					"payment_method": "qris",
					"completed_at":   "2026-12-25T12:00:00Z",
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// registerTransaction returns a registration callback for use with
// [grpctest.StartServer].
func registerTransaction(svc *Service) func(grpc.ServiceRegistrar) {
	return func(r grpc.ServiceRegistrar) {
		pakasirv1.RegisterTransactionServiceServer(r, svc)
	}
}

// --- E2E tests (happy path) ---

func TestE2ECreateTransaction(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()
	resp, err := txnClient.Create(context.Background(), &pakasirv1.CreateRequest{
		OrderId:       "E2E-001",
		Amount:        50000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	})
	elapsed := time.Since(start)
	t.Logf("Create RPC: %v", elapsed)

	require.NoError(t, err)
	require.NotNil(t, resp.GetPayment())
	assert.Equal(t, "testproject", resp.GetPayment().GetProject())
	assert.Equal(t, "E2E-001", resp.GetPayment().GetOrderId())
	assert.Equal(t, int64(51000), resp.GetPayment().GetTotalPayment())
	assert.Equal(t, pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS, resp.GetPayment().GetPaymentMethod())
	assert.NotNil(t, resp.GetPayment().GetExpiredAt())
}

func TestE2ECancelTransaction(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()
	resp, err := txnClient.Cancel(context.Background(), &pakasirv1.CancelRequest{
		OrderId: "E2E-001",
		Amount:  50000,
	})
	elapsed := time.Since(start)
	t.Logf("Cancel RPC: %v", elapsed)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestE2EGetTransactionDetail(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()
	resp, err := txnClient.Detail(context.Background(), &pakasirv1.DetailRequest{
		OrderId: "E2E-001",
		Amount:  50000,
	})
	elapsed := time.Since(start)
	t.Logf("Detail RPC: %v", elapsed)

	require.NoError(t, err)
	require.NotNil(t, resp.GetTransaction())
	assert.Equal(t, "E2E-001", resp.GetTransaction().GetOrderId())
	assert.Equal(t, pakasirv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED, resp.GetTransaction().GetStatus())
	assert.Equal(t, pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS, resp.GetTransaction().GetPaymentMethod())
	assert.NotNil(t, resp.GetTransaction().GetCompletedAt())
}

// --- E2E tests (error paths) ---

func TestE2ECreateTransactionError(t *testing.T) {
	mock := grpctest.MockErrorServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()
	_, err := txnClient.Create(context.Background(), &pakasirv1.CreateRequest{
		OrderId:       "ERR-001",
		Amount:        10000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	})
	elapsed := time.Since(start)
	t.Logf("Create RPC (error path): %v, err=%v", elapsed, err)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

func TestE2ECancelTransactionError(t *testing.T) {
	mock := grpctest.MockErrorServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()
	_, err := txnClient.Cancel(context.Background(), &pakasirv1.CancelRequest{
		OrderId: "ERR-002",
		Amount:  10000,
	})
	elapsed := time.Since(start)
	t.Logf("Cancel RPC (error path): %v, err=%v", elapsed, err)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

func TestE2EGetTransactionDetailError(t *testing.T) {
	mock := grpctest.MockErrorServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()
	_, err := txnClient.Detail(context.Background(), &pakasirv1.DetailRequest{
		OrderId: "ERR-003",
		Amount:  10000,
	})
	elapsed := time.Since(start)
	t.Logf("Detail RPC (error path): %v, err=%v", elapsed, err)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

// --- E2E tests (validation → InvalidArgument) ---

func TestE2ECreateValidationErrors(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *pakasirv1.CreateRequest
	}{
		{"empty order id", &pakasirv1.CreateRequest{
			OrderId: "", Amount: 10000,
			PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
		}},
		{"zero amount", &pakasirv1.CreateRequest{
			OrderId: "VAL-001", Amount: 0,
			PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
		}},
		{"negative amount", &pakasirv1.CreateRequest{
			OrderId: "VAL-002", Amount: -100,
			PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
		}},
		{"unspecified payment method", &pakasirv1.CreateRequest{
			OrderId: "VAL-003", Amount: 10000,
			PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := txnClient.Create(ctx, tt.req)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			t.Logf("  code=%s message=%s", st.Code(), st.Message())
		})
	}
}

func TestE2ECancelValidationErrors(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *pakasirv1.CancelRequest
	}{
		{"empty order id", &pakasirv1.CancelRequest{OrderId: "", Amount: 10000}},
		{"zero amount", &pakasirv1.CancelRequest{OrderId: "VAL-010", Amount: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := txnClient.Cancel(ctx, tt.req)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			t.Logf("  code=%s message=%s", st.Code(), st.Message())
		})
	}
}

func TestE2EDetailValidationErrors(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *pakasirv1.DetailRequest
	}{
		{"empty order id", &pakasirv1.DetailRequest{OrderId: "", Amount: 10000}},
		{"zero amount", &pakasirv1.DetailRequest{OrderId: "VAL-020", Amount: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := txnClient.Detail(ctx, tt.req)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			t.Logf("  code=%s message=%s", st.Code(), st.Message())
		})
	}
}

// --- E2E tests (APIError → gRPC status code mapping) ---

func TestE2ECreateAPIErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		httpBody   string
		grpcCode   codes.Code
	}{
		{"400 bad request", 400, `{"error":"bad request"}`, codes.InvalidArgument},
		{"401 unauthorized", 401, `{"error":"unauthorized"}`, codes.Unauthenticated},
		{"403 forbidden", 403, `{"error":"forbidden"}`, codes.PermissionDenied},
		{"404 not found", 404, `{"error":"not found"}`, codes.NotFound},
		{"409 conflict", 409, `{"error":"duplicate order"}`, codes.AlreadyExists},
		{"429 rate limited", 429, `{"error":"too many requests"}`, codes.Unavailable},
		{"502 bad gateway", 502, `{"error":"bad gateway"}`, codes.Unavailable},
		{"503 unavailable", 503, `{"error":"unavailable"}`, codes.Unavailable},
		{"504 gateway timeout", 504, `{"error":"gateway timeout"}`, codes.Unavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := grpctest.MockHTTPStatusServer(t, tt.httpStatus, tt.httpBody)
			defer mock.Close()

			c := client.New("testproject", "test-api-key",
				client.WithBaseURL(mock.URL),
				client.WithRetries(0),
			)
			grpcSvc := NewService(sdktxn.NewService(c))

			conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
			defer cleanup()

			txnClient := pakasirv1.NewTransactionServiceClient(conn)

			_, err := txnClient.Create(context.Background(), &pakasirv1.CreateRequest{
				OrderId:       "API-ERR-001",
				Amount:        10000,
				PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
			})
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.grpcCode, st.Code())
			t.Logf("  HTTP %d → gRPC %s: %s", tt.httpStatus, st.Code(), st.Message())
		})
	}
}

func TestE2ECancelAPIErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		httpBody   string
		grpcCode   codes.Code
	}{
		{"401 unauthorized", 401, `{"error":"unauthorized"}`, codes.Unauthenticated},
		{"404 not found", 404, `{"error":"not found"}`, codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := grpctest.MockHTTPStatusServer(t, tt.httpStatus, tt.httpBody)
			defer mock.Close()

			c := client.New("testproject", "test-api-key",
				client.WithBaseURL(mock.URL),
				client.WithRetries(0),
			)
			grpcSvc := NewService(sdktxn.NewService(c))

			conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
			defer cleanup()

			txnClient := pakasirv1.NewTransactionServiceClient(conn)

			_, err := txnClient.Cancel(context.Background(), &pakasirv1.CancelRequest{
				OrderId: "API-ERR-002",
				Amount:  10000,
			})
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.grpcCode, st.Code())
			t.Logf("  HTTP %d → gRPC %s: %s", tt.httpStatus, st.Code(), st.Message())
		})
	}
}

func TestE2EDetailAPIErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		httpBody   string
		grpcCode   codes.Code
	}{
		{"401 unauthorized", 401, `{"error":"unauthorized"}`, codes.Unauthenticated},
		{"404 not found", 404, `{"error":"not found"}`, codes.NotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := grpctest.MockHTTPStatusServer(t, tt.httpStatus, tt.httpBody)
			defer mock.Close()

			c := client.New("testproject", "test-api-key",
				client.WithBaseURL(mock.URL),
				client.WithRetries(0),
			)
			grpcSvc := NewService(sdktxn.NewService(c))

			conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
			defer cleanup()

			txnClient := pakasirv1.NewTransactionServiceClient(conn)

			_, err := txnClient.Detail(context.Background(), &pakasirv1.DetailRequest{
				OrderId: "API-ERR-003",
				Amount:  10000,
			})
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.grpcCode, st.Code())
			t.Logf("  HTTP %d → gRPC %s: %s", tt.httpStatus, st.Code(), st.Message())
		})
	}
}

// --- E2E tests (context cancellation) ---

func TestE2ECreateContextCanceled(t *testing.T) {
	mock := grpctest.SlowServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := txnClient.Create(ctx, &pakasirv1.CreateRequest{
		OrderId:       "CTX-001",
		Amount:        10000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Canceled, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

func TestE2ECreateContextDeadlineExceeded(t *testing.T) {
	mock := grpctest.SlowServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := txnClient.Create(ctx, &pakasirv1.CreateRequest{
		OrderId:       "CTX-002",
		Amount:        10000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.DeadlineExceeded, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

func TestE2ECancelContextCanceled(t *testing.T) {
	mock := grpctest.SlowServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := txnClient.Cancel(ctx, &pakasirv1.CancelRequest{
		OrderId: "CTX-003",
		Amount:  10000,
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Canceled, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

func TestE2EDetailContextDeadlineExceeded(t *testing.T) {
	mock := grpctest.SlowServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdktxn.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), nil)
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := txnClient.Detail(ctx, &pakasirv1.DetailRequest{
		OrderId: "CTX-004",
		Amount:  10000,
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.DeadlineExceeded, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

// --- Interceptor pluggability tests ---

func TestE2EWithLoggingInterceptor(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	var callCount atomic.Int64
	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.LoggingInterceptor(t, &callCount),
	})
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	start := time.Now()

	_, err := txnClient.Create(context.Background(), &pakasirv1.CreateRequest{
		OrderId: "LOG-001", Amount: 10000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	})
	require.NoError(t, err)

	_, err = txnClient.Cancel(context.Background(), &pakasirv1.CancelRequest{
		OrderId: "LOG-001", Amount: 10000,
	})
	require.NoError(t, err)

	_, err = txnClient.Detail(context.Background(), &pakasirv1.DetailRequest{
		OrderId: "LOG-001", Amount: 10000,
	})
	require.NoError(t, err)

	t.Logf("3 RPCs with logging interceptor: total=%v, calls=%d", time.Since(start), callCount.Load())
	assert.Equal(t, int64(3), callCount.Load())
}

func TestE2EWithAuthInterceptorSuccess(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.AuthInterceptor("secret-token"),
	})
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"authorization", "Bearer secret-token",
	)

	start := time.Now()
	resp, err := txnClient.Create(ctx, &pakasirv1.CreateRequest{
		OrderId: "AUTH-001", Amount: 25000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_BNI_VA,
	})
	t.Logf("Create RPC with auth (success): %v", time.Since(start))

	require.NoError(t, err)
	assert.NotNil(t, resp.GetPayment())
}

func TestE2EWithAuthInterceptorReject(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.AuthInterceptor("secret-token"),
	})
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	// No auth token.
	start := time.Now()
	_, err := txnClient.Create(context.Background(), &pakasirv1.CreateRequest{
		OrderId: "AUTH-002", Amount: 25000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_BNI_VA,
	})
	t.Logf("Create RPC with auth (no token): %v, err=%v", time.Since(start), err)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())

	// Wrong token.
	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"authorization", "Bearer wrong-token",
	)
	start = time.Now()
	_, err = txnClient.Create(ctx, &pakasirv1.CreateRequest{
		OrderId: "AUTH-003", Amount: 25000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_BNI_VA,
	})
	t.Logf("Create RPC with auth (wrong token): %v, err=%v", time.Since(start), err)

	require.Error(t, err)
	st, ok = status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestE2EWithChainedInterceptors(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdktxn.NewService(c)
	grpcSvc := NewService(sdkSvc)

	var callCount atomic.Int64
	conn, cleanup := grpctest.StartServer(t, registerTransaction(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.LoggingInterceptor(t, &callCount),
		grpctest.AuthInterceptor("chain-token"),
	})
	defer cleanup()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)

	// Valid auth.
	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"authorization", "Bearer chain-token",
	)
	start := time.Now()
	resp, err := txnClient.Detail(ctx, &pakasirv1.DetailRequest{
		OrderId: "CHAIN-001", Amount: 10000,
	})
	t.Logf("Detail RPC chained (auth pass): %v", time.Since(start))

	require.NoError(t, err)
	assert.NotNil(t, resp.GetTransaction())
	assert.Equal(t, int64(1), callCount.Load())

	// Invalid auth.
	start = time.Now()
	_, err = txnClient.Detail(context.Background(), &pakasirv1.DetailRequest{
		OrderId: "CHAIN-002", Amount: 10000,
	})
	t.Logf("Detail RPC chained (auth reject): %v, err=%v", time.Since(start), err)

	require.Error(t, err)
	assert.Equal(t, int64(2), callCount.Load())
}
