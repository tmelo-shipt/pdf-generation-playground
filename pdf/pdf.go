package pdf

import (
	"os"
	"path/filepath"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/border"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/pkg/errors"
)

// ShopperInfo identifies the Shipt shopper a report belongs to.
type ShopperInfo struct {
	Name string
	ID   string
}

// PeriodTotals summarizes earnings for a single pay period (or the grand total across periods).
type PeriodTotals struct {
	Start         string
	End           string
	GrossEarnings string
	Deductions    string
	NetEarnings   string
}

// PieceRateEntry is a single piece-rate line item within a pay period.
type PieceRateEntry struct {
	Date        string
	ReferenceID string
	PieceRate   string
	Earnings    string
}

// PaymentPeriod groups a pay period's totals with its piece-rate line items.
type PaymentPeriod struct {
	Totals            PeriodTotals
	PieceRateEarnings []PieceRateEntry
}

// PaymentsReport is the data behind the multi-period "payments.pdf" statement.
type PaymentsReport struct {
	Shopper    ShopperInfo
	GrandTotal PeriodTotals
	Periods    []PaymentPeriod
}

// SummaryLine is a single label/value row in the summary earnings table.
type SummaryLine struct {
	Label string
	Value string
}

// SummaryReport is the data behind the single-page "summary.pdf" annual earnings summary.
type SummaryReport struct {
	Shopper   ShopperInfo
	Heading   string
	Intro     string
	Lines     []SummaryLine
	Footnotes []string
}

var (
	borderColor = &props.Color{Red: 224, Green: 224, Blue: 224}
	shadedColor = &props.Color{Red: 242, Green: 242, Blue: 242}
)

type pdfGenerator struct {
	maroto core.Maroto
}

func NewGenerator(left, top, right float64) *pdfGenerator {
	cfg := config.NewBuilder().
		WithLeftMargin(left).
		WithTopMargin(top).
		WithRightMargin(right).
		Build()

	mrt := maroto.New(cfg)
	m := maroto.NewMetricsDecorator(mrt)
	return &pdfGenerator{maroto: m}
}

// GeneratePayments renders the multi pay-period Shipt payments statement.
func (p *pdfGenerator) GeneratePayments(outputDir, logoImagePath, filename string, report PaymentsReport) error {
	companyRows, err := p.companyHeaderRows(logoImagePath, []string{
		"Shipt, Inc",
		"420 20th St. N. Suite 100",
		"Birmingham, AL 35203",
	})
	if err != nil {
		return err
	}
	p.maroto.AddRows(companyRows...)

	p.maroto.AddRows(p.shopperInfoRows(report.Shopper)...)

	p.maroto.AddRows(p.sectionHeading(
		"Grand Totals for " + report.GrandTotal.Start + " - " + report.GrandTotal.End,
	))
	p.maroto.AddRows(p.totalsRows(report.GrandTotal)...)

	for _, period := range report.Periods {
		p.maroto.AddRows(p.sectionHeading(
			"Payments for for " + period.Totals.Start + " - " + period.Totals.End,
		))
		p.maroto.AddRows(p.totalsRows(period.Totals)...)

		if len(period.PieceRateEarnings) == 0 {
			continue
		}

		p.maroto.AddRows(row.New(15).Add(
			col.New(12).Add(text.New("Piece Rate Earnings", props.Text{Size: 11, Align: align.Center})),
		))

		for _, entry := range period.PieceRateEarnings {
			p.maroto.AddRows(p.pieceRateRows(entry)...)
			p.maroto.AddRows(row.New(8))
		}
	}

	return p.save(outputDir, filename)
}

// GenerateSummary renders the single-page Shipt annual earnings summary.
func (p *pdfGenerator) GenerateSummary(outputDir, logoImagePath, filename string, report SummaryReport) error {
	companyRows, err := p.companyHeaderRows(logoImagePath, []string{
		"Shipt, Inc",
		"420 20th St N",
		"Birmingham, AL 35203",
	})
	if err != nil {
		return err
	}
	p.maroto.AddRows(companyRows...)

	p.maroto.AddRows(p.shopperInfoRows(report.Shopper)...)

	p.maroto.AddRows(
		row.New(8).Add(col.New(12).Add(text.New(report.Heading, props.Text{Size: 10, Style: fontstyle.Bold}))),
		row.New(22).Add(col.New(12).Add(text.New(report.Intro, props.Text{Size: 8}))),
	)

	for i, line := range report.Lines {
		var bg *props.Color
		if i%2 == 1 {
			bg = shadedColor
		}
		p.maroto.AddRows(
			row.New(8).WithStyle(&props.Cell{
				BorderType:  border.Full,
				BorderColor: borderColor,
				BackgroundColor: bg,
			}).Add(
				col.New(8).Add(text.New(line.Label, props.Text{Size: 9})),
				col.New(4).Add(text.New(line.Value, props.Text{Size: 9, Align: align.Right})),
			),
		)
	}

	p.maroto.AddRows(row.New(6))
	for _, footnote := range report.Footnotes {
		p.maroto.AddRows(row.New(10).Add(col.New(12).Add(text.New(footnote, props.Text{Size: 7}))))
	}

	return p.save(outputDir, filename)
}

func (p *pdfGenerator) save(outputDir, filename string) error {
	doc, err := p.maroto.Generate()
	if err != nil {
		return errors.Wrap(err, "failed to generate PDF")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}

	outPath := filepath.Join(outputDir, filename)
	return doc.Save(outPath)
}

// companyHeaderRows renders the centered Shipt logo followed by the centered company address block.
func (p *pdfGenerator) companyHeaderRows(logoImagePath string, addressLines []string) ([]core.Row, error) {
	logo, err := os.ReadFile(logoImagePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read logo file")
	}

	// row height matches the logo's actual height in the target payments.pdf (~72.4mm,
	// measured from a 300 DPI render); with Percent: 100 the image fills that height exactly.
	rows := []core.Row{
		row.New(73).Add(
			col.New(12).Add(
				image.NewFromBytes(logo, extension.Png, props.Rect{Percent: 100, Center: true}),
			),
		),
	}
	for _, line := range addressLines {
		rows = append(rows, row.New(6).Add(
			col.New(12).Add(text.New(line, props.Text{Size: 10, Align: align.Center})),
		))
	}
	rows = append(rows, row.New(10))

	return rows, nil
}

// shopperInfoRows renders the "Shopper Name" / "Shopper Id" key-value table.
func (p *pdfGenerator) shopperInfoRows(shopper ShopperInfo) []core.Row {
	return []core.Row{
		p.kvRow("Shopper Name", shopper.Name, false),
		p.kvRow("Shopper Id", shopper.ID, true),
		row.New(10),
	}
}

// totalsRows renders the "Gross Earnings / Deductions / Net Earnings" key-value table.
func (p *pdfGenerator) totalsRows(totals PeriodTotals) []core.Row {
	return []core.Row{
		p.kvRow("Gross Earnings", totals.GrossEarnings, false),
		p.kvRow("Deductions", totals.Deductions, true),
		p.kvRow("Net Earnings", totals.NetEarnings, false),
		row.New(10),
	}
}

// pieceRateRows renders a single "Date / Reference Id / Piece Rate / Earnings" key-value table.
func (p *pdfGenerator) pieceRateRows(entry PieceRateEntry) []core.Row {
	return []core.Row{
		p.kvRow("Date", entry.Date, false),
		p.kvRow("Reference Id", entry.ReferenceID, true),
		p.kvRow("Piece Rate", entry.PieceRate, false),
		p.kvRow("Earnings", entry.Earnings, true),
	}
}

// sectionHeading renders a centered bold section title, e.g. "Grand Totals for ...".
func (p *pdfGenerator) sectionHeading(title string) core.Row {
	return row.New(15).Add(
		col.New(12).Add(text.New(title, props.Text{Size: 11, Style: fontstyle.Bold, Align: align.Center})),
	)
}

// kvRow renders a bordered, optionally shaded label/value row.
func (p *pdfGenerator) kvRow(label, value string, shaded bool) core.Row {
	var bg *props.Color
	if shaded {
		bg = shadedColor
	}

	return row.New(10).WithStyle(&props.Cell{
		BorderType:      border.Full,
		BorderColor:     borderColor,
		BackgroundColor: bg,
	}).Add(
		col.New(4).Add(text.New(label, props.Text{Size: 9, Style: fontstyle.Bold})),
		col.New(8).Add(text.New(value, props.Text{Size: 9})),
	)
}
