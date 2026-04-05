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

package constants

// TransactionStatus represents a transaction status returned by the Pakasir API.
type TransactionStatus string

// Supported transaction status values.
const (
	StatusCompleted TransactionStatus = "completed"
	StatusPending   TransactionStatus = "pending"
	StatusExpired   TransactionStatus = "expired"
	StatusCancelled TransactionStatus = "cancelled"
	StatusCanceled  TransactionStatus = "canceled"
)

// validStatuses is a set of all valid transaction statuses for O(1) lookup.
var validStatuses = map[TransactionStatus]struct{}{
	StatusCompleted: {},
	StatusPending:   {},
	StatusExpired:   {},
	StatusCancelled: {},
	StatusCanceled:  {},
}

// Valid reports whether the transaction status is a recognized Pakasir status.
func (s TransactionStatus) Valid() bool {
	_, ok := validStatuses[s]
	return ok
}

// String returns the string representation of the transaction status.
func (s TransactionStatus) String() string {
	return string(s)
}
