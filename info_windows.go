package main

import (
	"io/fs"
	"syscall"
	"time"
)

func GetFileInfo(path string, info fs.FileInfo) *FileInfo {
	stat, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return nil
	}

	return &FileInfo{
		Name:         info.Name(),
		Path:         path,
		Size:         info.Size(),
		IsDir:        info.IsDir(),
		CreatedTime:  time.Unix(0, stat.CreationTime.Nanoseconds()),
		ModifiedTime: info.ModTime(),
	}
}
