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
