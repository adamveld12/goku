# Base run commad
RUN=docker run -v /var/run/docker.sock:/var/run/docker.sock -v $$(which docker):/bin/docker
BUILD=docker build -t
UPLOAD=cat ./bin/authorized_keys | ssh root@$$(docker-machine ip default) "gitreceive upload-key test"

up: build
	docker-compose up

upload: 
	$(UPLOAD)

build: build_goku
	$(BUILD) adamveld12/goku .

build_goku:
	go build -o ./bin/goku .

dev:
	$(RUN) -v $$PWD/:/go/src/github.com/adamveld12/goku --rm -it -p 2222:22 --entrypoint /bin/bash adamveld12/goku

live: build
	$(RUN) -d -p 2222:22 adamveld12/goku && $(UPLOAD)

release: build
	docker push adamveld12/goku
