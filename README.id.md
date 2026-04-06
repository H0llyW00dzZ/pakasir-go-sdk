# Pakasir Go SDK (Tidak Resmi)

[![Go Reference](https://pkg.go.dev/badge/github.com/H0llyW00dzZ/pakasir-go-sdk.svg)](https://pkg.go.dev/github.com/H0llyW00dzZ/pakasir-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/H0llyW00dzZ/pakasir-go-sdk)](https://goreportcard.com/report/github.com/H0llyW00dzZ/pakasir-go-sdk)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![codecov](https://codecov.io/gh/H0llyW00dzZ/pakasir-go-sdk/graph/badge.svg?token=K6I5QCQPDA)](https://codecov.io/gh/H0llyW00dzZ/pakasir-go-sdk)
[![Read in English](https://img.shields.io/badge/%F0%9F%87%AC%F0%9F%87%A7_Read_in_English-blue)](README.md)

> **Catatan:** Ini adalah SDK Go **tidak resmi** untuk [Pakasir](https://pakasir.com). SDK ini tidak berafiliasi, didukung, atau dikelola secara resmi oleh Pakasir. SDK ini tidak resmi karena API resmi hanya menyediakan dokumentasi dan dukungan untuk REST API dan SDK Node.js mereka. Library ini dibuat untuk menambahkan dukungan Go yang layak, dan digunakan secara aktif oleh pemilik repositori ini.

> [!WARNING]
> SDK ini masih dalam tahap pengembangan aktif dan **belum disarankan untuk digunakan di lingkungan produksi**.
> API dapat berubah sewaktu-waktu tanpa pemberitahuan. Harap tunggu rilis resmi dari pemilik repositori sebelum menggunakan paket ini.

SDK Go idiomatik untuk payment gateway [Pakasir](https://pakasir.com). Dibangun dengan Functional Options, Arsitektur Berbasis Layanan (Service-Oriented), dan dukungan i18n penuh (Inggris & Indonesia).

## Instalasi

```bash
go get github.com/H0llyW00dzZ/pakasir-go-sdk
```

**Membutuhkan Go 1.26 atau lebih baru.**

## Mulai Cepat

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
    // Inisialisasi klien
    c := client.New("slug-proyek-anda", "api-key-anda")

    // Buat transaksi QRIS
    txnService := transaction.NewService(c)
    resp, err := txnService.Create(context.Background(), constants.MethodQRIS, &transaction.CreateRequest{
        OrderID: "INV123456",
        Amount:  99000,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Nomor Pembayaran: %s\n", resp.Payment.PaymentNumber)
    fmt.Printf("Total Pembayaran: %d\n", resp.Payment.TotalPayment)
}
```

## Fitur

- **Functional Options** — Konfigurasi klien yang bersih dan mudah diperluas
- **Berbasis Layanan** — Layanan terpisah untuk `transaction`, `simulation`, dan `webhook`
- **Context-First** — Semua operasi I/O menerima `context.Context`
- **Request/Response Bertipe** — Tanpa raw map; struct bertipe penuh dengan tag JSON
- **Buffer Pooling** — Serialisasi request yang hemat memori menggunakan [`bytebufferpool`](https://github.com/valyala/bytebufferpool)
- **Exponential Backoff dengan Jitter** — Retry otomatis untuk kegagalan sementara (429, 5xx, error jaringan)
- **i18n** — Pesan error dalam Bahasa Inggris dan Indonesia
- **Sentinel Errors** — Penanganan error secara programatik melalui `errors.Is` dan `errors.As`
- **Helper Parsing Waktu** — Method `ParseTime()` pada tipe response
- **URL Builder** — Helper untuk integrasi pembayaran berbasis redirect
- **Pembuatan Kode QR** — Render string pembayaran QRIS menjadi gambar PNG dengan ukuran, tingkat pemulihan, dan warna yang dapat dikonfigurasi

## Struktur Proyek

```
pakasir-go-sdk/
├── src/
│   ├── client/          # Klien HTTP inti, konfigurasi, buffer pool
│   ├── constants/       # Metode pembayaran, status transaksi bertipe
│   ├── errors/          # Sentinel errors, tipe APIError
│   ├── i18n/            # Internasionalisasi (EN, ID)
│   ├── transaction/     # Layanan transaksi (buat, batalkan, detail)
│   ├── simulation/      # Layanan simulasi pembayaran (sandbox)
│   ├── webhook/         # Helper parsing webhook
│   ├── helper/
│   │   ├── gc/          # Pengelolaan buffer pool
│   │   ├── qr/          # Pembuatan kode QR untuk pembayaran QRIS
│   │   └── url/         # Pembangun URL pembayaran
│   └── internal/
│       ├── request/     # Body request internal bersama
│       └── timefmt/     # Helper parsing waktu RFC3339 bersama
├── examples/            # Contoh penggunaan
├── LICENSE              # Lisensi Apache 2.0
└── README.md
```

## Cakupan API

| Endpoint API | Method SDK | Keterangan |
|---|---|---|
| `POST /api/transactioncreate/{method}` | `transaction.Service.Create()` | Membuat transaksi baru |
| `POST /api/transactioncancel` | `transaction.Service.Cancel()` | Membatalkan transaksi |
| `GET /api/transactiondetail` | `transaction.Service.Detail()` | Mendapatkan detail transaksi |
| `POST /api/paymentsimulation` | `simulation.Service.Pay()` | Simulasi pembayaran (sandbox) |
| Webhook POST | `webhook.Parse()` | Parsing notifikasi webhook |
| URL Pembayaran | `url.Build()` | Membangun URL redirect pembayaran |

## Metode Pembayaran

| Konstanta | Nilai |
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

## Status Transaksi

| Konstanta | Nilai |
|---|---|
| `constants.StatusCompleted` | `completed` |
| `constants.StatusPending` | `pending` |
| `constants.StatusExpired` | `expired` |
| `constants.StatusCancelled` | `cancelled` |
| `constants.StatusCanceled` | `canceled` |

## Opsi Klien

```go
c := client.New("proyek", "api-key",
    client.WithBaseURL("https://custom.api.com"),               // URL dasar kustom
    client.WithTimeout(10 * time.Second),                       // Timeout HTTP
    client.WithHTTPClient(customHTTPClient),                    // http.Client kustom (shallow-copied)
    client.WithLanguage(i18n.Indonesian),                       // Error dalam Bahasa Indonesia
    client.WithRetries(5),                                      // Jumlah percobaan ulang
    client.WithRetryWait(500*time.Millisecond, 1*time.Minute),  // Konfigurasi backoff
    client.WithBufferPool(customPool),                          // Buffer pool kustom
    client.WithMaxResponseSize(5 << 20),                        // Batas body response (default 1 MB)
    client.WithQRCodeOptions(qr.WithSize(512)),                 // Pengaturan kode QR
)
```

## Pembuatan Kode QR

Paket `qr` merender string pembayaran QRIS menjadi gambar PNG. Dapat digunakan melalui klien atau secara mandiri:

```go
// Melalui klien (dikonfigurasi dengan WithQRCodeOptions)
png, err := c.QR().Encode(resp.Payment.PaymentNumber)

// Mandiri
q := qr.New(qr.WithSize(512), qr.WithRecoveryLevel(qr.RecoveryHigh))
png, err := q.Encode(paymentNumber)
```

Sajikan kode QR langsung melalui HTTP dengan framework apapun:

```go
// net/http
w.Header().Set("Content-Type", "image/png")
err := c.QR().Write(w, resp.Payment.PaymentNumber)
```

Simpan kode QR ke file:

```go
err := c.QR().WriteFile("payment_qr.png", resp.Payment.PaymentNumber)
```

Semua method QR mengembalikan sentinel error untuk penanganan programatik melalui `errors.Is`:

| Sentinel | Kondisi |
|---|---|
| `qr.ErrEmptyContent` | String kosong diberikan ke `Encode`, `Write`, atau `WriteFile` |
| `qr.ErrEncodeFailed` | Encoding QR gagal (membungkus penyebab asli) |

| Opsi | Keterangan | Default |
|---|---|---|
| `qr.WithSize(pixels)` | Lebar/tinggi gambar dalam piksel | 256 |
| `qr.WithRecoveryLevel(level)` | Tingkat koreksi error | `RecoveryMedium` |
| `qr.WithForegroundColor(color)` | Warna modul QR | `color.Black` |
| `qr.WithBackgroundColor(color)` | Warna latar belakang | `color.White` |

## Penanganan Webhook

Paket webhook dapat digunakan dengan **framework Go apapun** melalui tiga entry point:

| Fungsi | Input | Digunakan Dengan |
|---|---|---|
| `webhook.Parse(r)` | `io.Reader` | Gin, Echo, framework apapun |
| `webhook.ParseRequest(r)` | `*http.Request` | net/http, Chi |
| `webhook.ParseBytes(b)` | `[]byte` | Fiber |

`Parse` dan `ParseRequest` menerima opsional `webhook.WithMaxBodySize(n)` untuk mengubah batas ukuran body default 1 MB.

Semua fungsi parse mengembalikan sentinel error untuk penanganan programatik melalui `errors.Is`:

| Sentinel | Kondisi |
|---|---|
| `webhook.ErrNilReader` | `io.Reader` nil diberikan ke `Parse` |
| `webhook.ErrNilRequest` | `*http.Request` nil atau body nil diberikan ke `ParseRequest` |
| `webhook.ErrEmptyBody` | Payload kosong diberikan ke `ParseBytes` |
| `webhook.ErrReadBody` | Gagal membaca body (membungkus penyebab asli) |
| `webhook.ErrBodyTooLarge` | Body melebihi batas ukuran yang dikonfigurasi |
| `webhook.ErrDecodeBody` | Gagal decode JSON (membungkus penyebab asli) |

```go
// net/http
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    event, err := webhook.ParseRequest(r)
    if err != nil {
        http.Error(w, "request tidak valid", http.StatusBadRequest)
        return
    }

    // PENTING: Validasi amount dan order_id dengan sistem Anda
    if event.OrderID != expectedOrderID || event.Amount != expectedAmount {
        http.Error(w, "tidak cocok", http.StatusBadRequest)
        return
    }

    // Event sandbox tidak boleh memicu fulfillment sesungguhnya
    if event.IsSandbox {
        log.Println("webhook sandbox diterima, melewati fulfillment")
        w.WriteHeader(http.StatusOK)
        return
    }

    if event.Status == constants.StatusCompleted {
        // Proses pembayaran yang sudah selesai...
    }

    w.WriteHeader(http.StatusOK)
}
```

## Penanganan Error

SDK ini menyediakan sentinel error untuk penanganan programatik melalui `errors.Is`:

| Sentinel | Paket | Kondisi |
|---|---|---|
| `errors.ErrInvalidProject` | `errors` | Slug proyek kosong |
| `errors.ErrInvalidAPIKey` | `errors` | API key kosong |
| `errors.ErrInvalidOrderID` | `errors` | Order ID kosong |
| `errors.ErrInvalidAmount` | `errors` | Amount tidak positif |
| `errors.ErrInvalidPaymentMethod` | `errors` | Metode pembayaran tidak didukung |
| `errors.ErrNilRequest` | `errors` | Pointer request nil diberikan ke method layanan |
| `errors.ErrEncodeJSON` | `errors` | Gagal marshaling JSON pada body request |
| `errors.ErrRequestFailed` | `errors` | Kegagalan request permanen (tidak bisa di-retry) |
| `errors.ErrRequestFailedAfterRetries` | `errors` | Semua percobaan retry habis |
| `client.ErrResponseTooLarge` | `client` | Body response melebihi batas ukuran yang dikonfigurasi |
| `webhook.ErrNilReader` | `webhook` | `io.Reader` nil diberikan ke `Parse` |
| `webhook.ErrNilRequest` | `webhook` | `*http.Request` nil atau body nil diberikan ke `ParseRequest` |
| `webhook.ErrEmptyBody` | `webhook` | Payload kosong |
| `webhook.ErrReadBody` | `webhook` | Gagal membaca body (membungkus penyebab asli) |
| `webhook.ErrBodyTooLarge` | `webhook` | Body melebihi batas ukuran yang dikonfigurasi |
| `webhook.ErrDecodeBody` | `webhook` | Gagal decode JSON (membungkus penyebab asli) |
| `qr.ErrEmptyContent` | `qr` | String kosong diberikan ke `Encode`, `Write`, atau `WriteFile` |
| `qr.ErrEncodeFailed` | `qr` | Encoding QR gagal (membungkus penyebab asli) |
| `url.ErrEmptyBaseURL` | `url` | Base URL kosong |
| `url.ErrEmptyProject` | `url` | Slug proyek kosong |
| `url.ErrEmptyOrderID` | `url` | ID pesanan kosong |
| `url.ErrInvalidAmount` | `url` | Jumlah tidak positif |

Response error API dikembalikan sebagai `*errors.APIError` dan dapat diperiksa melalui `errors.As`:

```go
var apiErr *sdkerrors.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Body: %s\n", apiErr.StatusCode, apiErr.Body)
}
```

## Tipe Response

SDK ini menyediakan struct response bertipe dengan method untuk parsing waktu:

| Tipe | Method Helper | Keterangan |
|---|---|---|
| `transaction.PaymentInfo` | `ParseTime()` | Parsing timestamp kedaluwarsa pembayaran |
| `transaction.TransactionInfo` | `ParseTime()` | Parsing timestamp penyelesaian transaksi |
| `webhook.Event` | `ParseTime()` | Parsing timestamp penyelesaian event webhook |

```go
// Parsing waktu kedaluwarsa dari response pembuatan transaksi
expiry, err := resp.Payment.ParseTime()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Kedaluwarsa pada: %s\n", expiry)
```

## Penyangkalan

Ini adalah SDK **tidak resmi**. SDK ini tidak berafiliasi, didukung, atau dikelola secara resmi oleh Pakasir. SDK ini tidak resmi karena API resmi hanya menyediakan dokumentasi dan dukungan untuk REST API dan SDK Node.js mereka. Library ini dibuat untuk menambahkan dukungan Go yang layak, dan digunakan secara aktif oleh pemilik repositori ini. Gunakan dengan risiko Anda sendiri.

## Lisensi

Proyek ini dilisensikan di bawah Lisensi Apache 2.0 — lihat berkas [LICENSE](LICENSE) untuk detail lengkap.
