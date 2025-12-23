package helper

import (
	"path/filepath"
	"strings"
	"sync"
)

type File struct {
	Path      string
	pathLower string
}

type Cache struct {
	mu    sync.RWMutex
	files []File
}

func (c *Cache) Add(file File) {
	c.mu.Lock()
	defer c.mu.Unlock()

	file.pathLower = strings.ToLower(file.Path)
	c.files = append(c.files, file)
}

func (c *Cache) Search(query string, maxResults int) []string {
	if query == "" {
		return []string{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	queryLower := strings.ToLower(query)
	results := make([]string, 0, maxResults)

	for _, file := range c.files {
		fileName := string(filepath.Base(file.Path))
		fileName = strings.ToLower(fileName)

		if FuzzyMatch(queryLower, fileName) {
			results = append(results, file.Path)

			if len(results) >= maxResults {
				break
			}
		}
	}

	return results
}