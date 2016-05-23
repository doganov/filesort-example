.PHONY: all clean build test bench

GO=go
PACKAGES=filegen filesort

export GOPATH := $(shell pwd)

all: build

build: bin/filegen bin/filesort

bin/filegen: $(wildcard src/filegen/*.go)
	$(GO) install filegen

bin/filesort: $(wildcard src/filesort/*.go)
	$(GO) install filesort

test:
	go test -v $(PACKAGES) && go vet $(PACKAGES)

bench:
	go test -run NONE -bench . $(PACKAGES)

clean:
	rm -rf bin

