FROM golang:1.5.1-wheezy


COPY . /go/src/github.com/adamveld12/goku/
WORKDIR /go/src/github.com/adamveld12/goku/

RUN go get

EXPOSE 22 80

ENTRYPOINT go run *.go -debug
