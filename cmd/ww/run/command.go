package run

// import (
// 	"context"
// 	"log/slog"
// 	"os"
// 	"os/signal"

// 	ipfs_path "github.com/ipfs/boxo/path"
// 	"github.com/ipfs/kubo/core"
// 	"github.com/ipfs/kubo/core/coreapi"
// 	iface "github.com/ipfs/kubo/core/coreiface"
// 	"github.com/ipfs/kubo/repo/fsrepo"
// 	"github.com/urfave/cli/v2"
// )

// func Command() *cli.Command {
// 	ctx, cancel := signal.NotifyContext(context.Background(),
// 		os.Interrupt,
// 		os.Kill)
// 	defer cancel()

// 	app := &cli.App{
// 		Name:   "ww-run",
// 		Usage:  "run a wetware cell",
// 		Action: Main,
// 	}

// 	if err := app.RunContext(ctx, os.Args); err != nil {
// 		slog.Error("execution failed",
// 			"reason", err.Error())
// 		os.Exit(1)
// 	}

// 	return nil
// }

// func Main(c *cli.Context) error {
// 	p, err := ipfs_path.NewPath(c.Args().First())
// 	if err != nil {
// 		return err
// 	}

// 	n, err := env.IPFS.ResolveNode(c.Context, p) // fs node
// 	if err != nil {
// 		return err
// 	}

// 	node, err := env.IPFS.Dag().Get(c.Context, n.Cid()) // dag node
// 	if err != nil {
// 		return err
// 	}

// }

// func setupIPFS() (iface.CoreAPI, error) {
// 	// Try to find IPFS repo
// 	repoPath, err := fsrepo.BestKnownPath()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Open the repo
// 	repo, err := fsrepo.Open(repoPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create IPFS node
// 	node, err := core.NewNode(context.Background(), &core.BuildCfg{
// 		Repo: repo,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create CoreAPI
// 	api, err := coreapi.NewCoreAPI(node)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return api, nil
// }
