# Go Wetware

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/wetware/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/wetware/go?style=flat-square)](https://goreportcard.com/report/github.com/wetware/go)
[![tests](https://github.com/wetware/go/workflows/Go/badge.svg)](https://github.com/wetware/go/actions/workflows/go.yml)
[![Matrix](https://img.shields.io/matrix/wetware:matrix.org?color=lightpink&label=support%20chat&logo=matrix&style=flat-square)](https://matrix.to/#/#wetware:matrix.org)
[![white paper](https://img.shields.io/badge/white%20paper-reading%20time%20--%207%20min-9cf?style=flat-square)](https://hackmd.io/@fCsHyW7yR3C5lGQFbh9KdQ/SJzOIt9k3)

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
