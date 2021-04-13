package cmd

import (
	"path"

	"github.com/lzambarda/hbt/graph/naive"
	"github.com/lzambarda/hbt/internal"
	"github.com/lzambarda/hbt/server"
	"github.com/urfave/cli/v2"
)

var (
	root = &cli.App{
		Name:        "hbt",
		Usage:       "a zsh suggestion system",
		Description: `Running hbt will spawn a TCP server listening on the local port 43111 (can be changed with HBT_PORT).`,
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
		Action: func(_ *cli.Context) error {
			cachePath := path.Join(internal.CachePath, internal.CacheName)
			g := naive.NewGraph(10, 3)
			err := g.Load(cachePath)
			if err != nil {
				return err
			}
			return server.Start(g, cachePath)
		},
	}
)

func Run(arguments []string) error {
	return root.Run(arguments)
}
