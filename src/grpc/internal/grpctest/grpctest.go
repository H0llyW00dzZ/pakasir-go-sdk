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
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// NewBufListener creates an in-memory listener for gRPC testing.
// No real TCP port is opened.
func NewBufListener() *bufconn.Listener {
	return bufconn.Listen(bufSize)
}

// DialBufNet creates a gRPC client connection to an in-memory bufconn
// listener. The caller is responsible for closing the returned connection.
// Additional dial options (e.g., interceptors) can be appended.
func DialBufNet(ctx context.Context, lis *bufconn.Listener, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	base := []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	return grpc.NewClient("passthrough:///bufconn", append(base, opts...)...)
}

// StartServer creates a gRPC server on a bufconn listener, registers
// services via the provided register callback, and returns a client
// connection plus a cleanup function. Optional unary interceptors are
// chained in order.
//
// Example:
//
//	conn, cleanup := grpctest.StartServer(t, func(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, txnSvc)
//	}, nil)
//	defer cleanup()
func StartServer(t *testing.T, register func(grpc.ServiceRegistrar), unary []grpc.UnaryServerInterceptor) (*grpc.ClientConn, func()) {
	t.Helper()
	lis := NewBufListener()

	var opts []grpc.ServerOption
	if len(unary) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unary...))
	}

	srv := grpc.NewServer(opts...)
	register(srv)

	go func() {
		if err := srv.Serve(lis); err != nil {
			// Server stopped; expected during cleanup.
		}
	}()

	conn, err := DialBufNet(context.Background(), lis)
	if err != nil {
		t.Fatalf("grpctest: dial bufconn: %v", err)
	}

	return conn, func() {
		conn.Close()
		srv.GracefulStop()
	}
}
