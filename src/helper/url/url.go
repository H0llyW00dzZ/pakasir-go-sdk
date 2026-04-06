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

package url

import (
	"errors"
	"fmt"
	neturl "net/url"
	"strings"
)

// Sentinel errors for programmatic handling via [errors.Is].
var (
	// ErrEmptyBaseURL is returned when the base URL is empty.
	ErrEmptyBaseURL = errors.New("url: base URL is required")

	// ErrEmptyProject is returned when the project slug is empty.
	ErrEmptyProject = errors.New("url: project is required")

	// ErrEmptyOrderID is returned when the order ID is empty.
	ErrEmptyOrderID = errors.New("url: order ID is required")

	// ErrInvalidAmount is returned when the amount is not positive.
	ErrInvalidAmount = errors.New("url: amount must be greater than 0")
)

// Options configures the payment redirect URL.
type Options struct {
	// OrderID is the unique transaction identifier (required).
	OrderID string

	// Redirect is an optional URL to redirect the customer to after payment.
	// If empty, the customer is redirected to the previous page.
	Redirect string

	// QRISOnly forces QRIS-only mode. The customer will see the QR code
	// immediately and cannot switch to other payment methods.
	QRISOnly bool

	// UsePaypal routes the customer to the Paypal payment page
	// by using /paypal/ instead of /pay/ in the URL path.
	UsePaypal bool
}

// Build generates a Pakasir payment redirect URL.
//
// The baseURL must be a fully qualified URL with a scheme
// (e.g., "https://app.pakasir.com"). Trailing slashes are trimmed.
//
// It returns an error if the baseURL is empty, the project is empty,
// the orderID is empty, or the amount is not positive.
func Build(baseURL, project string, amount int64, opts Options) (string, error) {
	if baseURL == "" {
		return "", ErrEmptyBaseURL
	}
	if project == "" {
		return "", ErrEmptyProject
	}
	if opts.OrderID == "" {
		return "", ErrEmptyOrderID
	}
	if amount <= 0 {
		return "", ErrInvalidAmount
	}

	pathPrefix := "pay"
	if opts.UsePaypal {
		pathPrefix = "paypal"
	}

	base := strings.TrimRight(baseURL, "/")
	u := fmt.Sprintf("%s/%s/%s/%d", base, pathPrefix, neturl.PathEscape(project), amount)

	params := neturl.Values{}
	params.Set("order_id", opts.OrderID)

	if opts.Redirect != "" {
		params.Set("redirect", opts.Redirect)
	}
	if opts.QRISOnly {
		params.Set("qris_only", "1")
	}

	return u + "?" + params.Encode(), nil
}
