.PHONY: clean

all: binary

clean:
	@if [ -f "ww" ]; then rm ww; fi

binary:
	@go generate ./...
	@go build -o ww cmd/main.go
