package pdf

import (
	"fmt"
	"testing"
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

func BenchmarkTestPDFGenerator(b *testing.B) {
	report := benchmarkPaymentsReport()
	for i := 0; i < b.N; i++ {
		pdfGenerator := NewGenerator(10, 10, 10)
		filename := fmt.Sprintf("payments_%d.pdf", i)
		err := pdfGenerator.GeneratePayments("benchmark/pdf", "../assets/logo.png", filename, report)
		if err != nil {
			b.Fatalf("Failed to generate PDF: %v", err)
		}
	}
}
