package export

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:   "export",
		Action: export,
	}
}

func export(c *cli.Context) error {
	dir := c.Args().First()

	addCmd := exec.CommandContext(c.Context, "ipfs", "add", "-Q", "-r", dir)
	defer addCmd.Cancel()

	out, err := addCmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer out.Close()

	if err := addCmd.Start(); err != nil {
		return err
	}

	var root string
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		root = scanner.Text()
	}
	if scanner.Err() != nil {
		return fmt.Errorf("scan: %w", scanner.Err())
	}
	if err := addCmd.Wait(); err != nil {
		return err
	}

	pubCmd := exec.CommandContext(c.Context, "ipfs", "name", "publish", "-Q", root)
	defer pubCmd.Cancel()

	out, err = pubCmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer out.Close()

	if err := pubCmd.Start(); err != nil {
		return err
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, out); err != nil {
		return err
	}
	if err := pubCmd.Wait(); err != nil {
		return err
	}

	exportID := strings.TrimSpace(buf.String())
	_, err = fmt.Fprintln(c.App.Writer, exportID)
	return err
}
