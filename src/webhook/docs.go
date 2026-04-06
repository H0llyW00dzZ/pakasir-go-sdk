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

// Package webhook provides helpers for parsing incoming Pakasir webhook
// notifications.
//
// When a customer completes a payment, Pakasir sends an HTTP POST to
// your configured webhook URL. The package offers three entry points
// to decode the payload into a typed [Event] struct:
//
//   - [Parse]: accepts an [io.Reader] — works with any Go HTTP framework
//   - [ParseRequest]: accepts an [http.Request] — convenience for net/http
//   - [ParseBytes]: accepts raw []byte — for frameworks like Fiber
//
// # Sentinel Errors
//
// All parse functions return sentinel errors for programmatic handling
// via [errors.Is]:
//
//   - [ErrNilReader]: nil reader passed to [Parse]
//   - [ErrNilRequest]: nil request or body passed to [ParseRequest] (wraps [errors.ErrNilRequest])
//   - [ErrEmptyBody]: empty payload passed to [ParseBytes]
//   - [ErrReadBody]: body read failure (wraps underlying cause)
//   - [ErrBodyTooLarge]: body exceeds configured size limit
//   - [ErrDecodeBody]: JSON decode failure (wraps underlying cause)
//   - [ErrInvalidOrderID]: empty order ID from [Event.Validate]
//   - [ErrInvalidAmount]: non-positive amount from [Event.Validate]
//
// # Body Size Limit
//
// The [Parse] and [ParseRequest] functions limit the maximum request body
// to [DefaultMaxBodySize] (1 MB) by default. Use [WithMaxBodySize] to
// adjust the limit.
//
// # Sandbox Mode
//
// The [Event.IsSandbox] field indicates whether the event was generated in
// sandbox (testing) mode. Production webhooks set this to false or omit
// the field entirely. Callers can use this flag to route sandbox events to
// a separate processing path or to skip real fulfillment logic.
//
// # Important Security Note
//
// As stated in the Pakasir documentation, you must always verify that
// the Amount and OrderID in the webhook event match a pending transaction
// in your system. Use [Event.Validate] for basic sanity checks (non-empty
// order ID, positive amount), then verify the values against your own
// records.
//
// # net/http
//
//	func webhookHandler(w http.ResponseWriter, r *http.Request) {
//	    event, err := webhook.ParseRequest(r)
//	    if err != nil {
//	        http.Error(w, "bad request", http.StatusBadRequest)
//	        return
//	    }
//
//	    // Validate against your system
//	    if event.OrderID != expectedOrderID || event.Amount != expectedAmount {
//	        http.Error(w, "mismatch", http.StatusBadRequest)
//	        return
//	    }
//
//	    // Process the completed payment...
//	    w.WriteHeader(http.StatusOK)
//	}
//
// # Gin / Echo / Chi
//
//	// Gin
//	event, err := webhook.Parse(c.Request.Body)
//
//	// Echo
//	event, err := webhook.Parse(c.Request().Body)
//
//	// Chi (uses net/http)
//	event, err := webhook.ParseRequest(r)
//
// # Fiber
//
//	event, err := webhook.ParseBytes(c.Body())
package webhook
