package system

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var _ IPFS_Server = (*IPFSConfig)(nil)

type IPFSConfig struct {
	API iface.CoreAPI
}

func (s IPFSConfig) New() IPFS {
	return IPFS_ServerToClient(s)
}

func (s IPFSConfig) Add(ctx context.Context, call IPFS_add) error {
	args := call.Args()
	data, err := args.Data()
	if err != nil {
		return err
	}

	// Create a file from the data
	file := files.NewBytesFile(data)

	// Add the file to IPFS
	path, err := s.API.Unixfs().Add(ctx, file)
	if err != nil {
		return err
	}

	// Get the results and set the CID
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	return results.SetCid(path.RootCid().String())
}

func (s IPFSConfig) Cat(ctx context.Context, call IPFS_cat) error {
	args := call.Args()
	cidStr, err := args.Cid()
	if err != nil {
		return err
	}

	// Parse the CID and create a path
	c, err := cid.Decode(cidStr)
	if err != nil {
		return err
	}
	p := path.FromCid(c)

	// Get the file from IPFS
	file, err := s.API.Unixfs().Get(ctx, p)
	if err != nil {
		return err
	}

	// Read the file data
	var buf bytes.Buffer
	f, ok := file.(files.File)
	if !ok {
		return errors.New("node is not a file")
	}
	if _, err := io.Copy(&buf, f); err != nil {
		return err
	}

	// Get the results and set the data
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	return results.SetBody(buf.Bytes())
}

func (s IPFSConfig) Ls(ctx context.Context, call IPFS_ls) error {
	args := call.Args()
	pathStr, err := args.Path()
	if err != nil {
		return err
	}

	// Parse the path as a CID and create a path
	c, err := cid.Decode(pathStr)
	if err != nil {
		return err
	}
	p := path.FromCid(c)

	// Get the directory listing
	links, err := s.API.Unixfs().Ls(ctx, p)
	if err != nil {
		return err
	}

	// Get the results and create entries list
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	// Collect all links first
	var linkList []iface.DirEntry
	for link := range links {
		if link.Err != nil {
			continue // Skip links with errors
		}
		linkList = append(linkList, link)
	}

	entries, err := results.NewEntries(int32(len(linkList)))
	if err != nil {
		return err
	}

	// Convert links to entries
	for i, link := range linkList {
		entry := entries.At(i)
		entry.SetName(link.Name)
		entry.SetSize(link.Size)
		entry.SetCid(link.Cid.String())

		// Set entry type based on link type
		switch {
		case link.Type == iface.TDirectory:
			entry.SetType(EntryType_directory)
		case link.Type == iface.TSymlink:
			entry.SetType(EntryType_symlink)
		default:
			entry.SetType(EntryType_file)
		}
	}

	return results.SetEntries(entries)
}

func (s IPFSConfig) Stat(ctx context.Context, call IPFS_stat) error {
	args := call.Args()
	cidStr, err := args.Cid()
	if err != nil {
		return err
	}

	// Parse the CID
	c, err := cid.Decode(cidStr)
	if err != nil {
		return err
	}

	// Get the results and create NodeInfo
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	info, err := results.NewInfo()
	if err != nil {
		return err
	}

	info.SetCid(c.String())
	info.SetSize(0) // We'll set this to 0 for now as getting actual size requires more complex logic
	info.SetCumulativeSize(0)
	info.SetType("file") // Default to file type

	// Set links if any (for now empty as we're just getting data)
	_, err = info.NewLinks(0)
	if err != nil {
		return err
	}

	return results.SetInfo(info)
}

func (s IPFSConfig) Pin(ctx context.Context, call IPFS_pin) error {
	args := call.Args()
	cidStr, err := args.Cid()
	if err != nil {
		return err
	}

	// Parse the CID and create a path
	c, err := cid.Decode(cidStr)
	if err != nil {
		return err
	}
	p := path.FromCid(c)

	// Pin the CID
	err = s.API.Pin().Add(ctx, p)
	if err != nil {
		return err
	}

	// Get the results and set success
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	results.SetSuccess(true)
	return nil
}

func (s IPFSConfig) Unpin(ctx context.Context, call IPFS_unpin) error {
	args := call.Args()
	cidStr, err := args.Cid()
	if err != nil {
		return err
	}

	// Parse the CID and create a path
	c, err := cid.Decode(cidStr)
	if err != nil {
		return err
	}
	p := path.FromCid(c)

	// Unpin the CID
	err = s.API.Pin().Rm(ctx, p)
	if err != nil {
		return err
	}

	// Get the results and set success
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	results.SetSuccess(true)
	return nil
}

func (s IPFSConfig) Pins(ctx context.Context, call IPFS_pins) error {
	// Get all pinned CIDs
	pins, err := s.API.Pin().Ls(ctx)
	if err != nil {
		return err
	}

	// Get the results and create CIDs list
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	// Convert pins to string list
	var cidStrings []string
	for pin := range pins {
		if pin.Err() != nil {
			continue // Skip pins with errors
		}
		cidStrings = append(cidStrings, pin.Path().RootCid().String())
	}

	cids, err := results.NewCids(int32(len(cidStrings)))
	if err != nil {
		return err
	}

	for i, cidStr := range cidStrings {
		cids.Set(i, cidStr)
	}

	return results.SetCids(cids)
}

func (s IPFSConfig) Id(ctx context.Context, call IPFS_id) error {
	// Get peer ID
	id, err := s.API.Key().Self(ctx)
	if err != nil {
		return err
	}

	// Get the results and create PeerInfo
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	info, err := results.NewPeerInfo()
	if err != nil {
		return err
	}

	info.SetId(id.ID().String())

	// Get addresses
	addrs, err := s.API.Swarm().LocalAddrs(ctx)
	if err != nil {
		return err
	}

	addrStrings := make([]string, len(addrs))
	for i, addr := range addrs {
		addrStrings[i] = addr.String()
	}

	addresses, err := info.NewAddresses(int32(len(addrStrings)))
	if err != nil {
		return err
	}

	for i, addrStr := range addrStrings {
		addresses.Set(i, addrStr)
	}

	// Set protocols (empty for now as this info might not be readily available)
	_, err = info.NewProtocols(0)
	if err != nil {
		return err
	}

	return results.SetPeerInfo(info)
}

func (s IPFSConfig) Connect(ctx context.Context, call IPFS_connect) error {
	args := call.Args()
	addrStr, err := args.Addr()
	if err != nil {
		return err
	}

	// Parse the multiaddr
	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		return err
	}

	// Extract peer ID from multiaddr
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return err
	}

	// Connect to the peer
	err = s.API.Swarm().Connect(ctx, *peerInfo)
	if err != nil {
		return err
	}

	// Get the results and set success
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	results.SetSuccess(true)
	return nil
}

func (s IPFSConfig) Peers(ctx context.Context, call IPFS_peers) error {
	// Get connected peers
	peers, err := s.API.Swarm().Peers(ctx)
	if err != nil {
		return err
	}

	// Get the results and create PeerInfo list
	results, err := call.AllocResults()
	if err != nil {
		return err
	}

	peerList, err := results.NewPeerList(int32(len(peers)))
	if err != nil {
		return err
	}

	// Convert peers to PeerInfo
	for i, p := range peers {
		info := peerList.At(i)
		info.SetId(p.ID().String())

		// Get addresses for this peer (using empty list for now as Addrs() method might not be available)
		addrStrings := []string{}

		addresses, err := info.NewAddresses(int32(len(addrStrings)))
		if err != nil {
			return err
		}

		for j, addrStr := range addrStrings {
			addresses.Set(j, addrStr)
		}

		// Set protocols (empty for now as this info might not be readily available)
		_, err = info.NewProtocols(0)
		if err != nil {
			return err
		}
	}

	return results.SetPeerList(peerList)
}
