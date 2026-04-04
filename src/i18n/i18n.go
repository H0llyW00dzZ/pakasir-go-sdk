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

package i18n

// Language represents a supported language for SDK messages.
type Language string

const (
	// English is the default language.
	English Language = "en"

	// Indonesian provides localized messages in Bahasa Indonesia.
	Indonesian Language = "id"
)

// Get retrieves the localized message for the given key and language.
// If the language or key is not found, it falls back to English.
func Get(lang Language, key MessageKey) string {
	if msgs, ok := translations[lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	// Fallback to English.
	if msgs, ok := translations[English]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return string(key)
}
