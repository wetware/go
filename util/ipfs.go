package util

import (
	"errors"
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	ma "github.com/multiformats/go-multiaddr"
)

// "path/filepath"

// "github.com/ipfs/boxo/files"

func LoadIPFSFromName(name string) (iface.CoreAPI, error) {
	switch name {
	case "local":
		return rpc.NewLocalApi()

	default:
		// // name is URI?
		// if u, err := url.ParseRequestURI(name); err == nil {
		// 	switch u.Scheme {
		// 	case "file":
		// 		return rpc.NewPathApi(u.Path)

		// 	default:
		// 		return rpc.NewURLApiWithClient(u.String(), http.DefaultClient)
		// 	}
		// }

		// Else attempt to load as multiaddr
		if a, err := ma.NewMultiaddr(name); err == nil {
			return rpc.NewApiWithClient(a, http.DefaultClient)
		}
	}

	return nil, errors.New("invalid name")
}

// // LoadByteCode loads the bytecode from the provided IPFS node.
// // If the node is a directory, it will walk the directory and
// // load the bytecode from the first file named "main.wasm". If
// // the node is a file, it will attempt to load the bytecode from
// // the file.  An error from Wazero usually indicates that the
// // bytecode is invalid.
// func LoadByteCode(ctx context.Context, node files.Node) ([]byte, error) {
// 	switch node := node.(type) {
// 	case files.File:
// 		return io.ReadAll(node)

// 	case files.Directory:
// 		return LoadByteCodeFromDir(ctx, node)

// 	default:
// 		panic(node) // unreachable
// 	}
// }

// func LoadByteCodeFromDir(ctx context.Context, d files.Directory) (b []byte, err error) {
// 	if err = files.Walk(d, func(fpath string, node files.Node) error {
// 		// Note:  early returns are used to short-circuit the walk. These
// 		// are signaled by returning errAbortWalk.

// 		// Already have the bytecode?
// 		if b != nil {
// 			return errAbortWalk
// 		}

// 		// File named "main.wasm"?
// 		if fname := filepath.Base(fpath); fname == "main.wasm" {
// 			if b, err = LoadByteCode(ctx, node); err != nil {
// 				return err
// 			}

// 			return errAbortWalk
// 		}

// 		// Keep walking.
// 		return nil
// 	}); err == errAbortWalk { // no error; we've just bottomed out
// 		err = nil
// 	}

// 	return
// }

// var errAbortWalk = errors.New("abort walk")
