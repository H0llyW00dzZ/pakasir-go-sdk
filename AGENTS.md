# AGENTS.md — Pakasir Go SDK

Unofficial Go SDK for the Pakasir payment gateway. Module: `github.com/H0llyW00dzZ/pakasir-go-sdk`. Go 1.26+.

All source code lives under `src/`. No Makefile or task runner — use standard Go tooling.

## Build / Lint / Test Commands

```bash
# Build all packages
go build ./...

# Run all tests with race detector and coverage
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./src/...

# Run a single package's tests
go test -v -race ./src/client/
go test -v -race ./src/transaction/

# Run a single test by name (regex match)
go test -v -race -run TestDoRetryOn5xxThenSuccess ./src/client/

# Static analysis
go vet ./src/...

# Format (must pass before submitting)
gofmt -s -w .

# Check formatting without writing (CI check)
gofmt -s -d .
```

CI runs on 8 OS targets (ubuntu, macos, windows, including ARM) with Go 1.26.0.
Coverage is uploaded to Codecov from ubuntu-latest only.

## Project Layout

```
src/
  client/           — Core HTTP client, functional options, retry/backoff
  constants/        — PaymentMethod/TransactionStatus enums, SDK version
  errors/           — Sentinel errors (ErrInvalid*), APIError type, i18n wrapping
  i18n/             — Language type (en/id), message keys, translation map
  transaction/      — Service: Create, Cancel, Detail
  simulation/       — Service: Pay (sandbox)
  webhook/          — Parse webhook HTTP requests into Event structs
  helper/gc/        — Buffer/Pool interfaces wrapping bytebufferpool
  helper/url/       — Payment redirect URL builder
  internal/request/ — Shared request body struct, validation, and JSON encoding (unexported)
  internal/timefmt/ — Shared RFC3339 time-parsing helper (unexported)
examples/           — Example usage (build-tagged with //go:build ignore)
```

Every package has a `docs.go` with package-level godoc. Every package has tests.

## Code Style

### License Header

Every `.go` file must start with the Apache 2.0 header using `//` line comments:

```go
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
```

### Imports

Two groups separated by a blank line: (1) stdlib, (2) external + internal module.
In test files, three groups: stdlib / third-party (testify) / internal module.
All groups sorted alphabetically.

```go
import (
    "context"
    "net/http"

    "github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
    sdkerrors "github.com/H0llyW00dzZ/pakasir-go-sdk/src/errors"
)
```

Alias `src/errors` as `sdkerrors` to avoid collision with stdlib `errors`.
Alias `net/url` as `neturl` when inside the `url` package.

### Naming

- **Packages**: lowercase single-word (`client`, `constants`, `gc`, `i18n`).
- **Types**: PascalCase (`Client`, `Service`, `APIError`, `PaymentInfo`).
- **Constants**: grouped in `const ()` blocks with consistent prefix — `Method*`, `Status*`, `Default*`, `Msg*`, `Err*`, `SDK*`.
- **Enums**: typed strings (`PaymentMethod string`, `TransactionStatus string`, `Language string`, `MessageKey string`) with a `Valid()` method and unexported validation map.
- **Constructors**: `New(...)` returns `*Client` (no error); `NewService(c)` for service types. Credential validation is deferred to `Do()`.
- **Functional options**: `type Option func(*Client)` with `With*` functions.
- **Getters**: exported read-only accessors for encapsulated fields — `Project()`, `APIKey()`, `Lang()`, `GetBufferPool()`. Service packages use these to read client state.
- **Receivers**: single-letter matching the type (`c *Client`, `s *Service`, `e *Event`, `m PaymentMethod`, `s TransactionStatus`).
- **Unexported helpers**: camelCase (`isRetryable`, `calculateBackoff`, `validateRequest`).

### Struct Tags

JSON tags use `snake_case`. Fields are column-aligned. No `omitempty` used.

```go
type PaymentInfo struct {
    Project       string `json:"project"`
    OrderID       string `json:"order_id"`
    TotalPayment  int64  `json:"total_payment"`
}
```

### Error Handling

- **Sentinel errors** defined with `errors.New()` in `src/errors/`, prefixed `Err*`.
- **Package-local sentinels** in standalone packages: `webhook.ErrNilReader`, `webhook.ErrEmptyBody`, etc.
- **Localized wrapping** via `sdkerrors.New(lang, sentinel, messageKey)` — always wraps with `%w` so `errors.Is()` works.
- **`fmt.Errorf` wrapping** for non-sentinel errors: `fmt.Errorf("context: %w", err)` with lowercase prefix.
- **`APIError`** struct for HTTP error responses; checked with `errors.As()`.
- **`errors.AsType[T]`** (Go 1.26 generics) for type-asserting errors without a separate variable — used in `isRetryable` to unwrap `*url.Error` and detect TLS certificate errors.
- **Package-prefixed messages** in standalone packages: `"webhook: ..."`, `"url: ..."`.
- **Validate early, return immediately** at the top of functions. `client.New` is an exception: it defers project/API-key validation to `Do()` so initialization is infallible.
- **Nil-guard request pointers** in service methods using `sdkerrors.ErrNilRequest` — distinct from `ErrInvalidOrderID`.

### Documentation

- Package docs live in `docs.go` files using `// Package name provides...` format.
- Use `# Heading` syntax and `[TypeName]` bracket references in godoc.
- Every exported symbol has a doc comment starting with its name.
- Code examples in comments are indented with a leading space.
- `README.md` and `CONTRIBUTING.md` have Indonesian translations: `README.id.md` and `CONTRIBUTING.id.md`.
- Each English doc includes a badge linking to the Indonesian version, and vice versa.
- When updating `README.md` or `CONTRIBUTING.md`, always update the corresponding `.id.md` file to keep translations in sync.

### i18n

All user-facing error messages must be localized. When adding a message:
1. Define a `MessageKey` constant in `src/i18n/messages.go`.
2. Add both English and Indonesian translations in the `translations` map.
3. Use via `sdkerrors.New(lang, sentinel, key)` or `i18n.Get(lang, key)`.

### Buffer Pool Discipline

Always return buffers to the pool after use:

```go
buf := s.client.GetBufferPool().Get()
defer func() {
    buf.Reset()
    s.client.GetBufferPool().Put(buf)
}()
```

## Testing Conventions

- **White-box tests** — test files use the same package (e.g., `package client`).
- **Framework**: `github.com/stretchr/testify` — `require` for fatal checks, `assert` for soft checks.
- **Test naming**: `TestXxxYyy` PascalCase, no underscores.
- **Section comments** separate groups: `// --- New ---`, `// --- Do ---`.
- **Table-driven tests** with `tt` as the loop variable and `t.Run(tt.name, ...)`.
- **Test helpers** use `t.Helper()` for clean stack traces.
- **HTTP mocking** via `httptest.NewServer` with inline `http.HandlerFunc`.
- **Error assertions**: `assert.ErrorIs` for sentinels, `assert.ErrorAs` for types, `assert.Contains` for messages.

### Dependencies

Only two direct dependencies — keep the footprint minimal:
- `github.com/stretchr/testify` (test only)
- `github.com/valyala/bytebufferpool`

### Patterns to Preserve

- No `init()` functions anywhere.
- No global mutable state (except the buffer pool `gc.Default`).
- Service-oriented architecture: each API domain is a `Service` wrapping `*client.Client`.
- Functional options for client configuration.
- Encapsulated client fields: all `Client` struct fields are unexported; use `Project()`, `APIKey()`, `Lang()`, `GetBufferPool()` getters from service packages.
- Internal packages (`src/internal/`) for shared types and validation not exposed to consumers.
- Shared validation via `request.ValidateOrderAndAmount` (avoid duplicating order/amount checks).
- Shared JSON encoding via `request.EncodeJSON` (centralizes buffer pool acquire/encode/release).
