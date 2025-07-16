package main

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PackageDef struct {
	path string
}

type Definition struct {
	Packages []string
	Host     *string
}

const FILE_EXTENSION = ".pkgs"

func NewPackageDef() *PackageDef {
	configPath, err := getConfigPath()
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}

	packagesPath := filepath.Join(filepath.Dir(configPath), "packages")
	return &PackageDef{path: packagesPath}
}

func (pd *PackageDef) LoadAllDefinitions() ([]Definition, error) {
	var defs []Definition

	err := filepath.WalkDir(pd.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if d.Type().IsRegular() && strings.HasSuffix(path, FILE_EXTENSION) {
			def, err := pd.parseDefFile(path)
			if err != nil {
				return err
			}
			defs = append(defs, def)
		}
		return nil
	})

	return defs, err
}

func (pd *PackageDef) parseDefFile(file string) (Definition, error) {
	f, err := os.Open(file)
	if err != nil {
		return Definition{}, err
	}
	defer f.Close()

	var pkgs []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pkgs = append(pkgs, line)
	}

	if err := scanner.Err(); err != nil {
		return Definition{}, err
	}

	if len(pkgs) == 0 {
		log.Printf("warning: no packages defined in %s", file)
	}

	var host *string
	relPath, err := filepath.Rel(pd.path, file)
	if err != nil {
		return Definition{}, err
	}

	parts := strings.Split(relPath, string(os.PathSeparator))
	if len(parts) >= 2 && parts[0] == "hosts" {
		if len(parts) == 2 {
			hostVal := strings.TrimSuffix(parts[1], FILE_EXTENSION)
			host = &hostVal
		} else if len(parts) > 2 {
			hostVal := parts[1]
			host = &hostVal
		}
	}

	return Definition{
		Packages: pkgs,
		Host:     host,
	}, nil
}
