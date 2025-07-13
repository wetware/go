package build

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "build",
		Usage:     "build and push wetware artifacts to IPFS",
		ArgsUsage: "<source-dir>",
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
		output, err := build(c, s)
		if err != nil {
			return fmt.Errorf("exec: failed to add %s: %v", s, err)
		}

		n, err := fmt.Fprintln(c.App.Writer, output)
		if err != nil {
			return fmt.Errorf("fprint: write failed at byte %d: %w", n, err)
		}
	}

	return nil
}

func build(c *cli.Context, src string) (string, error) {
	// Use system default for tempdir to avoid sticky bit or permission issues
	tmpdir, err := os.MkdirTemp("", "ww-build-*")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(tmpdir); err != nil {
			slog.ErrorContext(c.Context, "failed to clean up temporary directory",
				"reason", err,
				"path", tmpdir)
		}
	}()

	// Step 1: Build the image
	imghash, err := command(c, "podman", "build", "-q", src)
	if err != nil {
		return "", fmt.Errorf("build: %w", err)
	}
	imghash = strings.TrimSpace(imghash)

	// Step 2: Tag it with desired name
	tag := "wetware:" + imghash
	if err := exec.CommandContext(c.Context, "podman", "tag", imghash, tag).Run(); err != nil {
		return "", fmt.Errorf("tag: %w", err)
	}

	// Step 3: Save the tagged image as an OCI directory
	if err := exec.CommandContext(c.Context, "podman", "image", "save", "--format=oci-dir", "-o", tmpdir, tag).Run(); err != nil {
		return "", fmt.Errorf("save: %w", err)
	}

	// Step 4: Add to IPFS (returns root CID)
	return command(c, "ipfs", "add", "-Qr", tmpdir)
}

func command(c *cli.Context, name string, args ...string) (output string, err error) {
	cmd := exec.CommandContext(c.Context, name, args...)
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("exec: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}
