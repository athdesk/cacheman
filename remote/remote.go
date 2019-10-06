package remote

import (
	. "cacheman/shared"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

func ServeFile(w http.ResponseWriter, ReqPath string, Cfg *Config) {
	fmt.Println("Serving remotely")

	Halting := false
	NonExistent := false
	LocalPath := Cfg.CacheDir + "/" + ReqPath

	CurrentMirrorIndex := 0
	var httpClient = new(http.Client)
	var CurrentMirror url.URL
	var PackageURL url.URL

	OutFile, _ := os.Create(LocalPath)

	for { //execute cycle for each mirror, will break if download is successful
		CurrentMirror = *Cfg.MirrorList[CurrentMirrorIndex]
		PackageURL = CurrentMirror
		PackageURL.Path = path.Join(PackageURL.Path, ReqPath)
		GetResp, GetErr := httpClient.Get(PackageURL.String())
		MirrorBad := false

		if GetErr != nil || GetResp.StatusCode == 404 {
			MirrorBad = true
		} //is there a problem with the mirror?

		if MirrorBad || GetErr != nil { //moves to the next mirror, if possible
			CurrentMirrorIndex++
			if CurrentMirrorIndex >= len(Cfg.MirrorList) {
				CurrentMirrorIndex = 0
				NonExistent = true
				Halting = true
				break
			}
		} else { //mirror ok, start downloading

			FileSize := GetResp.Header.Get("Content-Length") //copy size from remote header
			w.Header().Add("Content-Length", FileSize)       //and send it to our client
			w.WriteHeader(200)

			SplitWr := io.MultiWriter(OutFile, w)
			StreamingError := CopyStream(SplitWr, GetResp.Body, Cfg)
			Halting = StreamingError != nil //if there's an error, delete the file
			_ = GetResp.Body.Close()
			break
		}

	}

	OutFile.Close()
	if Halting {
		_ = os.Remove(LocalPath)
	}
	if NonExistent {
		w.WriteHeader(404)
	}

}

func CopyStream(SplitWriter io.Writer, GetReader io.Reader, Cfg *Config) error {
	for { //cycle reads Get-Body, ChunkSize bytes at a time
		_, copyErr := io.CopyN(SplitWriter, GetReader, int64(Cfg.ChunkSize)) //read body

		if copyErr != nil && copyErr != io.EOF { //stream errored, not because file is over: delete file and move on
			return copyErr
		} else if copyErr == io.EOF { // stream errored because file is over: close gracefully
			return nil
		}

	}
}

//Returns -1 if can't get to package
func GetCorrectSize(ReqPath string, Cfg *Config) int64 {
	CurrentMirrorIndex := 0
	var httpClient = new(http.Client)
	var CurrentMirror url.URL
	var PackageURL url.URL

	for { //execute cycle for each mirror, will break if download is successful
		CurrentMirror = *Cfg.MirrorList[CurrentMirrorIndex]
		PackageURL = CurrentMirror
		PackageURL.Path = path.Join(PackageURL.Path, ReqPath)
		GetResp, getErr := httpClient.Get(PackageURL.String())
		MirrorBad := false

		if getErr != nil || GetResp.StatusCode == 404 {
			MirrorBad = true
		} //is there a problem with the mirror?

		if MirrorBad { //moves to the next mirror, if possible
			CurrentMirrorIndex++
			if CurrentMirrorIndex >= len(Cfg.MirrorList) {
				CurrentMirrorIndex = 0
				return -1 //
			}
		} else { //if mirror replied, get size header
			FileSize, _ := strconv.ParseInt(GetResp.Header.Get("Content-Length"), 10, 64)
			return FileSize
		}

	}
}
