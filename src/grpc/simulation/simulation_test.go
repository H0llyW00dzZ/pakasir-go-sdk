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
	sdksim "github.com/H0llyW00dzZ/pakasir-go-sdk/src/simulation"
)

// --- Unit tests ---

func TestNewService(t *testing.T) {
	svc := NewService(nil)
	require.NotNil(t, svc)
}

// --- Test helpers ---

// mockPakasirServer returns an httptest.Server that simulates the
// Pakasir sandbox payment simulation endpoint.
func mockPakasirServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == constants.PathPaymentSimulation && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// registerSimulation returns a registration callback for use with
// [grpctest.StartServer].
func registerSimulation(svc *Service) func(grpc.ServiceRegistrar) {
	return func(r grpc.ServiceRegistrar) {
		pakasirv1.RegisterSimulationServiceServer(r, svc)
	}
}

// --- E2E tests (happy path) ---

func TestE2ESimulatePay(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), nil)
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	start := time.Now()
	resp, err := simClient.Pay(context.Background(), &pakasirv1.PayRequest{
		OrderId: "SIM-001",
		Amount:  42000,
	})
	elapsed := time.Since(start)
	t.Logf("Pay RPC: %v", elapsed)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

// --- E2E tests (error path) ---

func TestE2ESimulatePayError(t *testing.T) {
	mock := grpctest.MockErrorServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), nil)
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	start := time.Now()
	_, err := simClient.Pay(context.Background(), &pakasirv1.PayRequest{
		OrderId: "SIM-ERR-001",
		Amount:  10000,
	})
	elapsed := time.Since(start)
	t.Logf("Pay RPC (error path): %v, err=%v", elapsed, err)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

// --- E2E tests (validation → InvalidArgument) ---

func TestE2EPayValidationErrors(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdksim.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), nil)
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)
	ctx := context.Background()

	tests := []struct {
		name string
		req  *pakasirv1.PayRequest
	}{
		{"empty order id", &pakasirv1.PayRequest{OrderId: "", Amount: 10000}},
		{"zero amount", &pakasirv1.PayRequest{OrderId: "VAL-SIM-001", Amount: 0}},
		{"negative amount", &pakasirv1.PayRequest{OrderId: "VAL-SIM-002", Amount: -100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := simClient.Pay(ctx, tt.req)
			require.Error(t, err)

			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, codes.InvalidArgument, st.Code())
			t.Logf("  code=%s message=%s", st.Code(), st.Message())
		})
	}
}

// --- E2E tests (APIError → gRPC status code mapping) ---

func TestE2EPayAPIErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		httpStatus int
		httpBody   string
		grpcCode   codes.Code
	}{
		{"400 bad request", http.StatusBadRequest, `{"error":"bad request"}`, codes.InvalidArgument},
		{"401 unauthorized", http.StatusUnauthorized, `{"error":"unauthorized"}`, codes.Unauthenticated},
		{"403 not sandbox", http.StatusForbidden, `{"error":"not sandbox"}`, codes.PermissionDenied},
		{"404 not found", http.StatusNotFound, `{"error":"not found"}`, codes.NotFound},
		{"409 conflict", http.StatusConflict, `{"error":"duplicate order"}`, codes.AlreadyExists},
		{"429 rate limited", http.StatusTooManyRequests, `{"error":"too many requests"}`, codes.Unavailable},
		{"500 internal", http.StatusInternalServerError, `{"error":"internal server error"}`, codes.Internal},
		{"502 bad gateway", http.StatusBadGateway, `{"error":"bad gateway"}`, codes.Unavailable},
		{"503 unavailable", http.StatusServiceUnavailable, `{"error":"unavailable"}`, codes.Unavailable},
		{"504 gateway timeout", http.StatusGatewayTimeout, `{"error":"gateway timeout"}`, codes.Unavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := grpctest.MockHTTPStatusServer(t, tt.httpStatus, tt.httpBody)
			defer mock.Close()

			c := client.New("testproject", "test-api-key",
				client.WithBaseURL(mock.URL),
				client.WithRetries(0),
			)
			grpcSvc := NewService(sdksim.NewService(c))

			conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), nil)
			defer cleanup()

			simClient := pakasirv1.NewSimulationServiceClient(conn)

			_, err := simClient.Pay(context.Background(), &pakasirv1.PayRequest{
				OrderId: "API-ERR-SIM-001",
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

func TestE2ESimulatePayContextCanceled(t *testing.T) {
	mock := grpctest.SlowServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdksim.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), nil)
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := simClient.Pay(ctx, &pakasirv1.PayRequest{
		OrderId: "SIM-CTX-001",
		Amount:  10000,
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Canceled, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

func TestE2ESimulatePayContextDeadlineExceeded(t *testing.T) {
	mock := grpctest.SlowServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	grpcSvc := NewService(sdksim.NewService(c))

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), nil)
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := simClient.Pay(ctx, &pakasirv1.PayRequest{
		OrderId: "SIM-CTX-002",
		Amount:  10000,
	})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.DeadlineExceeded, st.Code())
	t.Logf("  code=%s message=%s", st.Code(), st.Message())
}

// --- Interceptor pluggability tests ---

func TestE2ESimulatePayWithLogging(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	var callCount atomic.Int64
	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.LoggingInterceptor(t, &callCount),
	})
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	start := time.Now()

	_, err := simClient.Pay(context.Background(), &pakasirv1.PayRequest{
		OrderId: "SIM-LOG-001", Amount: 10000,
	})
	require.NoError(t, err)

	_, err = simClient.Pay(context.Background(), &pakasirv1.PayRequest{
		OrderId: "SIM-LOG-002", Amount: 20000,
	})
	require.NoError(t, err)

	t.Logf("2 RPCs with logging interceptor: total=%v, calls=%d", time.Since(start), callCount.Load())
	assert.Equal(t, int64(2), callCount.Load())
}

func TestE2ESimulatePayWithAuthSuccess(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.AuthInterceptor("sim-token"),
	})
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"authorization", "Bearer sim-token",
	)

	start := time.Now()
	resp, err := simClient.Pay(ctx, &pakasirv1.PayRequest{
		OrderId: "SIM-AUTH-001", Amount: 15000,
	})
	t.Logf("Pay RPC with auth (success): %v", time.Since(start))

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestE2ESimulatePayWithAuthReject(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.AuthInterceptor("sim-token"),
	})
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	start := time.Now()
	_, err := simClient.Pay(context.Background(), &pakasirv1.PayRequest{
		OrderId: "SIM-AUTH-002", Amount: 15000,
	})
	t.Logf("Pay RPC with auth (reject): %v, err=%v", time.Since(start), err)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestE2ESimulatePayWithChainedInterceptors(t *testing.T) {
	mock := mockPakasirServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	var callCount atomic.Int64
	conn, cleanup := grpctest.StartServer(t, registerSimulation(grpcSvc), []grpc.UnaryServerInterceptor{
		grpctest.LoggingInterceptor(t, &callCount),
		grpctest.AuthInterceptor("chain-sim"),
	})
	defer cleanup()

	simClient := pakasirv1.NewSimulationServiceClient(conn)

	// Valid auth.
	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"authorization", "Bearer chain-sim",
	)
	start := time.Now()
	_, err := simClient.Pay(ctx, &pakasirv1.PayRequest{
		OrderId: "SIM-CHAIN-001", Amount: 10000,
	})
	t.Logf("Pay RPC chained (auth pass): %v", time.Since(start))

	require.NoError(t, err)
	assert.Equal(t, int64(1), callCount.Load())

	// Invalid auth.
	start = time.Now()
	_, err = simClient.Pay(context.Background(), &pakasirv1.PayRequest{
		OrderId: "SIM-CHAIN-002", Amount: 10000,
	})
	t.Logf("Pay RPC chained (auth reject): %v, err=%v", time.Since(start), err)

	require.Error(t, err)
	assert.Equal(t, int64(2), callCount.Load())
}
