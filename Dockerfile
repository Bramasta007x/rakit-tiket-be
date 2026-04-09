# Stage 1: Build
FROM golang:1.24-bookworm AS builder
WORKDIR /app
COPY . .
RUN go mod download
# Build binary
RUN go build -o main cmd/main.go 

# Stage 2: Run
FROM debian:bookworm-slim
WORKDIR /app

# Runtime dependencies:
# - wkhtmltopdf: generate e-ticket PDF
# - font packages: avoid blank/garbled PDF text
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    fontconfig \
    fonts-dejavu-core \
    wkhtmltopdf \
    && rm -rf /var/lib/apt/lists/*

# Copy binary dari stage builder
COPY --from=builder /app/main .
# PENTING: Copy folder script migrasi agar bisa dibaca pgClient.Migration()
COPY --from=builder /app/internal/pkg/scripts ./internal/pkg/scripts
# Copy file .env (atau nanti di-inject via docker-compose)
COPY --from=builder /app/.env . 

CMD ["./main"]
