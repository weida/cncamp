# syntax=docker/dockerfile:1
FROM golang:1.16  AS builder
RUN  mkdir -p /bin/amd64
RUN  mkdir -p /bin/amd64/metrics
WORkDIR /bin/amd64
COPY httpserver.go  /bin/amd64/
RUN go env -w GOPROXY=https://goproxy.cn
RUN go mod init httpserver
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/amd64/httpserver

FROM alpine:latest as prod
ENV VERSION=1.3
COPY --from=builder /bin/amd64/httpserver /httpserver
EXPOSE 8090
ENTRYPOINT /httpserver
