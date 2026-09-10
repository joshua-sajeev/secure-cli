FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o secure-cli ./cmd/secure-cli

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/secure-cli /usr/local/bin/secure-cli

RUN mkdir -p /app/data

ENTRYPOINT ["secure-cli"]
