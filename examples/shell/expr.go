package main

import (
	"context"
	"runtime"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/libp2p/go-libp2p/core/record"
	"github.com/spy16/slurp/core"
	"github.com/wetware/go/system"
)

// ImportExpr implements core.Expr for import statements
type ImportExpr struct {
	Client   system.Importer
	Envelope *record.Envelope
	Timeout  time.Duration
}

func (e ImportExpr) NewEvalContext() (context.Context, context.CancelFunc) {
	if e.Timeout < 0 {
		return context.WithCancel(context.Background())
	} else if e.Timeout == 0 {
		e.Timeout = 10 * time.Second
	}

	return context.WithTimeout(context.Background(), e.Timeout)
}

func (e ImportExpr) Eval(env core.Env) (core.Any, error) {
	ctx, cancel := e.NewEvalContext()
	defer cancel()

	f, release := e.Client.Import(ctx, e.SetEnvelope)
	runtime.SetFinalizer(f.Future, func(*capnp.Future) {
		release()
	})

	return f, nil
}

func (e ImportExpr) SetEnvelope(p system.Importer_import_Params) error {
	b, err := e.Envelope.Marshal()
	if err != nil {
		return err
	}

	return p.SetEnvelope(b)
}
