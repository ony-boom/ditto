package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
	white := lipgloss.Color("15")
	green := lipgloss.Color("10")
	red := lipgloss.Color("9")

	headerStyle := lipgloss.NewStyle().
		Foreground(white).
		Padding(0, 1).
		Bold(true)

	installStyle := lipgloss.NewStyle().
		Foreground(green).
		Padding(0, 1).
		Width(24)

	removeStyle := lipgloss.NewStyle().
		Foreground(red).
		Padding(0, 1).
		Width(24)

	t := table.New().
		Border(lipgloss.MarkdownBorder()).
		BorderTop(false).
		BorderBottom(false).
		Headers("To Install", "To Remove").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			if col == 0 {
				return installStyle
			}
			return removeStyle
		})

	maxLen := len(diff.ToAdd)
	if strict && len(diff.ToRemove) > maxLen {
		maxLen = len(diff.ToRemove)
	}

	for i := 0; i < maxLen; i++ {
		var addPkg, removePkg string
		if i < len(diff.ToAdd) {
			addPkg = diff.ToAdd[i]
		}
		if strict && i < len(diff.ToRemove) {
			removePkg = diff.ToRemove[i]
		}
		t.Row(addPkg, removePkg)
	}

	if len(diff.ToAdd) == 0 && (!strict || len(diff.ToRemove) == 0) {
		fmt.Println("No package changes needed")
		return
	}

	fmt.Println()
	fmt.Println(t)
	fmt.Println()
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
