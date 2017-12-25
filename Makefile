
PACKAGE		= go-binutils

UTILS		= readelf objdump as
BINDIR		= $(GOPATH)/bin
TARGETS		= $(addprefix $(BINDIR)/, $(UTILS))

.nstPHONY: build install clean

all: build 

build: common $(UTILS) main.go
	go build

install: $(BINDIR)/$(PACKAGE) $(TARGETS)

$(BINDIR)/go-binutils: common $(UTILS) main.go
	go install

$(TARGETS): $(BINDIR)/%: %
	ln -s $(BINDIR)/$(PACKAGE) $(BINDIR)/$<

clean:
	rm -f main go-binutils $(BINDIR)/$(PACKAGE) $(TARGETS)
