package remote

import (
	"cacheman/local"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

//ServeFile serves the requested file to a http.ResponseWriter
func ServeFile(w http.ResponseWriter, ReqPath string, Cfg *local.Config) {

	//if there's no valid mirrors, wait till we have
	//if it's because of refreshing, everything good
	//if it's because of network issues, pacman will timeout us
	//once timeouted, we download the file anyway to keep it cached
	for len(Cfg.MirrorList) < 1 {
		time.Sleep(100 * time.Millisecond)
	}

	ServeStartTime := time.Now().Unix()
	TimeElapsed := func() int64 { return int64(time.Now().Unix() - ServeStartTime) }

	Halting := false
	LocalPath := Cfg.CacheDir + "/" + ReqPath

	CurrentMirrorIndex := 0
	var httpClient = new(http.Client)
	var CurrentMirror url.URL
	var PackageURL url.URL
	var ThisFile *local.CachingFile

	OutFile, _ := os.Create(LocalPath)

	for { // will break if either download is successful or no mirrors left
		NowStr := time.Now().Format(time.Kitchen)
		CurrentMirror = *Cfg.MirrorList[CurrentMirrorIndex]
		PackageURL = CurrentMirror
		PackageURL.Path = path.Join(PackageURL.Path, ReqPath)
		GetResp, GetErr := httpClient.Get(PackageURL.String())

		//fmt.Printf("[REMOTE %s] Downloading from mirror %d\n", NowStr, CurrentMirrorIndex)

		//is there a problem with the mirror?
		if GetErr != nil || GetResp.StatusCode != 200 { // moves to the next mirror, if possible
			CurrentMirrorIndex++
			if CurrentMirrorIndex >= len(Cfg.MirrorList) || TimeElapsed() > 3 { // Last mirror checked, or request taking too long TODO make timeout time an user choice
				CurrentMirrorIndex = 0
				Halting = true
				fmt.Printf("[REMOTE %s] File unavailable, closing connection for %s\n", NowStr, ReqPath)

				w.Header().Add("Server", Cfg.ServerAgent)
				StatusCodeErr := 500
				if GetResp != nil {
					StatusCodeErr = GetResp.StatusCode
				}
				w.WriteHeader(StatusCodeErr)
				break
			}
		} else { //mirror ok, start downloading

			FileSize := GetResp.Header.Get("Content-Length") //copy size from remote header
			w.Header().Add("Content-Length", FileSize)       //and send it to our client
			w.Header().Add("Server", Cfg.ServerAgent)
			w.WriteHeader(GetResp.StatusCode)

			ThisFile = local.AddFileToList(ReqPath, LocalPath, FileSize, Cfg)
			SplitWr := io.MultiWriter(OutFile, w)
			StreamingError := copyStream(SplitWr, GetResp.Body, ThisFile, Cfg)

			if InnerError, IsOpErr := StreamingError.(*net.OpError); IsOpErr { //unfolds net.OpError, which in our desired case contains
				StreamingError = InnerError.Err //os.SyscallError, which is itself unfolded to show our Errno
				if InnerError, IsSyscallErr := StreamingError.(*os.SyscallError); IsSyscallErr {
					StreamingError = InnerError.Err
				}
			}
			if StreamingError == syscall.EPIPE { //even when the original client stops downloading a file, we continue downloading it
				streamToFile(OutFile, GetResp.Body, ThisFile, Cfg)
			}
			if StreamingError != nil && StreamingError != syscall.EPIPE {
				Halting = true
			}

			_ = GetResp.Body.Close()
			Cfg.CachingFiles = local.WaitAndDelete(ThisFile, Cfg.CachingFiles)
			break
		}

	}

	_ = OutFile.Close()

	if Halting {
		_ = os.Remove(LocalPath)
	}

}

//ServeCachedFile Takes a requests and fulfills it with a cached file
func ServeCachedFile(w http.ResponseWriter, r *http.Request, path string, Cfg *local.Config) bool {
	AbsPath := strings.ReplaceAll(path, Cfg.CacheDir, "")

	ExpectedSize := GetCorrectSize(AbsPath, Cfg)
	RealSize := local.FileSize(path)
	if ExpectedSize != -1 && RealSize != ExpectedSize {
		return false // if filesize is mismatched serve it from remote server, this will redownload the file
	}

	NowStr := time.Now().Format(time.Kitchen)
	fmt.Printf("[LOCAL %s] Serving from storage \n", NowStr)
	http.ServeFile(w, r, path) //does not serve paths containing /../, supports byte ranges
	return true
}
