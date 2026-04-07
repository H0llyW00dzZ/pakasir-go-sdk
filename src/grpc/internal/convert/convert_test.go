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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	pakasirv1 "github.com/H0llyW00dzZ/pakasir-go-sdk/src/grpc/pakasir/v1"
)

// --- PaymentMethod ---

func TestPaymentMethodRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		sdk   constants.PaymentMethod
		proto pakasirv1.PaymentMethod
	}{
		{"CIMB Niaga VA", constants.MethodCIMBNiagaVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_CIMB_NIAGA_VA},
		{"BNI VA", constants.MethodBNIVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_BNI_VA},
		{"QRIS", constants.MethodQRIS, pakasirv1.PaymentMethod_PAYMENT_METHOD_QRIS},
		{"Sampoerna VA", constants.MethodSampoernaVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_SAMPOERNA_VA},
		{"BNC VA", constants.MethodBNCVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_BNC_VA},
		{"Maybank VA", constants.MethodMaybankVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_MAYBANK_VA},
		{"Permata VA", constants.MethodPermataVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_PERMATA_VA},
		{"ATM Bersama VA", constants.MethodATMBersamaVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_ATM_BERSAMA_VA},
		{"Artha Graha VA", constants.MethodArthaGrahaVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_ARTHA_GRAHA_VA},
		{"BRI VA", constants.MethodBRIVA, pakasirv1.PaymentMethod_PAYMENT_METHOD_BRI_VA},
		{"PayPal", constants.MethodPaypal, pakasirv1.PaymentMethod_PAYMENT_METHOD_PAYPAL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PaymentMethodProto(tt.sdk)
			assert.Equal(t, tt.proto, got)

			back := PaymentMethod(got)
			assert.Equal(t, tt.sdk, back)
		})
	}
}

func TestPaymentMethodUnspecified(t *testing.T) {
	assert.Equal(t, constants.PaymentMethod(""), PaymentMethod(pakasirv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED))
	assert.Equal(t, pakasirv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED, PaymentMethodProto("unknown"))
}

// --- TransactionStatus ---

func TestTransactionStatusRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		sdk   constants.TransactionStatus
		proto pakasirv1.TransactionStatus
	}{
		{"Pending", constants.StatusPending, pakasirv1.TransactionStatus_TRANSACTION_STATUS_PENDING},
		{"Completed", constants.StatusCompleted, pakasirv1.TransactionStatus_TRANSACTION_STATUS_COMPLETED},
		{"Expired", constants.StatusExpired, pakasirv1.TransactionStatus_TRANSACTION_STATUS_EXPIRED},
		{"Cancelled", constants.StatusCancelled, pakasirv1.TransactionStatus_TRANSACTION_STATUS_CANCELLED},
		{"Canceled", constants.StatusCanceled, pakasirv1.TransactionStatus_TRANSACTION_STATUS_CANCELED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransactionStatusProto(tt.sdk)
			assert.Equal(t, tt.proto, got)

			back := TransactionStatus(got)
			assert.Equal(t, tt.sdk, back)
		})
	}
}

func TestTransactionStatusUnspecified(t *testing.T) {
	assert.Equal(t, constants.TransactionStatus(""), TransactionStatus(pakasirv1.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED))
	assert.Equal(t, pakasirv1.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED, TransactionStatusProto("bogus"))
}

// --- Timestamp ---

func TestTimestamp(t *testing.T) {
	ts := Timestamp("2026-04-07T12:00:00Z")
	assert.NotNil(t, ts)
	assert.Equal(t, int64(2026), int64(ts.AsTime().Year()))
}

func TestTimestampInvalid(t *testing.T) {
	assert.Nil(t, Timestamp("not-a-time"))
	assert.Nil(t, Timestamp(""))
}

func TestTimeString(t *testing.T) {
	now := time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC)
	ts := timestamppb.New(now)
	s := TimeString(ts)
	assert.Contains(t, s, "2026-12-25")
}

func TestTimeStringNil(t *testing.T) {
	assert.Equal(t, "", TimeString(nil))
}
