dev: build
	./goku -debug -gitpath ./repositories -http ":8080" server

vagrant: build
	./goku -debug -gitpath ./repositories -http ":8080" -host "192.168.99.100.xip.io" server

build: clean
	go build -o ./goku ./cli/goku 

clean:
	rm -rf ./goku
	rm -rf ./repositories
