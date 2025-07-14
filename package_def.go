package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PackageDef struct {
	path string
}

/* const (
	createDirPerm = 0755
) */

func NewPackageDef() *PackageDef {
	conFigFile, err := getConfigPath()
	if err != nil {
		log.Fatalf("failed to get config path: %v", err)
	}

	path := filepath.Join(filepath.Dir(conFigFile), "packages")

	/* if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, createDirPerm); err != nil {
			log.Fatalf("failed to create packages directory: %v", err)
		}
	} */

	return &PackageDef{
		path: path,
	}
}

func (pd *PackageDef) LoadAllDefFiles() ([]string, error) {
	var files []string
	err := filepath.WalkDir(pd.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil // Ignore non-existent paths
			}
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".toml") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
