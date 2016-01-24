build:
	mkdir -p bin .gopath
	if [ ! -L .gopath/src ]; then ln -s "$(CURDIR)/vendor" .gopath/src; fi
	cd src && GOBIN="../bin/" GOPATH="$(CURDIR)/.gopath" go install && mv ../bin/src ../bin/wolfengo

errcheck:
	mkdir -p bin .gopath
	if [ ! -L .gopath/src ]; then ln -s "$(CURDIR)/vendor" .gopath/src; fi
	cd src && GOPATH="$(CURDIR)/.gopath" errcheck

test:
	mkdir -p bin .gopath
	if [ ! -L .gopath/src ]; then ln -s "$(CURDIR)/vendor" .gopath/src; fi
	cd src && GOPATH="$(CURDIR)/.gopath" go test -v

all: build errcheck test

.PHONY: build errcheck test
