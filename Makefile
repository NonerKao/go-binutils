.nstPHONY: build install clean

PACKAGE = go-binutils
GOPATH  = $(PWD)/../../../..
BASE    = $(GOPATH)/src/$(PACKAGE)

LIST    = $(shell ls)

all: build

build: $(LIST)
	go build 

install: 
	go install
	ln -s $(GOPATH)/bin/go-binutils $(GOPATH)/bin/readelf	

clean:
	rm -f main go-binutils $(GOPATH)/bin/go-binutils
