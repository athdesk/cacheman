package shared

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
)

func AddFileToList(ReqPath string, LocalPath string, SizeHeader string, Cfg *Config) *CachingFile {
	FileSlice := CachingFile{
		ReqPath:    ReqPath,
		LocalPath:  LocalPath,
		BytesRead:  0,
		SizeHeader: SizeHeader,
	}

	var Overwritten bool

	for _, CurrentFile := range Cfg.CachingFiles { //check if there's an overwriteable entry
		if CurrentFile.Completed && !CurrentFile.InUse {
			CurrentFile = &FileSlice
			Overwritten = true
			break
		}
	}

	if !Overwritten {
		Cfg.CachingFiles = append(Cfg.CachingFiles, &FileSlice)
	}
	return &FileSlice
}

func (FileDesc *CachingFile) ServeCachingFile(w http.ResponseWriter, Cfg *Config) {
	FileDesc.InUse = true
	File, _ := os.Open(FileDesc.LocalPath)
	defer File.Close()

	var TotalBytesRead int64

	FileReader := bufio.NewReader(File)
	DataBuffer := make([]byte, Cfg.ChunkSize)

	w.Header().Add("Content-Length", FileDesc.SizeHeader)
	w.WriteHeader(200)

	for {
		BytesRead, _ := FileReader.Read(DataBuffer)
		TotalBytesRead += int64(BytesRead)
		if BytesRead == 0 {
			if FileDesc.Completed || FileDesc.Errored {
				break
			}
		}
		BufferReader := bytes.NewReader(DataBuffer)
		_, _ = io.CopyN(w, BufferReader, int64(BytesRead))
	}

	FileDesc.InUse = false
}

func FindFile(DescSlice []*CachingFile, ReqPath string) *CachingFile {
	for _, SliceElement := range DescSlice {
		if SliceElement.ReqPath == ReqPath {
			return SliceElement
		}
	}
	return nil
}
