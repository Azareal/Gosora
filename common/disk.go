package common

import (
	"path/filepath"
	"os"
)

func DirSize(path string) (int, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, file os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !file.IsDir() {
			size += file.Size()
		}
		return err
	})
	return int(size), err
}