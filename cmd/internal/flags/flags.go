package flags

import (
	"time"

	"github.com/urfave/cli/v2"
)

// CapabilityFlags returns the capability control flags that can be shared across commands
func CapabilityFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:     "with-all",
			Category: "CAPABILITIES",
			Usage:    "grant all capabilities (console, IPFS, exec)",
			EnvVars:  []string{"WW_WITH_ALL"},
		},
		&cli.BoolFlag{
			Name:     "with-console",
			Category: "CAPABILITIES",
			Usage:    "grant console output capability",
			EnvVars:  []string{"WW_WITH_CONSOLE"},
		},
		&cli.BoolFlag{
			Name:     "with-ipfs",
			Category: "CAPABILITIES",
			Usage:    "grant IPFS capability",
			EnvVars:  []string{"WW_WITH_IPFS"},
		},
		&cli.BoolFlag{
			Name:     "with-exec",
			Category: "CAPABILITIES",
			Usage:    "grant process execution capability",
			EnvVars:  []string{"WW_WITH_EXEC"},
		},
	}
}

// P2PFlags returns the P2P networking flags that can be shared across commands
func P2PFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "join",
			Category: "P2P",
			Aliases:  []string{"j"},
			Usage:    "connect to cluster through specified peers",
			EnvVars:  []string{"WW_JOIN"},
		},
		&cli.StringFlag{
			Name:     "discover",
			Category: "P2P",
			Aliases:  []string{"d"},
			Usage:    "automatic peer discovery settings",
			Value:    "/mdns",
			EnvVars:  []string{"WW_DISCOVER"},
		},
		&cli.StringFlag{
			Name:     "namespace",
			Category: "P2P",
			Aliases:  []string{"ns"},
			Usage:    "cluster namespace (must match dial host)",
			Value:    "ww",
			EnvVars:  []string{"WW_NAMESPACE"},
		},
		&cli.BoolFlag{
			Name:     "dial",
			Category: "P2P",
			Usage:    "dial into a cluster using -join and -discover",
			EnvVars:  []string{"WW_AUTODIAL"},
		},
		&cli.DurationFlag{
			Name:     "timeout",
			Category: "P2P",
			Usage:    "timeout for -dial",
			Value:    time.Second * 10,
		},
	}
}

// OutputFlags returns the output control flags that can be shared across commands
func OutputFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:     "quiet",
			Category: "OUTPUT",
			Aliases:  []string{"q"},
			Usage:    "suppress banner message on interactive startup",
			EnvVars:  []string{"WW_QUIET"},
		},
	}
}
