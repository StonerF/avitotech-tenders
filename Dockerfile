FROM golang:alpine AS builder

WORKDIR /build

ADD go.mod .

COPY . .

EXPOSE 8080

RUN go build -o main cmd/main.go
