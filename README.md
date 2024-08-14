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

This section is for people who are hacking on Wetware itself.  If you're just using Wetware, you don't need to read this.

Before starting, it is recommended configure your shell to add `$GOPATH/bin` to your `PATH`, if you have not done so already.

### One-Liner

You can perform the entire step-by-step build in once command.

```bash
$ make
```



### Step-by-Step

Start by installing the binary to your `$GOPATH/bin`.

```bash
$ make install  # requires $GOPATH/bin to be part of $PATH
```


Then, publish the top-level repo to IPFS, so that we can later navigate down
to the desired `main.wasm`. 

```bash
$ make publish
```
Your output should look similar to this, but possibly with different hash values:
```
go generate ./...
ipfs add -r .
added Qma3D1VxRYhEGsd6QnJN14CKVbMDRLyYvXJppwaPBw4EUh go/Makefile
added QmS5dcfcBn1wz8iZ7VvAcSfqip86bub9AdHKvRFjBZQN7i go/README.md
added QmPevxpfFSHCHX32eD8n25VvEUPqkBFiSa13n9EHLei783 go/cmd/export/export.go
added QmP9ZDkC5VzBv3K34DbUJNDsKwiXgifE86DDWJhrUafkU1 go/cmd/main.go
added QmZp5uVuwkASacGdvAQiz5ymNqF3JcGqjYf2Tz824sGp1f go/cmd/run/run.go
added QmeqRn1nwjT55Ztb6WY7CgBnx2ApZQBuYf6bh6SLtFseep go/examples/hello-world/main.go
added QmfJUq8S8A4mGvbfpiy1D6owYLNwxtectgpZzfwSj5PGK7 go/examples/hello-world/main.wasm
added Qmdfmvq7by9JFg2fbfd4m6Ky9hnfAatpePjujH1Bn5mVvq go/go.mod
added QmcPC48ezbMPupFfiLS1sgaWi182iA4ar8MuDJ71ibgU3z go/go.sum
added QmWgB1pAGArYe5fg5ECQyNFJ2B348dXRcCTRh2dRFNSD4w go/system/cmd/nop/main.go
added QmPphwrbnDVSk7xP1QvQWBvfRBPrhMCpVf4RTR8bLkkbcN go/system/cmd/nop/main.wasm
added QmbUaEmsxkpgPNWD95xcnVn9tmsBHaXp1w5GeK96JqgTvo go/system/fs.go
added QmShXte4QUj3G1tcVULnqTahpcVGjDcSJ9KFJAGMgGRMNL go/system/fs_test.go
added QmSNbKtm6nToqcuQg55F2r1QC68cZVq9ZTRDd15winJeTA go/system/system.go
added QmSninJxE9nocVZZGn7XKjh9sTvJ5W9bY6GkX3uwmxWHG3 go/system/testdata/main.go
added QmegusupDqtTHptyPbW84BbtdwAhNaMuMoTXTu7KaGBMv1 go/system/testdata/main.wasm
added QmRf22bZar3WKmojipms22PkXH1MZGmvsqzQtuSvQE3uhm go/system/testdata/testdata
added QmZoiJbZdp9zop3LFycjw2e1mKYJRH6zugvMtcng6ZGUKw go/test/libp2p/libp2p.go
added QmeMJ4Hv7psSc5EP9yh9jpeecAsBGnYJQmxjymihgtPBDn go/test/libp2p.go
added QmRy4HmXbL4ChPwQfewWGE5SbxmGF1TvCN6LY2zFMrKxPk go/util/fail.go
added QmP4LuE6nQbHHYXnAComiYXCnCdL4CLKvxg2GivEkCxKUV go/util/service.go
added QmYhk1ajDk25HP7WhPDvSNEk5BPWS7BLTdWMBuXd6WfFBb go/ww
added QmTTKwXePzaeR8kS3kswW7qVQ6cFEKEfTZsc2GqE6PoGWJ go/ww.capnp
added QmXaBnfXSCEZDjsL7BTe8aQptGy4TXJkqNusCf1zKdNz22 go/ww.capnp.go
added QmdVDsNDP5VAyYrodz8QRP38muum6iYXeCPCMqXFm9Hu8w go/ww.go
added QmZ7dZXxGxriJCGtVrCWe5cYzaUQXLcYD2SfG7BraJqGHY go/ww_test.go
added Qmekz7CJ8V5EYUD87wC71Lna5QE2eUxb2H2gGNqFfrGnN8 go/cmd/export
added QmXy6KhyZSn9DpKs7neECb1FYBi9gp6XVMWtfk1wT9AENx go/cmd/run
added QmejzQJ2Fwks4AyxJGWoC9zAQKryFFf9y4BG79EsByvsS6 go/cmd
added QmQVE5xqVB1L2pK2LaRQVsBSz5XEcQZYWSvbEAdcZeWZWw go/examples/hello-world
added QmXvzhSyYxHntosXF2dRXKMSsSSuLtiEh97ezBZLwek8px go/examples
added QmQ6hzngTKT2CQjcRMuQyfmjhGAWKdeGp86wcdbSVDvN2P go/system/cmd/nop
added QmegPz76HZ3rJDaWCheN587bZerK1c8PDMdwsrgKPcXT7n go/system/cmd
added QmRecDLNaESeNY3oUFYZKK9ftdANBB8kuLaMdAXMD43yon go/system/testdata
added QmcJZeoNaJjqukQm2JKvi93hudombv96szXQ6jZ7DGEuYf go/system
added QmR3bBkJU5mBeKebhCpFWo36WGmrpNmGm5pDtanJpKa7uF go/test/libp2p
added QmSM2stNXuMTrYtU4mxvqmFJr9fDTUtSxtwZMp6CyqRWdR go/test
added QmNdWSBnMX7KEYvSESZj9Dos5tWmgRdpU4okZucHPh9X1H go/util
added QmTK1JSQnzsYfbrTWn359jtYBAeVAHtfknWa2Hgikinxjy go
```

The hash of the top-level `go` directory is the last line in the above.  If you cloned
the repo under a different name, you will see a different directory name and a different
hash. 

Copy it and assign it to `ROOT_CID`.

```bash
ROOT_CID=<hash-you-copied>  # e.g. QmTK1JSQnzsYfbrTWn359jtYBAeVAHtfknWa2Hgikinxjy
```

Now start a wetware node.  Note the path.
```bash
$ww run -load "/ipfs/QmWXvzjDBwjFcbLyiDZduqjk3RSxzfEkk5gn9uNX841XYR/system/cmd/nop"
```


### Building WASM Components
WASM artifacts are built by running `go generate ./...`.

WASM artifacts are into the main application or employed during testing.  The build-chain is managed by `go:generate` directives.  See `examples/hello-world/main.go` for one such example.

###  Application

The main application is built by running `go build cmd/main.go -o ww`.  New or modified WASM components will need to be (re)built beforehand.
