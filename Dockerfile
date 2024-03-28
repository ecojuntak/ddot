FROM golang:1.22 as builder
LABEL stage=builder
WORKDIR /builder

COPY . .
RUN GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bin/ddot main.go

FROM alpine:latest

COPY --from=builder /builder/bin/ddot /bin/ddot
