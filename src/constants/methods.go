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

// PaymentMethod represents a supported Pakasir payment method.
type PaymentMethod string

// Supported payment methods as defined by the Pakasir API.
const (
	MethodCIMBNiagaVA  PaymentMethod = "cimb_niaga_va"
	MethodBNIVA        PaymentMethod = "bni_va"
	MethodQRIS         PaymentMethod = "qris"
	MethodSampoernaVA  PaymentMethod = "sampoerna_va"
	MethodBNCVA        PaymentMethod = "bnc_va"
	MethodMaybankVA    PaymentMethod = "maybank_va"
	MethodPermataVA    PaymentMethod = "permata_va"
	MethodATMBersamaVA PaymentMethod = "atm_bersama_va"
	MethodArthaGrahaVA PaymentMethod = "artha_graha_va"
	MethodBRIVA        PaymentMethod = "bri_va"
	MethodPaypal       PaymentMethod = "paypal"
)

// validMethods is a set of all valid payment methods for O(1) lookup.
var validMethods = map[PaymentMethod]struct{}{
	MethodCIMBNiagaVA:  {},
	MethodBNIVA:        {},
	MethodQRIS:         {},
	MethodSampoernaVA:  {},
	MethodBNCVA:        {},
	MethodMaybankVA:    {},
	MethodPermataVA:    {},
	MethodATMBersamaVA: {},
	MethodArthaGrahaVA: {},
	MethodBRIVA:        {},
	MethodPaypal:       {},
}

// Valid reports whether the payment method is a recognized Pakasir method.
func (m PaymentMethod) Valid() bool {
	_, ok := validMethods[m]
	return ok
}

// String returns the string representation of the payment method.
func (m PaymentMethod) String() string {
	return string(m)
}
