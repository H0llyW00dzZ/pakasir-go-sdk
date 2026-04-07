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

// Version information for the Pakasir Go SDK.
//
// These values are used to construct the User-Agent header for HTTP requests.
const (
	// SDKName is the name of this SDK.
	SDKName = "pakasir-go-sdk"

	// SDKVersion is the current version of this SDK.
	// This should be updated with each release.
	SDKVersion = "0.2.0"

	// SDKRepository is the GitHub repository URL for this SDK.
	SDKRepository = "https://github.com/H0llyW00dzZ/pakasir-go-sdk"
)

// UserAgent returns the formatted User-Agent string for HTTP requests.
//
// Format: pakasir-go-sdk/0.0.0-dev (+https://github.com/H0llyW00dzZ/pakasir-go-sdk)
//
// This header is automatically set on every request by the [client.Client]
// and cannot be overridden, even when using a custom [http.Client].
func UserAgent() string { return SDKName + "/" + SDKVersion + " (+" + SDKRepository + ")" }
