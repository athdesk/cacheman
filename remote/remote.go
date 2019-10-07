package remote

import (
	. "cacheman/shared"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"syscall"
)

func ServeFile(w http.ResponseWriter, ReqPath string, Cfg *Config) {
	Halting := false
	NonExistent := false
	LocalPath := Cfg.CacheDir + "/" + ReqPath

	CurrentMirrorIndex := 0
	var httpClient = new(http.Client)
	var CurrentMirror url.URL
	var PackageURL url.URL
	var ThisFile *CachingFile

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

			ThisFile = AddFileToList(ReqPath, LocalPath, FileSize, Cfg)
			SplitWr := io.MultiWriter(OutFile, w)
			StreamingError := CopyStream(SplitWr, GetResp.Body, ThisFile, Cfg)

			if InnerError, IsOpErr := StreamingError.(*net.OpError); IsOpErr { //unfolds net.OpError, which in our desired case contains
				StreamingError = InnerError.Err //os.SyscallError, which is itself unfolded to show our Errno
				if InnerError, IsSyscallErr := StreamingError.(*os.SyscallError); IsSyscallErr {
					StreamingError = InnerError.Err
				}
			}
			if StreamingError == syscall.EPIPE { //even when the original client stops downloading a file, we continue downloading it
				StreamToFile(OutFile, GetResp.Body, ThisFile, Cfg)
			}
			if StreamingError != nil && StreamingError != syscall.EPIPE {
				Halting = true
			}

			_ = GetResp.Body.Close()
			Cfg.CachingFiles = WaitAndDelete(ThisFile, Cfg.CachingFiles)
			break
		}

	}

	_ = OutFile.Close()

	if Halting {
		_ = os.Remove(LocalPath)
	}
	if NonExistent {
		w.WriteHeader(404)
	}

}

func CopyStream(SplitWriter io.Writer, GetReader io.Reader, FileDesc *CachingFile, Cfg *Config) error {
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

func StreamToFile(FileWriter io.Writer, GetReader io.Reader, FileDesc *CachingFile, Cfg *Config) {
	for { //cycle reads Get-Body, ChunkSize bytes at a time
		BytesRead, CopyErr := io.CopyN(FileWriter, GetReader, int64(Cfg.ChunkSize)) //read body
		FileDesc.BytesRead += BytesRead
		if CopyErr != nil {
			break
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
