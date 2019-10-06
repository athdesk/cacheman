package main

import (
	"cacheman/local"
	"cacheman/remote"
	. "cacheman/shared"
	"fmt"
	"net/http"
)

var Cfg Config

func main() {
	local.GetConfig(&Cfg)
	local.GetMirrorList(&Cfg)
	//TODO: handle errors
	http.HandleFunc("/", HandleReq)
	http.ListenAndServe(":8080", nil)
}

func HandleReq(w http.ResponseWriter, r *http.Request) {
	//TODO: Sanitize request input
	//TODO: Handle case where 2+ computers are downloading the same file
	RequestedLocalPath := Cfg.CacheDir + "/" + r.URL.Path[1:] //add cachedir to path, to not check in /
	RequestedPath := r.URL.Path[1:]
	fmt.Printf("File requested: %s\n", RequestedPath)

	RemoteRequired := true

	if local.FileExists(RequestedLocalPath) { //is file cached?
		RemoteRequired = !local.ServeFile(w, RequestedLocalPath, &Cfg)
	}

	if RemoteRequired {
		local.BuildDirTreeForFile(RequestedLocalPath)
		remote.ServeFile(w, RequestedPath, &Cfg)
	}

}
