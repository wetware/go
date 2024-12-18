package serve

import (
	"github.com/hashicorp/go-memdb"
	"github.com/wetware/go/glia"
)

var _ glia.Router = (*Router)(nil)

type Router struct {
	DB *memdb.MemDB
}

func (r Router) GetProc(pid string) (glia.Proc, error) {
	tx := r.DB.Txn(false)
	defer tx.Abort()

	v, err := tx.First("proc", "id", pid)
	if err != nil {
		return nil, err
	} else if v == nil {
		return nil, glia.ErrNotFound
	}

	return v.(glia.Proc), nil
}
