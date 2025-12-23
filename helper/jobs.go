package helper

import (
	"os"
	"path/filepath"

)

func ProcessFiles(dir string, c *Cache, dirJobs chan File) {
	entries, err := os.ReadDir(dir)

	if err != nil {
		return
	}

	for _, entry := range entries {
		currPath := filepath.Join(dir, entry.Name())

		_, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := File{Path: currPath}
		c.Add(filePath)

		if entry.IsDir() {
			if _, ok := ignoredDirs[entry.Name()]; !ok {
				dirJobs <- filePath
			}
		}
	}
}

func FuzzyMatch(query, actual string) bool {
	if len(query) == 0 {
		return true
	}
	if len(query) > len(actual) {
		return false
	}

	j := 0
	for i := 0; i < len(actual); i++ {
		if actual[i] == query[j] {
			j++
			if j == len(query) {
				return true
			}
		}
	}
	return false
}