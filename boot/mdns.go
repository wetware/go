package boot

import (
	"context"
	"io"

	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/wetware/go/system"
)

type MDNS struct {
	NS  string
	Env *system.Env
}

func (m MDNS) String() string {
	return "mdns"
}

// Serve mDNS to discover peers on the local network
func (m MDNS) Serve(ctx context.Context) error {
	d, err := m.ListenAndServe()
	if err != nil {
		return err
	}
	defer d.Close() // defenisve; closed when context expires

	m.Env.Log().DebugContext(ctx, "service started",
		"service", m.String(),
		"ns", m.Name())
	<-ctx.Done()

	return d.Close()
}

func (m MDNS) ListenAndServe() (io.Closer, error) {
	d := mdns.NewMdnsService(m.Env.Host, m.Name(), m.Env)
	return d, d.Start()
}

func (m MDNS) Name() string {
	if m.NS == "" {
		return "ww"
	}

	return m.NS
}
