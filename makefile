dev: build
	./goku -debug -gitpath ./repositories -http ":8080" server

vagrantdev: build nginx
	./goku -debug -gitpath ./repositories -http ":8080" -host "192.168.50.4.xip.io" server

test_cover:
	rm -rf ./c.out
	go test -v -coverprofile=c.out
	go tool cover -html=c.out

nginx:
	cp ./nginx.conf /etc/nginx
	sudo service nginx reload

build: clean
	go build -o ./goku ./cli/goku 

clean:
	rm -rf ./goku
	rm -rf ./repositories

