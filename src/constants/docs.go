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

// Package constants defines the payment methods, transaction statuses,
// and other enumerated values used by the Pakasir API.
//
// # Payment Methods
//
// The [PaymentMethod] type provides type-safe constants for all supported
// Pakasir payment channels, including Virtual Account (VA) banks and QRIS.
// Use the [PaymentMethod.Valid] method to verify a method before sending
// it to the API.
//
// # Transaction Statuses
//
// The [TransactionStatus] type provides type-safe constants for transaction
// lifecycle states (e.g., [StatusCompleted], [StatusPending], [StatusCancelled]).
// Both [StatusCancelled] ("cancelled") and [StatusCanceled] ("canceled")
// spellings are accepted. Use the [TransactionStatus.Valid] method to verify
// a status value.
//
// # SDK Version
//
// The package also provides SDK version metadata ([SDKName], [SDKVersion],
// [SDKRepository]) and a pre-computed [UserAgent] function used by the
// HTTP client for all outgoing requests.
//
// # API Paths
//
// Centralized path constants ([PathTransactionCreate], [PathTransactionCancel],
// [PathTransactionDetail], [PathPaymentSimulation]) are used by the service
// packages to construct request URLs, avoiding hardcoded strings across
// multiple packages.
package constants
