GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
OS_PERMS=sudo
CWD=$(shell pwd)
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)
FULL_DOCKER_NAME=unixvoid/bitnuke

all: bitnuke

dependencies:
	go mod init github.com/unixvoid/bitnuke
	go mod tidy

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
		bitnuke/remove_shortlink.go \
		bitnuke/token_generator.go \
		bitnuke/upload.go

webdev:
	$(OS_PERMS) docker run \
		-d \
		--name bitnuke-webdev \
		-v ./deps/nginx/nginx-local.conf:/nginx/conf/nginx.conf:ro \
		-v ./deps/nginx/mime.types:/nginx/conf/mime.types:ro \
		-v ./deps/nginx/data:/nginx/data:ro \
		-p 8881:80 \
		nginx:1.19.6-alpine \
		nginx -g 'daemon off;' -c /nginx/conf/nginx.conf
	$(OS_PERMS) docker logs -f bitnuke-webdev

stop-webdev:
	$(OS_PERMS) docker stop -t 0 bitnuke-webdev && sudo docker rm bitnuke-webdev

docker: clean stat
	rm -rf stage.tmp/
	mkdir -p stage.tmp/
	cp deps/Dockerfile-bitnuke stage.tmp/Dockerfile
	cp bin/bitnuke* stage.tmp/bitnuke
	cd stage.tmp/ && \
		$(OS_PERMS) docker build -t $(FULL_DOCKER_NAME) .

run-stack:
	cd deps/ && \
		$(OS_PERMS) docker compose up -d && \
		$(OS_PERMS) docker compose logs -f
restart-stack:
	cd deps/ && \
		$(OS_PERMS) docker compose down && \
		$(OS_PERMS) docker rm `sudo docker ps -aq` 2> /dev/null && \
		$(OS_PERMS) docker compose up -d && \
		$(OS_PERMS) docker compose logs -f

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/bitnuke-$(GIT_HASH)-linux-amd64 bitnuke/*.go

clean:
	rm -rf bin/
	rm -rf stage.tmp/
