FROM golang:1.8 as golang

WORKDIR /go/src/github.com/koba789/steved
COPY . /go/src/github.com/koba789/steved
RUN go-wrapper download && CGO_ENABLED=0 GOOS=linux go-wrapper install

FROM alpine:latest

WORKDIR /app
COPY --from=golang /go/bin/steved /app

CMD ["/app/steved"]
