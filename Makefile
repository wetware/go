.PHONY: clean

all: binary

clean:
	@if [ -f "ww" ]; then rm ww; fi
	@rm -f $(GOPATH)/bin/ww

generate:
	go generate ./...

binary: generate
	go build -o ww cmd/main.go

install:
	go install github.com/wetware/go/cmd
