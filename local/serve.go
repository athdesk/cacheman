package local

import (
	"cacheman/remote"
	"cacheman/shared"
	"net/http"
	"strings"
)

//ServeCachedFile Takes a requests and fulfills it with a cached file
func ServeCachedFile(w http.ResponseWriter, r *http.Request, path string, Cfg *shared.Config) bool {
	AbsPath := strings.ReplaceAll(path, Cfg.CacheDir, "")

	ExpectedSize := remote.GetCorrectSize(AbsPath, Cfg)
	RealSize := FileSize(path)
	if ExpectedSize != -1 && RealSize != ExpectedSize {
		return false // if filesize is mismatched serve it from remote server, this will redownload the file
	}

	http.ServeFile(w, r, path) //does not serve paths containing /../, supports byte ranges
	return true
}
