// Package cmd contains the entrypoint of hbt.
package cmd

import (
	"fmt"
	"path"

	"github.com/lzambarda/hbt/graph/naive"
	"github.com/lzambarda/hbt/internal"
	"github.com/lzambarda/hbt/server"
	"github.com/urfave/cli/v2"
)

var (
	g         server.Graph
	cachePath string
	root      = &cli.App{
		Name:        "hbt",
		Usage:       "a zsh suggestion system",
		Description: `Spawn a TCP server listening on the local port 43111 (can be changed with HBT_PORT).`,
		Version:     internal.Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "debug",
				Aliases:     []string{"d"},
				DefaultText: "false",
				Destination: &internal.Debug,
				EnvVars:     []string{internal.DebugName},
			},
			&cli.StringFlag{
				Name:        "cache",
				Aliases:     []string{"c"},
				DefaultText: "binary location",
				Value:       internal.DefaultCachePath,
				Destination: &internal.CachePath,
				EnvVars:     []string{internal.CachePathName},
			},
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				DefaultText: internal.DefaultPort,
				Value:       internal.DefaultPort,
				Destination: &internal.Port,
				EnvVars:     []string{internal.PortName},
			},
			&cli.DurationFlag{
				Name:        "save-interval",
				Aliases:     []string{"s"},
				DefaultText: internal.DefaultSaveInterval.String(),
				Value:       internal.DefaultSaveInterval,
				Destination: &internal.SaveInterval,
				EnvVars:     []string{internal.SaveIntervalName},
			},
		},
		Before: func(_ *cli.Context) error {
			cachePath = path.Join(internal.CachePath, internal.CacheName)
			g = naive.NewGraph(10, 3)
			return g.Load(cachePath)
		},
		// By default start a server
		Action: func(_ *cli.Context) error {
			return server.Start(g, cachePath)
		},
		Commands: []*cli.Command{
			{
				Name:    "cli",
				Aliases: []string{"c"},
				Usage:   "run a command without starting a server",
				Action: func(c *cli.Context) error {
					result, err := server.ProcessCommand(c.Args().Slice(), g)
					if err != nil {
						return err
					}
					fmt.Println(result)
					return nil
				},
			},
		},
	}
)

// Run the root command with the given arguments.
func Run(arguments []string) error {
	return root.Run(arguments)
}
