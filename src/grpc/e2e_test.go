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

package grpc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/internal/grpctest"
	pakasirv1 "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/pakasir/v1"
	grpcsim "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/simulation"
	grpctxn "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/transaction"
	sdksim "github.com/H0llyW00dzZ/pakasir-go-sdk/src/simulation"
	sdktxn "github.com/H0llyW00dzZ/pakasir-go-sdk/src/transaction"
)

// orderStore is an in-memory store that tracks order state across
// the mock Pakasir API endpoints, simulating the real payment lifecycle.
type orderStore struct {
	mu     sync.Mutex
	orders map[string]string // order_id -> status
}

func newOrderStore() *orderStore {
	return &orderStore{orders: make(map[string]string)}
}

func (s *orderStore) create(id string) { s.mu.Lock(); s.orders[id] = "pending"; s.mu.Unlock() }
func (s *orderStore) pay(id string)    { s.mu.Lock(); s.orders[id] = "completed"; s.mu.Unlock() }
func (s *orderStore) status(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.orders[id]; ok {
		return st
	}
	return ""
}

// statefulPakasirServer returns a mock HTTP server that maintains order
// state across create, simulate-pay, and detail endpoints.
func statefulPakasirServer(t *testing.T, store *orderStore) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Parse order_id from the request body for POST endpoints.
		var body struct {
			OrderID string `json:"order_id"`
			Amount  int64  `json:"amount"`
		}
		if r.Method == http.MethodPost && r.Body != nil {
			json.NewDecoder(r.Body).Decode(&body)
		}

		switch {
		case strings.HasPrefix(r.URL.Path, "/api/transactioncreate/"):
			store.create(body.OrderID)
			json.NewEncoder(w).Encode(map[string]any{
				"payment": map[string]any{
					"project":        "testproject",
					"order_id":       body.OrderID,
					"amount":         body.Amount,
					"fee":            500,
					"total_payment":  body.Amount + 500,
					"payment_method": strings.TrimPrefix(r.URL.Path, "/api/transactioncreate/"),
					"payment_number": "0002010112345",
					"expired_at":     "2026-12-31T23:59:59Z",
				},
			})

		case r.URL.Path == "/api/paymentsimulation":
			store.pay(body.OrderID)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))

		case r.URL.Path == "/api/transactiondetail":
			orderID := r.URL.Query().Get("order_id")
			st := store.status(orderID)
			completedAt := ""
			if st == "completed" {
				completedAt = "2026-12-31T12:00:00Z"
			}
			json.NewEncoder(w).Encode(map[string]any{
				"transaction": map[string]any{
					"amount":         50000,
					"order_id":       orderID,
					"project":        "testproject",
					"status":         st,
					"payment_method": "qris",
					"completed_at":   completedAt,
				},
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// TestE2EPaymentFlowSuccess tests the full payment lifecycle over gRPC:
//
//  1. Create a transaction (status: pending)
//  2. Simulate payment in sandbox mode (status: completed)
//  3. Query transaction detail and verify it is completed
//
// Both TransactionService and SimulationService are registered on the
// same gRPC server, backed by a stateful mock Pakasir HTTP API.
func TestE2EPaymentFlowSuccess(t *testing.T) {
	store := newOrderStore()
	mock := statefulPakasirServer(t, store)
	defer mock.Close()

	// --- SDK + gRPC service setup ---
	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	txnSvc := grpctxn.NewService(sdktxn.NewService(c))
	simSvc := grpcsim.NewService(sdksim.NewService(c))

	lis := grpctest.NewBufListener()
	srv := grpc.NewServer()
	pakasirv1.RegisterTransactionServiceServer(srv, txnSvc)
	pakasirv1.RegisterSimulationServiceServer(srv, simSvc)

	go func() { srv.Serve(lis) }()
	defer srv.GracefulStop()

	conn, err := grpctest.DialBufNet(context.Background(), lis)
	require.NoError(t, err)
	defer conn.Close()

	txnClient := pakasirv1.NewTransactionServiceClient(conn)
	simClient := pakasirv1.NewSimulationServiceClient(conn)

	ctx := context.Background()
	totalStart := time.Now()

	// --- Step 1: Create transaction ---
	t.Log("step 1: creating transaction")
	start := time.Now()
	createResp, err := txnClient.Create(ctx, &pakasirv1.CreateRequest{
		OrderId:       "FLOW-001",
		Amount:        50000,
		PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	})
	t.Logf("  Create RPC: %v", time.Since(start))
	require.NoError(t, err)
	require.NotNil(t, createResp.GetPayment())
	assert.Equal(t, "FLOW-001", createResp.GetPayment().GetOrderId())
	assert.Equal(t, int64(50500), createResp.GetPayment().GetTotalPayment())
	assert.Equal(t, pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS, createResp.GetPayment().GetPaymentMethod())
	t.Logf("  payment_number=%s expired_at=%s",
		createResp.GetPayment().GetPaymentNumber(),
		createResp.GetPayment().GetExpiredAt().AsTime())

	// Verify order is pending in the store.
	assert.Equal(t, "pending", store.status("FLOW-001"))

	// --- Step 2: Query detail (should be pending) ---
	t.Log("step 2: verifying transaction is pending")
	start = time.Now()
	detailResp, err := txnClient.Detail(ctx, &pakasirv1.DetailRequest{
		OrderId: "FLOW-001",
		Amount:  50000,
	})
	t.Logf("  Detail RPC (pending): %v", time.Since(start))
	require.NoError(t, err)
	assert.Equal(t, pakasirv1.TransactionStatus_TRANSACTION_STATUS_PENDING, detailResp.GetTransaction().GetStatus())

	// --- Step 3: Simulate payment ---
	t.Log("step 3: simulating payment")
	start = time.Now()
	_, err = simClient.Pay(ctx, &pakasirv1.PayRequest{
		OrderId: "FLOW-001",
		Amount:  50000,
	})
	t.Logf("  Pay RPC: %v", time.Since(start))
	require.NoError(t, err)

	// Verify order is completed in the store.
	assert.Equal(t, "completed", store.status("FLOW-001"))

	// --- Step 4: Query detail (should be completed) ---
	t.Log("step 4: verifying transaction is completed")
	start = time.Now()
	detailResp, err = txnClient.Detail(ctx, &pakasirv1.DetailRequest{
		OrderId: "FLOW-001",
		Amount:  50000,
	})
	t.Logf("  Detail RPC (completed): %v", time.Since(start))
	require.NoError(t, err)
	assert.Equal(t, pakasirv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED, detailResp.GetTransaction().GetStatus())
	assert.NotNil(t, detailResp.GetTransaction().GetCompletedAt())

	t.Logf("full payment flow: %v", time.Since(totalStart))
}
