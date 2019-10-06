package local

import (
	"os"
	"path/filepath"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func FileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return 0
	} else {
		return info.Size()
	}
}

func BuildDirTreeForFile(path string) {
	realPath := filepath.Dir(path)
	if !DirExists(realPath) {
		os.MkdirAll(realPath, 0755)
	}
}
