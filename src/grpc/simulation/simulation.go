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

package simulation

import (
	"context"

	pakasirv1 "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/pakasir/v1"
	sdksim "github.com/H0llyW00dzZ/pakasir-go-sdk/src/simulation"
)

// Service implements [pakasirv1.SimulationServiceServer] by delegating
// to the SDK [sdksim.Service].
type Service struct {
	pakasirv1.UnimplementedSimulationServiceServer
	sdk *sdksim.Service
}

// NewService creates a new gRPC simulation [Service] backed by the given
// SDK [sdksim.Service].
func NewService(sdk *sdksim.Service) *Service {
	return &Service{sdk: sdk}
}

// Pay simulates a payment for a pending transaction in sandbox mode.
func (s *Service) Pay(ctx context.Context, req *pakasirv1.PayRequest) (*pakasirv1.PayResponse, error) {
	sdkReq := &sdksim.PayRequest{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	}

	if err := s.sdk.Pay(ctx, sdkReq); err != nil {
		return nil, err
	}

	return &pakasirv1.PayResponse{}, nil
}
