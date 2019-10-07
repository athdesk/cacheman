package shared

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"os"
	"time"
)

func AddFileToList(ReqPath string, LocalPath string, SizeHeader string, Cfg *Config) *CachingFile {
	FileSlice := CachingFile{
		ReqPath:    ReqPath,
		LocalPath:  LocalPath,
		BytesRead:  0,
		SizeHeader: SizeHeader,
	}

	Cfg.CachingFiles = append(Cfg.CachingFiles, &FileSlice)
	return &FileSlice
}

func WaitAndDelete(ElementPtr *CachingFile, Slice []*CachingFile) []*CachingFile {
	ElementPtr.Completed = true
	for ElementPtr.InUse {
		time.Sleep(10 * time.Millisecond)
	}
	return FindAndDelete(ElementPtr, Slice)
}

func FindAndDelete(ElementPtr *CachingFile, Slice []*CachingFile) []*CachingFile {
	var Index int
	var Exists bool
	var CurrentElemPtr *CachingFile
	for Index, CurrentElemPtr = range Slice {
		if ElementPtr == CurrentElemPtr {
			Exists = true
			break
		}
	}

	if Exists { //erasing element, order and place in memory will change, has no effect because we're using a pointer array
		Slice[Index] = Slice[len(Slice)-1] //last element overwrites Indexth element
		Slice[len(Slice)-1] = nil          //last element is erased
		Slice = Slice[:len(Slice)-1]       //last element is excluded from slice
	}

	return Slice

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
			if FileDesc.Completed {
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
