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

// Package grpc provides server-side gRPC service implementations for the
// Pakasir payment gateway SDK.
//
// The transaction/ and simulation/ sub-packages implement the generated
// [pakasirv1.TransactionServiceServer] and [pakasirv1.SimulationServiceServer]
// interfaces by delegating to the SDK's REST-based services. All proto-to-SDK
// conversion is handled internally — consumers only work with proto messages
// on the gRPC side.
//
// Client-side stubs ([pakasirv1.NewTransactionServiceClient],
// [pakasirv1.NewSimulationServiceClient]) are generated automatically in
// pakasir/v1/ and require no additional implementation.
//
// # Package Layout
//
//   - pakasir/v1/      — Generated protobuf code: server interfaces + client stubs (do not edit).
//   - transaction/     — Server: [pakasirv1.TransactionServiceServer] implementation.
//   - simulation/      — Server: [pakasirv1.SimulationServiceServer] implementation.
//   - internal/convert — Shared enum and timestamp mapping (unexported).
//   - internal/grpctest — In-memory bufconn test helpers (unexported).
//
// # Server Registration
//
// The services implement standard gRPC server interfaces and can be
// registered with any [grpc.ServiceRegistrar] — a plain [grpc.Server],
// the grpc-template's server package, or any other framework:
//
//	// SDK setup.
//	c := client.New("my-project", "api-key")
//	txnSvc := grpctxn.NewService(sdktxn.NewService(c))
//	simSvc := grpcsim.NewService(sdksim.NewService(c))
//
//	// Standard gRPC server.
//	pakasirv1.RegisterTransactionServiceServer(grpcServer, txnSvc)
//	pakasirv1.RegisterSimulationServiceServer(grpcServer, simSvc)
//
//	// Or grpc-template pattern.
//	srv.RegisterService(func(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, txnSvc)
//	    pakasirv1.RegisterSimulationServiceServer(r, simSvc)
//	})
//
// # Interceptor Compatibility
//
// Because the services are plain [pakasirv1.TransactionServiceServer] and
// [pakasirv1.SimulationServiceServer] implementations, they work with any
// gRPC interceptor chain — logging, auth, recovery, rate limiting, request
// ID tracing, and so on — without any SDK-specific middleware:
//
//	srv := server.New(
//	    server.WithUnaryInterceptors(
//	        interceptor.RequestID(),
//	        interceptor.Recovery(),
//	        interceptor.Logging(),
//	    ),
//	)
//
// # Client Usage
//
// On the client side, use the generated stubs directly with any gRPC
// client connection:
//
//	txn := pakasirv1.NewTransactionServiceClient(conn)
//	resp, err := txn.Create(ctx, &pakasirv1.CreateRequest{
//	    OrderId:       "INV-001",
//	    Amount:        50000,
//	    PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
//	})
package grpc
