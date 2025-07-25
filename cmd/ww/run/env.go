package run

import (
	"io"

	iface "github.com/ipfs/kubo/core/coreiface"
	"github.com/urfave/cli/v2"
)

func maybeIPFS(c *cli.Context) iface.CoreAPI {
	if c.Bool("with-full-rights") || c.Bool("with-ipfs") {
		return env.IPFS
	}

	return nil
}

func maybeConsoleWriter(c *cli.Context) io.Writer {
	if c.Bool("with-full-rights") || c.Bool("with-console") {
		return c.App.Writer
	}

	return nil
}

func maybeExec(c *cli.Context) bool {
	if c.Bool("with-full-rights") || c.Bool("with-exec") {
		return true
	}

	return false
}

// // identity configures libp2p host identity using a private key. The key can be provided
// // in several ways:
// //
// // 1. As a base58-encoded string via the "privkey" flag:
// //   - Attempts to decode the string as base58
// //   - If successful, unmarshals it as a private key
// //   - Returns a libp2p.Identity option with the decoded key
// //
// // 2. As a file path via the "privkey" flag:
// //   - If base58 decode fails, treats the string as a file path
// //   - Reads the private key bytes from the file
// //   - Unmarshals the bytes into a private key
// //   - Returns a libp2p.Identity option with the loaded key
// //
// // 3. Auto-generated if no key is provided:
// //   - Generates a new Ed25519 private key
// //   - Logs that a new identity was generated
// //   - Returns a libp2p.Identity option with the new key
// //
// // The function handles errors by wrapping them in a libp2p.Option that will
// // raise the error when applied to the libp2p config.
// func identity(c *cli.Context) libp2p.Option {
// 	// Check if a private key was provided via the "privkey" flag
// 	if pkStr := c.String("privkey"); pkStr != "" {
// 		// First attempt: try to decode the input as a base58-encoded key.
// 		// This allows passing keys directly via environment variables or command line.
// 		pkBytes, err := base58.Decode(pkStr)
// 		if err == nil {
// 			priv, err := crypto.UnmarshalPrivateKey(pkBytes)
// 			if err != nil {
// 				return erroptf("failed to unmarshal base58-encoded private key: %w", err)
// 			}
// 			return libp2p.Identity(priv)
// 		}

// 		// Second attempt: try to read the input as a file path.
// 		// This allows storing keys in files for better security.
// 		pkBytes, err = os.ReadFile(pkStr)
// 		if err != nil {
// 			return erroptf("failed to read private key file: %w", err)
// 		}

// 		// Try to unmarshal the file contents as a private key
// 		pk, err := crypto.UnmarshalPrivateKey(pkBytes)
// 		if err != nil {
// 			return erroptf("failed to unmarshal private key from file: %w", err)
// 		}

// 		// Successfully loaded key from file
// 		return libp2p.Identity(pk)
// 	}

// 	// No key provided - generate a new Ed25519 keypair.
// 	// Ed25519 is chosen for its security and performance characteristics.
// 	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
// 	if err != nil {
// 		return erroptf("failed to generate Ed25519 key: %w", err)
// 	}

// 	// Log that we generated a new identity, as this is important for debugging
// 	slog.InfoContext(c.Context, "generated new identity using Ed25519")
// 	return libp2p.Identity(priv)
// }

// // erroptf returns a libp2p.Option that will raise a formatted error when applied.
// // This is a helper function used to wrap errors that occur during host configuration
// // in a way that defers the error until the option is actually applied.
// //
// // Parameters:
// //   - format: A format string for the error message
// //   - args: Arguments to be formatted into the error message
// //
// // Returns a libp2p.Option that will return the formatted error when applied to
// // a libp2p.Config.
// func erroptf(format string, args ...any) libp2p.Option {
// 	err := fmt.Errorf(format, args...)
// 	return erropt(err)
// }

// // erropt returns a libp2p.Option that will raise an error when applied.
// // This is a helper function used to wrap errors that occur during host configuration
// // in a way that defers the error until the option is actually applied.
// //
// // Parameters:
// //   - err: The error to be returned when the option is applied
// //
// // Returns a libp2p.Option that will return the error when applied to
// // a libp2p.Config.
// func erropt(err error) libp2p.Option {
// 	return func(_ *libp2p.Config) error {
// 		return err
// 	}
// }

// func newLibp2pHost(c *cli.Context) (host.Host, error) {
// 	listenAddrs, err := listenAddrs(c)
// 	if err != nil {
// 		return nil, err
// 	}

// 	opts := []libp2p.Option{
// 		libp2p.ListenAddrs(listenAddrs...),
// 		libp2p.WithDialTimeout(time.Second * 15),
// 		identity(c)}

// 	return libp2p.New(opts...)
// }

// func listenAddrs(c *cli.Context) ([]ma.Multiaddr, error) {
// 	var addrs []ma.Multiaddr
// 	var errs []error
// 	for _, a := range c.StringSlice("listen") {
// 		if m, err := ma.NewMultiaddr(a); err != nil {
// 			errs = append(errs, err)
// 		} else {
// 			addrs = append(addrs, m)
// 		}
// 	}

// 	return addrs, multierr.Combine(errs...)
// }

// // addrs returns bootstrap addresses parsed from args
// func addrs(c *cli.Context) []peer.AddrInfo {
// 	ps := map[peer.ID][]ma.Multiaddr{}
// 	for _, a := range c.StringSlice("dial") {
// 		m, err := ma.NewMultiaddr(a)
// 		if err != nil {
// 			slog.DebugContext(c.Context, "failed to parse multiaddr",
// 				"addr", a,
// 				"reason", err)
// 			continue
// 		}

// 		s, err := m.ValueForProtocol(ma.P_P2P)
// 		if err != nil {
// 			slog.DebugContext(c.Context, "failed to parse value for protocol",
// 				"proto", "p2p",
// 				"addr", a,
// 				"reason", err)
// 			continue
// 		}
// 		id, err := peer.Decode(s)
// 		if err != nil {
// 			slog.DebugContext(c.Context, "failed to decode peer ID",
// 				"addr", a,
// 				"reason", err)
// 			continue
// 		}

// 		addr := m.Decapsulate(ma.StringCast("/p2p/" + s))
// 		ps[id] = append(ps[id], addr)
// 	}

// 	var dial []peer.AddrInfo
// 	for id, addrs := range ps {
// 		if len(addrs) > 0 {
// 			dial = append(dial, peer.AddrInfo{
// 				ID:    id,
// 				Addrs: addrs,
// 			})
// 		}
// 	}

// 	return dial
// }
