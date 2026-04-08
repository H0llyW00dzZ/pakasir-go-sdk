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

package convert

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
	pakasirv1 "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/pakasir/v1"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/internal/timefmt"
)

// protoToMethod maps proto enum values to SDK [constants.PaymentMethod] strings.
var protoToMethod = map[pakasirv1.PaymentMethod]constants.PaymentMethod{
	pakasirv1.PaymentMethod_PAYMENT_METHOD_CIMB_NIAGA_VA:  constants.MethodCIMBNiagaVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_BNI_VA:         constants.MethodBNIVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS:           constants.MethodQRIS,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_SAMPOERNA_VA:   constants.MethodSampoernaVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_BNC_VA:         constants.MethodBNCVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_MAYBANK_VA:     constants.MethodMaybankVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_PERMATA_VA:     constants.MethodPermataVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_ATM_BERSAMA_VA: constants.MethodATMBersamaVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_ARTHA_GRAHA_VA: constants.MethodArthaGrahaVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_BRI_VA:         constants.MethodBRIVA,
	pakasirv1.PaymentMethod_PAYMENT_METHOD_PAYPAL:         constants.MethodPaypal,
}

// methodToProto maps SDK [constants.PaymentMethod] strings to proto enum values.
var methodToProto = map[constants.PaymentMethod]pakasirv1.PaymentMethod{
	constants.MethodCIMBNiagaVA:  pakasirv1.PaymentMethod_PAYMENT_METHOD_CIMB_NIAGA_VA,
	constants.MethodBNIVA:        pakasirv1.PaymentMethod_PAYMENT_METHOD_BNI_VA,
	constants.MethodQRIS:         pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS,
	constants.MethodSampoernaVA:  pakasirv1.PaymentMethod_PAYMENT_METHOD_SAMPOERNA_VA,
	constants.MethodBNCVA:        pakasirv1.PaymentMethod_PAYMENT_METHOD_BNC_VA,
	constants.MethodMaybankVA:    pakasirv1.PaymentMethod_PAYMENT_METHOD_MAYBANK_VA,
	constants.MethodPermataVA:    pakasirv1.PaymentMethod_PAYMENT_METHOD_PERMATA_VA,
	constants.MethodATMBersamaVA: pakasirv1.PaymentMethod_PAYMENT_METHOD_ATM_BERSAMA_VA,
	constants.MethodArthaGrahaVA: pakasirv1.PaymentMethod_PAYMENT_METHOD_ARTHA_GRAHA_VA,
	constants.MethodBRIVA:        pakasirv1.PaymentMethod_PAYMENT_METHOD_BRI_VA,
	constants.MethodPaypal:       pakasirv1.PaymentMethod_PAYMENT_METHOD_PAYPAL,
}

// PaymentMethod converts a proto [pakasirv1.PaymentMethod] enum to
// the SDK [constants.PaymentMethod]. Unrecognized values return empty string.
func PaymentMethod(pm pakasirv1.PaymentMethod) constants.PaymentMethod {
	return protoToMethod[pm]
}

// PaymentMethodProto converts an SDK [constants.PaymentMethod] to the proto
// enum. Unrecognized values return PAYMENT_METHOD_UNSPECIFIED.
func PaymentMethodProto(pm constants.PaymentMethod) pakasirv1.PaymentMethod {
	return methodToProto[pm]
}

// protoToStatus maps proto enum values to SDK [constants.TransactionStatus] strings.
var protoToStatus = map[pakasirv1.TransactionStatus]constants.TransactionStatus{
	pakasirv1.TransactionStatus_TRANSACTION_STATUS_PENDING:   constants.StatusPending,
	pakasirv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED: constants.StatusCompleted,
	pakasirv1.TransactionStatus_TRANSACTION_STATUS_EXPIRED:   constants.StatusExpired,
	pakasirv1.TransactionStatus_TRANSACTION_STATUS_CANCELLED: constants.StatusCancelled,
	pakasirv1.TransactionStatus_TRANSACTION_STATUS_CANCELED:  constants.StatusCanceled,
}

// statusToProto maps SDK [constants.TransactionStatus] strings to proto enum values.
var statusToProto = map[constants.TransactionStatus]pakasirv1.TransactionStatus{
	constants.StatusPending:   pakasirv1.TransactionStatus_TRANSACTION_STATUS_PENDING,
	constants.StatusCompleted: pakasirv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED,
	constants.StatusExpired:   pakasirv1.TransactionStatus_TRANSACTION_STATUS_EXPIRED,
	constants.StatusCancelled: pakasirv1.TransactionStatus_TRANSACTION_STATUS_CANCELLED,
	constants.StatusCanceled:  pakasirv1.TransactionStatus_TRANSACTION_STATUS_CANCELED,
}

// TransactionStatus converts a proto [pakasirv1.TransactionStatus] enum to
// the SDK [constants.TransactionStatus]. Unrecognized values return empty string.
func TransactionStatus(s pakasirv1.TransactionStatus) constants.TransactionStatus {
	return protoToStatus[s]
}

// TransactionStatusProto converts an SDK [constants.TransactionStatus] to the
// proto enum. Unrecognized values return TRANSACTION_STATUS_UNSPECIFIED.
func TransactionStatusProto(s constants.TransactionStatus) pakasirv1.TransactionStatus {
	return statusToProto[s]
}

// Timestamp converts an RFC3339 time string to a proto [timestamppb.Timestamp].
// It returns nil if the string cannot be parsed.
func Timestamp(s string) *timestamppb.Timestamp {
	if t, err := timefmt.Parse(s); err == nil {
		return timestamppb.New(t)
	}
	return nil
}

// TimeString converts a proto [timestamppb.Timestamp] to an RFC3339Nano string.
// It returns an empty string if the timestamp is nil.
func TimeString(ts *timestamppb.Timestamp) string {
	if ts != nil {
		return ts.AsTime().Format(time.RFC3339Nano)
	}
	return ""
}

// sentinelCodes maps SDK sentinel errors to their corresponding gRPC
// status codes. The table is scanned linearly by [Error], so ordering
// matters only for readability — each sentinel is unique.
var sentinelCodes = [...]struct {
	target error
	code   codes.Code
}{
	// Validation sentinels → InvalidArgument.
	{sdkerrors.ErrNilRequest, codes.InvalidArgument},
	{sdkerrors.ErrInvalidOrderID, codes.InvalidArgument},
	{sdkerrors.ErrInvalidAmount, codes.InvalidArgument},
	{sdkerrors.ErrInvalidPaymentMethod, codes.InvalidArgument},
	{sdkerrors.ErrInvalidProject, codes.InvalidArgument},
	{sdkerrors.ErrInvalidAPIKey, codes.InvalidArgument},

	// Encoding/decoding → Internal.
	{sdkerrors.ErrEncodeJSON, codes.Internal},
	{sdkerrors.ErrDecodeJSON, codes.Internal},

	// Size limits → ResourceExhausted.
	{sdkerrors.ErrResponseTooLarge, codes.ResourceExhausted},
	{sdkerrors.ErrBodyTooLarge, codes.ResourceExhausted},

	// Transport failures → Unavailable.
	{sdkerrors.ErrRequestFailedAfterRetries, codes.Unavailable},
	{sdkerrors.ErrRequestFailed, codes.Unavailable},

	// Context errors.
	{context.Canceled, codes.Canceled},
	{context.DeadlineExceeded, codes.DeadlineExceeded},
}

// Error maps an SDK error to a gRPC [status.Error] with an appropriate
// [codes.Code]. The original error message is preserved.
//
// Mapping:
//
//   - Validation errors ([sdkerrors.ErrNilRequest], [sdkerrors.ErrInvalidOrderID],
//     [sdkerrors.ErrInvalidAmount], [sdkerrors.ErrInvalidPaymentMethod],
//     [sdkerrors.ErrInvalidProject], [sdkerrors.ErrInvalidAPIKey]) → [codes.InvalidArgument]
//   - Encoding/decoding errors ([sdkerrors.ErrEncodeJSON], [sdkerrors.ErrDecodeJSON]) → [codes.Internal]
//   - Size limit errors ([sdkerrors.ErrResponseTooLarge], [sdkerrors.ErrBodyTooLarge]) → [codes.ResourceExhausted]
//   - [sdkerrors.ErrRequestFailedAfterRetries] → [codes.Unavailable]
//   - [sdkerrors.ErrRequestFailed] (permanent network failures) → [codes.Unavailable]
//   - [context.Canceled] → [codes.Canceled]
//   - [context.DeadlineExceeded] → [codes.DeadlineExceeded]
//   - [sdkerrors.APIError] → mapped by HTTP status code (400 → [codes.InvalidArgument],
//     401 → [codes.Unauthenticated], 403 → [codes.PermissionDenied],
//     404 → [codes.NotFound], 409 → [codes.AlreadyExists],
//     502/503/504 → [codes.Unavailable], other 5xx → [codes.Internal],
//     other non-5xx → [codes.Unknown])
//   - All other errors → [codes.Internal]
func Error(err error) error {
	if err == nil {
		return nil
	}

	// Sentinel errors → corresponding gRPC code.
	for _, sc := range sentinelCodes {
		if errors.Is(err, sc.target) {
			return status.Error(sc.code, err.Error())
		}
	}

	// APIError → map HTTP status to gRPC code.
	if apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err); ok {
		return status.Error(httpStatusToCode(apiErr.StatusCode), err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

// httpStatusToCode maps an HTTP status code to the appropriate gRPC
// [codes.Code] following the conventions in gRPC documentation.
//
// Gateway/proxy errors (502, 503, 504) all map to [codes.Unavailable]
// because they indicate the upstream service is unreachable, not a bug
// in the server's logic. This is consistent regardless of whether retries
// are enabled: with retries the SDK exhausts attempts and returns
// [sdkerrors.ErrRequestFailedAfterRetries] (caught earlier by [Error]),
// but with retries disabled (WithRetries(0)) the raw [sdkerrors.APIError]
// reaches this function directly.
//
// HTTP 429 is intentionally absent: the SDK client retries 429 responses,
// so the error reaching this function is always
// [sdkerrors.ErrRequestFailedAfterRetries] (mapped to [codes.Unavailable]
// by [Error] before the [sdkerrors.APIError] branch is reached).
func httpStatusToCode(statusCode int) codes.Code {
	switch statusCode {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	case 502, 503, 504:
		return codes.Unavailable
	default:
		if statusCode >= 500 {
			return codes.Internal
		}
		return codes.Unknown
	}
}
