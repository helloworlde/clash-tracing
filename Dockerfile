FROM golang:alpine as builder

RUN apk add --no-cache git
WORKDIR /clash-tracing-src
COPY . /clash-tracing-src
RUN go mod download && \
    go build -o /clash-tracing .

FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /clash-tracing /
ENTRYPOINT ["/clash-tracing"]
