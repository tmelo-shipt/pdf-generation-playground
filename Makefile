.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]\n"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: test
## test: run unit tests
test:
	@ go test -v ./... -count=1

.PHONY: pdf
## pdf: generate PDF using maroto
pdf:
	@ go run cmd/main.go sample

.PHONY: htmlpdf
## htmlpdf: generate PDF using html2pdf
htmlpdf:
	@ go run cmd/main.go fromHTML

.PHONY: benchmark-maroto
## benchmark-maroto: run maroto benchmark with CPU and memory profiling
benchmark-maroto:
	@go test -bench=. ./pdf/ -benchmem \
	    -memprofile mem_maroto.out \
	    -cpuprofile cpu_maroto.out

.PHONY: benchmark-chromedp
## benchmark-chromedp: run chromedp benchmark with CPU and memory profiling
benchmark-chromedp:
	@go test -bench=. ./htmlpdf/ -benchmem \
	    -memprofile mem_chromedp.out \
	    -cpuprofile cpu_chromedp.out

.PHONY: docker-build
## docker-build: build Docker image for the application
docker-build:
	@docker buildx build --platform linux/amd64 -t pdf-benchmark .

.PHONY: pdf-docker
## pdf-docker: generate PDF using maroto in Docker
pdf-docker: docker-build
	@docker run --rm -v ${PWD}/samples:/samples pdf-benchmark sample -o /samples/pdf

.PHONY: htmlpdf-docker
## htmlpdf-docker: generate PDF using html2pdf in Docker
htmlpdf-docker: docker-build
	@docker run --rm  -v ${PWD}/samples:/samples pdf-benchmark fromHTML -o /samples/htmlpdf -d true