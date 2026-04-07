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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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
