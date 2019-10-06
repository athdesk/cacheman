package local

import (
	. "cacheman/shared"
	"os"
	"path/filepath"
	"strings"
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

func BuildDirTreeForFile(path string) error {
	realPath := filepath.Dir(path)
	if !DirExists(realPath) {
		return os.MkdirAll(realPath, 0755)
	}
	return nil
}

func IsFileExcluded(path string, Cfg *Config) bool {
	SplitPath := strings.Split(path, ".")
	for _, Excl := range Cfg.ExcludedExts {
		if SplitPath[len(SplitPath)-1] == Excl {
			return true
		}
	}
	return false
}
