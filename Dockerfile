FROM golang:latest AS builder
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rate-limiter cmd/rate-limiter/main.go

# FROM scratch não funciona pois não possui os Certificados Root e não valida o weatherapi.com
FROM scratch
COPY --from=builder /app/rate-limiter /app/rate-limiter

CMD ["/app/rate-limiter"]