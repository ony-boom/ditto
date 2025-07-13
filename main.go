package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	cfg    = LoadConfig()
	pacman = NewPacman(cfg)
)

func main() {
	app := &cli.Command{
		Name:  "ditto",
		Usage: "Declarative package sync tool for Arch-based systems",
		Description: `Ditto helps you keep your installed packages in sync across machines
by using a simple declarative config file.

You define the packages you want, and Ditto installs them — optionally removing everything else
if you enable strict mode. It also supports dry runs, so you can preview changes safely.`,
		Commands: []*cli.Command{
			{
				Name:    "sync",
				Usage:   "Synchronize installed packages with your desired package list",
				Aliases: []string{"s"},
				Description: `Sync compares your desired package list with what's currently installed.
By default, it will only install missing packages.

You can enable strict mode to remove all packages not listed in your config (excluding ignored ones).`,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Simulate the sync process without making changes. Shows what would be installed or removed.",
					},
					&cli.BoolFlag{
						Name:    "strict",
						Aliases: []string{"x"},
						Usage:   "Enable strict mode: remove packages that are not in the desired list (excluding ignored ones).",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return Sync(SyncOptions{
						Strict: cmd.Bool("strict"),
						DryRun: cmd.Bool("dry-run"),
					})
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}
