package idgen

import (
	"crypto/rand"
	"fmt"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mr-tron/base58"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "idgen",
		Usage: "generate a new Ed25519 private key",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "raw",
				Usage:   "output raw bytes instead of base58 encoding",
				Aliases: []string{"r"},
			},
		},
		Action: generate(),
	}
}

func generate() cli.ActionFunc {
	return func(c *cli.Context) error {
		priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}

		bytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			return fmt.Errorf("failed to marshal private key: %w", err)
		}

		if c.Bool("raw") {
			_, err = c.App.Writer.Write(bytes)
			return err
		}

		// Default to base58 encoding for human-readable output
		encoded := base58.Encode(bytes)
		_, err = fmt.Fprintln(c.App.Writer, encoded)
		return err
	}
}
