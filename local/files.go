package local

import (
	"os"
	"path/filepath"
	"strings"
)

//FileExists checks if a file exists, and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

//DirExists checks if a directory exists, and is a directory
func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

//FileSize safely returns a file size, if it exists
func FileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return 0
	}
	return info.Size()
}

//BuildDirTreeForFile makes sure that the proper directory structure for a chosen filepath exists
func BuildDirTreeForFile(fpath string) error {
	realPath := filepath.Dir(fpath)
	if !DirExists(realPath) {
		return os.MkdirAll(realPath, 0755)
	}
	return nil
}

//IsFileExcluded checks if a file has a blacklisted extension
func IsFileExcluded(path string, Cfg *Config) bool {
	SplitPath := strings.Split(path, ".")
	for _, Excl := range Cfg.ExcludedExts {
		if SplitPath[len(SplitPath)-1] == Excl {
			return true
		}
	}
	return false
}
