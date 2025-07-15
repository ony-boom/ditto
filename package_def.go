package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type PackageDef struct {
	path string
}

type DefinitionFile struct {
	Packages *[]string `toml:"packages"`
}

type Definition struct {
	Packages []string `toml:"packages"`
	Host     *string
}

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

		if d.Type().IsRegular() && strings.HasSuffix(path, ".toml") {
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
	data, err := os.ReadFile(file)
	if err != nil {
		return Definition{}, err
	}

	var defFile DefinitionFile
	if err := toml.Unmarshal(data, &defFile); err != nil {
		return Definition{}, err
	}

	if defFile.Packages == nil {
		log.Printf("warning: 'packages' field missing in %s", file)
	}

	var host *string

	relPath, err := filepath.Rel(pd.path, file)
	if err != nil {
		return Definition{}, err
	}

	parts := strings.Split(relPath, string(os.PathSeparator))
	if len(parts) >= 2 && parts[0] == "hosts" {
		if len(parts) == 2 {
			hostVal := strings.TrimSuffix(parts[1], ".toml")
			host = &hostVal
		} else if len(parts) > 2 {
			hostVal := parts[1]
			host = &hostVal
		}
	}

	return Definition{
		Packages: ptrValueOrDefault(defFile.Packages, []string{}),
		Host:     host,
	}, nil
}
