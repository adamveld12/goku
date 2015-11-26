# Base run commad
RUN=docker run -v /var/run/docker.sock:/var/run/docker.sock -v $$(which docker-compose):/bin/docker-compose
BUILD=docker build -t

up: build
	docker-compose up

build:
	$(BUILD) adamveld12/goku .

build_goku:
	go build -o ./bin/goku .

dev:
	$(RUN) -v $$PWD/:/go/src/github.com/adamveld12/goku --rm -it -p 3000:80 -p 2222:22 --entrypoint /bin/bash adamveld12/goku

live: build
	$(RUN) -d -p 2222:22 adamveld12/goku && $(UPLOAD)

release: build
	docker push adamveld12/goku
