package scanner

import (
	"io/fs"
	"github.com/mzavhorodnii/mac-cleaner-go/internal/model"
	"path/filepath"
	"syscall"
)

var cleanableDirs = map[string]bool{
	"Caches": true,
	"Logs":   true,
	".Trash": true,
}

func Scan(root string) ([]model.Dir, error) {
	sizes := make(map[string]int64)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			var size int64
			info, err := d.Info()
			if err == nil {
				if stat, ok := info.Sys().(*syscall.Stat_t); ok {
					size = stat.Blocks * 512
				} else {
					size = info.Size()
				}

				dir := filepath.Dir(path)
				for {
					sizes[dir] += size
					if dir == root || dir == "/" || dir == "." {
						break
					}
					parent := filepath.Dir(dir)
					if parent == dir {
						break
					}
					dir = parent
				}
			}
		} else {
			if _, exists := sizes[path]; !exists {
				sizes[path] = 0
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	var output []model.Dir
	for p, s := range sizes {
		base := filepath.Base(p)
		if cleanableDirs[base] && s > 0 {
			output = append(output, model.Dir{Path: p, Size: s})
		}
	}
	return output, nil
}
