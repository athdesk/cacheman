package main

import (
	"cacheman/local"
	"cacheman/remote"
	"fmt"
	"net/http"
)

var Config local.Config

func main() {
	local.GetConfig(&Config)
	local.GetMirrorList(&Config)
	//TODO: handle errors
	http.HandleFunc("/", HandleReq)
	http.ListenAndServe(":8080", nil)
}

func HandleReq(w http.ResponseWriter, r *http.Request) {
	//TODO: Sanitize request input
	//TODO: Handle case where 2+ computers are downloading the same file
	RequestedLocalPath := Config.CacheDir + "/" + r.URL.Path[1:] //add cachedir to path, to not check in /
	RequestedPath := r.URL.Path[1:]
	fmt.Printf("File requested: %s\n", RequestedPath)

	RemoteRequired := true

	if local.FileExists(RequestedLocalPath) { //is file cached?
		RemoteRequired = !local.ServeFile(w, RequestedLocalPath, &Config)
	}

	if RemoteRequired {
		remote.ServeFile(w, RequestedPath, &Config)
	}

}
