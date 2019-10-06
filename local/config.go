package local

import (
	"net/url"
	"strings"
)

type Config struct {
	CacheDir     string
	HostAddr     string
	ChunkSize    int
	MirrorList   []*url.URL
	MirrorSuffix string
}

func GetConfig(Cfg *Config) {
	Cfg.CacheDir = "/home/mario/cacheman" //TODO: get config from a file
	Cfg.HostAddr = ":8080"
	Cfg.ChunkSize = 1024
	Cfg.MirrorSuffix = "$repo/os/$arch"
}

func GetMirrorList(Cfg *Config) {
	Cfg.MirrorList = make([]*url.URL, 1) //TODO: get mirrorlist from a file
	StrMirrorList := make([]string, 1)
	StrMirrorList[0] = "http://mirrors.prometeus.net/archlinux/$repo/os/$arch"

	for index := 0; index < len(Cfg.MirrorList); index++ { //strips suffix from mirror urls, parses them
		Cfg.MirrorList[index], _ = url.Parse(strings.ReplaceAll(StrMirrorList[index], Cfg.MirrorSuffix, ""))
	}
}
