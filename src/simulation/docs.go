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

// Package simulation provides a sandbox payment simulation service
// for the Pakasir SDK.
//
// Use this package to test webhook integrations without processing
// real payments. The simulation endpoint is only available when the
// Pakasir project is in Sandbox mode.
//
// # Basic Usage
//
//	c := client.New("my-project", "api-key-xxx")
//	simService := simulation.NewService(c)
//
//	err := simService.Pay(ctx, &simulation.PayRequest{
//	    OrderID: "INV123456",
//	    Amount:  99000,
//	})
//
// # Error Handling
//
// A nil request pointer returns [errors.ErrNilRequest]. Invalid fields
// return localized errors via [errors.Is] (e.g., [errors.ErrInvalidOrderID]).
//
// [errors.ErrNilRequest]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#ErrNilRequest
// [errors.ErrInvalidOrderID]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#ErrInvalidOrderID
// [errors.Is]: https://pkg.go.dev/errors#Is
package simulation
