package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"
)

type SyncOptions struct {
	DryRun bool
	Strict bool
}

type PackageDiff struct {
	ToAdd    []string
	ToRemove []string
}

func Sync(opts SyncOptions) error {
	installedPackages, err := pacman.ListInstalled()
	if err != nil {
		return fmt.Errorf("failed to list installed packages: %w", err)
	}

	defs, err := packageDef.LoadAllDefinitions()
	if err != nil {
		return fmt.Errorf("failed to load package definitions: %w", err)
	}

	desiredPackages := buildDesiredPackagesFromDefs(defs)
	diff := calculateDiff(desiredPackages, installedPackages)

	printPackageChanges(diff, opts.Strict)

	if opts.DryRun {
		fmt.Println("Dry run mode — no changes made")
		return nil
	}

	return applyPackageChanges(diff, opts)
}

func buildDesiredPackagesFromDefs(defs []Definition) []string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("cannot get current hostname: %v", err)
	}

	unique := make(map[string]struct{})

	for _, def := range defs {
		if def.Host == nil || *def.Host == hostname {
			for _, pkg := range def.Packages {
				unique[pkg] = struct{}{}
			}
		}
	}

	packages := make([]string, 0, len(unique))
	for pkg := range unique {
		packages = append(packages, pkg)
	}

	sort.Strings(packages)

	return packages
}

func calculateDiff(desired, installed []string) PackageDiff {
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

	sort.Strings(toAdd)
	sort.Strings(toRemove)

	return PackageDiff{
		ToAdd:    toAdd,
		ToRemove: toRemove,
	}
}

func printPackageChanges(diff PackageDiff, strict bool) {
	if len(diff.ToAdd) == 0 && (!strict || len(diff.ToRemove) == 0) {
		return
	}

	t := buildDiffTable(diff, strict)

	var out bytes.Buffer
	out.WriteString("\n")
	out.WriteString(t.String())
	out.WriteString("\n")

	displayWithOptionalPager(&out)
}

func applyPackageChanges(diff PackageDiff, opts SyncOptions) error {
	if len(diff.ToAdd) == 0 && (!opts.Strict || len(diff.ToRemove) == 0) {
		fmt.Println("Nothing to apply.")
		return nil
	}

	fmt.Print("Proceed with applying changes? [y/N]: ")

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil || strings.ToLower(strings.TrimSpace(input)) != "y" {
		fmt.Println("Aborted.")
		return nil
	}

	if len(diff.ToAdd) > 0 {
		if err := pacman.Install(diff.ToAdd); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}

	if len(diff.ToRemove) > 0 && opts.Strict {
		if err := pacman.Remove(diff.ToRemove); err != nil {
			return fmt.Errorf("remove failed: %w", err)
		}
	}

	fmt.Println("Changes applied.")
	return nil
}
