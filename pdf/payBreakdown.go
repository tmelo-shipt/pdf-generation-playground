package pdf

import (
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/pkg/errors"
	"os"
	"strings"
)

// HeadlineTotals are the three big figures shown at the top of the breakdown:
// total pay, total pay excluding tips, and total tips.
type HeadlineTotals struct {
	TotalPay            string
	TotalPayWithoutTips string
	TotalTips           string
}

// SummaryRow is a single line of the "Current / Year-to-date" summary table.
type SummaryRow struct {
	Description string
	Current     string
	YearToDate  string
	Bold        bool
}

// PayoutRow is one payout in the payouts table.
type PayoutRow struct {
	Date    string
	Method  string
	Account string
	Amount  string
	Fee     string
}

// BundleOrderLine is a nested order line under a bundle (order id + tip delta).
type BundleOrderLine struct {
	OrderID  string
	TipDelta string
}

// BundleRow is one line in the "Bundle and Order Pay" table. When IsTotal is
// set it renders the bold totals row.
type BundleRow struct {
	Date      string
	Reference string
	Tag       string
	BasePay   string
	Tips      string
	Total     string
	Orders    []BundleOrderLine
	IsTotal   bool
}

// OtherPayRow is one line in the "Other Pay" table (incentive, compliance,
// special pay, referral). When IsTotal is set it renders the bold totals row.
type OtherPayRow struct {
	Date      string
	Category  string
	TypeNotes string
	Amount    string
	IsTotal   bool
}

// PayBreakdownReport is the data behind the modern per-category pay stub
// breakdown (the direction shown in Zach's Lovable prototype), built from the
// payroll-report PayStubWithDetails response.
type PayBreakdownReport struct {
	ShopperID      string
	PeriodStart    string
	PeriodEnd      string
	PayPeriods     string
	Payday         string
	Totals         HeadlineTotals
	Summary        []SummaryRow
	DeductionsNote string
	Payouts        []PayoutRow
	Bundles        []BundleRow
	OtherPay       []OtherPayRow
}

var (
	mutedColor   = &props.Color{Red: 110, Green: 110, Blue: 110}
	accentColor  = &props.Color{Red: 24, Green: 24, Blue: 24}
	tipColor     = &props.Color{Red: 5, Green: 150, Blue: 105}  // #059669 order tip delta
	orderIDColor = &props.Color{Red: 99, Green: 102, Blue: 241} // #6366f1 order id

	// Status badge palette, mirroring the Lovable mockup badges
	// (background + foreground per status tag).
	badgeReturnedBG  = &props.Color{Red: 255, Green: 237, Blue: 213} // #ffedd5
	badgeReturnedFG  = &props.Color{Red: 194, Green: 65, Blue: 12}   // #c2410c
	badgeCancelledBG = &props.Color{Red: 254, Green: 226, Blue: 226} // #fee2e2
	badgeCancelledFG = &props.Color{Red: 185, Green: 28, Blue: 28}   // #b91c1c
	badgeLateTipBG   = &props.Color{Red: 224, Green: 231, Blue: 255} // #e0e7ff
	badgeLateTipFG   = &props.Color{Red: 67, Green: 56, Blue: 202}   // #4338ca
	badgeDefaultBG   = &props.Color{Red: 241, Green: 245, Blue: 249} // #f1f5f9
	badgeDefaultFG   = &props.Color{Red: 71, Green: 85, Blue: 105}   // #475569
)

// tagColors maps a bundle status tag to its badge background and text colors.
func tagColors(tag string) (bg, fg *props.Color) {
	switch strings.ToLower(strings.TrimSpace(tag)) {
	case "returned":
		return badgeReturnedBG, badgeReturnedFG
	case "cancelled", "canceled":
		return badgeCancelledBG, badgeCancelledFG
	case "late tip", "late-tip", "latetip":
		return badgeLateTipBG, badgeLateTipFG
	default:
		return badgeDefaultBG, badgeDefaultFG
	}
}

// GeneratePayBreakdown renders the per-category pay stub breakdown statement.
func (p *pdfGenerator) GeneratePayBreakdown(outputDir, logoImagePath, filename string, report PayBreakdownReport) error {
	headerRows, err := p.breakdownHeader(logoImagePath, report)
	if err != nil {
		return err
	}
	p.maroto.AddRows(headerRows...)

	p.maroto.AddRows(p.headlineTotals(report.Totals)...)

	p.maroto.AddRows(p.summaryTable(report.Summary)...)

	p.maroto.AddRows(row.New(6).Add(
		col.New(12).Add(text.New(report.DeductionsNote, props.Text{Size: 8, Style: fontstyle.Italic, Color: mutedColor})),
	))

	p.maroto.AddRows(row.New(4))
	p.maroto.AddRows(p.periodHeading(report)...)

	p.maroto.AddRows(p.payoutsTable(report.Payouts)...)
	p.maroto.AddRows(p.bundlesTable(report.Bundles)...)
	p.maroto.AddRows(p.otherPayTable(report.OtherPay)...)

	return p.save(outputDir, filename)
}

func (p *pdfGenerator) breakdownHeader(logoImagePath string, report PayBreakdownReport) ([]core.Row, error) {
	logo, err := os.ReadFile(logoImagePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read logo file")
	}

	return []core.Row{
		row.New(14).Add(
			col.New(3).Add(image.NewFromBytes(logo, extension.Png, props.Rect{Percent: 100})),
			col.New(9).Add(text.New("Summary  ·  Shopper #"+report.ShopperID,
				props.Text{Size: 13, Style: fontstyle.Bold, Align: align.Right, Top: 4})),
		),
		row.New(6).Add(
			col.New(12).Add(text.New(report.PeriodStart+" – "+report.PeriodEnd+"  ·  "+report.PayPeriods,
				props.Text{Size: 9, Align: align.Right, Color: mutedColor})),
		),
		row.New(4),
	}, nil
}

func (p *pdfGenerator) headlineTotals(t HeadlineTotals) []core.Row {
	cell := func(value, label string) core.Col {
		return col.New(4).Add(
			text.New(value, props.Text{Size: 16, Style: fontstyle.Bold, Align: align.Center, Color: accentColor, Top: 5}),
			text.New(label, props.Text{Size: 8, Align: align.Center, Color: mutedColor, Top: 14}),
		)
	}

	return []core.Row{
		row.New(23).WithStyle(&props.Cell{BorderType: border.Full, BorderColor: borderColor, BackgroundColor: shadedColor}).Add(
			cell(t.TotalPay, "Total pay"),
			cell(t.TotalPayWithoutTips, "Total pay without tips"),
			cell(t.TotalTips, "Total tips"),
		),
		row.New(6),
	}
}

func (p *pdfGenerator) summaryTable(rows []SummaryRow) []core.Row {
	out := []core.Row{
		p.tableHeaderRow([]bcell{
			{text: "Description", span: 6},
			{text: "Current", span: 3, align: align.Right},
			{text: "Year-to-date", span: 3, align: align.Right},
		}),
	}
	for _, r := range rows {
		out = append(out, p.tableDataRow(9, []bcell{
			{text: r.Description, span: 6, bold: r.Bold},
			{text: r.Current, span: 3, align: align.Right, bold: r.Bold},
			{text: r.YearToDate, span: 3, align: align.Right, bold: r.Bold},
		}))
	}
	out = append(out, row.New(6))
	return out
}

func (p *pdfGenerator) periodHeading(report PayBreakdownReport) []core.Row {
	return []core.Row{
		row.New(9).WithStyle(&props.Cell{BorderType: border.Full, BorderColor: borderColor, BackgroundColor: shadedColor}).Add(
			col.New(9).Add(text.New(report.PeriodStart+" – "+report.PeriodEnd+"   Payday "+report.Payday,
				props.Text{Size: 9, Style: fontstyle.Bold, Left: 2, Top: 2})),
			col.New(3).Add(text.New(report.Totals.TotalPay,
				props.Text{Size: 9, Style: fontstyle.Bold, Align: align.Right, Right: 2, Top: 2})),
		),
		row.New(4),
	}
}

func (p *pdfGenerator) payoutsTable(payouts []PayoutRow) []core.Row {
	out := []core.Row{
		p.sectionTitle("Payouts (" + itoa(len(payouts)) + ")"),
		p.tableHeaderRow([]bcell{
			{text: "Date", span: 3},
			{text: "Method", span: 3},
			{text: "Account", span: 3},
			{text: "Amount", span: 3, align: align.Right},
		}),
	}
	for _, r := range payouts {
		amount := r.Amount
		if r.Fee != "" {
			amount = r.Amount + "  " + r.Fee
		}
		out = append(out, p.tableDataRow(8, []bcell{
			{text: r.Date, span: 3},
			{text: r.Method, span: 3},
			{text: r.Account, span: 3},
			{text: amount, span: 3, align: align.Right},
		}))
	}
	out = append(out, row.New(6))
	return out
}

func (p *pdfGenerator) bundlesTable(bundles []BundleRow) []core.Row {
	out := []core.Row{
		p.sectionTitle("Bundle and Order Pay"),
		row.New(5).Add(col.New(12).Add(text.New(
			"Base pay includes base, promo, extra, canceled, returned, and returns adjustment.",
			props.Text{Size: 7, Style: fontstyle.Italic, Color: mutedColor}))),
		p.tableHeaderRow([]bcell{
			{text: "Date", span: 2},
			{text: "ID / Reference", span: 5},
			{text: "Base Pay", span: 2, align: align.Right},
			{text: "Tips", span: 1, align: align.Right},
			{text: "Total", span: 2, align: align.Right},
		}),
	}

	for _, b := range bundles {
		out = append(out, p.bundleRow(b))

		for _, o := range b.Orders {
			out = append(out, row.New(6).Add(
				col.New(2),
				col.New(5).Add(text.New("Order #"+o.OrderID, props.Text{Size: 7, Left: 4, Color: orderIDColor})),
				col.New(2),
				col.New(1).Add(text.New(o.TipDelta, props.Text{Size: 7, Align: align.Right, Color: tipColor})),
				col.New(2),
			))
		}
	}
	out = append(out, row.New(6))
	return out
}

// bundleRow renders one bundle line. When the bundle carries a status tag
// (Returned/Cancelled/Late tip) the tag is drawn as a colored badge column,
// mirroring the Lovable mockup; otherwise the reference spans the full column.
// The bottom border is applied per column (rather than on the row) so the tag
// column can paint its own background — maroto only draws per-column cells when
// the row itself has no style.
func (p *pdfGenerator) bundleRow(b BundleRow) core.Row {
	style := fontstyle.Normal
	if b.IsTotal {
		style = fontstyle.Bold
	}
	rule := func() *props.Cell {
		return &props.Cell{BorderType: border.Bottom, BorderColor: borderColor}
	}

	cols := []core.Col{
		col.New(2).WithStyle(rule()).Add(text.New(b.Date, props.Text{Size: 8, Left: 2, Top: 1.5})),
	}
	if b.Tag == "" {
		cols = append(cols, col.New(5).WithStyle(rule()).Add(
			text.New(b.Reference, props.Text{Size: 8, Style: style, Left: 2, Top: 1.5})))
	} else {
		bg, fg := tagColors(b.Tag)
		badge := rule()
		badge.BackgroundColor = bg
		cols = append(cols,
			col.New(3).WithStyle(rule()).Add(
				text.New(b.Reference, props.Text{Size: 8, Style: style, Left: 2, Top: 1.5})),
			col.New(1).WithStyle(badge).Add(
				text.New(b.Tag, props.Text{Size: 6.5, Style: fontstyle.Bold, Align: align.Center, Color: fg, Top: 2.2})),
			col.New(1).WithStyle(rule()),
		)
	}
	cols = append(cols,
		col.New(2).WithStyle(rule()).Add(text.New(b.BasePay, props.Text{Size: 8, Align: align.Right, Right: 2, Top: 1.5, Style: style})),
		col.New(1).WithStyle(rule()).Add(text.New(b.Tips, props.Text{Size: 8, Align: align.Right, Right: 2, Top: 1.5, Style: style})),
		col.New(2).WithStyle(rule()).Add(text.New(b.Total, props.Text{Size: 8, Align: align.Right, Right: 2, Top: 1.5, Style: style})),
	)

	return row.New(8).Add(cols...)
}

func (p *pdfGenerator) otherPayTable(rows []OtherPayRow) []core.Row {
	out := []core.Row{
		p.sectionTitle("Other Pay"),
		p.tableHeaderRow([]bcell{
			{text: "Date", span: 2},
			{text: "Category", span: 3},
			{text: "Type / Notes", span: 5},
			{text: "Amount", span: 2, align: align.Right},
		}),
	}
	for _, r := range rows {
		out = append(out, p.tableDataRow(8, []bcell{
			{text: r.Date, span: 2, bold: r.IsTotal},
			{text: r.Category, span: 3, bold: r.IsTotal},
			{text: r.TypeNotes, span: 5, bold: r.IsTotal},
			{text: r.Amount, span: 2, align: align.Right, bold: r.IsTotal},
		}))
	}
	out = append(out, row.New(6))
	return out
}

// bcell is a single cell in a breakdown table row.
type bcell struct {
	text  string
	span  int
	align align.Type
	bold  bool
}

func (p *pdfGenerator) sectionTitle(title string) core.Row {
	return row.New(9).Add(col.New(12).Add(text.New(title, props.Text{Size: 11, Style: fontstyle.Bold, Top: 2})))
}

func (p *pdfGenerator) tableHeaderRow(cells []bcell) core.Row {
	return p.breakdownRow(7, border.Full, shadedColor, true, cells)
}

func (p *pdfGenerator) tableDataRow(height float64, cells []bcell) core.Row {
	return p.breakdownRow(height, border.Bottom, nil, false, cells)
}

func (p *pdfGenerator) breakdownRow(height float64, bd border.Type, bg *props.Color, header bool, cells []bcell) core.Row {
	cols := make([]core.Col, 0, len(cells))
	for _, c := range cells {
		style := fontstyle.Normal
		if c.bold || header {
			style = fontstyle.Bold
		}
		cols = append(cols, col.New(c.span).Add(
			text.New(c.text, props.Text{Size: 8, Align: c.align, Style: style, Left: 2, Right: 2, Top: 1.5}),
		))
	}
	return row.New(height).WithStyle(&props.Cell{
		BorderType:      bd,
		BorderColor:     borderColor,
		BackgroundColor: bg,
	}).Add(cols...)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
