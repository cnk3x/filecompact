//go:build windows

package main

import (
	"strings"
	"syscall"
	"time"
)

// 文件遍历
func ListFiles(dir string, files *[]*FileInfo, exclude []string) (err error) {
	dir = strings.TrimRight(dir, `\`)

	var p *uint16
	if p, err = syscall.UTF16PtrFromString(dir + `\*`); err != nil {
		return
	}

	var fdata syscall.Win32finddata
	var fd syscall.Handle

	if fd, err = syscall.FindFirstFile(p, &fdata); err != nil {
		if err == syscall.ERROR_NO_MORE_FILES || err == syscall.ERROR_ACCESS_DENIED {
			err = nil
		}
		return
	}
	defer syscall.FindClose(fd)

	for fd != 0 && int(fd) != -1 {
		name := syscall.UTF16ToString(fdata.FileName[:])
		if name != "." && name != ".." {
			path := dir + `\` + name
			if !isMatched(path, exclude) {
				if fdata.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0 {
					if err = ListFiles(path, files, exclude); err != nil {
						return
					}
				} else {
					file := &FileInfo{
						Name:         name,
						Path:         path,
						Size:         int64(fdata.FileSizeHigh)<<32 + int64(fdata.FileSizeLow),
						IsDir:        false,
						CreatedTime:  time.Unix(0, fdata.CreationTime.Nanoseconds()),
						ModifiedTime: time.Unix(0, fdata.LastWriteTime.Nanoseconds()),
					}
					*files = append(*files, file)
				}
			}
		}

		if err = syscall.FindNextFile(fd, &fdata); err != nil {
			if err == syscall.ERROR_NO_MORE_FILES {
				err = nil
				break
			}

			if err == syscall.ERROR_ACCESS_DENIED {
				err = nil
				continue
			}
		}
	}

	return
}
