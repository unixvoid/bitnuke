GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0

all: bitnuke

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

stage: bitnuke.go
	make stat
	mv bitnuke docker/

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/bitnuke bitnuke/*.go

install: stat
	cp bitnuke /usr/bin

clean:
	rm -f bitnuke
	rm -f docker/bitnuke
	rm -rf bin/

#CGO_ENABLED=0 go build -a -ldflags '-s' bitnuke.go
