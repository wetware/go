package boot

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type IPFS struct {
	Client iface.CoreAPI
}

func (ipfs IPFS) FindPeers(ctx context.Context, ns string, opt ...discovery.Option) (<-chan peer.AddrInfo, error) {
	p, err := path.NewPath(ns)
	if err != nil {
		return nil, err
	}

	n, err := ipfs.ResolveNode(ctx, p)
	if err != nil {
		return nil, err
	}

	out := make(chan peer.AddrInfo)
	go func() {
		defer close(out)

		// Iterate through the IPLD node as a map
		// TODO:  add logging or some kind of improved error handling.
		for _, link := range n.Links() {
			// Get the linked node containing peer info
			peerNode, err := ipfs.Client.Dag().Get(ctx, link.Cid)
			if err != nil {
				continue
			}

			id, err := peer.Decode(link.Name)
			if err != nil {
				continue
			}
			info := peer.AddrInfo{ID: id}

			// Resolve the addresses from the peer node
			for i := 0; ; i++ {
				// Try to resolve each index as a string address
				v, rest, err := peerNode.Resolve([]string{fmt.Sprint(i)})
				if err != nil || len(rest) > 0 {
					break // end of list or error
				}

				addrStr, ok := v.(string)
				if !ok {
					continue
				}

				// Parse the multiaddr string
				addr, err := multiaddr.NewMultiaddr(addrStr)
				if err != nil {
					continue
				}

				info.Addrs = append(info.Addrs, addr)
			}

			// Send the peer info through the channel
			select {
			case out <- info:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (ipfs IPFS) ResolveNode(ctx context.Context, p path.Path) (format.Node, error) {
	parts := p.Segments()
	if len(parts) < 2 {
		return nil, fs.ErrInvalid
	}

	switch parts[0] {
	case "ipns":
		p, err := ipfs.Client.Name().Resolve(ctx, p.String())
		if err != nil {
			return nil, err
		}
		return ipfs.ResolveNode(ctx, p) // should be an ipfs path second time around

	case "ipfs":
		return ipfs.Client.ResolveNode(ctx, p)

	case "ipld":
		cid, err := cid.Decode(parts[1])
		if err != nil {
			return nil, err
		}
		return ipfs.Client.Dag().Get(ctx, cid)
	}

	return nil, fmt.Errorf("invalid node type: %s", parts[0])
}
