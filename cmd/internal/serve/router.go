package serve

import (
	"github.com/hashicorp/go-memdb"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/system"
)

type Router struct {
	DB *memdb.MemDB
}

func (r Router) GetProc(pid string) (system.Proc, error) {
	tx := r.DB.Txn(false)
	defer tx.Abort()

	v, err := tx.First("proc", "id", pid)
	if err != nil {
		return nil, err
	} else if v == nil {
		return nil, glia.ErrNotFound
	}

	return v.(system.Proc), nil
}
