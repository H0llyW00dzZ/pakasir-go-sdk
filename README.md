# Pakasir Go SDK (Unofficial)

[![Go Reference](https://pkg.go.dev/badge/github.com/H0llyW00dzZ/pakasir-go-sdk.svg)](https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/H0llyW00dzZ/pakasir-go-sdk)](https://goreportcard.com/report/github.com/H0llyW00dzZ/pakasir-go-sdk)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![codecov](https://codecov.io/gh/H0llyW00dzZ/pakasir-go-sdk/graph/badge.svg?token=K6I5QCQPDA)](https://codecov.io/gh/H0llyW00dzZ/pakasir-go-sdk)

> **Note:** This is an **unofficial** Go SDK for [Pakasir](https://pakasir.com). It is not affiliated with, endorsed by, or officially supported by Pakasir.

> [!WARNING]
> This SDK is still under active development and is **not recommended for production use**.
> APIs may change without notice. Please wait for an official release from the repository owner before depending on this package.

An idiomatic Go SDK for the [Pakasir](https://pakasir.com) payment gateway. Built with Functional Options, Service-Oriented Architecture, and full i18n support (English & Indonesian).

## Installation

```bash
go get github.com/H0llyW00dzZ/pakasir-go-sdk
```

**Requires Go 1.26 or later.**

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
    "github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
    "github.com/H0llyW00dzZ/pakasir-go-sdk/src/transaction"
)

func main() {
    // Initialize client
    c := client.New("your-project-slug", "your-api-key")

    // Create a QRIS transaction
    txnService := transaction.NewService(c)
    resp, err := txnService.Create(context.Background(), constants.MethodQRIS, &transaction.CreateRequest{
        OrderID: "INV123456",
        Amount:  99000,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Payment Number: %s\n", resp.Payment.PaymentNumber)
    fmt.Printf("Total Payment:  %d\n", resp.Payment.TotalPayment)
}
```

## Features

- **Functional Options** — Clean, extensible client configuration
- **Service-Oriented** — Separate `transaction`, `simulation`, and `webhook` services
- **Context-First** — All I/O operations accept `context.Context`
- **Typed Requests/Responses** — No raw maps; fully typed structs with JSON tags
- **Buffer Pooling** — Memory-efficient request serialization via [`bytebufferpool`](https://github.com/valyala/bytebufferpool)
- **Exponential Backoff with Jitter** — Automatic retry for transient failures (5xx, network errors)
- **i18n** — Localized error messages in English and Indonesian
- **Sentinel Errors** — Programmatic error handling via `errors.Is` and `errors.As`
- **Time Parsing Helpers** — Unified `ParseTime()` on response types
- **URL Builder** — Helper for redirect-based payment integrations

## Project Structure

```
pakasir-go-sdk/
├── src/
│   ├── client/          # Core HTTP client, configuration, buffer pool
│   ├── constants/       # Payment methods, typed transaction statuses
│   ├── errors/          # Sentinel errors, APIError type
│   ├── i18n/            # Internationalization (EN, ID)
│   ├── transaction/     # Transaction service (create, cancel, detail)
│   ├── simulation/      # Payment simulation service (sandbox)
│   ├── webhook/         # Webhook parsing helper
│   ├── helper/
│   │   ├── gc/          # Buffer pool management
│   │   └── url/         # Payment URL builder
│   └── internal/
│       ├── request/     # Shared request body and validation
│       └── timefmt/     # Shared RFC3339 time-parsing helper
├── examples/            # Usage examples
├── LICENSE              # Apache License 2.0
└── README.md
```

## API Coverage

| API Endpoint | SDK Method | Description |
|---|---|---|
| `POST /api/transactioncreate/{method}` | `transaction.Service.Create()` | Create a new transaction |
| `POST /api/transactioncancel` | `transaction.Service.Cancel()` | Cancel a transaction |
| `GET /api/transactiondetail` | `transaction.Service.Detail()` | Get transaction details |
| `POST /api/paymentsimulation` | `simulation.Service.Pay()` | Simulate payment (sandbox) |
| Webhook POST | `webhook.Parse()` | Parse webhook notification |
| Payment URL | `url.Build()` | Build payment redirect URL |

## Payment Methods

| Constant | Value |
|---|---|
| `constants.MethodQRIS` | `qris` |
| `constants.MethodBNIVA` | `bni_va` |
| `constants.MethodBRIVA` | `bri_va` |
| `constants.MethodCIMBNiagaVA` | `cimb_niaga_va` |
| `constants.MethodPermataVA` | `permata_va` |
| `constants.MethodMaybankVA` | `maybank_va` |
| `constants.MethodBNCVA` | `bnc_va` |
| `constants.MethodSampoernaVA` | `sampoerna_va` |
| `constants.MethodATMBersamaVA` | `atm_bersama_va` |
| `constants.MethodArthaGrahaVA` | `artha_graha_va` |
| `constants.MethodPaypal` | `paypal` |

## Transaction Statuses

| Constant | Value |
|---|---|
| `constants.StatusCompleted` | `completed` |
| `constants.StatusPending` | `pending` |
| `constants.StatusExpired` | `expired` |
| `constants.StatusCancelled` | `cancelled` |

## Client Options

```go
c := client.New("project", "api-key",
    client.WithBaseURL("https://custom.api.com"),     // Custom base URL
    client.WithTimeout(10 * time.Second),              // HTTP timeout
    client.WithHTTPClient(customHTTPClient),            // Custom http.Client
    client.WithLanguage(i18n.Indonesian),               // Localized errors
    client.WithRetries(5),                              // Retry attempts
    client.WithRetryWait(500*time.Millisecond, 1*time.Minute), // Backoff config
    client.WithBufferPool(customPool),                  // Custom buffer pool
)
```

## Webhook Handling

The webhook package works with **any Go HTTP framework** via three entry points:

| Function | Input | Use With |
|---|---|---|
| `webhook.Parse(r)` | `io.Reader` | Gin, Echo, any framework |
| `webhook.ParseRequest(r)` | `*http.Request` | net/http, Chi |
| `webhook.ParseBytes(b)` | `[]byte` | Fiber |

All parse functions return sentinel errors for programmatic handling via `errors.Is`:

| Sentinel | Condition |
|---|---|
| `webhook.ErrNilReader` | nil `io.Reader` passed to `Parse` |
| `webhook.ErrNilRequest` | nil `*http.Request` or nil body passed to `ParseRequest` |
| `webhook.ErrEmptyBody` | empty payload passed to `ParseBytes` |
| `webhook.ErrReadBody` | body read failure (wraps underlying cause) |
| `webhook.ErrDecodeBody` | JSON decode failure (wraps underlying cause) |

```go
// net/http
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    event, err := webhook.ParseRequest(r)
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }

    // IMPORTANT: Validate amount and order_id against your system
    if event.OrderID != expectedOrderID || event.Amount != expectedAmount {
        http.Error(w, "mismatch", http.StatusBadRequest)
        return
    }

    if event.Status == constants.StatusCompleted {
        // Process the completed payment...
    }

    w.WriteHeader(http.StatusOK)
}
```

## Error Handling

The SDK provides sentinel errors for programmatic handling via `errors.Is`:

| Sentinel | Package | Condition |
|---|---|---|
| `errors.ErrInvalidProject` | `errors` | Empty project slug |
| `errors.ErrInvalidAPIKey` | `errors` | Empty API key |
| `errors.ErrInvalidOrderID` | `errors` | Empty order ID |
| `errors.ErrInvalidAmount` | `errors` | Non-positive amount |
| `errors.ErrInvalidPaymentMethod` | `errors` | Unsupported payment method |
| `errors.ErrNilRequest` | `errors` | Nil request pointer passed to a service method |
| `errors.ErrRequestFailed` | `errors` | Permanent request failure (non-retryable) |
| `errors.ErrRequestFailedAfterRetries` | `errors` | All retry attempts exhausted |

API error responses are returned as `*errors.APIError` and can be inspected via `errors.As`:

```go
var apiErr *sdkerrors.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Body: %s\n", apiErr.StatusCode, apiErr.Body)
}
```

## Response Types

The SDK provides typed response structs with convenience methods for time parsing:

| Type | Helper Method | Description |
|---|---|---|
| `transaction.PaymentInfo` | `ParseTime()` | Parse payment expiration timestamp |
| `transaction.TransactionInfo` | `ParseTime()` | Parse transaction completion timestamp |
| `webhook.Event` | `ParseTime()` | Parse webhook event completion timestamp |

```go
// Parse expiration time from a create response
expiry, err := resp.Payment.ParseTime()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Expires at: %s\n", expiry)
```

## Disclaimer

This is an **unofficial** SDK. It is not affiliated with, endorsed by, or officially supported by Pakasir. Use at your own risk.

## License

This project is licensed under the Apache License 2.0 — see the [LICENSE](LICENSE) file for details.
