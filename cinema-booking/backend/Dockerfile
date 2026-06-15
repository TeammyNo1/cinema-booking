FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Copy ทุกอย่างก่อน แล้วค่อย tidy + download + build ในขั้นตอนเดียว
COPY . .

# go mod tidy จะสร้าง go.sum อัตโนมัติ แล้ว download + build เลย
RUN go mod tidy && \
    go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cinema-api ./cmd/server

# ─── Runtime image ────────────────────────────────────────────────────────────
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/cinema-api .

EXPOSE 8080

CMD ["./cinema-api"]
