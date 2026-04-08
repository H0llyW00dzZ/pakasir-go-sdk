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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// MockErrorServer returns an httptest.Server that always returns
// HTTP 500 with a JSON error body.
func MockErrorServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
}

// MockHTTPStatusServer returns an httptest.Server that always returns
// the given status code and body.
func MockHTTPStatusServer(t *testing.T, statusCode int, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

// LoggingInterceptor returns a unary server interceptor that increments
// counter for every RPC handled and logs the method and duration via t.
func LoggingInterceptor(t *testing.T, counter *atomic.Int64) grpc.UnaryServerInterceptor {
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

// SlowServer returns an httptest.Server that delays responses long
// enough for context cancellation to trigger. The handler uses a
// bounded sleep to avoid blocking server cleanup.
func SlowServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(1 * time.Second):
		}
	}))
}

// APIErrorCase defines an HTTP status code and its expected gRPC status
// code mapping for E2E API error tests.
type APIErrorCase struct {
	Name       string
	HTTPStatus int
	HTTPBody   string
	GRPCCode   codes.Code
}

// APIErrorStatusCases is the standard set of HTTP status → gRPC code
// mappings tested across all E2E API error tests. It covers client
// errors (4xx), server errors (5xx), and gateway/proxy errors.
//
// HTTP 408 maps to [codes.DeadlineExceeded] because the server timed
// out waiting for the request (not retried by the SDK client).
//
// HTTP 429 maps to [codes.Unavailable] because the SDK client retries
// 429 responses; with retries disabled the resulting
// [sdkerrors.ErrRequestFailedAfterRetries] is caught by the sentinel
// scan in [convert.Error] before the [sdkerrors.APIError] branch.
var APIErrorStatusCases = []APIErrorCase{
	{"400 bad request", http.StatusBadRequest, `{"error":"bad request"}`, codes.InvalidArgument},
	{"401 unauthorized", http.StatusUnauthorized, `{"error":"unauthorized"}`, codes.Unauthenticated},
	{"403 forbidden", http.StatusForbidden, `{"error":"forbidden"}`, codes.PermissionDenied},
	{"404 not found", http.StatusNotFound, `{"error":"not found"}`, codes.NotFound},
	{"408 request timeout", http.StatusRequestTimeout, `{"error":"request timeout"}`, codes.DeadlineExceeded},
	{"409 conflict", http.StatusConflict, `{"error":"duplicate order"}`, codes.AlreadyExists},
	{"429 rate limited", http.StatusTooManyRequests, `{"error":"too many requests"}`, codes.Unavailable},
	{"500 internal", http.StatusInternalServerError, `{"error":"internal server error"}`, codes.Internal},
	{"502 bad gateway", http.StatusBadGateway, `{"error":"bad gateway"}`, codes.Unavailable},
	{"503 unavailable", http.StatusServiceUnavailable, `{"error":"unavailable"}`, codes.Unavailable},
	{"504 gateway timeout", http.StatusGatewayTimeout, `{"error":"gateway timeout"}`, codes.Unavailable},
}

// AuthInterceptor returns a unary server interceptor that rejects
// requests without a valid "authorization" metadata key.
func AuthInterceptor(validToken string) grpc.UnaryServerInterceptor {
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
