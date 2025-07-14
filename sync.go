package main

import (
	"fmt"
	"log"
	"slices"
)

var desiredPackages = []string{
	"sl",
}

type SyncOptions struct {
	DryRun bool
	Strict bool
}

type PackageDiff struct {
	ToAdd    []string
	ToRemove []string
}

func Sync(opts SyncOptions) error {
	fmt.Println("Syncing packages...")

	installedPackages, err := pacman.ListInstalled()
	if err != nil {
		return fmt.Errorf("failed to list installed packages: %w", err)
	}

	def, err := packageDef.LoadAllDefFiles()
	if err != nil {
		log.Fatalf("failed to load package definitions: %v", err)
	}

	fmt.Printf("Loaded package definitions: %v\n", def)

	diff := diff(desiredPackages, installedPackages)

	printPackageChanges(diff, opts.Strict)

	if opts.DryRun {
		fmt.Println("Dry run mode - no changes made")
		return nil
	}

	return applyPackageChanges(diff, opts)
}

func diff(desired, installed []string) PackageDiff {
	installedSet := make(map[string]bool, len(installed))
	for _, pkg := range installed {
		installedSet[pkg] = true
	}

	desiredSet := make(map[string]bool, len(desired))
	for _, pkg := range desired {
		desiredSet[pkg] = true
	}

	var toAdd, toRemove []string

	for _, pkg := range desired {
		if !installedSet[pkg] {
			toAdd = append(toAdd, pkg)
		}
	}

	for _, pkg := range installed {
		if !desiredSet[pkg] && !slices.Contains(cfg.UninstallIgnore, pkg) {
			toRemove = append(toRemove, pkg)
		}
	}

	return PackageDiff{
		ToAdd:    toAdd,
		ToRemove: toRemove,
	}
}

func printPackageChanges(diff PackageDiff, strict bool) {
	if len(diff.ToAdd) > 0 {
		fmt.Printf("Packages to install: %v\n", diff.ToAdd)
	}

	if len(diff.ToRemove) > 0 && strict {
		fmt.Printf("Packages to remove: %v\n", diff.ToRemove)
	}

	if len(diff.ToAdd) == 0 && len(diff.ToRemove) == 0 {
		fmt.Println("No package changes needed")
	}
}

func applyPackageChanges(diff PackageDiff, opts SyncOptions) error {
	fmt.Println("Applying package changes...")

	if len(diff.ToAdd) > 0 {
		fmt.Printf("Installing: %v\n", diff.ToAdd)
		pacman.Install(diff.ToAdd)
	} else {
		fmt.Println("No packages to install")
	}

	if len(diff.ToRemove) > 0 && opts.Strict {
		fmt.Printf("Removing: %v\n", diff.ToRemove)
	} else {
		fmt.Println("No packages to remove")
	}

	return nil
}
