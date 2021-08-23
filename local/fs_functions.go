package local

import (
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

func FindFile(DescSlice []*CachingFile, ReqPath string) *CachingFile {
	for _, SliceElement := range DescSlice {
		if SliceElement.ReqPath == ReqPath {
			return SliceElement
		}
	}
	return nil
}
