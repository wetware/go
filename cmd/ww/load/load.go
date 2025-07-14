package load

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/wetware/go/util"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "load",
		Usage:     "fetch a package from IPFS",
		ArgsUsage: "<image.tar>",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "path",
				EnvVars: []string{"WW_PATH"},
				Value:   "~/.ww",
				Usage:   "path to local wetware storage",
			},
		},
		Before: setup,
		Action: Main,
		// After:     teardown,
	}
}

func Main(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return cli.Exit("missing IPFS hash argument", 1)
	}

	// Get and expand the storage path
	pathArg := c.String("path")
	wwpath, err := util.ExpandHome(pathArg)
	if err != nil {
		return fmt.Errorf("failed to expand WW_PATH: %w", err)
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(wwpath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory %s: %w", wwpath, err)
	}

	for _, hash := range c.Args().Slice() {
		if err := load(c, hash); err != nil {
			return fmt.Errorf("load: %w", err)
		}
	}

	return nil
}

func load(c *cli.Context, hash string) error {
	wwpath := c.String("path")
	if strings.HasPrefix(wwpath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		wwpath = filepath.Join(homeDir, wwpath[2:])
	}

	imageDir := filepath.Join(wwpath, hash)
	imageTag := "wetware:" + hash
	indexPath := filepath.Join(imageDir, "index.json")

	if err := exec.Command("podman", "image", "exists", imageTag).Run(); err == nil {
		fmt.Printf("✅ found %s\n", imageTag)
		return nil
	}

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		fmt.Printf("⬇️  Fetching %s from IPFS...\n", hash)
		_ = os.RemoveAll(imageDir)
		cmd := exec.Command("ipfs", "get", hash)
		cmd.Dir = wwpath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("ipfs get %s failed: %v\nOutput: %s", hash, err, string(output))
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat %s: %w", indexPath, err)
	} else {
		fmt.Printf("ℹ️  OCI dir already exists at %s, skipping fetch.\n", imageDir)
	}

	return loadOCIWithPipe(c.Context, imageDir)

	// Fallback: tar the directory and podman load
	// return loadOCIWithTar(c, imageDir)
}

func loadOCIWithPipe(ctx context.Context, imageDir string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tarCmd := exec.CommandContext(ctx, "tar", "-C", imageDir, "-cf", "-", ".")
	podmanCmd := exec.CommandContext(ctx, "podman", "image", "load")

	pipeReader, pipeWriter := io.Pipe()
	tarCmd.Stdout = pipeWriter
	podmanCmd.Stdin = pipeReader

	var stderr bytes.Buffer
	podmanCmd.Stderr = &stderr

	if err := tarCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tar: %w", err)
	}
	if err := podmanCmd.Start(); err != nil {
		return fmt.Errorf("failed to start podman: %w", err)
	}

	cherr := make(chan error, 1)
	go func() {
		defer close(cherr)
		cherr <- tarCmd.Wait()
		_ = pipeWriter.Close()
	}()

	if err := podmanCmd.Wait(); err != nil {
		return fmt.Errorf("podman image load failed: %w\n%s", err, stderr.String())
	}

	// Wait for tar to finish
	if err := <-cherr; err != nil {
		return fmt.Errorf("tar failed: %w", err)
	}

	// 🔍 Get ID of most recently loaded image
	out, err := exec.Command("podman", "images", "--format", "{{.ID}}", "--no-trunc").Output()
	if err != nil {
		return fmt.Errorf("cannot find loaded image: %w", err)
	}
	ids := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(ids) == 0 || ids[0] == "" {
		return fmt.Errorf("no image ID found after load")
	}
	imageID := ids[0]

	// 🏷️ Tag with the expected name
	cid := filepath.Base(imageDir)
	imageTag := "wetware:" + cid
	if err := exec.Command("podman", "tag", imageID, imageTag).Run(); err != nil {
		return fmt.Errorf("podman tag failed: %w", err)
	}

	fmt.Printf("✅ loaded %s", imageTag)
	return nil
}
