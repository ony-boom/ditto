package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ony-boom/ditto/database"
	"github.com/urfave/cli/v3"
)

type AppContext struct {
	Config      *Config
	Pacman      *Pacman
	QueryClient *database.Queries
	PackageDef  *PackageDef
}

func main() {
	// Build our application context
	appCtx := &AppContext{
		Config:      LoadConfig(),
		Pacman:      NewPacman(LoadConfig()),
		PackageDef:  NewPackageDef(),
		QueryClient: NewQueryClient(),
	}

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
			newSyncCommand(appCtx),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func newSyncCommand(appCtx *AppContext) *cli.Command {
	return &cli.Command{
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
			return syncAction(appCtx, cmd)
		},
	}
}

func syncAction(appCtx *AppContext, cmd *cli.Command) error {
	installArgs, removeArgs := splitInstallRemoveArgs(cmd.Args().Slice())

	if len(removeArgs) > 0 && !cmd.Bool("strict") {
		fmt.Fprintln(os.Stderr, "Warning: remove args provided but --strict is disabled, ignoring them.")
		removeArgs = nil
	}

	return Sync(SyncOptions{
		Strict:      cmd.Bool("strict"),
		DryRun:      cmd.Bool("dry-run"),
		InstallArgs: installArgs,
		RemoveArgs:  removeArgs,
	}, appCtx)
}

func splitInstallRemoveArgs(args []string) (installArgs, removeArgs []string) {
	for i, a := range args {
		if a == "::" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}
