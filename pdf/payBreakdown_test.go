package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func samplePayBreakdown() PayBreakdownReport {
	return PayBreakdownReport{
		ShopperID:   "102727067",
		PeriodStart: "07/13/26",
		PeriodEnd:   "07/19/26",
		PayPeriods:  "1 pay period",
		Payday:      "07/24/26",
		Totals: HeadlineTotals{
			TotalPay: "$259.88", TotalPayWithoutTips: "$208.50", TotalTips: "$51.38",
		},
		Summary: []SummaryRow{
			{Description: "Order Pay – Bundle & Order Pay", Current: "$102.50", YearToDate: "$102.50"},
			{Description: "Total pay", Current: "$259.88", YearToDate: "$259.88", Bold: true},
		},
		DeductionsNote: "Deductions: None – paid as a 1099 independent contractor.",
		Payouts: []PayoutRow{
			{Date: "07/19/26", Method: "standard", Account: "•••• 9400", Amount: "$150.00"},
			{Date: "07/15/26", Method: "instant", Account: "•••• 9400", Amount: "$80.00", Fee: "-$0.49 fee"},
		},
		Bundles: []BundleRow{
			{Date: "07/21/26", Reference: "Bundle: e9b5b07a…", BasePay: "$25.00", Tips: "$14.63", Total: "$39.63", Orders: []BundleOrderLine{
				{OrderID: "280033103", TipDelta: "+$8.13"},
			}},
			{Reference: "Total", BasePay: "$102.50", Tips: "$51.38", Total: "$153.88", IsTotal: true},
		},
		OtherPay: []OtherPayRow{
			{Date: "07/23/26", Category: "Incentive Pay", TypeNotes: "—", Amount: "$10.00"},
			{Category: "Total", Amount: "$106.00", IsTotal: true},
		},
	}
}

func TestGeneratePayBreakdown(t *testing.T) {
	dir := t.TempDir()

	err := NewGenerator(10, 10, 10).GeneratePayBreakdown(dir, "../assets/logo.png", "pay_breakdown.pdf", samplePayBreakdown())
	if err != nil {
		t.Fatalf("GeneratePayBreakdown returned error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "pay_breakdown.pdf"))
	if err != nil {
		t.Fatalf("failed to read generated PDF: %v", err)
	}
	if !bytes.HasPrefix(got, []byte("%PDF")) {
		t.Fatalf("expected PDF magic bytes, got %q", got[:min(4, len(got))])
	}
}
