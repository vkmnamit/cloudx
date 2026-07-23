# ---------- Build Stage ----------
FROM golang:1.26.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o cloudx ./cmd/server

# ---------- Runtime Stage ----------
FROM alpine:latest

# Install git for repository cloning and docker-cli for container orchestration
RUN apk add --no-cache git docker-cli curl

WORKDIR /app

COPY --from=builder /app/cloudx .

EXPOSE 8080

CMD ["./cloudx"]