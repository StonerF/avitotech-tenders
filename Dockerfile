FROM golang:alpine AS builder
EXPOSE 8080

WORKDIR /build

ADD go.mod .

COPY . .

RUN go build -o main cmd/main.go

CMD [". /cmd/main"]