package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var CacheDir = "/home/mario/cacheman" //TODO: get config from a file
var HostAddr = ":8080"
var ChunkSize = 1024
var MirrorList []*url.URL
var MirrorSuffix = "$repo/os/$arch"

func main() {
	GetMirrorList()
	//TODO: handle errors
	http.HandleFunc("/", HandleReq)
	http.ListenAndServe(":8080", nil)
}

func GetMirrorList() {
	MirrorList = make([]*url.URL, 1) //TODO: get mirrorlist from a file
	StrMirrorList := make([]string, 1)
	StrMirrorList[0] = "http://mirrors.prometeus.net/archlinux/$repo/os/$arch"

	for index := 0; index < len(MirrorList); index++ { //strips suffix from mirror urls, parses them
		MirrorList[index], _ = url.Parse(strings.ReplaceAll(StrMirrorList[index], MirrorSuffix, ""))
	}
}

func HandleReq(w http.ResponseWriter, r *http.Request) {
	//TODO: Sanitize request input
	requestedPath := CacheDir + "/" + r.URL.Path[1:] //Check if file exists in cache directory, not /
	fmt.Printf("File requested: %s\n", requestedPath)

	if FileExists(requestedPath) { //is file cached?
		ServeCachedFile(w, requestedPath)
	} else {
		ServeRemoteFile(w, requestedPath)
	}
}

func ServeRemoteFile(w http.ResponseWriter, reqPath string) {
	fmt.Println("Serving remotely")
	BuildDirTreeForFile(reqPath)
	currentMirrorIndex := 0
	halting := false
	nonExistent := false

	remotePath := strings.ReplaceAll(reqPath, CacheDir, "")

	var httpClient = new(http.Client)
	var currentMirror url.URL
	var packageURL url.URL

	out, _ := os.Create(reqPath)

	for { //execute cycle for each mirror, will break if download is successful
		currentMirror = *MirrorList[currentMirrorIndex]
		packageURL = currentMirror
		packageURL.Path = path.Join(packageURL.Path, remotePath)
		getResp, getErr := httpClient.Get(packageURL.String())
		mirrorBad := false

		if getErr != nil || getResp.StatusCode == 404 { //is there a problem with the mirror?
			mirrorBad = true
		} else { //mirror ok, start downloading
			fileSize := getResp.Header.Get("Content-Length")
			w.Header().Add("Content-Length", fileSize)
			w.WriteHeader(200)
			splitWr := io.MultiWriter(out, w)

			for { //cycle reads Get>Body, ChunkSize bytes at a time
				_, copyErr := io.CopyN(splitWr, getResp.Body, int64(ChunkSize)) //read body

				if copyErr != nil && copyErr != io.EOF { //stream errored, not because file is over: delete file and move on
					halting = true
					fmt.Println(copyErr)
					break
				} else if copyErr == io.EOF { // stream errored because file is over: close gracefully
					getResp.Body.Close()
					break
				}

			}

		}

		if mirrorBad { //moves to the next mirror, if possible
			currentMirrorIndex++
			if currentMirrorIndex >= len(MirrorList) {
				currentMirrorIndex = 0
				nonExistent = true
				halting = true
				break
			}
		}

	}

	out.Close()
	if halting {
		os.Remove(reqPath)
	}
	if nonExistent {
		w.WriteHeader(404)
	}

}

func BuildDirTreeForFile(path string) {
	realPath := filepath.Dir(path)
	if !DirExists(realPath) {
		os.MkdirAll(realPath, 0755)
	}
}

func ServeCachedFile(w http.ResponseWriter, path string) {

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

func DirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
