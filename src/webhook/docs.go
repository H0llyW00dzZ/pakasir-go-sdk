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
// your configured webhook URL. Use [Parse] to decode the request body
// into a typed [Event] struct.
//
// # Important Security Note
//
// As stated in the Pakasir documentation, you must always verify that
// the Amount and OrderID in the webhook event match a pending transaction
// in your system. The SDK parses the webhook payload but does not perform
// this validation — that is the caller's responsibility.
//
// # Basic Usage
//
//	func webhookHandler(w http.ResponseWriter, r *http.Request) {
//	    event, err := webhook.Parse(r)
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
package webhook
