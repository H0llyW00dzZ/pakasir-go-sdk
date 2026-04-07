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

package transaction

import (
	"context"

	conv "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/internal/convert"
	pakasirv1 "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/pakasir/v1"
	sdktxn "github.com/H0llyW00dzZ/pakasir-go-sdk/src/transaction"
)

// Service implements [pakasirv1.TransactionServiceServer] by delegating
// to the SDK [sdktxn.Service].
type Service struct {
	pakasirv1.UnimplementedTransactionServiceServer
	sdk *sdktxn.Service
}

// NewService creates a new gRPC transaction [Service] backed by the given
// SDK [sdktxn.Service].
func NewService(sdk *sdktxn.Service) *Service {
	return &Service{sdk: sdk}
}

// Create creates a new payment transaction.
func (s *Service) Create(ctx context.Context, req *pakasirv1.CreateRequest) (*pakasirv1.CreateResponse, error) {
	method := conv.PaymentMethod(req.GetPaymentMethod())
	sdkReq := &sdktxn.CreateRequest{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	}

	resp, err := s.sdk.Create(ctx, method, sdkReq)
	if err != nil {
		return nil, err
	}

	return &pakasirv1.CreateResponse{
		Payment: paymentInfoToProto(&resp.Payment),
	}, nil
}

// Cancel cancels an existing transaction.
func (s *Service) Cancel(ctx context.Context, req *pakasirv1.CancelRequest) (*pakasirv1.CancelResponse, error) {
	sdkReq := &sdktxn.CancelRequest{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	}

	if err := s.sdk.Cancel(ctx, sdkReq); err != nil {
		return nil, err
	}

	return &pakasirv1.CancelResponse{}, nil
}

// Detail retrieves the details of a transaction.
func (s *Service) Detail(ctx context.Context, req *pakasirv1.DetailRequest) (*pakasirv1.DetailResponse, error) {
	sdkReq := &sdktxn.DetailRequest{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	}

	resp, err := s.sdk.Detail(ctx, sdkReq)
	if err != nil {
		return nil, err
	}

	return &pakasirv1.DetailResponse{
		Transaction: transactionInfoToProto(&resp.Transaction),
	}, nil
}

// paymentInfoToProto converts an SDK [sdktxn.PaymentInfo] to proto.
func paymentInfoToProto(p *sdktxn.PaymentInfo) *pakasirv1.PaymentInfo {
	if p == nil {
		return nil
	}
	return &pakasirv1.PaymentInfo{
		Project:       p.Project,
		OrderId:       p.OrderID,
		Amount:        p.Amount,
		Fee:           p.Fee,
		TotalPayment:  p.TotalPayment,
		PaymentMethod: conv.PaymentMethodProto(p.PaymentMethod),
		PaymentNumber: p.PaymentNumber,
		ExpiredAt:     conv.Timestamp(p.ExpiredAt),
	}
}

// transactionInfoToProto converts an SDK [sdktxn.TransactionInfo] to proto.
func transactionInfoToProto(t *sdktxn.TransactionInfo) *pakasirv1.TransactionInfo {
	if t == nil {
		return nil
	}
	return &pakasirv1.TransactionInfo{
		Amount:        t.Amount,
		OrderId:       t.OrderID,
		Project:       t.Project,
		Status:        conv.TransactionStatusProto(t.Status),
		PaymentMethod: conv.PaymentMethodProto(t.PaymentMethod),
		CompletedAt:   conv.Timestamp(t.CompletedAt),
	}
}
