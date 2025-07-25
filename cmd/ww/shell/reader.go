package shell

import (
	"io"

	"github.com/spy16/slurp/reader"
	"github.com/spy16/slurp/repl"
	"github.com/wetware/go/lang"
	"github.com/wetware/go/system"
)

// readerFactory creates a custom reader factory for the shell
func readerFactory(ipfs system.IPFS) repl.ReaderFactoryFunc {
	return func(r io.Reader) *reader.Reader {
		// Create a reader with hex support
		rd := lang.NewReaderWithHexSupport(r)

		// Set up the Unix path reader macro for '/' character
		rd.SetMacro('/', false, lang.UnixPathReader())
		// Set up the custom list reader macro for '(' character
		rd.SetMacro('(', false, lang.ListReader(ipfs))

		return rd
	}
}
