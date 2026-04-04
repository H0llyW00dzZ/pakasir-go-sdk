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

//go:build ignore
// +build ignore

// This file provides a usage example for the Pakasir Go SDK.
// Run it with: go run examples/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/client"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/constants"
	urlhelper "github.com/H0llyW00dzZ/pakasir-go-sdk/src/helper/url"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/i18n"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/simulation"
	"github.com/H0llyW00dzZ/pakasir-go-sdk/src/transaction"
)

func main() {
	// 1. Initialize the client with functional options.
	c := client.New("depodomain", "xxx123",
		client.WithTimeout(15*time.Second),
		client.WithLanguage(i18n.Indonesian),
		client.WithRetries(2),
	)

	ctx := context.Background()

	// 2. Create a QRIS transaction.
	txnService := transaction.NewService(c)
	createResp, err := txnService.Create(ctx, constants.MethodQRIS, &transaction.CreateRequest{
		OrderID: "INV123123",
		Amount:  99000,
	})
	if err != nil {
		log.Fatalf("Failed to create transaction: %v", err)
	}

	fmt.Println("=== Transaction Created ===")
	fmt.Printf("Order ID:       %s\n", createResp.Payment.OrderID)
	fmt.Printf("Amount:         %d\n", createResp.Payment.Amount)
	fmt.Printf("Fee:            %d\n", createResp.Payment.Fee)
	fmt.Printf("Total Payment:  %d\n", createResp.Payment.TotalPayment)
	fmt.Printf("Payment Method: %s\n", createResp.Payment.PaymentMethod)
	fmt.Printf("Payment Number: %s\n", createResp.Payment.PaymentNumber)
	fmt.Printf("Expired At:     %s\n", createResp.Payment.ExpiredAt)

	// 3. Simulate payment (sandbox only).
	simService := simulation.NewService(c)
	if err := simService.Pay(ctx, &simulation.PayRequest{
		OrderID: "INV123123",
		Amount:  99000,
	}); err != nil {
		log.Fatalf("Failed to simulate payment: %v", err)
	}
	fmt.Println("\n=== Payment Simulated ===")

	// 4. Check transaction detail.
	detail, err := txnService.Detail(ctx, &transaction.DetailRequest{
		OrderID: "INV123123",
		Amount:  99000,
	})
	if err != nil {
		log.Fatalf("Failed to get detail: %v", err)
	}

	fmt.Println("\n=== Transaction Detail ===")
	fmt.Printf("Status:         %s\n", detail.Transaction.Status)
	fmt.Printf("Payment Method: %s\n", detail.Transaction.PaymentMethod)
	fmt.Printf("Completed At:   %s\n", detail.Transaction.CompletedAt)

	// 5. Build a payment redirect URL.
	payURL, err := urlhelper.Build(client.DefaultBaseURL, "depodomain", 22000, urlhelper.Options{
		OrderID:  "240910HDE7C9",
		QRISOnly: true,
	})
	if err != nil {
		log.Fatalf("Failed to build URL: %v", err)
	}

	fmt.Println("\n=== Payment URL ===")
	fmt.Println(payURL)
}
