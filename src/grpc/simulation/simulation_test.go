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
		if r.URL.Path == "/api/paymentsimulation" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// mockErrorServer returns an httptest.Server that always returns 500.
func mockErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
}

// startGRPCServer creates a gRPC server with the simulation service and
// optional interceptors, starts it on a bufconn listener, and returns
// the client connection and a cleanup function.
func startGRPCServer(t *testing.T, svc *Service, unary []grpc.UnaryServerInterceptor) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := grpctest.NewBufListener()

	opts := []grpc.ServerOption{}
	if len(unary) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unary...))
	}

	srv := grpc.NewServer(opts...)
	pakasirv1.RegisterSimulationServiceServer(srv, svc)

	go func() {
		if err := srv.Serve(lis); err != nil {
			// Server stopped; expected during cleanup.
		}
	}()

	conn, err := grpctest.DialBufNet(context.Background(), lis)
	require.NoError(t, err)

	return conn, func() {
		conn.Close()
		srv.GracefulStop()
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

	conn, cleanup := startGRPCServer(t, grpcSvc, nil)
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
	mock := mockErrorServer(t)
	defer mock.Close()

	c := client.New("testproject", "test-api-key",
		client.WithBaseURL(mock.URL),
		client.WithRetries(0),
	)
	sdkSvc := sdksim.NewService(c)
	grpcSvc := NewService(sdkSvc)

	conn, cleanup := startGRPCServer(t, grpcSvc, nil)
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
}

// --- Interceptor pluggability tests ---

// loggingInterceptor increments a counter for every RPC handled and
// logs the method and duration.
func loggingInterceptor(t *testing.T, counter *atomic.Int64) grpc.UnaryServerInterceptor {
	t.Helper()
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		counter.Add(1)
		resp, err := handler(ctx, req)
		t.Logf("[interceptor/logging] method=%s duration=%v err=%v",
			info.FullMethod, time.Since(start), err)
		return resp, err
	}
}

// authInterceptor rejects requests without a valid "authorization" metadata key.
func authInterceptor(validToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		tokens := md.Get("authorization")
		if len(tokens) == 0 || tokens[0] != "Bearer "+validToken {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}
}

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
	conn, cleanup := startGRPCServer(t, grpcSvc, []grpc.UnaryServerInterceptor{
		loggingInterceptor(t, &callCount),
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

	conn, cleanup := startGRPCServer(t, grpcSvc, []grpc.UnaryServerInterceptor{
		authInterceptor("sim-token"),
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

	conn, cleanup := startGRPCServer(t, grpcSvc, []grpc.UnaryServerInterceptor{
		authInterceptor("sim-token"),
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
	conn, cleanup := startGRPCServer(t, grpcSvc, []grpc.UnaryServerInterceptor{
		loggingInterceptor(t, &callCount),
		authInterceptor("chain-sim"),
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
