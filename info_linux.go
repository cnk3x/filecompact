package main

import (
	"io/fs"
	"syscall"
	"time"
)

func GetFileInfo(path string, info fs.FileInfo) *FileInfo {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}

	return &FileInfo{
		Name:         info.Name(),
		Path:         path,
		Size:         info.Size(),
		IsDir:        info.IsDir(),
		CreatedTime:  time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)),
		ModifiedTime: info.ModTime(),
	}
}
