package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	cfg        = LoadConfig()
	pacman     = NewPacman(cfg)
	packageDef = NewPackageDef()
)

func main() {
	app := &cli.Command{
		Name:  "ditto",
		Usage: "Declarative package sync tool for Arch-based systems",
		Description: `Ditto helps you keep your installed packages in sync across machines
by using a simple declarative config file.

You define the packages you want, and Ditto installs them — optionally removing everything else
if you enable strict mode. It also supports dry runs, so you can preview changes safely.

By the way, just running 'ditto' will create the config file at <XDG_CONFIG_HOME>/ditto
`,
		Commands: []*cli.Command{
			{
				Name:      "sync",
				Usage:     "Synchronize installed packages with your desired package list",
				Aliases:   []string{"s"},
				ArgsUsage: "[pacman install args...] :: [pacman remove args...]",
				Description: `Sync compares your desired package list with what's currently installed.
By default, it only installs missing packages.

You can pass additional pacman arguments after '--'. 
Use '::' to separate install args from remove args.

Examples:
  ditto sync -- -Syu --needed
  ditto sync -- -Syu --needed :: -Rns`,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "dry-run",
						Aliases: []string{"n"},
						Usage:   "Simulate the sync process without making changes.",
					},
					&cli.BoolFlag{
						Name:    "strict",
						Aliases: []string{"x"},
						Usage:   "Enable strict mode: remove packages not in the desired list.",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					args := cmd.Args().Slice()

					var installArgs, removeArgs []string
					sepIndex := -1
					for i, a := range args {
						if a == "::" {
							sepIndex = i
							break
						}
					}

					if sepIndex != -1 {
						installArgs = args[:sepIndex]
						removeArgs = args[sepIndex+1:]
						if !cmd.Bool("strict") && len(removeArgs) > 0 {
							fmt.Fprintln(os.Stderr, "Warning: remove args provided but --strict is disabled, ignoring them.")
						}
					} else {
						installArgs = args
					}

					return Sync(SyncOptions{
						Strict:      cmd.Bool("strict"),
						DryRun:      cmd.Bool("dry-run"),
						InstallArgs: installArgs,
						RemoveArgs:  removeArgs,
					})
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		panic(err)
	}
}
