package htmlpdf

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
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

type pdfGenerator struct {
	dockerEnv bool
}

func NewGenerator(dockerEnv bool) *pdfGenerator {
	return &pdfGenerator{dockerEnv: dockerEnv}
}

const sharedStyle = `
body {
  font-family: Arial, sans-serif;
  margin: 40px;
  color: #333;
}
.header {
  text-align: center;
  margin-bottom: 30px;
}
.header img {
  /* matches the logo's actual height in the target payments.pdf (~72.4mm, measured from a 300 DPI render) */
  height: 72mm;
}
.header .address {
  margin-top: 15px;
  font-size: 14px;
  line-height: 1.5;
}
.section-heading {
  text-align: center;
  font-size: 15px;
  font-weight: bold;
  margin: 20px 0 10px;
}
.subsection-heading {
  text-align: center;
  font-size: 14px;
  margin: 20px 0 10px;
}
.kv-table {
  border: 1px solid #e0e0e0;
  margin-bottom: 15px;
}
.kv-table .kv-row {
  display: flex;
  padding: 10px 12px;
  border-bottom: 1px solid #e0e0e0;
}
.kv-table .kv-row:last-child {
  border-bottom: none;
}
.kv-table .kv-row.shaded {
  background-color: #f2f2f2;
}
.kv-table .kv-label {
  flex: 0 0 30%;
  font-weight: bold;
  font-size: 13px;
}
.kv-table .kv-value {
  font-size: 13px;
  word-break: break-all;
}
.piece-rate-block {
  margin-bottom: 20px;
}
.summary-table {
  border-collapse: collapse;
  width: 100%;
  margin: 10px 0 15px;
}
.summary-table td {
  border: 1px solid #ccc;
  padding: 6px 8px;
  font-size: 12px;
}
.summary-table td:last-child {
  text-align: right;
}
.intro {
  font-size: 12px;
  line-height: 1.5;
}
.footnote {
  font-size: 11px;
  line-height: 1.5;
  color: #555;
  margin-top: 8px;
}
.footnote .learn-more {
  color: #1a73e8;
}
`

// GeneratePayments renders the multi pay-period Shipt payments statement.
func (p *pdfGenerator) GeneratePayments(
	ctx context.Context,
	outputDir, logoImagePath, filename string,
	report PaymentsReport,
) error {
	headerHTML, err := p.getHeader(logoImagePath, []string{
		"Shipt, Inc",
		"420 20th St. N. Suite 100",
		"Birmingham, AL 35203",
	})
	if err != nil {
		return err
	}

	var body strings.Builder
	body.WriteString(headerHTML)
	body.WriteString(kvTable([]kvRow{
		{"Shopper Name", report.Shopper.Name, false},
		{"Shopper Id", report.Shopper.ID, true},
	}))
	fmt.Fprintf(&body, `<div class="section-heading">Grand Totals for %s - %s</div>`,
		html.EscapeString(report.GrandTotal.Start), html.EscapeString(report.GrandTotal.End))
	body.WriteString(kvTable(totalsRows(report.GrandTotal)))

	for _, period := range report.Periods {
		fmt.Fprintf(&body, `<div class="section-heading">Payments for for %s - %s</div>`,
			html.EscapeString(period.Totals.Start), html.EscapeString(period.Totals.End))
		body.WriteString(kvTable(totalsRows(period.Totals)))

		if len(period.PieceRateEarnings) == 0 {
			continue
		}

		body.WriteString(`<div class="subsection-heading">Piece Rate Earnings</div>`)
		for _, entry := range period.PieceRateEarnings {
			body.WriteString(`<div class="piece-rate-block">`)
			body.WriteString(kvTable([]kvRow{
				{"Date", entry.Date, false},
				{"Reference Id", entry.ReferenceID, true},
				{"Piece Rate", entry.PieceRate, false},
				{"Earnings", entry.Earnings, true},
			}))
			body.WriteString(`</div>`)
		}
	}

	return p.render(ctx, outputDir, filename, body.String())
}

// GenerateSummary renders the single-page Shipt annual earnings summary.
func (p *pdfGenerator) GenerateSummary(
	ctx context.Context,
	outputDir, logoImagePath, filename string,
	report SummaryReport,
) error {
	headerHTML, err := p.getHeader(logoImagePath, []string{
		"Shipt, Inc",
		"420 20th St N",
		"Birmingham, AL 35203",
	})
	if err != nil {
		return err
	}

	var body strings.Builder
	body.WriteString(headerHTML)
	body.WriteString(kvTable([]kvRow{
		{"Shopper Name", report.Shopper.Name, false},
		{"Shopper Id", report.Shopper.ID, false},
	}))
	fmt.Fprintf(&body, `<div style="font-weight:bold;font-size:13px;margin-top:10px;">%s</div>`,
		html.EscapeString(report.Heading))
	fmt.Fprintf(&body, `<div class="intro">%s</div>`, html.EscapeString(report.Intro))

	body.WriteString(`<table class="summary-table">`)
	for _, line := range report.Lines {
		fmt.Fprintf(&body, `<tr><td>%s</td><td>%s</td></tr>`,
			html.EscapeString(line.Label), html.EscapeString(line.Value))
	}
	body.WriteString(`</table>`)

	for _, footnote := range report.Footnotes {
		escaped := html.EscapeString(footnote)
		escaped = strings.Replace(escaped, "Learn More", `<span class="learn-more">Learn More</span>`, 1)
		fmt.Fprintf(&body, `<div class="footnote">%s</div>`, escaped)
	}

	return p.render(ctx, outputDir, filename, body.String())
}

func (p *pdfGenerator) render(ctx context.Context, outputDir, filename, bodyHTML string) error {
	fullHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>%s</style>
</head>
<body>
%s
</body>
</html>
`, sharedStyle, bodyHTML)

	// Docker-friendly chromedp.
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
	)
	if p.dockerEnv {
		// in a Docker environment, specify the path to the Chrome binary.
		allocOpts = append(allocOpts, chromedp.ExecPath("/usr/bin/google-chrome"))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	var pdfBuf []byte
	if err := chromedp.Run(taskCtx,
		chromedp.Navigate("data:text/html,"+url.PathEscape(fullHTML)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).   // A4 width in inches.
				WithPaperHeight(11.69). // A4 height in inches.
				Do(ctx)
			return err
		}),
	); err != nil {
		return errors.Wrap(err, "failed to generate PDF")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(outputDir, filename), pdfBuf, 0o644)
}

func (p *pdfGenerator) getHeader(logoImagePath string, addressLines []string) (string, error) {
	if _, err := os.Stat(logoImagePath); os.IsNotExist(err) {
		logoImagePath = "assets/logo.png"
	}

	logoBytes, err := os.ReadFile(logoImagePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read logo file")
	}
	logoBase64 := base64.StdEncoding.EncodeToString(logoBytes)

	address := make([]string, len(addressLines))
	for i, line := range addressLines {
		address[i] = html.EscapeString(line)
	}

	return fmt.Sprintf(`
<div class="header">
    <img src="data:image/png;base64,%s">
    <div class="address">%s</div>
</div>
`, logoBase64, strings.Join(address, "<br>")), nil
}

type kvRow struct {
	Label  string
	Value  string
	Shaded bool
}

func totalsRows(totals PeriodTotals) []kvRow {
	return []kvRow{
		{"Gross Earnings", totals.GrossEarnings, false},
		{"Deductions", totals.Deductions, true},
		{"Net Earnings", totals.NetEarnings, false},
	}
}

func kvTable(rows []kvRow) string {
	var b strings.Builder
	b.WriteString(`<div class="kv-table">`)
	for _, r := range rows {
		class := "kv-row"
		if r.Shaded {
			class += " shaded"
		}
		fmt.Fprintf(&b,
			`<div class="%s"><div class="kv-label">%s</div><div class="kv-value">%s</div></div>`,
			class, html.EscapeString(r.Label), html.EscapeString(r.Value),
		)
	}
	b.WriteString(`</div>`)
	return b.String()
}
