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
	//TODO: Handle case where 2 computers are downloading the same file
	RequestedPath := Config.CacheDir + "/" + r.URL.Path[1:] //Check if file exists in cache directory, not /
	fmt.Printf("File requested: %s\n", RequestedPath)

	if local.FileExists(RequestedPath) { //is file cached?
		local.ServeFile(w, RequestedPath, &Config)
	} else {
		remote.ServeFile(w, RequestedPath, &Config)
	}
}
