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
//   - pakasir/v1/       — Generated protobuf code: server interfaces + client stubs (do not edit).
//   - transaction/      — Server: [pakasirv1.TransactionServiceServer] implementation.
//   - simulation/       — Server: [pakasirv1.SimulationServiceServer] implementation.
//   - internal/convert  — Shared enum and timestamp mapping (unexported).
//   - internal/grpctest — In-memory bufconn test helpers (unexported).
//
// # Server Registration
//
// The services implement standard gRPC server interfaces and can be
// registered with any [grpc.ServiceRegistrar] — a plain [grpc.Server],
// the [grpc-template]'s server package, or any other framework:
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
// # Embedding into Existing Codebases
//
// The [Service] types in transaction/ and simulation/ are exported structs
// that can be embedded into your own types. This lets you combine the
// Pakasir RPCs with your own dependencies — databases, caches, message
// queues, metrics, loggers — while delegating the core payment logic to
// the SDK. Use [NewService] to initialize the embedded service — the
// unexported sdk field requires the constructor.
//
// Embed a service alongside application dependencies:
//
//	type PaymentService struct {
//	    *grpctxn.Service
//	    db     *sql.DB
//	    cache  *redis.Client
//	    logger logging.Handler
//	}
//
//	func NewPaymentService(sdk *sdktxn.Service, db *sql.DB, cache *redis.Client, l logging.Handler) *PaymentService {
//	    return &PaymentService{
//	        Service: grpctxn.NewService(sdk),
//	        db:      db,
//	        cache:   cache,
//	        logger:  l,
//	    }
//	}
//
//	func (s *PaymentService) Register(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, s)
//	}
//
// Override individual methods to add custom behavior — unoverridden
// methods are inherited from the embedded [Service]. For example,
// persist the transaction to your database after creation:
//
//	func (s *PaymentService) Create(ctx context.Context, req *pakasirv1.CreateRequest) (*pakasirv1.CreateResponse, error) {
//	    resp, err := s.Service.Create(ctx, req) // delegate to SDK
//	    if err != nil {
//	        return nil, err
//	    }
//	    // Store in your database, update cache, publish event, etc.
//	    _ = s.db.ExecContext(ctx, "INSERT INTO orders ...", req.GetOrderId())
//	    return resp, nil
//	}
//
// Compose multiple Pakasir services into a single registration unit:
//
//	type PakasirServices struct {
//	    txn *grpctxn.Service
//	    sim *grpcsim.Service
//	}
//
//	func NewPakasirServices(c *client.Client) *PakasirServices {
//	    return &PakasirServices{
//	        txn: grpctxn.NewService(sdktxn.NewService(c)),
//	        sim: grpcsim.NewService(sdksim.NewService(c)),
//	    }
//	}
//
//	func (s *PakasirServices) Register(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, s.txn)
//	    pakasirv1.RegisterSimulationServiceServer(r, s.sim)
//	}
//
// Then register alongside other [grpc-template] services:
//
//	pakasirSvc := NewPakasirServices(sdkClient)
//	greeterSvc := greeter.NewService(srv.Logger())
//
//	srv.RegisterService(
//	    greeterSvc.Register,
//	    pakasirSvc.Register,
//	)
//
// # Internal Dependency in Custom Services
//
// The gRPC service types can also be used as internal dependencies inside
// your own proto-defined services without exposing /pakasir.v1.* routes.
// Call methods on *grpctxn.Service or *grpcsim.Service directly — these
// are plain Go method calls, not gRPC calls. The SDK handles proto-to-REST
// conversion internally:
//
//	// proto: package myapp.v1; service OrderService { rpc PlaceOrder(...) ... }
//	type OrderService struct {
//	    myappv1.UnimplementedOrderServiceServer
//	    db      *sql.DB
//	    pakasir *grpctxn.Service // in-process dependency, not a gRPC route
//	}
//
//	func (s *OrderService) PlaceOrder(ctx context.Context, req *myappv1.PlaceOrderRequest) (*myappv1.PlaceOrderResponse, error) {
//	    resp, err := s.pakasir.Create(ctx, &pakasirv1.CreateRequest{
//	        OrderId:       req.GetOrderId(),
//	        Amount:        req.GetAmount(),
//	        PaymentMethod: pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
//	    })
//	    if err != nil {
//	        return nil, err
//	    }
//	    payment := resp.GetPayment()
//	    _ = s.db.ExecContext(ctx, "INSERT INTO orders ...", req.GetOrderId())
//	    return &myappv1.PlaceOrderResponse{
//	        PaymentNumber: payment.GetPaymentNumber(),
//	    }, nil
//	}
//
// The route is /myapp.v1.OrderService/PlaceOrder — determined by your
// proto definition. The Pakasir SDK is an implementation detail:
//
//	orderSvc := &OrderService{db: db, pakasir: grpctxn.NewService(sdktxn.NewService(c))}
//	srv.RegisterService(orderSvc.Register) // only /myapp.v1.OrderService/* routes
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
//
// # End-to-End Test
//
// To see the full payment lifecycle (create -> simulate pay -> verify
// completed) running over gRPC with automatic proto-to-SDK conversion:
//
//	go test -v -race -run TestE2EPaymentFlowSuccess ./src/grpc/
//
// [grpc-template]: https://github.com/H0llyW00dzZ/grpc-template
package grpc
