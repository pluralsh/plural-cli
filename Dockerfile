FROM gcr.io/pluralsh/golang:alpine AS cmd

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/plural/
COPY . .
RUN go build -o /go/bin/plural -ldflags '-s -w' ./cmd/plural/
RUN /go/bin/plural --help