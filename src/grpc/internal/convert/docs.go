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

// Package convert provides shared mapping between SDK constants and proto
// enum types, as well as SDK-to-gRPC error conversion. It is internal to
// the grpc packages and not exported to consumers.
//
// # Enum Mapping
//
// Bidirectional maps convert between SDK [constants.PaymentMethod] /
// [constants.TransactionStatus] strings and their proto enum counterparts.
// Unrecognized values map to the zero value (empty string or _UNSPECIFIED):
//
//   - [PaymentMethod] — proto → SDK
//   - [PaymentMethodProto] — SDK → proto
//   - [TransactionStatus] — proto → SDK
//   - [TransactionStatusProto] — SDK → proto
//
// # Timestamp Conversion
//
// [Timestamp] converts an RFC3339 time string to a proto
// [timestamppb.Timestamp], returning nil on parse failure.
// [TimeString] converts the other direction, returning an empty string
// for nil timestamps.
//
// # Error Mapping
//
// [Error] maps an SDK error to a gRPC [status.Error] with the appropriate
// [codes.Code]. The original error message is preserved. The mapping
// follows gRPC conventions:
//
//   - Validation sentinels → [codes.InvalidArgument]
//   - Encoding/decoding errors → [codes.Internal]
//   - Size limit errors → [codes.ResourceExhausted]
//   - Retries exhausted → [codes.Unavailable]
//   - Permanent network failures → [codes.Unavailable]
//   - [context.Canceled] → [codes.Canceled]
//   - [context.DeadlineExceeded] → [codes.DeadlineExceeded]
//   - [sdkerrors.APIError] → mapped by HTTP status code
//   - All other errors → [codes.Internal]
//
// [constants.PaymentMethod]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants#PaymentMethod
// [constants.TransactionStatus]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants#TransactionStatus
// [timestamppb.Timestamp]: https://pkg.go.dev/google.golang.org/protobuf/types/known/timestamppb#Timestamp
// [status.Error]: https://pkg.go.dev/google.golang.org/grpc/status#Error
// [codes.Code]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.InvalidArgument]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.ResourceExhausted]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.Unavailable]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.Internal]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.Canceled]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.DeadlineExceeded]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [sdkerrors.APIError]: https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors#APIError
//
// [codes.Unauthenticated]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.PermissionDenied]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.NotFound]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.AlreadyExists]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
// [codes.Unknown]: https://pkg.go.dev/google.golang.org/grpc/codes#Code
package convert
