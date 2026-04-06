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

// Package request provides shared request types and helpers used
// internally by the Pakasir SDK service packages.
//
// # Body
//
// [Body] is the standard JSON request payload sent to all Pakasir API
// endpoints that accept POST requests. It includes project, order ID,
// amount, and API key fields.
//
// # Validation
//
// [ValidateOrderAndAmount] performs common input validation (non-empty
// order ID, positive amount) with localized error messages. Service
// packages call this before building a request.
//
// # Encoding
//
// [EncodeJSON] encodes a value as JSON using a pooled buffer, copies
// the result into an independent []byte, and returns the buffer to the
// pool. On failure it returns a localized [errors.ErrEncodeJSON] error
// wrapping the original marshal cause. This centralizes the
// buffer acquire/encode/release lifecycle so service packages do not
// manage it directly.
//
// This package is internal and not part of the public SDK API.
package request
