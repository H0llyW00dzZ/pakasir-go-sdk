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

// Package pakasirv1 contains generated Protocol Buffer and gRPC code for
// the Pakasir payment gateway API v1.
//
// All code in this package is generated from the proto definitions in
// proto/pakasir/v1/. Do not edit the *.pb.go or *_grpc.pb.go files
// directly; instead modify the .proto sources and run:
//
//	make proto
//
// # Services
//
// Two gRPC services are defined:
//
//   - [TransactionServiceServer] / [TransactionServiceClient] — create,
//     cancel, and query payment transactions.
//   - [SimulationServiceServer] / [SimulationServiceClient] — simulate
//     payments in sandbox mode.
//
// # Shared Types
//
//   - [PaymentMethod] — enum of supported payment channels (QRIS, BNI VA, etc.).
//   - [TransactionStatus] — enum of transaction states (pending, completed, etc.).
//   - [PaymentInfo] — payment details returned after creating a transaction.
//   - [TransactionInfo] — transaction details returned by a detail query.
//
// # Server Implementations
//
// Server-side implementations of these interfaces live in the parent
// grpc/transaction and grpc/simulation packages, which delegate to the
// SDK's REST-based services. See the parent [grpc] package documentation
// for registration examples and interceptor compatibility.
//
// # Client Usage
//
//	conn, err := grpc.NewClient("localhost:50051",
//	    grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
//	txn := pakasirv1.NewTransactionServiceClient(conn)
//	resp, err := txn.Create(ctx, &pakasirv1.CreateRequest{
//	    OrderId:       "INV-001",
//	    Amount:        50000,
//	    PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
//	})
//
// [grpc]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc
package pakasirv1
