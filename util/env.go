package util

import (
	"os"
	"path/filepath"
	"strings"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/urfave/cli/v2"
)

func ExpandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

type Env struct {
	IPFS iface.CoreAPI
	Host host.Host
}

func (env *Env) Setup(c *cli.Context) (err error) {
	for _, bind := range []bindFunc{
		func(c *cli.Context) (err error) {
			env.IPFS, err = LoadIPFSFromName(c.String("ipfs"))
			return
		},
		// func(ctx *cli.Context) (err error) {
		// 	env.Host, err = LoadHost(env.IPFS)
		// 	return
		// },
	} {
		if err = bind(c); bind != nil {
			break
		}
	}

	return
}

type bindFunc func(*cli.Context) (err error)

func (env *Env) Teardown(c *cli.Context) (err error) {
	if env.Host != nil {
		err = env.Host.Close()
	}

	return
}
