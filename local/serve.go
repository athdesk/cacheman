package local

import (
	"bufio"
	"bytes"
	"cacheman/remote"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

//ServeCachedFile Takes a requests and fulfills it with a cached file
func ServeCachedFile(w http.ResponseWriter, r *http.Request, path string, Cfg *Config) bool {
	AbsPath := strings.ReplaceAll(path, Cfg.CacheDir, "")

	ExpectedSize := remote.GetCorrectSize(AbsPath, Cfg)
	RealSize := FileSize(path)
	if ExpectedSize != -1 && RealSize != ExpectedSize {
		return false // if filesize is mismatched serve it from remote server, this will redownload the file
	}

	NowStr := time.Now().Format(time.Kitchen)
	fmt.Printf("[LOCAL %s] Serving from storage \n", NowStr)
	http.ServeFile(w, r, path) //does not serve paths containing /../, supports byte ranges
	return true
}

func (FileDesc *CachingFile) ServeCachingFile(w http.ResponseWriter, Cfg *Config) {
	FileDesc.InUse = true
	File, Err := os.Open(FileDesc.LocalPath)

	if Err != nil { // could not open the file; drop with status 500
		w.Header().Add("Server", Cfg.ServerAgent)
		w.WriteHeader(500)
		return
	}

	defer File.Close()

	var TotalBytesRead int64

	FileReader := bufio.NewReader(File)
	DataBuffer := make([]byte, Cfg.ChunkSize)

	NowStr := time.Now().Format(time.Kitchen)
	fmt.Printf("[SERVER %s] Client attached to ongoing download\n", NowStr)

	w.Header().Add("Content-Length", FileDesc.SizeHeader)
	w.Header().Add("Server", Cfg.ServerAgent)
	w.WriteHeader(200)

	for {
		BytesRead, _ := FileReader.Read(DataBuffer) // This serves the file that's being written, directly from disk
		TotalBytesRead += int64(BytesRead)          // Is it I/O efficient? No. Is it easy and memory efficient? Yes.
		if BytesRead == 0 {
			if FileDesc.Completed {
				break
			}
		}
		BufferReader := bytes.NewReader(DataBuffer)
		_, _ = io.CopyN(w, BufferReader, int64(BytesRead))
	}

	FileDesc.InUse = false
}
