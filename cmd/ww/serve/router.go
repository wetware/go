package serve

import (
	"github.com/hashicorp/go-memdb"
	"github.com/wetware/go/glia"
	"github.com/wetware/go/system"
)

// Router provides an interface for routing messages between processes in the system.
// It uses an in-memory database with software transactional memory (STM) semantics
// to safely coordinate access to process state.
type Router struct {
	// DB is the underlying MemDB instance that provides STM capabilities.
	// All process state is stored in tables within this database.
	DB *memdb.MemDB
}

// GetProc looks up a process by its PID and returns it if found.
// The lookup is performed as a read-only transaction to ensure consistency.
//
// The process lookup works as follows:
// 1. Start a read-only transaction
// 2. Query the "proc" table using the "id" index
// 3. Return the first matching process, or an error if not found
//
// Parameters:
//   - pid: The process ID to look up
//
// Returns:
//   - The process if found
//   - glia.ErrNotFound if no process exists with the given PID
//   - Any other errors that occur during the lookup
func (r Router) GetProc(pid string) (system.Proc, error) {
	// Start a read-only transaction. This ensures we get a consistent view
	// of the database without blocking writers. The transaction is automatically
	// aborted when we're done via defer.
	tx := r.DB.Txn(false)
	defer tx.Abort()

	// Look up the process in the "proc" table using the "id" index.
	// This is an O(1) operation since "id" is the primary key.
	v, err := tx.First("proc", "id", pid)
	if err != nil {
		return nil, err
	} else if v == nil {
		return nil, glia.ErrNotFound
	}

	// Type assert the result to system.Proc since we know all values
	// in the "proc" table implement this interface
	return v.(system.Proc), nil
}
