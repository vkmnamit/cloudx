# ---------- Build Stage ----------
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o cloudx ./cmd/server

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/cloudx .

EXPOSE 8080

CMD ["./cloudx"]