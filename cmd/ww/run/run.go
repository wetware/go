package run

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/boxo/path"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/experimental/sysfs"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/urfave/cli/v2"
	"github.com/wetware/go/system"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "run",
		ArgsUsage: "<source-dir>",
		Before:    setup,
		Action:    Main,
		// After: teardown,
	}
}

type MountError struct {
	Path   string
	Status int
}

func (err MountError) Error() string {
	return fmt.Sprintf("mount %s: %d", err.Path, err.Status)
}

func Main(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.Exit("missing tarball argument", 1)
	}

	r := wazero.NewRuntimeWithConfig(c.Context, wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true))
	defer r.Close(c.Context)

	wasi, err := wasi_snapshot_preview1.Instantiate(c.Context, r)
	if err != nil {
		return err
	}
	defer wasi.Close(c.Context)

	// TODO:  load WASM src from IPFS
	bin, err := load(c)
	if err != nil {
		return err
	}

	cm, err := r.CompileModule(c.Context, bin)
	if err != nil {
		return err
	}
	defer cm.Close(c.Context)

	rc, path, fsConfig := validateMounts(c)
	if rc != 0 {
		return cli.Exit(&MountError{
			Path:   path,
			Status: rc,
		}, rc)
	}

	mod, err := r.InstantiateModule(c.Context, cm, wazero.NewModuleConfig().
		WithArgs(c.Args().Tail()...).
		// WithEnv().
		// WithName()
		WithFSConfig(fsConfig).
		WithStdin(c.App.Reader).     // FIXME(security):  require --with-stdin flag
		WithStdout(c.App.Writer).    // FIXME(security):  require --with-stdin flag
		WithStderr(c.App.ErrWriter). // FIXME(security):  require --with-stdin flag
		WithRandSource(rand.Reader)) // FIXME(security):  require --with-rand flag
	if err != nil {
		return err
	}
	defer mod.Close(c.Context)

	return nil
}

func load(c *cli.Context) ([]byte, error) {
	p, err := path.NewPath(c.Args().First())
	if err != nil {
		return nil, err
	}

	n, err := env.IPFS.Unixfs().Get(c.Context, p)
	if err != nil {
		return nil, err
	}
	defer n.Close()

	return util.LoadByteCode(c.Context, n)
}

func validateMounts(c *cli.Context) (rc int, rootPath string, config wazero.FSConfig) {
	config = wazero.NewFSConfig().
		WithFSMount(&system.FS{
			Ctx:  c.Context,
			IPFS: env.IPFS,
		}, "/")

	for _, mount := range c.StringSlice("mount") {
		if len(mount) == 0 {
			fmt.Fprintln(c.App.ErrWriter, "invalid mount: empty string")
			return 1, rootPath, config
		}

		readOnly := false
		if trimmed := strings.TrimSuffix(mount, ":ro"); trimmed != mount {
			mount = trimmed
			readOnly = true
		}

		// TODO: Support wasm paths with colon in them.
		var dir, guestPath string
		if clnIdx := strings.LastIndexByte(mount, ':'); clnIdx != -1 {
			dir, guestPath = mount[:clnIdx], mount[clnIdx+1:]
		} else {
			dir = mount
			guestPath = dir
		}

		// Eagerly validate the mounts as we know they should be on the host.
		if abs, err := filepath.Abs(dir); err != nil {
			fmt.Fprintf(c.App.ErrWriter, "invalid mount: path %q invalid: %v\n", dir, err)
			return 1, rootPath, config
		} else {
			dir = abs
		}

		if stat, err := os.Stat(dir); err != nil {
			fmt.Fprintf(c.App.ErrWriter, "invalid mount: path %q error: %v\n", dir, err)
			return 1, rootPath, config
		} else if !stat.IsDir() {
			fmt.Fprintf(c.App.ErrWriter, "invalid mount: path %q is not a directory\n", dir)
		}

		root := sysfs.DirFS(dir)
		if readOnly {
			root = &sysfs.ReadFS{FS: root}
		}

		config = config.(sysfs.FSConfig).WithSysFSMount(root, guestPath)

		if util.StripPrefixesAndTrailingSlash(guestPath) == "" {
			rootPath = dir
		}
	}
	return 0, rootPath, config
}
