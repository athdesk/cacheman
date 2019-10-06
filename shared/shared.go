package shared

import "net/url"

type Config struct {
	CacheDir     string
	HostAddr     string
	ChunkSize    int
	MirrorList   []*url.URL
	MirrorSuffix string
}
