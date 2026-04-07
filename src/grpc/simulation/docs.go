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

// Package simulation provides a gRPC service implementation for the
// Pakasir sandbox payment simulation API.
//
// The [Service] type implements [pakasirv1.SimulationServiceServer] by
// delegating to the SDK's [sdksim.Service]. It handles all proto-to-SDK
// conversion internally.
//
// # Usage
//
// Create the service by wrapping an SDK simulation service, then register
// it with a gRPC server:
//
//	c := client.New("my-project", "api-key")
//	simSDK := sdksim.NewService(c)
//	simGRPC := simulation.NewService(simSDK)
//
//	// Register with a standard gRPC server:
//	pakasirv1.RegisterSimulationServiceServer(grpcServer, simGRPC)
//
//	// Or with the grpc-template pattern:
//	srv.RegisterService(func(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterSimulationServiceServer(r, simGRPC)
//	})
package simulation
