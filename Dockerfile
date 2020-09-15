FROM golang:alpine AS cmd

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
RUN go build -o /go/bin/forge
RUN /go/bin/forge --help