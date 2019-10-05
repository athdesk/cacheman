package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"time"
)

var CacheDir = "/home/mario/Scaricati"
var HostAddr = ":8080"
var ChunkSize = 16384

func main() {
	http.HandleFunc("/", HandleReq)
	http.ListenAndServe(":8080", nil)
}

func HandleReq(w http.ResponseWriter, r *http.Request) {
	//TODO: Sanitize request input
	requestedPath := CacheDir + "/" + r.URL.Path[1:] //Check if file exists in cache directory, not
	fmt.Printf("File requested: %s\n", requestedPath)

	if FileExists(requestedPath) { //is file cached?
		ServeCachedFile(w, requestedPath)
	} else {
		//TODO: Get actual file from remote server, save a copy while serving
		SendNotFound(w)
	}
}

func ServeCachedFile(w http.ResponseWriter, path string) {
	fmt.Printf("Serving file: %s\n", path)

	file, _ := os.Open(path)
	defer file.Close()
	reader := bufio.NewReader(file)
	buffer := make([]byte, ChunkSize)

	startTime := time.Now().UnixNano()
	var totalBytesServed float64

	for {
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

}

func SendNotFound(w http.ResponseWriter) {
	w.WriteHeader(404)
	fmt.Fprintln(w, "Not Found")
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
