# syntax=docker/dockerfile:1

FROM golang:1.23-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]