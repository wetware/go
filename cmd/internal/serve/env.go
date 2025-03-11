package serve

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	insecure_rand "math/rand"
	"net/http"
	"os"
	"time"

	"github.com/ipfs/kubo/client/rpc"
	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-kad-dht/dual"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/mr-tron/base58"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
	syncutils "github.com/wetware/go/util/sync"
	"go.uber.org/multierr"
	"golang.org/x/time/rate"
)

func teardown(c *cli.Context) error {
	// host was started?
	if env.Host != nil {
		return env.Host.Close()
	}

	return nil
}

func setup(c *cli.Context) (err error) {
	// Set up IPFS and libp2p
	////
	if env.IPFS, err = newIPFSClient(c); err != nil {
		return
	} else if env.Host, err = newLibp2pHost(c); err != nil {
		return
	}

	// Initialize a Dual DHT, which maintains two separate Kademlia routing tables:
	// one for the Wide Area Network (WAN) and another for the Local Area Network
	// (LAN). This separation allows for efficient routing in both local and global
	// contexts.
	//
	// The WAN DHT is bootstrapped with public bootstrap peers - well-known nodes
	// that help new peers join the network by providing initial routing table
	// entries. These peers are typically operated by IPFS infrastructure providers
	// and are accessible from anywhere on the internet.
	//
	// The LAN DHT is bootstrapped with private bootstrap peers - nodes that are
	// only accessible within the local network. This enables peer discovery and
	// content routing to work efficiently in isolated/local network environments,
	// without having to rely on public infrastructure.
	//
	// Both DHTs shuffle their bootstrap peers to prevent hotspots and ensure even
	// load distribution across the bootstrap nodes. User-provided bootstrap addresses
	// (-addr flag) are added to both DHTs to support custom network topologies.
	////
	env.DHT, err = dual.New(c.Context, env.Host,
		dual.WanDHTOption(dht.BootstrapPeersFunc(func() []peer.AddrInfo {
			public := append(
				env.PublicBootstrapPeers(),
				dht.GetDefaultBootstrapPeerAddrInfos()...)
			insecure_rand.Shuffle(len(public), func(i, j int) {
				public[i], public[j] = public[j], public[i]
			})

			return append(addrs(c), public...)
		})),
		dual.LanDHTOption(dht.BootstrapPeersFunc(func() []peer.AddrInfo {
			private := env.PrivateBootstrapPeers()
			insecure_rand.Shuffle(len(private), func(i, j int) {
				private[i], private[j] = private[j], private[i]
			})

			return append(addrs(c), private...)
		})))
	if err != nil {
		return
	}

	// Wrap the host in a routed host, which intercepts all network operations
	// and uses the DHT for peer routing. This enables automatic peer discovery
	// and routing through the DHT when direct connections aren't available.
	// When the host attempts to dial a peer, the routed host will first check
	// if it has a direct connection. If not, it will query the DHT to find
	// the peer's addresses before attempting the connection.
	////
	env.Host = routedhost.Wrap(env.Host, env.DHT)
	env.Host.Peerstore().AddAddrs(
		env.Host.ID(),
		env.Host.Addrs(),
		peerstore.PermanentAddrTTL)

	return bootstrap(c)
}

// bootstrap attempts to establish initial connectivity with the P2P network by
// connecting to bootstrap peers and initializing the DHT. The process works as
// follows:
//
// 1. For each bootstrap address provided via command line flags:
//   - Launch a concurrent connection attempt using env.Host.Connect()
//   - Use a syncutils.Any join strategy, which succeeds if ANY connection works
//   - This provides fault tolerance - we only need one working bootstrap peer
//
// 2. Wait for all connection attempts to complete:
//   - If no connections succeeded, return the error from syncutils.Any.Wait()
//   - Otherwise proceed with DHT bootstrap
//
// 3. Call env.DHT.Bootstrap() to:
//   - Initialize the DHT routing tables
//   - Begin the background DHT maintenance processes
//   - Enable peer discovery and content routing
func bootstrap(c *cli.Context) error {
	var peers []peer.AddrInfo
	peers = append(peers, addrs(c)...)                               // User-provided peers
	peers = append(peers, env.PublicBootstrapPeers()...)             // Public peers
	peers = append(peers, dht.GetDefaultBootstrapPeerAddrInfos()...) // Default IPFS peers

	// Try to connect to bootstrap peers concurrently with rate limiting
	limiter := rate.NewLimiter(rate.Limit(32), 8) // Allow up to 8 concurrent connection attempts
	var join syncutils.Any                        // join strategy: any successful connection => ok
	for _, info := range peers {
		info := info // capture for closure
		join.Go(func() error {
			if err := limiter.Wait(c.Context); err != nil {
				return err
			}

			return env.Host.Connect(c.Context, info)
		})
	}

	if err := join.Wait(); err != nil {
		return fmt.Errorf("failed to connect to any bootstrap peers: %w", err)
	}

	// Now that we have connections, bootstrap the DHT
	return env.DHT.Bootstrap(c.Context)
}

// newIPFSClient creates and returns an IPFS CoreAPI client based on the
// configuration provided. The client can be configured in two ways:
//
// 1. Default local client:
//   - Used when no "ipfs" flag is set
//   - Creates a new local API client with default settings
//   - Suitable for embedded IPFS nodes
//
// 2. Remote client:
//   - Used when "ipfs" flag contains a multiaddr
//   - Parses the multiaddr to determine connection endpoint
//   - Creates HTTP client to connect to remote IPFS node
//   - Enables integration with external IPFS daemons
func newIPFSClient(c *cli.Context) (iface.CoreAPI, error) {
	if !c.IsSet("ipfs") {
		return rpc.NewLocalApi()
	}

	a, err := ma.NewMultiaddr(c.String("ipfs"))
	if err != nil {
		return nil, err
	}

	return rpc.NewApiWithClient(a, http.DefaultClient)
}

// identity configures libp2p host identity using a private key. The key can be provided
// in several ways:
//
// 1. As a base58-encoded string via the "privkey" flag:
//   - Attempts to decode the string as base58
//   - If successful, unmarshals it as a private key
//   - Returns a libp2p.Identity option with the decoded key
//
// 2. As a file path via the "privkey" flag:
//   - If base58 decode fails, treats the string as a file path
//   - Reads the private key bytes from the file
//   - Unmarshals the bytes into a private key
//   - Returns a libp2p.Identity option with the loaded key
//
// 3. Auto-generated if no key is provided:
//   - Generates a new Ed25519 private key
//   - Logs that a new identity was generated
//   - Returns a libp2p.Identity option with the new key
//
// The function handles errors by wrapping them in a libp2p.Option that will
// raise the error when applied to the libp2p config.
func identity(c *cli.Context) libp2p.Option {
	// Check if a private key was provided via the "privkey" flag
	if pkStr := c.String("privkey"); pkStr != "" {
		// First attempt: try to decode the input as a base58-encoded key.
		// This allows passing keys directly via environment variables or command line.
		pkBytes, err := base58.Decode(pkStr)
		if err == nil {
			priv, err := crypto.UnmarshalPrivateKey(pkBytes)
			if err != nil {
				return erroptf("failed to unmarshal base58-encoded private key: %w", err)
			}
			return libp2p.Identity(priv)
		}

		// Second attempt: try to read the input as a file path.
		// This allows storing keys in files for better security.
		pkBytes, err = os.ReadFile(pkStr)
		if err != nil {
			return erroptf("failed to read private key file: %w", err)
		}

		// Try to unmarshal the file contents as a private key
		pk, err := crypto.UnmarshalPrivateKey(pkBytes)
		if err != nil {
			return erroptf("failed to unmarshal private key from file: %w", err)
		}

		// Successfully loaded key from file
		return libp2p.Identity(pk)
	}

	// No key provided - generate a new Ed25519 keypair.
	// Ed25519 is chosen for its security and performance characteristics.
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return erroptf("failed to generate Ed25519 key: %w", err)
	}

	// Log that we generated a new identity, as this is important for debugging
	slog.InfoContext(c.Context, "generated new identity using Ed25519")
	return libp2p.Identity(priv)
}

// erroptf returns a libp2p.Option that will raise a formatted error when applied.
// This is a helper function used to wrap errors that occur during host configuration
// in a way that defers the error until the option is actually applied.
//
// Parameters:
//   - format: A format string for the error message
//   - args: Arguments to be formatted into the error message
//
// Returns a libp2p.Option that will return the formatted error when applied to
// a libp2p.Config.
func erroptf(format string, args ...any) libp2p.Option {
	err := fmt.Errorf(format, args...)
	return erropt(err)
}

// erropt returns a libp2p.Option that will raise an error when applied.
// This is a helper function used to wrap errors that occur during host configuration
// in a way that defers the error until the option is actually applied.
//
// Parameters:
//   - err: The error to be returned when the option is applied
//
// Returns a libp2p.Option that will return the error when applied to
// a libp2p.Config.
func erropt(err error) libp2p.Option {
	return func(_ *libp2p.Config) error {
		return err
	}
}

func newLibp2pHost(c *cli.Context) (host.Host, error) {
	listenAddrs, err := listenAddrs(c)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.WithDialTimeout(time.Second * 15),
		identity(c)}

	return libp2p.New(opts...)
}

func listenAddrs(c *cli.Context) ([]ma.Multiaddr, error) {
	var addrs []ma.Multiaddr
	var errs []error
	for _, a := range c.StringSlice("listen") {
		if m, err := ma.NewMultiaddr(a); err != nil {
			errs = append(errs, err)
		} else {
			addrs = append(addrs, m)
		}
	}

	return addrs, multierr.Combine(errs...)
}

// addrs returns bootstrap addresses parsed from args
func addrs(c *cli.Context) []peer.AddrInfo {
	ps := map[peer.ID][]ma.Multiaddr{}
	for _, a := range c.StringSlice("dial") {
		m, err := ma.NewMultiaddr(a)
		if err != nil {
			slog.DebugContext(c.Context, "failed to parse multiaddr",
				"addr", a,
				"reason", err)
			continue
		}

		s, err := m.ValueForProtocol(ma.P_P2P)
		if err != nil {
			slog.DebugContext(c.Context, "failed to parse value for protocol",
				"proto", "p2p",
				"addr", a,
				"reason", err)
			continue
		}
		id, err := peer.Decode(s)
		if err != nil {
			slog.DebugContext(c.Context, "failed to decode peer ID",
				"addr", a,
				"reason", err)
			continue
		}

		addr := m.Decapsulate(ma.StringCast("/p2p/" + s))
		ps[id] = append(ps[id], addr)
	}

	var dial []peer.AddrInfo
	for id, addrs := range ps {
		if len(addrs) > 0 {
			dial = append(dial, peer.AddrInfo{
				ID:    id,
				Addrs: addrs,
			})
		}
	}

	return dial
}
