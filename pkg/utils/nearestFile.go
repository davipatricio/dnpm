package utils

import (
	"os"
	"path/filepath"
)

// NearestPackageJSON finds the nearest "package.json" file starting from a given directory.
func NearestPackageJSON(dir string) (string, error) {
	return NearestFile(dir, "package.json")
}

// NearestFile finds the nearest file with the given filename starting from a given directory.
func NearestFile(dir, filename string) (string, error) {
	for {
		file := filepath.Join(dir, filename)
		_, err := os.Stat(file)

		if err == nil {
			return file, nil
		}

		if dir == "/" || !os.IsNotExist(err) {
			return "", os.ErrNotExist
		}

		dir = filepath.Dir(dir)
	}
}
