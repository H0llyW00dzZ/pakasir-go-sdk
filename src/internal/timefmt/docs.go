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

// Package timefmt provides a shared time-parsing helper for the Pakasir SDK.
//
// The Pakasir API returns timestamps in varying RFC3339 formats.
// [Parse] normalizes this by trying RFC3339Nano first, then falling back
// to RFC3339.
//
// This package is internal and not part of the public SDK API.
package timefmt
