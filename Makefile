.PHONY: clean

all: binary

clean:
	@if [ -f "ww" ]; then rm ww; fi

generate:
	@go generate ./...

binary: generate
	@go build -o ww cmd/main.go

install: generate
	@go install github.com/wetware/go/cmd
