FROM golang:alpine as builder

RUN apk add --no-cache git
WORKDIR /clash-reporter-src
COPY . /clash-reporter-src
RUN go mod download && \
    go build -o /clash-reporter .

FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /clash-reporter /
ENTRYPOINT ["/clash-reporter"]
