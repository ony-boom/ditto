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
	pkgs, err := readPackagesFromFile(file)
	if err != nil {
		return Definition{}, err
	}

	host, err := inferHost(pd.path, file)
	if err != nil {
		return Definition{}, err
	}

	return Definition{
		Packages: pkgs,
		Host:     host,
	}, nil
}

func readPackagesFromFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var pkgs []string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := parseLine(scanner.Text())
		if line != "" {
			pkgs = append(pkgs, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(pkgs) == 0 {
		log.Printf("warning: no packages defined in %s", file)
	}

	return pkgs, nil
}

// parseLine trims whitespace and removes comments from a single line.
func parseLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	if idx := strings.Index(line, "#"); idx != -1 {
		line = strings.TrimSpace(line[:idx])
	}
	return line
}

// inferHost extracts the host name based on the file's relative path.
func inferHost(basePath, file string) (*string, error) {
	relPath, err := filepath.Rel(basePath, file)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(relPath, string(os.PathSeparator))
	if len(parts) >= 2 && parts[0] == "hosts" {
		var hostVal string
		if len(parts) == 2 {
			hostVal = strings.TrimSuffix(parts[1], FILE_EXTENSION)
		} else {
			hostVal = parts[1]
		}
		return &hostVal, nil
	}

	return nil, nil
}
