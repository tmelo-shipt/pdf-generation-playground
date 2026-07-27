package htmlpdf

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func benchmarkPaymentsReport() PaymentsReport {
	return PaymentsReport{
		Shopper: ShopperInfo{Name: "Kathryn Lester", ID: "234629"},
		GrandTotal: PeriodTotals{
			Start: "2026-06-29", End: "2026-07-26",
			GrossEarnings: "$46.64", Deductions: "0", NetEarnings: "$46.64",
		},
		Periods: []PaymentPeriod{
			{
				Totals: PeriodTotals{
					Start: "2026-07-20", End: "2026-07-26",
					GrossEarnings: "$13.11", Deductions: "0", NetEarnings: "$13.11",
				},
				PieceRateEarnings: []PieceRateEntry{
					{Date: "2026-07-23", ReferenceID: "67152731-d843-4ebb-8808-89636748feab", PieceRate: "1 Bundle at $8.11", Earnings: "$8.11"},
					{Date: "2026-07-23", ReferenceID: "234629", PieceRate: "1 Metro lead at $5.00", Earnings: "$5.00"},
				},
			},
		},
	}
}

func BenchmarkHTMLToPDFGenerator_Parallel4(b *testing.B) {
	// we’ll ignore b.N and just do 10 PDFs for timing analysis.
	const total = 10

	outputDir := "benchmark/pdf_parallel"
	_ = os.MkdirAll(outputDir, 0o755)

	// allocate Chrome only once (avoid cold start per PDF).
	// we'll need a similar strategy in production to avoid
	// the overhead of starting Chrome for each PDF.
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	defer cancelAlloc()

	parentCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	pdfGenerator := NewGenerator(false)

	concurrency := 4
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	errs := make([]error, total)

	b.ResetTimer()
	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			start := time.Now()
			tabCtx, tabCancel := chromedp.NewContext(parentCtx)
			defer tabCancel()

			filename := fmt.Sprintf("payments_parallel_%d.pdf", i)
			if err := pdfGenerator.GeneratePayments(
				tabCtx,
				outputDir,
				"../assets/logo.png",
				filename,
				benchmarkPaymentsReport(),
			); err != nil {
				errs[i] = fmt.Errorf("Failed to generate PDF: %v", err)
				return
			}

			b.Logf("PDF %d generated in %v", i, time.Since(start))
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			b.Fatalf("PDF %d: %v", i, err)
		}
	}
}

func BenchmarkHTMLToPDFGenerator_Sequential10(b *testing.B) {
	outputDir := "benchmark/pdf"
	_ = os.MkdirAll(outputDir, 0o755)

	// allocate Chrome only once (avoid cold start per PDF).
	// we'll need a similar strategy in production to avoid
	// the overhead of starting Chrome for each PDF.
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	pdfGenerator := NewGenerator(false)

	// we’ll ignore b.N and just do 10 PDFs for timing analysis.
	numPDFs := 10
	b.ResetTimer()

	for i := 0; i < numPDFs; i++ {
		start := time.Now()
		filename := fmt.Sprintf("payments_%d.pdf", i)
		if err := pdfGenerator.GeneratePayments(ctx, outputDir, "../assets/logo.png", filename, benchmarkPaymentsReport()); err != nil {
			b.Fatalf("Failed to generate PDF: %v", err)
		}
		elapsed := time.Since(start)
		b.Logf("PDF %d generated in %v", i, elapsed)
	}
}
