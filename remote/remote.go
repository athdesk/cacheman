package remote

import (
	"cacheman/local"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func ServeFile(w http.ResponseWriter, ReqPath string, Cfg *local.Config) {
	fmt.Println("Serving remotely")
	local.BuildDirTreeForFile(ReqPath)
	CurrentMirrorIndex := 0
	Halting := false
	NonExistent := false

	remotePath := strings.ReplaceAll(ReqPath, Cfg.CacheDir, "") //get remote path from local

	var httpClient = new(http.Client)
	var CurrentMirror url.URL
	var PackageURL url.URL

	OutFile, _ := os.Create(ReqPath)

	for { //execute cycle for each mirror, will break if download is successful
		CurrentMirror = *Cfg.MirrorList[CurrentMirrorIndex]
		PackageURL = CurrentMirror
		PackageURL.Path = path.Join(PackageURL.Path, remotePath)
		GetResp, getErr := httpClient.Get(PackageURL.String())
		MirrorBad := false

		if getErr != nil || GetResp.StatusCode == 404 { //is there a problem with the mirror?
			MirrorBad = true
		} else { //mirror ok, start downloading

			FileSize := GetResp.Header.Get("Content-Length") //copy size from remote header
			w.Header().Add("Content-Length", FileSize)       //and send it to our client
			w.WriteHeader(200)

			SplitWr := io.MultiWriter(OutFile, w)
			for { //cycle reads Get-Body, ChunkSize bytes at a time
				_, copyErr := io.CopyN(SplitWr, GetResp.Body, int64(Cfg.ChunkSize)) //read body

				if copyErr != nil && copyErr != io.EOF { //stream errored, not because file is over: delete file and move on
					Halting = true
					fmt.Println(copyErr)
					break
				} else if copyErr == io.EOF { // stream errored because file is over: close gracefully
					GetResp.Body.Close()
					break
				}

			}

		}

		if MirrorBad { //moves to the next mirror, if possible
			CurrentMirrorIndex++
			if CurrentMirrorIndex >= len(Cfg.MirrorList) {
				CurrentMirrorIndex = 0
				NonExistent = true
				Halting = true
				break
			}
		}

	}

	OutFile.Close()
	if Halting {
		os.Remove(ReqPath)
	}
	if NonExistent {
		w.WriteHeader(404)
	}

}
