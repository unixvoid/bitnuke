.PHONY: docker
GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX=sudo
IMAGE_NAME=unixvoid/bitnuke
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

all: bitnuke

dependencies:
	go get github.com/gorilla/mux
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1
	go get gopkg.in/redis.v3

daemon:
	bin/bitnuke &
bitnuke:
	$(GOC) bitnuke.go

run:
	cd bitnuke && go run \
		bitnuke.go \
		dynamic_handler.go \
		link_compressor.go \
		remove.go \
		token_generator.go \
		upload.go

docker:
	$(MAKE) stat
	mkdir -p stage.tmp/
	cp bin/bitnuke stage.tmp/
	cp deps/Dockerfile stage.tmp/
	cp deps/run.sh stage.tmp/
	sed -i "s/<DIFF>/$(GIT_HASH)/g" stage.tmp/Dockerfile
	cd stage.tmp && \
		$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/bitnuke bitnuke/*.go

install: stat
	cp bitnuke /usr/bin

clean:
	rm -rf bin/
	rm -rf stage.tmp/
