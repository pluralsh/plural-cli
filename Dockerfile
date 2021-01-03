FROM golang:alpine AS cmd

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/forge/
COPY . .
RUN go build -o /go/bin/forge -ldflags '-s -w' ./cmd/forge/
RUN /go/bin/forge --help