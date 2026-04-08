# Berkontribusi pada Pakasir Go SDK

[![Read in English](https://img.shields.io/badge/%F0%9F%87%AC%F0%9F%87%A7_Read_in_English-blue)](CONTRIBUTING.md)

Terima kasih atas minat Anda untuk berkontribusi! Panduan ini akan membantu Anda mempersiapkan lingkungan pengembangan dan berkontribusi secara efektif.

## Prasyarat

- **Go 1.26** atau lebih baru
- `git`
- `buf` (untuk pembuatan proto — instal melalui `make deps`)
- `gocyclo` (untuk analisis kompleksitas — instal melalui `make deps`)

## Memulai

1. **Fork** repositori di GitHub.

2. **Clone** fork Anda:

   ```bash
   git clone https://github.com/USERNAME_ANDA/pakasir-go-sdk.git
   cd pakasir-go-sdk
   ```

3. **Verifikasi** bahwa semuanya berjalan dengan baik:

   ```bash
   make build
   make test
   make vet
   ```

## Struktur Proyek

```
pakasir-go-sdk/
├── src/
│   ├── client/          # Klien HTTP inti dan konfigurasi
│   ├── constants/       # Metode pembayaran, status transaksi
│   ├── errors/          # Sentinel errors dan tipe APIError
│   ├── i18n/            # Internasionalisasi (EN, ID)
│   ├── transaction/     # Layanan transaksi
│   ├── simulation/      # Layanan simulasi sandbox
│   ├── webhook/         # Parsing webhook
│   ├── grpc/
│   │   ├── pakasir/v1/  # Kode protobuf yang di-generate (jangan diedit)
│   │   ├── transaction/ # Server gRPC TransactionService
│   │   ├── simulation/  # Server gRPC SimulationService
│   │   └── internal/    # Konversi enum bersama dan helper pengujian
│   ├── helper/
│   │   ├── gc/          # Pengelolaan buffer pool
│   │   ├── qr/          # Pembuatan kode QR untuk pembayaran QRIS
│   │   └── url/         # Pembangun URL pembayaran
│   └── internal/
│       ├── request/     # Body request internal bersama
│       └── timefmt/     # Helper parsing waktu RFC3339 bersama
├── Makefile             # Target build, test, pembuatan proto, analisis kualitas
├── buf.yaml             # Konfigurasi modul Buf untuk linting proto
├── buf.gen.yaml         # Konfigurasi pembuatan kode Buf
├── proto/               # Definisi Protobuf (berkas .proto)
├── examples/            # Contoh penggunaan
├── LICENSE              # Lisensi Apache 2.0
└── README.md
```

## Panduan Pengembangan

### Format Kode

Semua kode harus diformat sebelum dikirim:

```bash
gofmt -s -w .
```

### Header Lisensi

Setiap berkas `.go` harus diawali dengan header lisensi Apache 2.0:

```go
// Copyright 2026 H0llyW00dzZ
//
// Licensed under the Apache License, Version 2.0 (the "License");
// ...
```

### Dokumentasi

- Setiap paket harus memiliki berkas `docs.go` dengan dokumentasi tingkat paket.
- Ikuti format header standar `// Package name provides...`.
- Sertakan contoh penggunaan dalam dokumentasi.

### Menambahkan Pesan i18n

Saat menambahkan pesan baru yang ditampilkan ke pengguna:

1. Definisikan konstanta `MessageKey` di `src/i18n/messages.go`.
2. Tambahkan terjemahan untuk **Bahasa Inggris dan Indonesia** di map `translations`.
3. Gunakan melalui `errors.New(lang, sentinel, key)` atau `i18n.Get(lang, key)`.

### Penanganan Error

- Gunakan sentinel error dari `src/errors/` untuk penanganan programatik.
- Bungkus error dengan pesan yang telah diterjemahkan menggunakan `errors.New()`.
- Semua error harus mendukung `errors.Is()` terhadap sentinel-nya.

### Layanan gRPC

- **Jangan** mengedit berkas yang di-generate di `src/grpc/pakasir/v1/`. Generate ulang dari definisi `proto/` menggunakan `make proto` (menjalankan `buf generate`).
- Implementasi layanan gRPC mendelegasikan ke layanan berbasis REST SDK — tidak boleh mengandung logika bisnis.
- Konversi enum antara tipe proto dan SDK berada di `src/grpc/internal/convert/`.
- Gunakan helper `src/grpc/internal/grpctest/` (bufconn) untuk pengujian gRPC in-memory. Gunakan tabel `APIErrorStatusCases` yang dibagikan untuk pengujian E2E pemetaan status HTTP → kode gRPC.

## Mengirim Perubahan

1. Buat branch fitur:

   ```bash
   git checkout -b feature/nama-fitur-anda
   ```

2. Lakukan perubahan dengan commit yang jelas dan deskriptif.

3. Pastikan semua pengecekan berhasil:

   ```bash
    make build
    make test
    make vet
    make fmt
    make gocyclo
    ```

4. Push dan buat Pull Request ke branch `master`.

## Lisensi

Dengan berkontribusi, Anda menyetujui bahwa kontribusi Anda akan dilisensikan di bawah Lisensi Apache 2.0.
