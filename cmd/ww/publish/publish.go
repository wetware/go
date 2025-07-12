package publish

import (
	"fmt"
	"os/exec"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "publish",
		Aliases:   []string{"pub"},
		Usage:     "push an OCI image tarball to IPFS (as UnixFS)",
		ArgsUsage: "<image.tar>",
		Before:    setup,
		Action:    Main,
		// After: teardown,
	}
}

func Main(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.Exit("missing tarball argument", 1)
	}

	for _, s := range c.Args().Slice() {
		cmd := exec.Command("ipfs", "add", "-Qr", "--wrap-with-directory", s)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("exec: failed to add %s: %v", s, err)
		}

		n, err := fmt.Fprint(c.App.Writer, string(output))
		if err != nil {
			return fmt.Errorf("fprint: write failed at byte %d: %w", n, err)
		}
	}

	return nil
}
