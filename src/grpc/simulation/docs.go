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
//	// Or with the [grpc-template] pattern:
//	srv.RegisterService(func(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterSimulationServiceServer(r, simGRPC)
//	})
//
// # Embedding
//
// The [Service] struct is exported and can be embedded into your own types
// to add custom behavior while delegating the core [Service.Pay] RPC to the SDK.
// Use [NewService] to initialize the embedded service — the unexported sdk
// field requires the constructor:
//
//	type SandboxService struct {
//	    *simulation.Service
//	    logger logging.Handler
//	}
//
//	func NewSandboxService(sdk *sdksim.Service, l logging.Handler) *SandboxService {
//	    return &SandboxService{
//	        Service: simulation.NewService(sdk),
//	        logger:  l,
//	    }
//	}
//
//	// Register provides a [grpc-template]-compatible registration method.
//	func (s *SandboxService) Register(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterSimulationServiceServer(r, s)
//	}
//
// # Internal Dependency in Custom Services
//
// The [Service] can also be used as an in-process dependency inside your
// own proto-defined services. Call [Service.Pay] directly — it is a plain
// Go method call that delegates to the SDK's REST client, not a gRPC call.
// No /pakasir.v1.* routes are created unless you explicitly register them:
//
//	// proto: package myapp.v1; service TestingService { rpc SimulatePayment(...) ... }
//	type TestingService struct {
//	    myappv1.UnimplementedTestingServiceServer
//	    sandbox *simulation.Service // in-process dependency
//	    logger  logging.Handler
//	}
//
//	func (s *TestingService) SimulatePayment(ctx context.Context, req *myappv1.SimulatePaymentRequest) (*myappv1.SimulatePaymentResponse, error) {
//	    _, err := s.sandbox.Pay(ctx, &pakasirv1.PayRequest{
//	        OrderId: req.GetOrderId(),
//	        Amount:  req.GetAmount(),
//	    })
//	    if err != nil {
//	        return nil, err
//	    }
//	    s.logger.Info("payment simulated", "order_id", req.GetOrderId())
//	    return &myappv1.SimulatePaymentResponse{}, nil
//	}
//
// The route is /myapp.v1.TestingService/SimulatePayment — determined
// entirely by your proto definition, not by the Pakasir SDK.
//
// [grpc-template]: https://github.com/H0llyW00dzZ/grpc-template
package simulation
