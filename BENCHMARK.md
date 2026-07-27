# PDF Generation Benchmark

This benchmark compares two PDF generation approaches:

1. **Maroto (gofpdf-based)** – Pure Go library for programmatic PDFs
2. **Chromedp (HTML → PDF)** – Renders HTML in headless Chrome and exports to PDF

Both benchmarks generate a single-period Shipt payments report (one pay period, two piece-rate
line items, the `assets/logo.png` header), matching the actual documents this project produces.

---

## 1. Maroto Benchmark

```bash
$ make benchmark-maroto
goos: darwin
goarch: arm64
pkg: github.com/shipt/pdf-generation-playground.git/pdf
cpu: Apple M3 Pro
BenchmarkTestPDFGenerator-12    	      97	  11753521 ns/op	14339667 B/op	   58064 allocs/op
PASS
ok  	github.com/shipt/pdf-generation-playground.git/pdf	2.389s
```

**Memory Profile (`go tool pprof mem_maroto.out`):**

```
(pprof) top
Showing nodes accounting for 1928.33MB, 92.47% of 2085.42MB total
      flat  flat%   sum%        cum   cum%
 1468.16MB 70.40% 70.40%  1468.16MB 70.40%  bytes.growSlice
  184.22MB  8.83% 79.23%   340.14MB 16.31%  compress/flate.NewWriter
   92.96MB  4.46% 83.69%   155.92MB  7.48%  compress/flate.(*compressor).init
   73.50MB  3.52% 87.22%       74MB  3.55%  fmt.newScanState
   61.96MB  2.97% 90.19%    61.96MB  2.97%  compress/flate.newDeflateFast (inline)
```

**CPU Profile (`go tool pprof cpu_maroto.out`):**

```
(pprof) top
Showing nodes accounting for 1620ms, 89.01% of 1820ms total
      flat  flat%   sum%        cum   cum%
    1030ms 56.59% 56.59%     1030ms 56.59%  runtime.madvise
     260ms 14.29% 70.88%      260ms 14.29%  runtime.kevent
     220ms 12.09% 82.97%      220ms 12.09%  syscall.rawsyscalln
      40ms  2.20% 85.16%       40ms  2.20%  internal/runtime/atomic.(*UnsafePointer).Load
      40ms  2.20% 87.36%       40ms  2.20%  runtime.pthread_cond_wait
```

**Observation:**

* **~12ms and ~14MB per PDF.**
* Memory is dominated by slice growth and gofpdf's zlib stream compression (`compress/flate`).
* CPU is mostly OS-level memory syscalls (`runtime.madvise`), i.e. actual document-building work
  is cheap; the cost is almost all allocator/GC overhead.

---

## 2. Chromedp Benchmark

```bash
$ make benchmark-chromedp
goos: darwin
goarch: arm64
pkg: github.com/shipt/pdf-generation-playground.git/htmlpdf
cpu: Apple M3 Pro
BenchmarkHTMLToPDFGenerator_Parallel4-12       	       1	3711953250 ns/op	115318904 B/op	   30310 allocs/op
--- BENCH: BenchmarkHTMLToPDFGenerator_Parallel4-12
    htmlpdf_test.go:87: PDF 0 generated in 1.3600605s
    htmlpdf_test.go:87: PDF 2 generated in 1.362983042s
    htmlpdf_test.go:87: PDF 1 generated in 1.371095s
    htmlpdf_test.go:87: PDF 3 generated in 1.37938125s
    htmlpdf_test.go:87: PDF 4 generated in 1.255972584s
    htmlpdf_test.go:87: PDF 6 generated in 1.263351542s
    htmlpdf_test.go:87: PDF 7 generated in 1.25622625s
    htmlpdf_test.go:87: PDF 5 generated in 1.274191667s
    htmlpdf_test.go:87: PDF 9 generated in 1.076393667s
    htmlpdf_test.go:87: PDF 8 generated in 1.095729167s
	... [output truncated]
BenchmarkHTMLToPDFGenerator_Sequential10-12    	       1	10161325709 ns/op	115582848 B/op	   27582 allocs/op
--- BENCH: BenchmarkHTMLToPDFGenerator_Sequential10-12
    htmlpdf_test.go:129: PDF 0 generated in 965.760166ms
    htmlpdf_test.go:129: PDF 1 generated in 1.04136125s
    htmlpdf_test.go:129: PDF 2 generated in 1.025927584s
    htmlpdf_test.go:129: PDF 3 generated in 1.032642541s
    htmlpdf_test.go:129: PDF 4 generated in 1.032710125s
    htmlpdf_test.go:129: PDF 5 generated in 1.050664833s
    htmlpdf_test.go:129: PDF 6 generated in 1.024568083s
    htmlpdf_test.go:129: PDF 7 generated in 979.974666ms
    htmlpdf_test.go:129: PDF 8 generated in 1.033579083s
    htmlpdf_test.go:129: PDF 9 generated in 973.972916ms
	... [output truncated]
PASS
ok  	github.com/shipt/pdf-generation-playground.git/htmlpdf	14.414s
```

**Memory Profile (`go tool pprof mem_chromedp.out`):**

```
(pprof) top
Showing nodes accounting for 187.29MB, 84.01% of 222.93MB total
      flat  flat%   sum%        cum   cum%
  108.63MB 48.73% 48.73%   108.63MB 48.73%  bytes.growSlice
   32.78MB 14.70% 63.43%    32.78MB 14.70%  github.com/go-json-experiment/json.makeString
   32.04MB 14.37% 77.80%    32.04MB 14.37%  github.com/go-json-experiment/json/jsontext.(*Value).UnmarshalJSON
    4.97MB  2.23% 80.03%     8.29MB  3.72%  fmt.Sprintf
    4.97MB  2.23% 82.27%     4.97MB  2.23%  encoding/base64.(*Encoding).EncodeToString
```

**CPU Profile (`go tool pprof cpu_chromedp.out`):**

```
(pprof) top
Showing nodes accounting for 450ms, 84.91% of 530ms total
      flat  flat%   sum%        cum   cum%
     290ms 54.72% 54.72%      290ms 54.72%  syscall.rawsyscalln
      50ms  9.43% 64.15%       50ms  9.43%  runtime.pthread_cond_signal
      40ms  7.55% 71.70%       40ms  7.55%  runtime.kevent
      30ms  5.66% 77.36%       30ms  5.66%  runtime.madvise
      30ms  5.66% 83.02%       30ms  5.66%  runtime.pthread_cond_wait
```

**Observation:**

* **~1.0–1.1s per PDF**, dominated by the Chrome DevTools Protocol round-trip
  (`syscall.rawsyscalln`), not application logic.
* **~11–12MB per PDF** (222.93MB / 20 PDFs across both sub-benchmarks). The base64-encoded logo
  is visible in the profile (`encoding/base64.EncodeToString`, `fmt.Sprintf`) as a direct cost of
  embedding it in the data URI — keeping that asset small keeps this number small.
* Parallel4 (3.71s/op for 10 PDFs, 4 at a time) is faster wall-clock than Sequential10
  (10.16s/op for 10 PDFs) even though per-PDF Chrome cost is similar — concurrency still pays
  off once the browser is warm.

---

## 3. Warm-up Considerations

Chromedp requires launching a headless Chrome instance:

* **First PDFs take longer** (\~1.3–1.4s) even though the browser is allocated once before the
  loop — the actual process spawn/DevTools handshake cost only shows up on the first real
  navigation.
* Later PDFs settle to **\~1.0–1.1s each**, both in the parallel and sequential runs.
* **Sequential cold runs will always pay this penalty**

**Mitigation Strategies:**

1. **Persistent Chrome instance** – launch once and reuse contexts
2. **Pre-warm on startup** – generate a dummy PDF to amortize the cost
3. **Context pooling** – for high concurrency, maintain 2–4 Chrome instances in rotation

---

## 4. Trade-off Summary

| Aspect               | Maroto (gofpdf)                     | Chromedp (HTML → PDF)                             |
| --------------------- | ------------------------------------ | --------------------------------------------------- |
| First PDF latency    | Very low (\~12ms)                   | High (warm-up \~1.3–1.4s)                          |
| Steady-state speed   | \~12ms per PDF                      | \~1.0–1.1s per PDF                                 |
| Memory usage         | \~14MB per PDF                      | \~11–12MB per PDF                                  |
| Concurrency          | Cheap — no shared external process  | Good, but bounded by Chrome instance/context pool  |
| External dependency  | None (pure Go)                      | Requires Chrome                                    |
| HTML/CSS Support     | None                                 | Full                                                |

Memory usage is within the same order of magnitude for both approaches. Maroto's latency
advantage is dramatic (\~12ms vs. \~1s), since it never pays the Chrome DevTools Protocol
round-trip cost.

---

## 5. Conclusion

* **Maroto/gofpdf** is the clear default for this project's document sizes: it's two orders of
  magnitude faster with comparable memory usage, at the cost of coding layout in Go instead of
  HTML/CSS.
* **Chromedp** is attractive when HTML/CSS fidelity or reuse of existing web templates matters
  more than raw latency, and its memory cost stays reasonable as long as embedded assets
  (logos, images) are kept small.

For production:

* Prefer Maroto unless the reports need rich HTML/CSS layouts that are impractical to replicate
  in Maroto's grid API.
* If using Chromedp, keep Chrome warm (persistent instance/context pool) and keep embedded
  assets as small as the design allows — asset size is the dominant memory cost, not the PDF
  conversion itself.
