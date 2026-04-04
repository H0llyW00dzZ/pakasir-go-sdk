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

package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildBasic(t *testing.T) {
	u, err := Build("https://app.pakasir.com", "depodomain", 22000, Options{OrderID: "240910HDE7C9"})
	require.NoError(t, err)
	assert.Contains(t, u, "https://app.pakasir.com/pay/depodomain/22000?")
	assert.Contains(t, u, "order_id=240910HDE7C9")
}

func TestBuildWithRedirect(t *testing.T) {
	u, err := Build("https://app.pakasir.com", "depodomain", 22000, Options{
		OrderID: "INV123", Redirect: "https://example.com/done",
	})
	require.NoError(t, err)
	assert.Contains(t, u, "redirect=")
}

func TestBuildQRISOnly(t *testing.T) {
	u, err := Build("https://app.pakasir.com", "depodomain", 22000, Options{
		OrderID: "INV123", QRISOnly: true,
	})
	require.NoError(t, err)
	assert.Contains(t, u, "qris_only=1")
}

func TestBuildPaypal(t *testing.T) {
	u, err := Build("https://app.pakasir.com", "depodomain", 22000, Options{
		OrderID: "INV123", UsePaypal: true,
	})
	require.NoError(t, err)
	assert.Contains(t, u, "/paypal/depodomain/")
}

func TestBuildAllOptions(t *testing.T) {
	u, err := Build("https://app.pakasir.com", "proj", 50000, Options{
		OrderID: "INV999", Redirect: "https://mysite.com/thanks",
		QRISOnly: true, UsePaypal: true,
	})
	require.NoError(t, err)
	assert.Contains(t, u, "/paypal/proj/50000")
	assert.Contains(t, u, "order_id=INV999")
	assert.Contains(t, u, "qris_only=1")
	assert.Contains(t, u, "redirect=")
}

func TestBuildEmptyOrderID(t *testing.T) {
	_, err := Build("https://app.pakasir.com", "proj", 22000, Options{OrderID: ""})
	require.Error(t, err)
}

func TestBuildZeroAmount(t *testing.T) {
	_, err := Build("https://app.pakasir.com", "proj", 0, Options{OrderID: "INV1"})
	require.Error(t, err)
}

func TestBuildNegativeAmount(t *testing.T) {
	_, err := Build("https://app.pakasir.com", "proj", -100, Options{OrderID: "INV1"})
	require.Error(t, err)
}

func TestBuildEmptyProject(t *testing.T) {
	_, err := Build("https://app.pakasir.com", "", 22000, Options{OrderID: "INV1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project is required")
}

func TestBuildTrailingSlashBaseURL(t *testing.T) {
	u, err := Build("https://app.pakasir.com/", "proj", 22000, Options{OrderID: "INV1"})
	require.NoError(t, err)
	assert.Contains(t, u, "https://app.pakasir.com/pay/proj/22000?")
	assert.NotContains(t, u, "//pay")
}

func TestBuildProjectWithSpecialChars(t *testing.T) {
	u, err := Build("https://app.pakasir.com", "my project/test", 22000, Options{OrderID: "INV1"})
	require.NoError(t, err)
	assert.Contains(t, u, "my%20project%2Ftest")
}
