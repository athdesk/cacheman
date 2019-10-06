package local

import (
	"bufio"
	"cacheman/remote"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func ServeFile(w http.ResponseWriter, path string, Cfg *Config) bool {
	AbsPath := strings.ReplaceAll(path, Cfg.CacheDir, "")
	ExpectedSize := remote.GetCorrectSize(AbsPath, Cfg)
	if ExpectedSize != -1 && FileSize(path) != ExpectedSize {
		return false
	}

	file, _ := os.Open(path)
	defer file.Close()
	reader := bufio.NewReader(file)
	buffer := make([]byte, Cfg.ChunkSize) //file is read only ChunkSize bytes at a time

	startTime := time.Now().UnixNano() //time is taken for speed calculation
	var totalBytesServed float64

	for { //cycle is executed until file is over
		bytesRead, _ := reader.Read(buffer)
		if bytesRead == 0 {
			break
		}
		w.Write(buffer)
		totalBytesServed += float64(bytesRead)
	}

	endTime := time.Now().UnixNano()
	deltaTime := endTime - startTime
	avgSpeed := totalBytesServed / float64(deltaTime) // GB/s
	avgSpeed = avgSpeed * 1000000                     // kB/s

	fmt.Printf("Served file %s. Average speed: %f kB/s \n", path, avgSpeed)
	return true
}
