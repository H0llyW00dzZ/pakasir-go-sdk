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
//	// Or with the [grpc-template] pattern:
//	srv.RegisterService(func(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, txnGRPC)
//	})
//
// # Embedding
//
// The [Service] struct is exported and can be embedded into your own types
// to combine Pakasir payment RPCs with application-specific dependencies
// such as databases, caches, message queues, or loggers. Use [NewService]
// to initialize the embedded service — the unexported sdk field requires
// the constructor:
//
//	type PaymentService struct {
//	    *transaction.Service
//	    db     *sql.DB
//	    cache  *redis.Client
//	    logger logging.Handler
//	}
//
//	func NewPaymentService(sdk *sdktxn.Service, db *sql.DB, cache *redis.Client, l logging.Handler) *PaymentService {
//	    return &PaymentService{
//	        Service: transaction.NewService(sdk),
//	        db:      db,
//	        cache:   cache,
//	        logger:  l,
//	    }
//	}
//
//	// Register provides a [grpc-template]-compatible registration method.
//	func (s *PaymentService) Register(r grpc.ServiceRegistrar) {
//	    pakasirv1.RegisterTransactionServiceServer(r, s)
//	}
//
// Override individual methods as needed — [Create], [Cancel], and [Detail]
// are inherited from the embedded [Service] unless explicitly overridden.
// For example, persist the transaction to your database after creation:
//
//	func (s *PaymentService) Create(ctx context.Context, req *pakasirv1.CreateRequest) (*pakasirv1.CreateResponse, error) {
//	    resp, err := s.Service.Create(ctx, req)
//	    if err != nil {
//	        return nil, err
//	    }
//	    _ = s.db.ExecContext(ctx, "INSERT INTO orders ...", req.GetOrderId())
//	    return resp, nil
//	}
//
// # Internal Dependency in Custom Services
//
// The [Service] can also be used as an in-process dependency inside your
// own proto-defined services. Call [Service.Create], [Service.Cancel], or
// [Service.Detail] directly — these are plain Go method calls that
// delegate to the SDK's REST client, not gRPC calls. No /pakasir.v1.*
// routes are created unless you explicitly register them:
//
//	// proto: package myapp.v1; service OrderService { rpc PlaceOrder(...) ... }
//	type OrderService struct {
//	    myappv1.UnimplementedOrderServiceServer
//	    db      *sql.DB
//	    pakasir *transaction.Service // in-process dependency
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
// The route is /myapp.v1.OrderService/PlaceOrder — determined entirely
// by your proto definition, not by the Pakasir SDK.
//
// [grpc-template]: https://github.com/H0llyW00dzZ/grpc-template
package transaction
