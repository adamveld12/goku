# Base run commad
RUN=docker run -v /var/run/docker.sock:/var/run/docker.sock

dev: build
	./goku server -debug -ssh=:2223

build: clean
	go build .

install:
	go get

clean:
	rm -rf ./goku
	rm -rf ./repositories
