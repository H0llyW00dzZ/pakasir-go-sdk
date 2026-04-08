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

// Package grpctest provides shared in-memory gRPC test helpers.
//
// It uses [bufconn] for zero-port in-memory connections, following the
// same pattern as the grpc-template's testutil package.
//
// # Bufconn Helpers
//
// [NewBufListener] creates a 1 MB in-memory listener. [DialBufNet]
// creates a gRPC client connection to a bufconn listener with
// insecure credentials. [StartServer] combines both into a single call
// that creates a server, registers services via a callback, and returns
// a client connection with a cleanup function.
//
// # Mock HTTP Servers
//
// Reusable mock HTTP servers simulate upstream API responses for
// gRPC service E2E tests:
//
//   - [MockErrorServer] — always returns HTTP 500 with a JSON error body.
//   - [MockHTTPStatusServer] — returns the given status code and body.
//   - [SlowServer] — delays responses for context cancellation tests.
//
// # Interceptor Factories
//
// Interceptor factories verify interceptor pluggability in E2E tests:
//
//   - [LoggingInterceptor] — increments a counter and logs method/duration.
//   - [AuthInterceptor] — rejects requests without a valid "authorization"
//     metadata key.
//
// [bufconn]: https://pkg.go.dev/google.golang.org/grpc/test/bufconn
package grpctest
