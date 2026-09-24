package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/shipt/pdf-generation-playground.git/htmlpdf"
	"github.com/shipt/pdf-generation-playground.git/pdf"
)

func samplePaymentsReport() pdf.PaymentsReport {
	return pdf.PaymentsReport{
		Shopper: pdf.ShopperInfo{Name: "Kathryn Lester", ID: "234629"},
		GrandTotal: pdf.PeriodTotals{
			Start: "2026-06-29", End: "2026-07-26",
			GrossEarnings: "$46.64", Deductions: "0", NetEarnings: "$46.64",
		},
		Periods: []pdf.PaymentPeriod{
			{
				Totals: pdf.PeriodTotals{
					Start: "2026-07-20", End: "2026-07-26",
					GrossEarnings: "$13.11", Deductions: "0", NetEarnings: "$13.11",
				},
				PieceRateEarnings: []pdf.PieceRateEntry{
					{Date: "2026-07-23", ReferenceID: "67152731-d843-4ebb-8808-89636748feab", PieceRate: "1 Bundle at $8.11", Earnings: "$8.11"},
					{Date: "2026-07-23", ReferenceID: "234629", PieceRate: "1 Metro lead at $5.00", Earnings: "$5.00"},
				},
			},
			{
				Totals: pdf.PeriodTotals{
					Start: "2026-07-13", End: "2026-07-19",
					GrossEarnings: "$28.53", Deductions: "0", NetEarnings: "$28.53",
				},
				PieceRateEarnings: []pdf.PieceRateEntry{
					{Date: "2026-07-13", ReferenceID: "eb561dea-40e9-402b-bcb8-c51024657f38", PieceRate: "1 Bundle at $6.00", Earnings: "$6.00"},
					{Date: "2026-07-13", ReferenceID: "8096e182-b8c9-4565-b875-cb8e7fbfa514", PieceRate: "1 Bundle at $6.00", Earnings: "$6.00"},
					{Date: "2026-07-13", ReferenceID: "089df207-397c-48ad-a9eb-2ae533d7abcf", PieceRate: "1 Bundle at $8.94", Earnings: "$8.94"},
					{Date: "2026-07-13", ReferenceID: "1f068214-50fd-48c5-abfb-7d3805bc987d", PieceRate: "1 Bundle at $7.59", Earnings: "$7.59"},
				},
			},
			{
				Totals: pdf.PeriodTotals{
					Start: "2026-06-29", End: "2026-07-05",
					GrossEarnings: "$5.00", Deductions: "0", NetEarnings: "$5.00",
				},
				PieceRateEarnings: []pdf.PieceRateEntry{
					{Date: "2026-07-02", ReferenceID: "183082474", PieceRate: "1 Cancelled order at $5.00", Earnings: "$5.00"},
				},
			},
		},
	}
}

func sampleSummaryReport() pdf.SummaryReport {
	return pdf.SummaryReport{
		Shopper: pdf.ShopperInfo{Name: "Zach Serre", ID: "100597155"},
		Heading: "2025 total earnings",
		Intro: "2025 total earnings are based on payouts received during calendar year 2025 (pay " +
			"periods:12/30/24-12/28/25). Total earnings include order pay, special pay, tips, bonuses, and " +
			"milestones. You will receive a 1099 if your total earnings exceeded $600 in 2025.",
		Lines: []pdf.SummaryLine{
			{Label: "Total Non-tip Earnings:", Value: "$217.56"},
			{Label: "12/29/25-12/31/25 instant payouts*:", Value: "$0.00"},
			{Label: "12/30/24-12/31/24 instant payouts**:", Value: "$0.00"},
			{Label: "Total Tips:", Value: "$115.53"},
			{Label: "Total Earnings:", Value: "$333.09"},
		},
		Footnotes: []string{
			"*2025 total earnings include instant payouts withdrawn from 12/29/25-12/31/25.",
			"**2025 total earnings exclude instant payouts withdrawn from 12/30/24-12/31/24, as they were " +
				"included in your 2024 total earnings. Learn More",
		},
	}
}

func toHTMLPaymentsReport(r pdf.PaymentsReport) htmlpdf.PaymentsReport {
	periods := make([]htmlpdf.PaymentPeriod, len(r.Periods))
	for i, period := range r.Periods {
		entries := make([]htmlpdf.PieceRateEntry, len(period.PieceRateEarnings))
		for j, e := range period.PieceRateEarnings {
			entries[j] = htmlpdf.PieceRateEntry(e)
		}
		periods[i] = htmlpdf.PaymentPeriod{
			Totals:            htmlpdf.PeriodTotals(period.Totals),
			PieceRateEarnings: entries,
		}
	}
	return htmlpdf.PaymentsReport{
		Shopper:    htmlpdf.ShopperInfo(r.Shopper),
		GrandTotal: htmlpdf.PeriodTotals(r.GrandTotal),
		Periods:    periods,
	}
}

func toHTMLSummaryReport(r pdf.SummaryReport) htmlpdf.SummaryReport {
	lines := make([]htmlpdf.SummaryLine, len(r.Lines))
	for i, l := range r.Lines {
		lines[i] = htmlpdf.SummaryLine(l)
	}
	return htmlpdf.SummaryReport{
		Shopper:   htmlpdf.ShopperInfo(r.Shopper),
		Heading:   r.Heading,
		Intro:     r.Intro,
		Lines:     lines,
		Footnotes: r.Footnotes,
	}
}

func samplePayBreakdownReport() pdf.PayBreakdownReport {
	return pdf.PayBreakdownReport{
		ShopperID:   "102727067",
		PeriodStart: "07/13/26",
		PeriodEnd:   "07/19/26",
		PayPeriods:  "1 pay period",
		Payday:      "07/24/26",
		Totals: pdf.HeadlineTotals{
			TotalPay:            "$259.88",
			TotalPayWithoutTips: "$208.50",
			TotalTips:           "$51.38",
		},
		Summary: []pdf.SummaryRow{
			{Description: "Order Pay – Bundle & Order Pay", Current: "$102.50", YearToDate: "$102.50"},
			{Description: "Incentive pay", Current: "$10.00", YearToDate: "$10.00"},
			{Description: "Referral pay", Current: "$25.00", YearToDate: "$25.00"},
			{Description: "Shopper Special Pay", Current: "$59.00", YearToDate: "$59.00"},
			{Description: "Compliance pay", Current: "$12.00", YearToDate: "$12.00"},
			{Description: "Total Non-tip Pay", Current: "$208.50", YearToDate: "$208.50", Bold: true},
			{Description: "Tips", Current: "$51.38", YearToDate: "$51.38"},
			{Description: "Total pay", Current: "$259.88", YearToDate: "$259.88", Bold: true},
		},
		DeductionsNote: "Deductions: None – paid as a 1099 independent contractor; no tax withholding applies.",
		Payouts: []pdf.PayoutRow{
			{Date: "07/19/26", Method: "standard", Account: "•••• 9400", Amount: "$150.00"},
			{Date: "07/15/26", Method: "instant", Account: "•••• 9400", Amount: "$80.00", Fee: "-$0.49 fee"},
			{Date: "07/17/26", Method: "instant", Account: "•••• 1234", Amount: "$49.73", Fee: "-$0.49 fee"},
			{Date: "07/14/26", Method: "instant", Account: "•••• 5678", Amount: "$8.00", Fee: "-$0.49 fee"},
		},
		Bundles: []pdf.BundleRow{
			{Date: "07/21/26", Reference: "Bundle: e9b5b07a…", BasePay: "$25.00", Tips: "$14.63", Total: "$39.63", Orders: []pdf.BundleOrderLine{
				{OrderID: "280033103", TipDelta: "+$8.13"},
				{OrderID: "280033104", TipDelta: "+$6.50"},
			}},
			{Date: "07/22/26", Reference: "Order: #280044200", BasePay: "$18.00", Tips: "—", Total: "$18.00"},
			{Date: "07/21/26", Reference: "Bundle: c1d2e3f4…", BasePay: "$30.00", Tips: "$16.25", Total: "$46.25", Orders: []pdf.BundleOrderLine{
				{OrderID: "280055301", TipDelta: "+$9.25"},
				{OrderID: "280055302", TipDelta: "+$4.00"},
				{OrderID: "280055303", TipDelta: "+$3.00"},
			}},
			{Date: "07/23/26", Reference: "Order: #280066400", BasePay: "$12.00", Tips: "—", Total: "$12.00"},
			{Date: "07/21/26", Reference: "Order: #279443100", Tag: "Returned", BasePay: "$5.50", Tips: "—", Total: "$5.50"},
			{Date: "07/21/26", Reference: "Order: #279551200", Tag: "Cancelled", BasePay: "$4.00", Tips: "—", Total: "$4.00"},
			{Date: "07/22/26", Reference: "Order: #279662300", Tag: "Cancelled", BasePay: "$3.00", Tips: "—", Total: "$3.00"},
			{Date: "07/24/26", Reference: "Order: #279773400", Tag: "Cancelled", BasePay: "$5.00", Tips: "—", Total: "$5.00"},
			{Date: "07/21/26", Reference: "Order: #279900001", Tag: "Late tip", BasePay: "—", Tips: "$5.00", Total: "$5.00"},
			{Date: "07/22/26", Reference: "Order: #279900042", Tag: "Late tip", BasePay: "—", Tips: "$12.00", Total: "$12.00"},
			{Date: "07/21/26", Reference: "Order: #279900087", Tag: "Late tip", BasePay: "—", Tips: "$3.50", Total: "$3.50"},
			{Reference: "Total", BasePay: "$102.50", Tips: "$51.38", Total: "$153.88", IsTotal: true},
		},
		OtherPay: []pdf.OtherPayRow{
			{Date: "07/23/26", Category: "Incentive Pay", TypeNotes: "—", Amount: "$10.00"},
			{Date: "07/21/26", Category: "Compliance", TypeNotes: "Guaranteed earnings adjustment", Amount: "$12.00"},
			{Date: "07/21/26", Category: "Special Pay", TypeNotes: "Shopper bonus", Amount: "$20.00"},
			{Date: "07/21/26", Category: "Special Pay", TypeNotes: "Metro lead", Amount: "$9.00"},
			{Date: "07/21/26", Category: "Special Pay", TypeNotes: "Milestones", Amount: "$12.00"},
			{Date: "07/21/26", Category: "Special Pay", TypeNotes: "Marketing", Amount: "$7.00"},
			{Date: "07/21/26", Category: "Special Pay", TypeNotes: "Supplemental order pay", Amount: "$11.00"},
			{Date: "07/23/26", Category: "Referral / Recruitment", TypeNotes: "—", Amount: "$25.00"},
			{Category: "Total", Amount: "$106.00", IsTotal: true},
		},
	}
}

type PDFSampleCmd struct {
	OutputDir string `short:"o" long:"outputDir" description:"Path to the output directory" default:"samples/pdf"`
}

func (s *PDFSampleCmd) Execute(args []string) error {
	if err := pdf.NewGenerator(10, 10, 10).GeneratePayments(s.OutputDir, "assets/logo.png", "payments.pdf", samplePaymentsReport()); err != nil {
		return err
	}
	if err := pdf.NewGenerator(10, 10, 10).GenerateSummary(s.OutputDir, "assets/logo.png", "summary.pdf", sampleSummaryReport()); err != nil {
		return err
	}
	if err := pdf.NewGenerator(10, 10, 10).GeneratePayBreakdown(s.OutputDir, "assets/logo.png", "pay_breakdown.pdf", samplePayBreakdownReport()); err != nil {
		return err
	}
	fmt.Println("Sample PDFs generated successfully at:", s.OutputDir)
	return nil
}

type FromHTMLCmd struct {
	OutputDir string `short:"o" long:"outputDir" description:"Path to the output directory" default:"samples/htmlpdf"`
	DockerEnv bool   `short:"d" long:"dockerEnv" description:"Run in Docker environment"`
}

func (s *FromHTMLCmd) Execute(args []string) error {
	ctx := context.Background()
	pdfGenerator := htmlpdf.NewGenerator(s.DockerEnv)
	if err := pdfGenerator.GeneratePayments(ctx, s.OutputDir, "assets/logo.png", "payments.pdf", toHTMLPaymentsReport(samplePaymentsReport())); err != nil {
		return err
	}
	if err := pdfGenerator.GenerateSummary(ctx, s.OutputDir, "assets/logo.png", "summary.pdf", toHTMLSummaryReport(sampleSummaryReport())); err != nil {
		return err
	}
	fmt.Println("PDFs generated successfully from HTML at:", s.OutputDir)
	return nil
}

type options struct {
	Pdf      PDFSampleCmd `command:"sample" description:"Generate sample PDFs"`
	FromHTML FromHTMLCmd  `command:"fromHTML" description:"Generate PDFs from HTML"`
}

func main() {
	var opts options
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				fmt.Println(err)
				os.Exit(0)
			}
			fmt.Println(err)
			os.Exit(1)
		default:
			fmt.Println("Error parsing flags:", err)
			os.Exit(1)
		}
	}
}
