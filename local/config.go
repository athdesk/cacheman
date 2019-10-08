package local

import (
	. "cacheman/shared"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net/url"
	"strings"
	"time"
)

type basicCfg struct {
	CacheDir             string
	HostAddr             string
	ChunkSize            int
	MirrorSuffix         string
	MirrorRefreshTimeout int
	ExcludedExts         []string
}

func GetConfig(Cfg *Config) {

	var Intermediary basicCfg

	ConfigData, Err := ioutil.ReadFile("/etc/cacheman/cacheman.conf")
	if Err != nil {
		panic(Err)
	}

	Err = toml.Unmarshal(ConfigData, &Intermediary)
	if Err != nil {
		panic(Err)
	}

	Cfg.CacheDir = Intermediary.CacheDir
	Cfg.HostAddr = Intermediary.HostAddr
	Cfg.ChunkSize = Intermediary.ChunkSize
	Cfg.MirrorSuffix = Intermediary.MirrorSuffix
	Cfg.MirrorRefreshTimeout = time.Duration(Intermediary.MirrorRefreshTimeout) * time.Second
	Cfg.ExcludedExts = Intermediary.ExcludedExts
	Cfg.CachingFiles = make([]*CachingFile, 0)
	getMirrorList(Cfg)
}

func getMirrorList(Cfg *Config) {
	//TODO: get mirrorlist from a file
	Cfg.FullMirrorList = make([]*url.URL, 4)
	StrMirrorList := make([]string, 4)
	StrMirrorList[0] = "http://mirrors.nonexistent.com/archlinux/$repo/os/$arch"
	StrMirrorList[1] = "http://mirrors.prometeusaa.net/archlinux/$repo/os/$arch"
	StrMirrorList[2] = "http://mirrors.nonexistentasda.com/archlinux/$repo/os/$arch"
	StrMirrorList[3] = "http://mirrors.prometeus.net/archlinux/$repo/os/$arch"

	for Index := 0; Index < len(Cfg.FullMirrorList); Index++ { //strips suffix from mirror urls, parses them
		Cfg.FullMirrorList[Index], _ = url.Parse(strings.ReplaceAll(StrMirrorList[Index], Cfg.MirrorSuffix, ""))
	}
	go checkMirrorStatus(Cfg)
}
