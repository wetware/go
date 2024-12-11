package util

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/ipfs/boxo/files"
)

// LoadByteCode loads the bytecode from the provided IPFS node.
// If the node is a directory, it will walk the directory and
// load the bytecode from the first file named "main.wasm". If
// the node is a file, it will attempt to load the bytecode from
// the file.  An error from Wazero usually indicates that the
// bytecode is invalid.
func LoadByteCode(ctx context.Context, node files.Node) ([]byte, error) {
	switch node := node.(type) {
	case files.File:
		return io.ReadAll(node)

	case files.Directory:
		return LoadByteCodeFromDir(ctx, node)

	default:
		panic(node) // unreachable
	}
}

func LoadByteCodeFromDir(ctx context.Context, d files.Directory) (b []byte, err error) {
	if err = files.Walk(d, func(fpath string, node files.Node) error {
		// Note:  early returns are used to short-circuit the walk. These
		// are signaled by returning errAbortWalk.

		// Already have the bytecode?
		if b != nil {
			return errAbortWalk
		}

		// File named "main.wasm"?
		if fname := filepath.Base(fpath); fname == "main.wasm" {
			if b, err = LoadByteCode(ctx, node); err != nil {
				return err
			}

			return errAbortWalk
		}

		// Keep walking.
		return nil
	}); err == errAbortWalk { // no error; we've just bottomed out
		err = nil
	}

	return
}

var errAbortWalk = errors.New("abort walk")
