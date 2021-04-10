package cmd

import (
	"os"
	"path"
	"sort"

	"github.com/lzambarda/hbt/client"
	"github.com/lzambarda/hbt/graph/naive"
	"github.com/lzambarda/hbt/internal"
	"github.com/lzambarda/hbt/server"
	"github.com/urfave/cli/v2"
)

var (
	root = &cli.App{
		Name:        "hbt",
		Description: "a zsh suggestion system",
		Version:     internal.Version,
		Flags: []cli.Flag{&cli.BoolFlag{
			Name:        "debug",
			Aliases:     []string{"d"},
			DefaultText: "true",
			Value:       true,
			Destination: &internal.Debug,
		},
			&cli.StringFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				DefaultText: internal.DefaultPort,
				Value:       internal.DefaultPort,
				Destination: &internal.Port,
			},
			&cli.StringFlag{
				Name:        "cache",
				Aliases:     []string{"c"},
				DefaultText: "binary location",
				Value:       ".",
				Destination: &internal.CachePath,
			}},
		Commands: []*cli.Command{
			{
				Name:        "start",
				Description: "start the TCP server",
				Usage:       "start",
				Action: func(_ *cli.Context) error {
					cachePath := path.Join(internal.CachePath, internal.CacheName)
					g := naive.NewGraph(10, 3)
					err := g.Load(cachePath)
					if err != nil {
						return err
					}
					return server.Start(g, cachePath)
				},
			},
			{
				Name:        "stop",
				Description: "stop the TCP server",
				Usage:       "stop",
				Action: func(c *cli.Context) error {
					return client.SendStop()
				},
			},
			{
				Name:        "track",
				Description: "track the executed command and directory",
				Usage:       "track <session_id> <current_dir> <current_command>",
				Action: func(c *cli.Context) error {
					return client.SendTrack(c.Args().Slice())
				},
			},
			{
				Name:        "hint",
				Description: "try to suggest a command to execute given the tracked history",
				Usage:       "hint <session_id> <current_dir>",
				Action: func(c *cli.Context) error {
					return client.SendHint(c.Args().Slice())
				},
			},
			{
				Name:        "end",
				Description: "end a tracking session",
				Usage:       "end <session_id>",
				Action: func(c *cli.Context) error {
					return client.SendEnd(c.Args().Slice())
				},
			},
		},
	}
)

func init() {
	sort.Sort(cli.FlagsByName(root.Flags))
	sort.Sort(cli.CommandsByName(root.Commands))
}

func Run(arguments []string) error {
	return root.Run(os.Args)
}
