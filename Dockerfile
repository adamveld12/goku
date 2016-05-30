FROM golang:1.6.2-wheezy

RUN apt-get update && apt-get install -y nginx

COPY . /go/src/github.com/adamveld12/goku/
WORKDIR /go/src/github.com/adamveld12/goku/

RUN go get

EXPOSE 6789 80

ENTRYPOINT go run *.go -debug
