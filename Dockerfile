FROM golang:tip-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pdf-benchmark ./cmd

FROM debian:12.11

# Install Chrome + required deps + fonts
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget gnupg ca-certificates \
    fonts-liberation fonts-dejavu fonts-noto-core \
    libasound2 libatk1.0-0 libatk-bridge2.0-0 \
    libatspi2.0-0 libc6 libcairo2 libcups2 libcurl3-gnutls \
    libdbus-1-3 libexpat1 libgbm1 libglib2.0-0 libgtk-3-0 \
    libnspr4 libnss3 libpango-1.0-0 libudev1 libvulkan1 \
    libx11-6 libxcb1 libxcomposite1 libxdamage1 libxext6 \
    libxfixes3 libxkbcommon0 libxrandr2 xdg-utils \
    && rm -rf /var/lib/apt/lists/*

# Install Google Chrome
RUN wget -q https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt-get update && apt-get install -y --no-install-recommends ./google-chrome-stable_current_amd64.deb \
    && rm google-chrome-stable_current_amd64.deb \
    && rm -rf /var/lib/apt/lists/*

# Force correct Chrome path (overwrite if needed)
RUN ln -sf /opt/google/chrome/google-chrome /usr/bin/google-chrome

WORKDIR /app
COPY --from=builder /app/pdf-benchmark /usr/local/bin/pdf-benchmark
COPY assets ./assets

RUN useradd -m appuser
USER appuser

ENTRYPOINT ["pdf-benchmark"]
