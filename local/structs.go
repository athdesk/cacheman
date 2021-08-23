package local

import (
	"net/url"
	"time"
)

type Config struct {
	CacheDir             string
	HostAddr             string
	ChunkSize            int
	MirrorList           []*url.URL
	FullMirrorList       []*url.URL
	MirrorSuffix         string
	MirrorRefreshTimeout time.Duration
	ExcludedExts         []string
	CachingFiles         []*CachingFile //Caching files array is getting carried by Config, as it's program global
	ServerAgent          string
}

type basicCfg struct {
	CacheDir             string
	HostAddr             string
	MirrorlistPath       string
	ChunkSize            int
	MirrorSuffix         string
	MirrorRefreshTimeout int
	ExcludedExts         []string
}

type CachingFile struct {
	ReqPath    string
	LocalPath  string
	BytesRead  int64
	SizeHeader string
	Completed  bool
	InUse      bool
}
