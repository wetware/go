package run

import (
	"os"
	"path/filepath"
	"strings"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/wetware/go/util"
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

func (env *Env) Boot(addr string) (err error) {
	for _, bind := range []func() error{
		func() (err error) {
			env.IPFS, err = util.LoadIPFSFromName(addr)
			return
		},
		func() (err error) {
			env.Host, err = HostConfig{IPFS: env.IPFS}.New()
			return
		},
	} {
		if err = bind(); err != nil {
			break
		}
	}

	return
}

func (env *Env) Close() error {
	if env.Host != nil {
		return env.Host.Close()
	}

	return nil
}

type HostConfig struct {
	IPFS    iface.CoreAPI
	Options []libp2p.Option
}

func (cfg HostConfig) New() (host.Host, error) {
	return libp2p.New(cfg.CombinedOptions()...)
}

func (cfg HostConfig) CombinedOptions() []libp2p.Option {
	return append(cfg.DefaultOptions(), cfg.Options...)
}

func (c HostConfig) DefaultOptions() []libp2p.Option {
	return []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2020"),
	}
}
