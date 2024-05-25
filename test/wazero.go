//go:generate mockgen -source=wazero.go -destination=wazero/wazero.go -package=test_wazero

package test

import "github.com/tetratelabs/wazero/api"

type (
	Module   interface{ api.Module }
	Function interface{ api.Function }
)
