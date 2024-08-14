.PHONY: clean

all: install

clean:
	@if [ -f "ww" ]; then rm ww; fi
	@rm -f $(GOPATH)/bin/ww

generate:
	go generate ./...

publish: generate
	ipfs add -r .

install:
	go install github.com/wetware/go/cmd/ww
