//go:generate capnp compile -I$GOPATH/src/capnproto.org/go/capnp/std -ogo capnp

package anchor

import (
	"context"

	format "github.com/ipfs/go-ipld-format"
)

var _ Node_Server = (*DefaultNode)(nil)
var _ Block_Server = (*DefaultNode)(nil)
var _ Resolver_Server = (*DefaultNode)(nil)

type DefaultNode struct {
	format.Node
}

func (a DefaultNode) Cid(ctx context.Context, call Block_cid) error {
	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	return res.SetCid(a.Node.Cid().Bytes())
}

func (a DefaultNode) ResolveLink(ctx context.Context, call Node_resolveLink) error {
	call.Go()

	path, err := call.Args().Path()
	if err != nil {
		return err
	}

	p := make([]string, path.Len())
	for i := range p {
		p[i], err = path.At(i)
		if err != nil {
			return err
		}
	}

	link, remaining, err := a.Node.ResolveLink(p)
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	resLink, err := res.NewLink()
	if err != nil {
		return err
	}

	if err := resLink.SetName(link.Name); err != nil {
		return err
	}

	resLink.SetSize(link.Size)

	if err := resLink.SetCid(link.Cid.Bytes()); err != nil {
		return err
	}

	size := int32(len(remaining))
	remainingPath, err := res.NewRemainingPath(size)
	if err != nil {
		return err
	}
	for i, p := range remaining {
		if err = remainingPath.Set(i, p); err != nil {
			break
		}
	}
	return err
}

func (a DefaultNode) Copy(ctx context.Context, call Node_copy) error {
	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	clone := DefaultNode{Node: a.Node.Copy()}
	node := Node_ServerToClient(clone)
	return res.SetNode(node)
}

func (a DefaultNode) Links(ctx context.Context, call Node_links) error {
	links := a.Node.Links()

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	size := int32(len(links))
	resLinks, err := res.NewLinks(size)
	if err != nil {
		return err
	}

	for i, link := range links {
		rl := resLinks.At(i)
		if err = rl.SetName(link.Name); err != nil {
			break
		}

		rawBytes := link.Cid.Bytes()
		if err = rl.SetCid(rawBytes); err != nil {
			break
		}
	}

	return err
}

func (a DefaultNode) Stat(ctx context.Context, call Node_stat) error {
	ns, err := a.Node.Stat()
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	stat, err := res.NewStat()
	if err != nil {
		return err
	}
	stat.SetBlockSize(uint64(ns.BlockSize))
	stat.SetLinksSize(uint64(ns.LinksSize))
	stat.SetDataSize(uint64(ns.DataSize))
	stat.SetCumulativeSize(uint64(ns.CumulativeSize))
	return stat.SetHash(a.Node.Cid().Bytes())
}

func (a DefaultNode) Size(ctx context.Context, call Node_size) error {
	size, err := a.Node.Size()
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err == nil {
		res.SetSize(size)
	}
	return err
}

func (a DefaultNode) RawData(ctx context.Context, call Block_rawData) error {
	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	body := a.Node.RawData()
	return res.SetData(body)
}

func (a DefaultNode) ResolvePath(ctx context.Context, call Resolver_resolvePath) error {
	path, err := call.Args().Path()
	if err != nil {
		return err
	}

	p := make([]string, path.Len())
	for i := range p {
		p[i], err = path.At(i)
		if err != nil {
			return err
		}
	}

	v, remainder, err := a.Node.Resolve(p)
	if err != nil {
		return err
	}

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	server := DefaultNode{Node: v.(format.Node)}
	node := Node_ServerToClient(server)
	if err := res.SetNode(node); err != nil {
		return err
	}

	remainingPath, err := res.NewRemainingPath(int32(len(remainder)))
	if err != nil {
		return err
	}
	for i, p := range remainder {
		if err = remainingPath.Set(i, p); err != nil {
			break
		}
	}
	return err
}

func (a DefaultNode) Tree(ctx context.Context, call Resolver_tree) error {
	path, err := call.Args().Path()
	if err != nil {
		return err
	}
	depth := call.Args().Depth()

	tree := a.Node.Tree(path, int(depth))

	res, err := call.AllocResults()
	if err != nil {
		return err
	}

	ps, err := res.NewPaths(int32(len(tree)))
	if err != nil {
		return err
	}
	for i, path := range tree {
		if err = ps.Set(i, path); err != nil {
			break
		}
	}
	return err
}
