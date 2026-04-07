# AGENTS.md — Pakasir Go SDK

Unofficial Go SDK for the Pakasir payment gateway. Module: `github.com/H0llyW00dzZ/pakasir-go-sdk`. Go 1.26+.

All source code lives under `src/`. A `Makefile` is provided for common tasks; standard Go tooling also works directly.

## Build / Lint / Test Commands

```bash
# ── Makefile targets (preferred) ─────────────────────────
make build        # Compile check all packages
make test         # Run all tests with race detector
make test-cover   # Tests + coverage report
make test-e2e     # gRPC end-to-end payment flow test
make vet          # Static analysis
make fmt          # Check formatting (fails if unformatted)
make proto        # Regenerate Go code from proto files (buf generate)
make lint-proto   # Lint proto files (buf lint)
make deps         # Install buf + protoc-gen-go tools
make clean        # Remove coverage artifacts

# ── Raw Go commands ──────────────────────────────────────
# Build all packages
go build ./...

# Run all tests with race detector and coverage
# (excludes generated proto package — same as Makefile)
go test -v -race -coverprofile=coverage.txt -covermode=atomic $(go list ./src/... | grep -v /grpc/pakasir/)

# Run a single package's tests
go test -v -race ./src/client/
go test -v -race ./src/transaction/

# Run a single test by name (regex match)
go test -v -race -run TestDoRetryOn5xxThenSuccess ./src/client/

# Run the gRPC E2E payment flow test
go test -v -race -run TestE2EPaymentFlowSuccess ./src/grpc/

# Static analysis
go vet ./src/...

# Format (must pass before submitting)
gofmt -s -w .

# Check formatting without writing (CI check)
gofmt -s -d .
```

CI runs on 8 OS matrix combinations (ubuntu x86/ARM, macOS, Windows x86/ARM) testing Go 1.26.0 and 1.26.1.
Race detector is skipped on `windows-11-arm`. Coverage is uploaded to Codecov from `ubuntu-latest` only.
CI and Makefile test targets exclude the generated proto package (`grpc/pakasir/v1`) from test runs via `go list ./src/... | grep -v /grpc/pakasir/`.

## Project Layout

```
Makefile                    — Build, test, proto generation, benchmarks
buf.yaml                    — Buf module config for proto linting/breaking changes
buf.gen.yaml                — Buf code generation config (Go output to src/grpc)
src/
  client/                   — Core HTTP client, functional options, retry/backoff
  constants/                — PaymentMethod/TransactionStatus enums, SDK version, API paths
  errors/                   — Sentinel errors (ErrInvalid*, ErrResponseTooLarge, ErrBodyTooLarge, etc.), APIError type, i18n wrapping
  i18n/                     — Language type (en/id), message keys, translation map
  transaction/              — Service: Create, Cancel, Detail
  simulation/               — Service: Pay (sandbox)
  webhook/                  — Parse webhook HTTP requests into Event structs
  grpc/                     — Server-side gRPC service implementations (top-level docs + e2e tests)
  grpc/pakasir/v1/          — Generated protobuf Go code: server interfaces + client stubs (do not edit)
  grpc/transaction/         — gRPC TransactionServiceServer implementation (delegates to SDK transaction.Service)
  grpc/simulation/          — gRPC SimulationServiceServer implementation (delegates to SDK simulation.Service)
  grpc/internal/convert/    — Shared enum/timestamp mapping between SDK constants and proto types (unexported)
  grpc/internal/grpctest/   — In-memory bufconn test helpers, shared mock HTTP servers, and interceptor factories (unexported)
  helper/gc/                — Buffer/Pool interfaces wrapping bytebufferpool
  helper/qr/                — QR code generation for QRIS payment strings (go-qrcode)
  helper/url/               — Payment redirect URL builder
  internal/request/         — Shared request body struct, validation, and JSON encoding (unexported)
  internal/timefmt/         — Shared RFC3339 time-parsing helper (unexported)
proto/                      — Protobuf definitions (.proto files) for gRPC services
examples/                   — Example usage (build-tagged with //go:build ignore)
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
- **Constants**: grouped in `const ()` blocks with consistent prefix — `Method*`, `Status*`, `Default*`, `Msg*`, `Err*`, `SDK*`, `Path*`.
- **Enums**: typed strings (`PaymentMethod string`, `TransactionStatus string`, `Language string`, `MessageKey string`) with a `Valid()` method and unexported validation map.
- **Constructors**: `New(...)` returns `*Client` (no error); `NewService(c)` for service types. Credential validation is deferred to `Do()`.
- **Functional options**: `type Option func(*Client)` with `With*` functions. Webhook parsing uses `type ParseOption func(*parseConfig)` with `WithMaxBodySize`.
- **Getters**: exported read-only accessors for encapsulated fields — `Project()`, `APIKey()`, `Lang()`, `GetBufferPool()`, `QR()`. Service packages use these to read client state.
- **Receivers**: single-letter matching the type (`c *Client`, `s *Service`, `e *Event`, `m PaymentMethod`, `s TransactionStatus`).
- **Unexported helpers**: camelCase (`isRetryable`, `calculateBackoff`, `parseRetryAfter`, `validateCredentials`, `executeAttempt`, `handleResponse`, `validateRequest`).

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

- **Sentinel errors** defined with `errors.New()` in `src/errors/`, prefixed `Err*`. This includes transport/size sentinels (`ErrResponseTooLarge`, `ErrBodyTooLarge`, `ErrNilReader`, `ErrEmptyBody`, `ErrReadBody`, `ErrDecodeBody`) that serve as central definitions for the entire SDK.
- **Package-local sentinels** in standalone packages: `webhook.ErrNilReader`, `webhook.ErrEmptyBody`, `webhook.ErrInvalidOrderID`, `webhook.ErrInvalidAmount`, `url.ErrEmptyBaseURL`, `url.ErrEmptyOrderID`, `qr.ErrEmptyContent`, etc. All webhook sentinels wrap their central `sdkerrors` counterpart via `fmt.Errorf("webhook: %w", sdkerrors.ErrNilReader)` so callers can match either sentinel with `errors.Is`.
- **Localized wrapping** via `sdkerrors.New(lang, sentinel, messageKey, args...)` — always wraps with `%w` so `errors.Is()` works. The variadic `args` accept an error cause (wrapped with `%w`) and/or a string context (substituted into `%s` or appended as suffix). Used for both encoding errors (`ErrEncodeJSON`) and decoding errors (`ErrDecodeJSON`).
- **`fmt.Errorf` wrapping** for non-sentinel errors: `fmt.Errorf("context: %w", err)` with lowercase prefix.
- **`APIError`** struct for HTTP error responses; checked with `errors.As()` or the re-exported `sdkerrors.AsType[*sdkerrors.APIError](err)`.
- **`errors.AsType[T]`** (Go 1.26 generics) for type-asserting errors without a separate variable — used in `isRetryable` to unwrap `*url.Error` and detect TLS certificate errors and permanent DNS failures. Re-exported as `sdkerrors.AsType` so consumers don't need a separate stdlib `errors` import.
- **`errors.HasType[T]`** — boolean shorthand for `AsType` when only presence matters. Re-exported as `sdkerrors.HasType`. Used internally in `isRetryable` to consolidate TLS/x509 checks into a single `switch` case.
- **Package-prefixed messages** in standalone packages: `"webhook: ..."`, `"url: ..."`, `"client: ..."`.
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

In `readResponseBody`, manual pool management (explicit `Reset`+`Put` on each return path) is used instead of `defer` to allow early buffer release on oversize rejection before allocating a copy.

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

Five direct dependencies — keep the footprint minimal:
- `github.com/stretchr/testify` (test only)
- `github.com/valyala/bytebufferpool`
- `github.com/skip2/go-qrcode`
- `google.golang.org/grpc`
- `google.golang.org/protobuf`

### Patterns to Preserve

- No `init()` functions anywhere.
- No global mutable state (except the buffer pool `gc.Default`).
- Service-oriented architecture: each API domain is a `Service` wrapping `*client.Client`.
- Functional options for client configuration.
- Encapsulated client fields: all `Client` struct fields are unexported; use `Project()`, `APIKey()`, `Lang()`, `GetBufferPool()`, `QR()` getters from service packages.
- Internal packages (`src/internal/`) for shared types and validation not exposed to consumers.
- Shared validation via `request.ValidateOrderAndAmount` (avoid duplicating order/amount checks).
- Shared JSON encoding via `request.EncodeJSON` (centralizes buffer pool acquire/encode/release). Internally uses `json.NewEncoder.Encode`, which appends a trailing `\n` to the payload — this is intentional and correct behavior; HTTP servers accept it and RFC 7159 permits trailing whitespace.
- Response body limiting: `client.Do` caps reads at `DefaultMaxResponseSize` (1 MB) configurable via `WithMaxResponseSize`; `webhook.Parse`/`ParseRequest` cap at `DefaultMaxBodySize` (1 MB) configurable via `WithMaxBodySize`.
- Centralized sentinel errors: all SDK-wide sentinels live in `src/errors/`. Standalone packages (client, webhook) do not define their own `errors.New()` sentinels; they reference or wrap the central ones via `fmt.Errorf("webhook: %w", sdkerrors.Err*)`.
- API path constants: all endpoint paths live in `src/constants/paths.go` (`PathTransactionCreate`, `PathTransactionCancel`, `PathTransactionDetail`, `PathPaymentSimulation`). Service packages reference these instead of hardcoding path strings.
- Retry on 429 Too Many Requests in addition to 5xx and network errors; 4xx (other than 429), TLS certificate/handshake errors (`tls.AlertError`, `tls.RecordHeaderError`), and x509 verification errors are never retried.
- Retry-After header support: when a 429 response includes a `Retry-After` header (seconds or HTTP-date), the client uses the indicated delay (clamped to `retryWaitMax`) instead of calculated backoff. Parsed by `parseRetryAfter`, which caps delay-seconds at 24 hours to prevent `time.Duration` overflow before the clamp is applied.
- Permanent DNS failures (`*net.DNSError` with `IsNotFound: true`) are classified as non-retryable; DNS timeouts remain retryable.
- `WithBaseURL` strips trailing slashes to prevent double-slash paths in constructed URLs.
- `Accept: application/json` header is set on all requests by `buildRequest`.
- `Client.Do` is decomposed into small helpers (`validateCredentials`, `executeAttempt`, `handleResponse`, `retriesExhaustedError`) to keep cyclomatic complexity low. Non-retryable errors are propagated via the unexported `stopRetry` wrapper, which `Do` unwraps before returning to the caller.
- Response JSON decode errors are localized via `sdkerrors.New(lang, sdkerrors.ErrDecodeJSON, i18n.MsgFailedToDecode, err)` in transaction service methods.
- gRPC error mapping: `conv.Error` in `grpc/internal/convert` maps SDK errors to proper gRPC status codes. Validation sentinels → `codes.InvalidArgument`, `APIError` → mapped by HTTP status (400→InvalidArgument, 401→Unauthenticated, 403→PermissionDenied, 404→NotFound, 429→ResourceExhausted, 5xx→Internal), encode/decode → `codes.Internal`, size limits → `codes.ResourceExhausted`, retries exhausted → `codes.Unavailable`, `context.Canceled` → `codes.Canceled`, `context.DeadlineExceeded` → `codes.DeadlineExceeded`. All gRPC service methods call `conv.Error(err)` instead of returning raw SDK errors.
- gRPC early enum validation: `grpc/transaction.Create` validates `PAYMENT_METHOD_UNSPECIFIED` before delegating to the SDK, returning the proto enum name (e.g., `PAYMENT_METHOD_UNSPECIFIED`) in the gRPC status message instead of the empty SDK string.
- `formatMessage` in `src/errors/` trims trailing `": "` when `%s` is replaced with an empty string, preventing dangling separators in error messages.
- Invalid payment method errors use `strconv.Quote(method.String())` to make the invalid input visible in error messages (e.g., `"bitcoin"` instead of bare `bitcoin`).
- Context cancellation short-circuits the retry loop: `context.Canceled` and `context.DeadlineExceeded` are never wrapped in `ErrRequestFailedAfterRetries` — they propagate directly via `stopRetry` (from `executeAttempt` when `ctx.Err() != nil`) or from `waitForRetry`'s `select` on `ctx.Done()`.

### Security

- **API key in Detail query string**: The `transaction.Service.Detail` method passes `api_key` as a URL query parameter because the upstream [Pakasir API](https://pakasir.com/p/docs) requires it as a `GET` endpoint. All other endpoints use `POST` with the key in the JSON body. This is an upstream API design constraint, not an SDK choice. The `Detail` godoc and both READMEs document this exposure risk.
