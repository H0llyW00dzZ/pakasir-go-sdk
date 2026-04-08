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

package grpctest

import (
	"context"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewBufListener(t *testing.T) {
	lis := NewBufListener()
	require.NotNil(t, lis)
	t.Logf("bufconn listener created: addr=%s", lis.Addr())
}

func TestDialBufNet(t *testing.T) {
	lis := NewBufListener()

	// Start a dummy gRPC server so the dialer closure is actually invoked.
	srv := grpc.NewServer()
	go func() { srv.Serve(lis) }()
	defer srv.GracefulStop()

	conn, err := DialBufNet(context.Background(), lis)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Force the connection to dial (triggers the context dialer).
	conn.Connect()
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	conn.WaitForStateChange(ctx, connectivity.Idle)

	t.Logf("bufconn client connected: target=%s state=%s", conn.Target(), conn.GetState())
	assert.NoError(t, conn.Close())
}

func TestDialBufNetWithExtraOpts(t *testing.T) {
	lis := NewBufListener()

	srv := grpc.NewServer()
	go func() { srv.Serve(lis) }()
	defer srv.GracefulStop()

	conn, err := DialBufNet(context.Background(), lis,
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4*1024*1024)),
	)
	require.NoError(t, err)
	require.NotNil(t, conn)

	conn.Connect()
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	conn.WaitForStateChange(ctx, connectivity.Idle)

	t.Logf("bufconn client connected with extra opts: target=%s state=%s", conn.Target(), conn.GetState())
	assert.NoError(t, conn.Close())
}

// --- MockErrorServer ---

func TestMockErrorServer(t *testing.T) {
	srv := MockErrorServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/any-path")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"error":"internal server error"}`, string(body))
}

// --- MockHTTPStatusServer ---

func TestMockHTTPStatusServer(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"400 bad request", http.StatusBadRequest, `{"error":"bad"}`},
		{"401 unauthorized", http.StatusUnauthorized, `{"error":"unauthorized"}`},
		{"200 ok", http.StatusOK, `{"status":"ok"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := MockHTTPStatusServer(t, tt.statusCode, tt.body)
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/test")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.statusCode, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.JSONEq(t, tt.body, string(body))
		})
	}
}

// --- LoggingInterceptor ---

func TestLoggingInterceptor(t *testing.T) {
	var counter atomic.Int64
	interceptor := LoggingInterceptor(t, &counter)
	require.NotNil(t, interceptor)

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.Equal(t, int64(1), counter.Load())

	// Second call increments counter.
	_, err = interceptor(context.Background(), "request", info, handler)
	require.NoError(t, err)
	assert.Equal(t, int64(2), counter.Load())
}

func TestLoggingInterceptorPropagatesError(t *testing.T) {
	var counter atomic.Int64
	interceptor := LoggingInterceptor(t, &counter)

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "not found")
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Equal(t, int64(1), counter.Load())

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

// --- AuthInterceptor ---

func TestAuthInterceptorSuccess(t *testing.T) {
	interceptor := AuthInterceptor("secret-token")
	require.NotNil(t, interceptor)

	handler := func(ctx context.Context, req any) (any, error) {
		return "authorized", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer secret-token"),
	)

	resp, err := interceptor(ctx, "request", info, handler)
	require.NoError(t, err)
	assert.Equal(t, "authorized", resp)
}

func TestAuthInterceptorMissingMetadata(t *testing.T) {
	interceptor := AuthInterceptor("secret-token")
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	handler := func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	// No metadata at all.
	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Nil(t, resp)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "missing metadata")
}

func TestAuthInterceptorNoAuthHeader(t *testing.T) {
	interceptor := AuthInterceptor("secret-token")
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	handler := func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	// Metadata present but no authorization key.
	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("other-key", "value"),
	)

	resp, err := interceptor(ctx, "request", info, handler)
	assert.Nil(t, resp)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid token")
}

func TestAuthInterceptorWrongToken(t *testing.T) {
	interceptor := AuthInterceptor("secret-token")
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	handler := func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	ctx := metadata.NewIncomingContext(context.Background(),
		metadata.Pairs("authorization", "Bearer wrong-token"),
	)

	resp, err := interceptor(ctx, "request", info, handler)
	assert.Nil(t, resp)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
	assert.Contains(t, st.Message(), "invalid token")
}

// --- StartServer ---

func TestStartServer(t *testing.T) {
	registered := false
	conn, cleanup := StartServer(t, func(r grpc.ServiceRegistrar) {
		registered = true
	}, nil)
	defer cleanup()

	assert.True(t, registered)
	require.NotNil(t, conn)

	conn.Connect()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	conn.WaitForStateChange(ctx, connectivity.Idle)

	t.Logf("StartServer: target=%s state=%s", conn.Target(), conn.GetState())
}

func TestStartServerWithInterceptors(t *testing.T) {
	var counter atomic.Int64
	conn, cleanup := StartServer(t, func(r grpc.ServiceRegistrar) {
		// No services; just testing that interceptor opts are applied.
	}, []grpc.UnaryServerInterceptor{
		LoggingInterceptor(t, &counter),
	})
	defer cleanup()

	require.NotNil(t, conn)
	t.Logf("StartServer with interceptors: target=%s", conn.Target())
}

// --- SlowServer ---

func TestSlowServer(t *testing.T) {
	srv := SlowServer(t)
	defer srv.Close()

	require.NotNil(t, srv)

	// A request with an already-cancelled context should return
	// immediately (handler sees r.Context().Done()) instead of
	// blocking for the full 1-second timeout.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	start := time.Now()
	_, err = http.DefaultClient.Do(req)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Less(t, elapsed, 500*time.Millisecond)
	t.Logf("SlowServer cancelled request returned in %v", elapsed)
}

func TestSlowServerRespondsAfterDelay(t *testing.T) {
	srv := SlowServer(t)
	defer srv.Close()

	// With a generous timeout the server responds after its 1s delay.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond)
	t.Logf("SlowServer responded after %v", elapsed)
}
