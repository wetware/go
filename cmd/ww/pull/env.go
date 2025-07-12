package pull

import (
	"net/http"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
)

var env struct {
	IPFS iface.CoreAPI
}

func setup(c *cli.Context) (err error) {
	env.IPFS, err = newIPFSClient(c)
	return
}

// newIPFSClient creates and returns an IPFS CoreAPI client based on the
// configuration provided. The client can be configured in two ways:
//
// 1. Default local client:
//   - Used when no "ipfs" flag is set
//   - Creates a new local API client with default settings
//   - Suitable for embedded IPFS nodes
//
// 2. Remote client:
//   - Used when "ipfs" flag contains a multiaddr
//   - Parses the multiaddr to determine connection endpoint
//   - Creates HTTP client to connect to remote IPFS node
//   - Enables integration with external IPFS daemons
func newIPFSClient(c *cli.Context) (iface.CoreAPI, error) {
	if !c.IsSet("ipfs") {
		return rpc.NewLocalApi()
	}

	a, err := ma.NewMultiaddr(c.String("ipfs"))
	if err != nil {
		return nil, err
	}

	return rpc.NewApiWithClient(a, http.DefaultClient)
}
