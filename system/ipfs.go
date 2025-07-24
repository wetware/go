package system

import (
	context "context"

	iface "github.com/ipfs/kubo/core/coreiface"
)

var _ IPFS_Server = (*IPFS_Provider)(nil)

type IPFS_Provider struct {
	API iface.CoreAPI
}

func (s *IPFS_Provider) Add(ctx context.Context, call IPFS_add) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Cat(ctx context.Context, call IPFS_cat) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Ls(ctx context.Context, call IPFS_ls) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Stat(ctx context.Context, call IPFS_stat) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Pin(ctx context.Context, call IPFS_pin) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Connect(ctx context.Context, call IPFS_connect) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Peers(ctx context.Context, call IPFS_peers) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Unpin(ctx context.Context, call IPFS_unpin) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Pins(ctx context.Context, call IPFS_pins) error {
	panic("NOT IMPLEMENTED")
}

func (s *IPFS_Provider) Id(ctx context.Context, call IPFS_id) error {
	panic("NOT IMPLEMENTED")
}
