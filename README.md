# pdf-generation-playground

This repository demonstrates different approaches to generating PDF documents in Go, with a focus on reproducing **Shipt shopper earnings statements** (`payments.pdf` and `summary.pdf`).  
It serves as a **proof-of-concept / discovery project** for evaluating PDF generation strategies for backend services.

---

## features

- **option A – Native Go PDF generation**
  - Uses [`maroto`](https://github.com/johnfercher/maroto) v2 for layout-based PDF creation.
  - Supports:
    - Company logo (PNG, from file or base64)
    - Shopper info, pay-period totals, and piece-rate line items as bordered/shaded key-value tables
    - Multi-period payments statement (`payments.pdf`) and single-page annual summary (`summary.pdf`)
  - Generates PDFs **fully from Go code**, no browser dependency.

- **option B – HTML → PDF conversion**
  - Uses [`chromedp`](https://github.com/chromedp/chromedp) and headless Chrome.
  - Converts **HTML templates to PDF** with:
    - CSS-styled key-value tables and summary tables
    - Print backgrounds
  - Enables reuse of existing HTML/CSS designs for these reports.

---

## benchmark

See the [benchmark](./BENCHMARK.md) session.

---

## project structure

```
pdf-generation-playground/
├─ cmd/
│  └─ main.go           # CLI entrypoint
├─ pdf/
│  └─ pdf.go            # Maroto-based PDF generator
├─ htmlpdf/
│  └─ htmlpdf.go        # Chromedp HTML → PDF generator
├─ assets/
│  └─ logo.png          # Shipt logo
├─ samples/
│  ├─ pdf/              # Output of Maroto-generated reports
│  └─ htmlpdf/          # Output of HTML→PDF-generated reports
└─ README.md
```

---

## usage

### 1. generate sample PDFs using **Maroto**

```bash
make pdf
```

Output:

```
Sample PDFs generated successfully at: samples/pdf
```

---

### 2. Generate PDFs from **HTML template**

```bash
make htmlpdf
```

Output:

```
PDFs generated successfully from HTML at: samples/htmlpdf
```

Both commands produce `payments.pdf` and `summary.pdf` in the output directory.

#### running in Docker

1. generate sample PDFs using **Maroto**

```
make pdf-docker
```

2. generate PDFs from **HTML template**

```
make htmlpdf-docker
```

---

## dependencies

* **Go ≥ 1.22**
* **[maroto v2](https://github.com/johnfercher/maroto)**
* **[chromedp](https://github.com/chromedp/chromedp)** (requires Chrome/Chromium installed)
* **pkg/errors** for error wrapping

---

## notes & considerations

* **Maroto (Option A)**

  * Great for simple, structured reports without external dependencies.
  * Fully controlled in Go, but layouting is code-driven.

* **Chromedp HTML → PDF (Option B)**

  * Flexible for rich styling (CSS, fonts, logos).
  * Requires Chrome/Chromium on the host.
  * Ideal if existing HTML templates already exist for the report.

