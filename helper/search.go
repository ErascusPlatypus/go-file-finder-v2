package helper

import (
    "os"
    "path/filepath"
    "strings"
    "sync"
)

var ignoredDirs = map[string]struct{}{
    ".git":         {},
    "node_modules": {},
    ".svn":         {},
    ".hg":          {},
    "vendor":       {},
    "__pycache__":  {},
    ".cache":       {},
    ".vscode":      {},
    ".idea":        {},
    "target":       {},
    "build":        {},
    "dist":         {},
	"site-packages": {},
    "Python-3.10.12": {}, 
    "venv": {},
    ".venv": {},
}

func SemFinder(dir, fileName string) []string {
    var (
        wg      sync.WaitGroup
        mu      sync.Mutex
        sem     = make(chan struct{}, 50)
        out     []string
        visited = make(map[string]struct{})
        visitMu sync.Mutex
    )

    var search func(string)
    search = func(dir string) {
        defer wg.Done()

        realPath, err := filepath.EvalSymlinks(dir)
        if err != nil {
            return
        }

        visitMu.Lock()
        if _, seen := visited[realPath]; seen {
            visitMu.Unlock()
            return
        }
        visited[realPath] = struct{}{}
        visitMu.Unlock()

        sem <- struct{}{}
        entries, err := os.ReadDir(dir)
        <-sem
        if err != nil {
            return
        }

        for _, e := range entries {
            name := e.Name()
            path := filepath.Join(dir, name)

            if e.IsDir() || (e.Type()&os.ModeSymlink != 0) {
                if e.Type()&os.ModeSymlink != 0 {
                    info, err := os.Stat(path)
                    if err != nil || !info.IsDir() {
                        continue
                    }
                }

                if _, ok := ignoredDirs[name]; !ok {
                    wg.Add(1)
                    go search(path)
                }
                continue
            }

            if FuzzyMatch(strings.ToLower(fileName), strings.ToLower(name)) {
                mu.Lock()
                out = append(out, path)
                mu.Unlock()
            }
        }
    }

    wg.Add(1)
    search(dir)
    wg.Wait()
    return out
}