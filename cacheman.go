package main

import (
	"cacheman/local"
	"cacheman/remote"
	"fmt"
	"log"
	"net/http"
	"time"
)

var cfg local.Config

func main() {
	local.PutConfig(&cfg)
	http.HandleFunc("/", handleReq)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	NowStr := time.Now().Format(time.Kitchen)
	RequestedLocalPath := cfg.CacheDir + "/" + r.URL.Path[1:] //add cachedir to path, to not check in /
	RequestedPath := r.URL.Path[1:]
	fmt.Printf("[SERVER %s] File requested: %s\n", NowStr, RequestedPath)

	RemoteRequired := true

	if local.FileExists(RequestedLocalPath) { //is file cached?
		ThisFile := local.FindFile(cfg.CachingFiles, RequestedPath)
		if ThisFile == nil || ThisFile.Completed {
			if !local.IsFileExcluded(RequestedLocalPath, &cfg) {
				RemoteRequired = !remote.ServeCachedFile(w, r, RequestedLocalPath, &cfg) //if file has been already cached, ...
			}
		} else { //if file is BEING cached right now, ...
			ThisFile.ServeCachingFile(w, &cfg)
		}
	}

	if RemoteRequired {
		_ = local.BuildDirTreeForFile(RequestedLocalPath)
		remote.ServeFile(w, RequestedPath, &cfg)
	}

}
