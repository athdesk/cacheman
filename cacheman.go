package main

import (
	"cacheman/local"
	"cacheman/remote"
	. "cacheman/shared"
	"fmt"
	"log"
	"net/http"
	"time"
)

var Cfg Config

func main() {
	local.GetConfig(&Cfg)
	go func() {
		for {
			fmt.Println(Cfg)
			time.Sleep(10 * time.Second)
		}
	}()
	local.GetMirrorList(&Cfg)
	//TODO: handle errors
	http.HandleFunc("/", HandleReq)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HandleReq(w http.ResponseWriter, r *http.Request) {
	//TODO: Handle case where 2+ computers are downloading the same file
	RequestedLocalPath := Cfg.CacheDir + "/" + r.URL.Path[1:] //add cachedir to path, to not check in /
	RequestedPath := r.URL.Path[1:]
	fmt.Printf("File requested: %s\n", RequestedPath)

	RemoteRequired := true

	if local.FileExists(RequestedLocalPath) { //is file cached?
		ThisFile := FindFile(Cfg.CachingFiles, RequestedPath)
		if ThisFile == nil || ThisFile.Completed {
			if !local.IsFileExcluded(RequestedLocalPath, &Cfg) {
				RemoteRequired = !local.ServeCachedFile(w, r, RequestedLocalPath, &Cfg) //if file has been already cached, ...
			}
		} else { //if file is BEING cached right now, ...
			ThisFile.ServeCachingFile(w, &Cfg)
		}
	}

	if RemoteRequired {
		_ = local.BuildDirTreeForFile(RequestedLocalPath)
		remote.ServeFile(w, RequestedPath, &Cfg)
	}

}
