//go:build !windows

package main

import (
	"errors"
	"io/fs"
	"path/filepath"
	"syscall"
)

// 文件遍历
func ListFiles(dir string, files *[]*FileInfo, exclude []string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, e error) (err error) {
		if err = e; err != nil {
			if errors.Is(e, syscall.EACCES) {
				if d != nil && d.IsDir() {
					e = fs.SkipDir
				} else {
					e = nil
				}
			}
			return e
		}

		if isMatched(path, exclude) {
			return Iif(d.IsDir(), fs.SkipDir, nil)
		}

		if d.IsDir() {
			return nil
		}

		info, e := d.Info()
		if e != nil {
			return e
		}

		if stat := GetFileInfo(path, info); stat != nil {
			*files = append(*files, stat)
		}
		return nil
	})
}
