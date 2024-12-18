# syntax=docker/dockerfile:1

FROM golang:1.23-alpine

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

CMD ["./main"]