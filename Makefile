GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
OS_PERMS=sudo
CWD=$(shell pwd)
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

all: bitnuke

dependencies:
	go get github.com/gorilla/mux
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1
	go get gopkg.in/redis.v5
	go get golang.org/x/crypto/sha3

daemon:
	bin/bitnuke &

bitnuke:
	$(GOC) bitnuke.go

run:
	go run \
		bitnuke/bitnuke.go \
		bitnuke/dynamic_handler.go \
		bitnuke/link_compressor.go \
		bitnuke/remove.go \
		bitnuke/token_generator.go \
		bitnuke/upload.go

prep_aci: stat
	mkdir -p stage.tmp/bitnuke-layout/rootfs/
	cp bin/bitnuke* stage.tmp/bitnuke-layout/rootfs/bitnuke
	cp config.gcfg stage.tmp/bitnuke-layout/rootfs/
	cp deps/manifest.json stage.tmp/bitnuke-layout/manifest

build_aci: prep_aci
	# build image
	cd stage.tmp/ && \
		actool build bitnuke-layout bitnuke-api.aci && \
		mv bitnuke-api.aci ../
	@echo "bitnuke-api.aci built"

build_travis_aci: prep_aci
	wget https://github.com/appc/spec/releases/download/v0.8.7/appc-v0.8.7.tar.gz
	tar -zxf appc-v0.8.7.tar.gz
	# build image
	cd stage.tmp/ && \
		../appc-v0.8.7/actool build bitnuke-layout bitnuke-api.aci && \
		mv bitnuke-api.aci ../
	rm -rf appc-v0.8.7*
	@echo "bitnuke-api.aci built"

test: clean build_aci
	mkdir -p /tmp/redis
	mkdir -p /tmp/nginx/log
	$(OS_PERMS) rkt run \
		--port=web-http:8080 \
		--volume redis-data,kind=host,source=/tmp/redis \
		--volume nginx-data,kind=host,source=$(CWD)/deps/data/ \
		--volume nginx-conf,kind=host,source=$(CWD)/deps/conf/nginx.conf \
		--volume nginx-log,kind=host,source=/tmp/nginx/log \
		--volume nginx-mime,kind=host,source=$(CWD)/deps/conf/mime.types \
		unixvoid.com/redis \
			--mount volume=redis-data,target=/redisbak \
		unixvoid.com/nginx-1.13.11 \
			--mount volume=nginx-data,target=/data \
			--mount volume=nginx-conf,target=/nginx/nginx.conf \
			--mount volume=nginx-mime,target=/conf/mime.types \
			--mount volume=nginx-log,target=/nginx/log/ \
		./bitnuke-api.aci \
			--insecure-options=image



stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/bitnuke-$(GIT_HASH)-linux-amd64 bitnuke/*.go

install: stat
	cp bitnuke /usr/bin

clean:
	rm -rf bin/
	rm -rf stage.tmp/
	rm -f bitnuke-api.aci
