package run

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "run",
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
		if err := run(c, s); err != nil {
			return err
		}
	}

	return nil
}

func run(c *cli.Context, src string) error {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "ww-run-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tmpdir); err != nil {
			slog.ErrorContext(c.Context, "failed to clean up temporary directory",
				"reason", err,
				"path", tmpdir)
		}
	}()

	// Step 1: Ensure the image is loaded using `ww load <src>`
	loadCmd := exec.CommandContext(c.Context, os.Args[0], "load", src)
	loadCmd.Stdout = os.Stdout
	loadCmd.Stderr = os.Stderr

	if err := loadCmd.Run(); err != nil {
		return fmt.Errorf("load: %w", err)
	}

	imageTag := "wetware:" + src

	// Step 2: Run the container (attached, auto-remove)
	runCmd := exec.CommandContext(c.Context, "podman", "run", "--rm", "--pull=never", imageTag)
	runCmd.Stdin = c.App.Reader
	runCmd.Stdout = c.App.Writer
	runCmd.Stderr = c.App.ErrWriter

	if err := runCmd.Run(); err != nil { // ✅ actually run the command
		return fmt.Errorf("podman run failed: %w", err)
	}

	return nil
}
