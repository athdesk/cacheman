package remote

import (
	"cacheman/local"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

func copyStream(SplitWriter io.Writer, GetReader io.Reader, FileDesc *local.CachingFile, Cfg *local.Config) error {
	for { //cycle reads Get-Body, ChunkSize bytes at a time
		BytesRead, CopyErr := io.CopyN(SplitWriter, GetReader, int64(Cfg.ChunkSize)) //read body
		FileDesc.BytesRead += BytesRead

		if CopyErr != nil && CopyErr != io.EOF && CopyErr != io.ErrClosedPipe {
			return CopyErr //stream errored, not because file is over/connection closed: delete file and move on
		} else if CopyErr == io.EOF { // stream errored because file is over: close gracefully
			return nil
		} else if CopyErr != nil {
			return CopyErr
		}

	}
}

func streamToFile(FileWriter io.Writer, GetReader io.Reader, FileDesc *local.CachingFile, Cfg *local.Config) {
	for { //cycle reads Get-Body, ChunkSize bytes at a time
		BytesRead, CopyErr := io.CopyN(FileWriter, GetReader, int64(Cfg.ChunkSize)) //read body
		FileDesc.BytesRead += BytesRead
		if CopyErr != nil {
			break
		}
	}
}

//GetCorrectSize returns the Content-Length of a file, returns -1 if can't get to package
func GetCorrectSize(ReqPath string, Cfg *local.Config) int64 {
	CurrentMirrorIndex := 0
	var httpClient = new(http.Client)
	var CurrentMirror url.URL
	var PackageURL url.URL

	ServeStartTime := time.Now().Unix()
	TimeElapsed := func() int64 { return int64(time.Now().Unix() - ServeStartTime) }

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
			if CurrentMirrorIndex >= len(Cfg.MirrorList) || TimeElapsed() > 3 {
				CurrentMirrorIndex = 0
				return -1
			}
		} else { //if mirror replied, get size header
			FileSize, _ := strconv.ParseInt(GetResp.Header.Get("Content-Length"), 10, 64)
			return FileSize
		}

	}
}
