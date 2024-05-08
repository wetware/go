# Go Wetware

Wetware node implementation and client libraries for Go programmers.

## Installation

The latest binaries can be installed by running `go install github.com/wetware/cmd`.

For the smoothest developer experience, we encouraged you to check that `$GOPATH` is set and that `$GOPATH/bin` is in your `$PATH`.

## Building from Source

>This section is for people who are hacking on Wetware itself.  If you're just using Wetware, you can safely skip this section.

### Building WASM Components
WASM artifacts are built by running `go generate ./...`.

WASM artifacts are into the main application or employed during testing.  The build-chain is managed by `go:generate` directives.  See `examples/hello-world/main.go` for one such example.

###  Application

The main application is built by running `go build cmd/main.go -o ww`.  New or modified WASM components will need to be (re)built beforehand.
