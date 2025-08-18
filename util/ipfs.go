package util

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	ma "github.com/multiformats/go-multiaddr"
)

func LoadIPFSFromName(name string) (iface.CoreAPI, error) {
	if name == "" {
		name = rpc.DefaultPathRoot
	}

	// attempt to load as multiaddr
	if a, err := ma.NewMultiaddr(name); err == nil {
		if api, err := rpc.NewApiWithClient(a, http.DefaultClient); err == nil {
			return api, nil
		}
	}

	// attempt to load as URL
	if u, err := url.ParseRequestURI(name); err == nil {
		return rpc.NewURLApiWithClient(u.String(), http.DefaultClient)
	}

	if api, err := rpc.NewPathApi(name); err == nil {
		return api, nil
	}

	return nil, fmt.Errorf("invalid ipfs addr: %s", name)
}
