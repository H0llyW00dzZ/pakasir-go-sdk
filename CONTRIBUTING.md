# Contributing to Pakasir Go SDK

[![Baca dalam Bahasa Indonesia](https://img.shields.io/badge/%F0%9F%87%AE%F0%9F%87%A9_Baca_dalam_Bahasa_Indonesia-red)](CONTRIBUTING.id.md)

Thank you for your interest in contributing! This guide will help you get set up and contribute effectively.

## Prerequisites

- **Go 1.26** or later
- `git`
- `buf` (for proto generation — install via `make deps`)
- `gocyclo` (for complexity analysis — install via `make deps`)

## Getting Started

1. **Fork** the repository on GitHub.

2. **Clone** your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/pakasir-go-sdk.git
   cd pakasir-go-sdk
   ```

3. **Verify** the setup:

   ```bash
   make build
   make test
   make vet
   ```

## Project Structure

```
pakasir-go-sdk/
├── src/
│   ├── client/          # Core HTTP client and configuration
│   ├── constants/       # Payment methods, transaction statuses
│   ├── errors/          # Sentinel errors and APIError type
│   ├── i18n/            # Internationalization (EN, ID)
│   ├── transaction/     # Transaction service
│   ├── simulation/      # Sandbox simulation service
│   ├── webhook/         # Webhook parsing
│   ├── grpc/
│   │   ├── pakasir/v1/  # Generated protobuf code (do not edit)
│   │   ├── transaction/ # gRPC TransactionService server
│   │   ├── simulation/  # gRPC SimulationService server
│   │   └── internal/    # Shared enum conversion and test helpers
│   ├── helper/
│   │   ├── gc/          # Buffer pool management
│   │   ├── qr/          # QR code generation
│   │   └── url/         # Payment URL builder
│   └── internal/
│       ├── request/     # Shared internal request body
│       └── timefmt/     # Shared RFC3339 time-parsing helper
├── Makefile             # Build, test, proto generation, quality analysis targets
├── buf.yaml             # Buf module config for proto linting
├── buf.gen.yaml         # Buf code generation config
├── proto/               # Protobuf definitions (.proto files)
├── examples/            # Usage examples
├── LICENSE              # Apache License 2.0
└── README.md
```

## Development Guidelines

### Code Formatting

All code must be formatted before submission:

```bash
gofmt -s -w .
```

### License Header

Every `.go` file must start with the Apache 2.0 license header:

```go
// Copyright 2026 H0llyW00dzZ
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...
```

### Documentation

- Every package must have a `docs.go` file with package-level documentation.
- Follow the standard `// Package name provides...` header format.
- Include usage examples in documentation.

### Adding i18n Messages

When adding new user-facing messages:

1. Define a `MessageKey` constant in `src/i18n/messages.go`.
2. Add translations for **both** English and Indonesian in the `translations` map.
3. Use the key via `errors.New(lang, sentinel, key)` or `i18n.Get(lang, key)`.

### Error Handling

- Use sentinel errors from `src/errors/` for programmatic handling.
- Wrap errors with localized messages using `errors.New()`.
- All errors must support `errors.Is()` against their sentinel.

### gRPC Services

- Do **not** edit generated files in `src/grpc/pakasir/v1/`. Regenerate from `proto/` definitions using `make proto` (runs `buf generate`).
- gRPC service implementations delegate to the SDK's REST-based services — they should not contain business logic.
- Enum conversions between proto and SDK types live in `src/grpc/internal/convert/`.
- Use `src/grpc/internal/grpctest/` helpers (bufconn) for in-memory gRPC tests. Use the shared `APIErrorStatusCases` table for HTTP status → gRPC code mapping E2E tests.

## Submitting Changes

1. Create a feature branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes with clear, descriptive commits.

3. Ensure all checks pass:

   ```bash
    make build
    make test
    make vet
    make fmt
    make gocyclo
    ```

4. Push and open a Pull Request against `master`.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
