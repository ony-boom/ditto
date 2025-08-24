package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/ony-boom/ditto/database"
)

type PackageManager interface {
	ListInstalled() ([]string, error)
	Install(pkgs []string, args ...string) error
	Remove(pkgs []string, args ...string) error
}

type PackageDefLoader interface {
	LoadAllDefinitions() ([]Definition, error)
}

type SyncOptions struct {
	DryRun      bool
	Strict      bool
	InstallArgs []string
	RemoveArgs  []string
}

type PackageDiff struct {
	ToAdd             []string
	ToRemove          []string
	ToRemoveFromDitto []string
}

func Sync(
	opts SyncOptions,
	appCtx *AppContext,
) error {
	ctx := context.Background()

	installedPackages, err := appCtx.Pacman.ListInstalled()
	if err != nil {
		return fmt.Errorf("failed to list installed packages: %w", err)
	}

	defs, err := appCtx.PackageDef.LoadAllDefinitions()
	if err != nil {
		return fmt.Errorf("failed to load package definitions: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("cannot get current hostname: %v", err)
	}

	desiredPackages := buildDesiredPackagesFromDefs(defs)

	previouslyManaged, err := getPreviouslyManagedPackages(ctx, appCtx.QueryClient, hostname)
	if err != nil {
		return fmt.Errorf("failed to get previously managed packages: %w", err)
	}

	diff := calculateDiffWithDatabase(desiredPackages, installedPackages, previouslyManaged, *appCtx.Config)

	printPackageChanges(appCtx, diff, opts.Strict)

	if opts.DryRun {
		fmt.Println("Dry run mode — no changes made")
		return nil
	}

	if err := applyPackageChanges(diff, opts, appCtx.Pacman); err != nil {
		return err
	}

	return updateManagedPackages(ctx, queries, desiredPackages, hostname)
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

func getPreviouslyManagedPackages(ctx context.Context, queries *database.Queries, hostname string) ([]string, error) {
	hostPackages, err := queries.GetPackagesByHost(ctx, database.GetPackagesByHostParams{
		Host: sql.NullString{String: hostname},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get host packages: %w", err)
	}

	globalPackages, err := queries.GetPackagesByHost(ctx, database.GetPackagesByHostParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get global packages: %w", err)
	}

	packageSet := make(map[string]struct{})
	for _, pkg := range hostPackages {
		packageSet[pkg.Name] = struct{}{}
	}
	for _, pkg := range globalPackages {
		packageSet[pkg.Name] = struct{}{}
	}

	packages := make([]string, 0, len(packageSet))
	for pkg := range packageSet {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	return packages, nil
}

func calculateDiffWithDatabase(desired, installed, previouslyManaged []string, cfg Config) PackageDiff {
	installedSet := make(map[string]bool, len(installed))
	for _, pkg := range installed {
		installedSet[pkg] = true
	}

	desiredSet := make(map[string]bool, len(desired))
	for _, pkg := range desired {
		desiredSet[pkg] = true
	}

	previouslyManagedSet := make(map[string]bool, len(previouslyManaged))
	for _, pkg := range previouslyManaged {
		previouslyManagedSet[pkg] = true
	}

	var toAdd, toRemove, toRemoveFromDitto []string

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

	for _, pkg := range previouslyManaged {
		if !desiredSet[pkg] && installedSet[pkg] && !slices.Contains(cfg.UninstallIgnore, pkg) {
			toRemoveFromDitto = append(toRemoveFromDitto, pkg)
		}
	}

	sort.Strings(toAdd)
	sort.Strings(toRemove)
	sort.Strings(toRemoveFromDitto)

	return PackageDiff{toAdd, toRemove, toRemoveFromDitto}
}

func printPackageChanges(appCtx *AppContext, diff PackageDiff, strict bool) {
	hasChanges := len(diff.ToAdd) > 0 ||
		(strict && len(diff.ToRemove) > 0) ||
		len(diff.ToRemoveFromDitto) > 0

	if !hasChanges {
		return
	}

	t := buildDiffTable(diff, strict)
	var out bytes.Buffer
	out.WriteString("\n")
	out.WriteString(t.String())
	out.WriteString("\n")
	displayWithOptionalPager(appCtx, &out)
}

func applyPackageChanges(diff PackageDiff, opts SyncOptions, pm PackageManager) error {
	hasChanges := len(diff.ToAdd) > 0 ||
		(opts.Strict && len(diff.ToRemove) > 0) ||
		len(diff.ToRemoveFromDitto) > 0

	if !hasChanges {
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
		if err := pm.Install(diff.ToAdd, opts.InstallArgs...); err != nil {
			return fmt.Errorf("install failed: %w", err)
		}
	}

	if len(diff.ToRemove) > 0 && opts.Strict {
		if err := pm.Remove(diff.ToRemove, opts.RemoveArgs...); err != nil {
			return fmt.Errorf("remove failed: %w", err)
		}
	}

	if len(diff.ToRemoveFromDitto) > 0 {
		fmt.Printf("Removing packages no longer managed by ditto: %v\n", diff.ToRemoveFromDitto)
		if err := pm.Remove(diff.ToRemoveFromDitto, opts.RemoveArgs...); err != nil {
			return fmt.Errorf("ditto package removal failed: %w", err)
		}
	}

	fmt.Println("Changes applied.")
	return nil
}

func updateManagedPackages(ctx context.Context, queries *database.Queries, desiredPackages []string, hostname string) error {
	if err := queries.DeletePackagesByHost(ctx, database.DeletePackagesByHostParams{
		Host: sql.NullString{String: hostname},
	}); err != nil {
		return fmt.Errorf("failed to clear existing packages for host: %w", err)
	}

	for _, pkg := range desiredPackages {
		_, err := queries.CreatePackage(ctx, database.CreatePackageParams{
			Name: pkg,
			Host: sql.NullString{String: hostname},
		})
		if err != nil {
			return fmt.Errorf("failed to insert package %s: %w", pkg, err)
		}
	}

	return nil
}
