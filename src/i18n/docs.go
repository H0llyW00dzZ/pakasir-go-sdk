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

// Package i18n provides internationalization support for the Pakasir SDK.
//
// It supports English and Indonesian localization for all SDK-generated
// error messages and validation feedback.
//
// # Supported Languages
//
//   - [English] (default fallback)
//   - [Indonesian] (Bahasa Indonesia)
//
// # Basic Usage
//
//	msg := i18n.Get(i18n.Indonesian, i18n.MsgInvalidAmount)
//	// Output: "jumlah harus lebih dari 0"
//
// If a key is not found for the requested language, the system falls back
// to English. If the key is not found in English either, the raw key string
// is returned.
package i18n
