package local

import (
	. "cacheman/shared"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"net/url"
	"strings"
)

type basicCfg struct {
	CacheDir     string
	HostAddr     string
	ChunkSize    int
	MirrorSuffix string
	ExcludedExts []string
}

func GetConfig(Cfg *Config) {

	var Intermediary basicCfg
	ConfigData, _ := ioutil.ReadFile("/etc/cacheman/cacheman.conf")
	_ = toml.Unmarshal(ConfigData, &Intermediary)

	Cfg.CacheDir = Intermediary.CacheDir
	Cfg.HostAddr = Intermediary.HostAddr
	Cfg.ChunkSize = Intermediary.ChunkSize
	Cfg.MirrorSuffix = Intermediary.MirrorSuffix
	Cfg.ExcludedExts = Intermediary.ExcludedExts
	Cfg.CachingFiles = make([]*CachingFile, 0)
}

func GetMirrorList(Cfg *Config) {
	Cfg.MirrorList = make([]*url.URL, 1) //TODO: get mirrorlist from a file
	StrMirrorList := make([]string, 1)
	StrMirrorList[0] = "http://mirrors.prometeus.net/archlinux/$repo/os/$arch"

	for index := 0; index < len(Cfg.MirrorList); index++ { //strips suffix from mirror urls, parses them
		Cfg.MirrorList[index], _ = url.Parse(strings.ReplaceAll(StrMirrorList[index], Cfg.MirrorSuffix, ""))
	}
}
