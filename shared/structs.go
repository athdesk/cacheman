package shared

import (
	"net/url"
)

type Config struct {
	CacheDir     string
	HostAddr     string
	ChunkSize    int
	MirrorList   []*url.URL
	MirrorSuffix string
	ExcludedExts []string
	CachingFiles []*CachingFile //Caching files array is getting carried by Config, as it's program global
}

type CachingFile struct {
	ReqPath    string
	LocalPath  string
	BytesRead  int64
	SizeHeader string
	Completed  bool
	Errored    bool
	InUse      bool
}
