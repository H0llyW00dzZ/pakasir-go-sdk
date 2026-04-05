# Contributing to Pakasir Go SDK

[![Baca dalam Bahasa Indonesia](https://img.shields.io/badge/%F0%9F%87%AE%F0%9F%87%A9_Baca_dalam_Bahasa_Indonesia-red)](CONTRIBUTING.id.md)

Thank you for your interest in contributing! This guide will help you get set up and contribute effectively.

## Prerequisites

- **Go 1.26** or later
- `git`

## Getting Started

1. **Fork** the repository on GitHub.

2. **Clone** your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/pakasir-go-sdk.git
   cd pakasir-go-sdk
   ```

3. **Verify** the setup:

   ```bash
   go build ./...
   go test ./...
   go vet ./...
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
│   ├── helper/
│   │   ├── gc/          # Buffer pool management
│   │   ├── qr/          # QR code generation
│   │   └── url/         # Payment URL builder
│   └── internal/
│       ├── request/     # Shared internal request body
│       └── timefmt/     # Shared RFC3339 time-parsing helper
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

## Submitting Changes

1. Create a feature branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes with clear, descriptive commits.

3. Ensure all checks pass:

   ```bash
   go build ./...
   go test ./...
   go vet ./...
   gofmt -s -d .
   ```

4. Push and open a Pull Request against `main`.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
