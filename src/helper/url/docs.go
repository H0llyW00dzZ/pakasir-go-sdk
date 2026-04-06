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

// Package url provides a builder for Pakasir payment redirect URLs.
//
// Use this package for redirect-based (non-API) integrations where you
// direct customers to the Pakasir hosted payment page.
//
// # Basic Usage
//
//	payURL, err := url.Build("https://app.pakasir.com", "my-project", 99000, url.Options{
//	    OrderID: "INV123456",
//	})
//	// payURL => "https://app.pakasir.com/pay/my-project/99000?order_id=INV123456"
//
// # Options
//
//   - Redirect: Set a custom redirect URL after payment completion
//   - QRISOnly: Force QRIS-only mode (customer sees QR code immediately)
//   - UsePaypal: Use the /paypal/ endpoint instead of /pay/
//
// # Sentinel Errors
//
// [Build] returns sentinel errors for programmatic handling via [errors.Is]:
//
//   - [ErrEmptyBaseURL]: base URL is empty
//   - [ErrEmptyProject]: project slug is empty
//   - [ErrEmptyOrderID]: order ID is empty
//   - [ErrInvalidAmount]: amount is not positive
package url
