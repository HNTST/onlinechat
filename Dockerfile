FROM golang:1.26.2 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/app ./cmd/app/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/app .

COPY --from=builder /build/internal/core/config ./internal/core/config

EXPOSE 8080

CMD ["./app"]
