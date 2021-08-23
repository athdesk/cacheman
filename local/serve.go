package local

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

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
