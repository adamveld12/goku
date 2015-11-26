FROM golang:1.5.1-wheezy

RUN apt-get update && apt-get install -y nginx

COPY . /go/src/github.com/adamveld12/goku/
WORKDIR /go/src/github.com/adamveld12/goku/

RUN go get

EXPOSE 22 80

ENTRYPOINT go run *.go -debug
