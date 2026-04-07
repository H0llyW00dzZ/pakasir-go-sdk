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

// Package transaction provides the Transaction service for the Pakasir SDK.
//
// It supports creating, cancelling, and querying payment transactions
// via the Pakasir API.
//
// # Basic Usage
//
//	c := client.New("my-project", "api-key-xxx")
//	txnService := transaction.NewService(c)
//
//	// Create a QRIS transaction
//	resp, err := txnService.Create(ctx, constants.MethodQRIS, &transaction.CreateRequest{
//	    OrderID: "INV123456",
//	    Amount:  99000,
//	})
//
//	// Check transaction status
//	detail, err := txnService.Detail(ctx, &transaction.DetailRequest{
//	    OrderID: "INV123456",
//	    Amount:  99000,
//	})
//
//	// Cancel a transaction
//	err = txnService.Cancel(ctx, &transaction.CancelRequest{
//	    OrderID: "INV123456",
//	    Amount:  99000,
//	})
//
// # Error Handling
//
// All methods validate inputs before sending requests. A nil request pointer
// returns [errors.ErrNilRequest]. Invalid fields return localized errors that
// support [errors.Is] against sentinel values from the errors package
// (e.g., [errors.ErrInvalidOrderID], [errors.ErrInvalidAmount]).
//
// # Time Parsing
//
// Response types include a unified [PaymentInfo.ParseTime] and
// [TransactionInfo.ParseTime] method for parsing API timestamp fields.
// Both attempt RFC3339Nano first, then fall back to RFC3339.
//
// # Security: API Key Exposure in Detail Requests
//
// The [Service.Detail] method transmits the API key as a URL query
// parameter because the upstream [Pakasir API] defines this endpoint as a
// GET request. This is an upstream API design constraint, not an SDK
// choice. The key may therefore appear in server access logs, reverse
// proxy / CDN logs, and network monitoring tools.
//
// All other endpoints ([Service.Create], [Service.Cancel]) use POST with
// the API key in the JSON request body, which is not logged by default.
//
// Recommendations:
//   - Redact or exclude query strings from access logs
//   - Rotate API keys periodically via the Pakasir dashboard
//   - Never call Detail from client-side / browser code
//
// [Pakasir API]: https://pakasir.com/p/docs
// [errors.ErrNilRequest]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#ErrNilRequest
// [errors.ErrInvalidOrderID]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#ErrInvalidOrderID
// [errors.ErrInvalidAmount]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#ErrInvalidAmount
// [errors.Is]: https://pkg.go.dev/errors#Is
package transaction
