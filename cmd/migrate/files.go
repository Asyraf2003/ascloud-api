package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type migrationFile struct {
	Name string
	Path string
}

func listMigrations(dir string) ([]migrationFile, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	out := make([]migrationFile, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		out = append(out, migrationFile{Name: name, Path: filepath.Join(dir, name)})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
