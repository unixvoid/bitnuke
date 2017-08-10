.PHONY: docker
GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX=sudo
IMAGE_NAME=unixvoid/bitnuke
FULL_IMAGE_NAME=unixvoid/bitnuke
NGINX_IMAGE_NAME=unixvoid/bitnuke:nginx
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)
HOST_IP=172.17.0.1

all: bitnuke

dependencies:
	go get github.com/gorilla/mux
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1
	go get gopkg.in/redis.v3
	go get golang.org/x/crypto/sha3

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
	cp bin/bitnuke* stage.tmp/
	cp deps/Dockerfile stage.tmp/
	cp deps/run.sh stage.tmp/
	sed -i "s/<DIFF>/$(GIT_HASH)/g" stage.tmp/Dockerfile
	cd stage.tmp && \
		$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

fulldocker:
	$(MAKE) stat
	mkdir -p stage.tmp/
	cp bin/bitnuke* stage.tmp/bitnuke
	cp bitnuke/config.gcfg stage.tmp/
	cp deps/Dockerfile.full stage.tmp/Dockerfile
	cp -R deps/conf stage.tmp/
	cp -R deps/data stage.tmp/
	cp deps/full.run.sh stage.tmp/run.sh
	mv stage.tmp/conf/daemon.nginx.conf stage.tmp/conf/nginx.conf
	wget -O stage.tmp/nginx https://cryo.unixvoid.com/bin/nginx/libressl/nginx-1.11.10-linux-amd64
	chmod +x stage.tmp/nginx
	sed -i "s/<DIFF>/$(GIT_HASH)/g" stage.tmp/Dockerfile
	cd stage.tmp && \
		$(DOCKER_PREFIX) docker build -t $(FULL_IMAGE_NAME) .

runfull:
	$(DOCKER_PREFIX) docker run \
		-it \
		--rm \
		-p 9009:9009 \
		--name bitnuke-nginx \
		$(FULL_IMAGE_NAME)
	#$(DOCKER_PREFIX) docker logs -f bitnuke-nginx

nginx:
	mkdir -p stage.tmp/
	cp deps/Dockerfile.nginx stage.tmp/Dockerfile
	cp -R deps/conf stage.tmp/
	cp -R deps/data stage.tmp/
	sed -i "s/<SERVER_IP>/$(HOST_IP)/g" stage.tmp/conf/nginx.conf
	cd stage.tmp && \
		$(DOCKER_PREFIX) docker build -t $(NGINX_IMAGE_NAME) .

runnginx:
	$(DOCKER_PREFIX) docker run \
		-it \
		--rm \
		-p 9009:9009 \
		--name bitnuke-nginx \
		$(NGINX_IMAGE_NAME)
	#$(DOCKER_PREFIX) docker logs -f bitnuke-nginx

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/bitnuke-$(GIT_HASH)-linux-amd64 bitnuke/*.go

install: stat
	cp bitnuke /usr/bin

clean:
	rm -rf bin/
	rm -rf stage.tmp/
