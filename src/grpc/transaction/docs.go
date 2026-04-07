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

// Package transaction provides a gRPC service implementation for the
// Pakasir transaction API.
//
// The [Service] type implements [pakasirv1.TransactionServiceServer] by
// delegating to the SDK's [sdktxn.Service]. It handles all proto-to-SDK
// and SDK-to-proto conversion internally, so gRPC handlers work with
// proto messages while the underlying SDK communicates with the Pakasir
// REST API.
//
// # Usage
//
// Create the service by wrapping an SDK transaction service, then register
// it with a gRPC server:
//
//	c := client.New("my-project", "api-key")
//	txnSDK := sdktxn.NewService(c)
//	txnGRPC := transaction.NewService(txnSDK)
//
//	// Register with a standard gRPC server:
//	pakasirv1.RegisterTransactionServiceServer(grpcServer, txnGRPC)
//
//	// Or with the grpc-template pattern:
//	srv.RegisterService(func(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, txnGRPC)
//	})
package transaction
