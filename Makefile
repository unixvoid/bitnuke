GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0

all: bitnuke

bitnuke: bitnuke.go
	$(GOC) bitnuke.go

run: bitnuke.go
	go run bitnuke.go

stage: bitnuke.go
	make stat
	make statremote
	mv bitnuke builddeps/
	mv remotemanager builddeps/

stat: bitnuke.go
	$(CGOR) $(GOC) $(GOFLAGS) bitnuke.go

install: stat
	cp bitnuke /usr/bin

clean:
	rm -f bitnuke
	rm -r tmpnuke/*

#CGO_ENABLED=0 go build -a -ldflags '-s' bitnuke.go
