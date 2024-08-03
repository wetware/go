package guest

import (
	"context"
	"crypto/rand"
	"io"
	"os"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type Config struct {
	IPFS iface.CoreAPI
	Root path.Path
	Sys  interface {
		Stdin() io.Reader
	}
}

// Compile and instantiate the guest module from the namespace path.
// Note that CompiledModule is produced in an intermediate step, and
// that it is not closed until r is closed.
func (c Config) Instanatiate(ctx context.Context, r wazero.Runtime) (api.Module, error) {
	cm, err := c.CompileModule(ctx, r)
	if err != nil {
		return nil, err
	}

	return r.InstantiateModule(ctx, cm, wazero.NewModuleConfig().
		// WithName().
		// WithArgs().
		// WithEnv().
		WithRandSource(rand.Reader).
		WithFS(FS{UNIX: c.IPFS.Unixfs(), Root: c.Root}).
		// WithFSConfig().
		// WithStartFunctions(). // remove _start so that we can call it later
		WithStdin(c.Sys.Stdin()).
		WithStdout(os.Stdout). // FIXME
		WithStderr(os.Stderr). // FIXME
		WithSysNanotime())
}

func (c Config) CompileModule(ctx context.Context, r wazero.Runtime) (wazero.CompiledModule, error) {
	bytecode, err := c.LoadByteCode(ctx)
	if err != nil {
		return nil, err
	}

	return r.CompileModule(ctx, bytecode)
}

func (c Config) LoadByteCode(ctx context.Context) ([]byte, error) {
	n, err := c.IPFS.Unixfs().Get(ctx, c.Root)
	if err != nil {
		return nil, err
	}
	defer n.Close()

	// FIXME:  address the obvious DoS vector
	return io.ReadAll(n.(files.File))
}
