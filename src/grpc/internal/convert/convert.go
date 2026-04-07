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

// Package convert provides shared enum mapping between SDK constants and
// proto enum types. It is internal to the grpc packages and not exported
// to consumers.
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
//   - [context.Canceled] → [codes.Canceled]
//   - [context.DeadlineExceeded] → [codes.DeadlineExceeded]
//   - [sdkerrors.APIError] → mapped by HTTP status code (503 → [codes.Unavailable],
//     other 5xx → [codes.Internal])
//   - All other errors → [codes.Internal]
func Error(err error) error {
	if err == nil {
		return nil
	}

	// Validation sentinels → InvalidArgument.
	switch {
	case errors.Is(err, sdkerrors.ErrNilRequest),
		errors.Is(err, sdkerrors.ErrInvalidOrderID),
		errors.Is(err, sdkerrors.ErrInvalidAmount),
		errors.Is(err, sdkerrors.ErrInvalidPaymentMethod),
		errors.Is(err, sdkerrors.ErrInvalidProject),
		errors.Is(err, sdkerrors.ErrInvalidAPIKey):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, sdkerrors.ErrEncodeJSON),
		errors.Is(err, sdkerrors.ErrDecodeJSON):
		return status.Error(codes.Internal, err.Error())

	case errors.Is(err, sdkerrors.ErrResponseTooLarge),
		errors.Is(err, sdkerrors.ErrBodyTooLarge):
		return status.Error(codes.ResourceExhausted, err.Error())

	case errors.Is(err, sdkerrors.ErrRequestFailedAfterRetries):
		return status.Error(codes.Unavailable, err.Error())

	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())

	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	}

	// APIError → map HTTP status to gRPC code.
	if apiErr, ok := sdkerrors.AsType[*sdkerrors.APIError](err); ok {
		return status.Error(httpStatusToCode(apiErr.StatusCode), err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}

// httpStatusToCode maps an HTTP status code to the appropriate gRPC
// [codes.Code] following the conventions in gRPC documentation.
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
	case 429:
		return codes.ResourceExhausted
	case 503:
		return codes.Unavailable
	default:
		if statusCode >= 500 {
			return codes.Internal
		}
		return codes.Unknown
	}
}
