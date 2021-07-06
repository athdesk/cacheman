package local

import (
	"cacheman/shared"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type basicCfg struct {
	CacheDir             string
	HostAddr             string
	MirrorlistPath       string
	ChunkSize            int
	MirrorSuffix         string
	MirrorRefreshTimeout int
	ExcludedExts         []string
}

//PutConfig populates a Cfg struct with settings from the config files
func PutConfig(Cfg *shared.Config) {

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
	Cfg.CachingFiles = make([]*shared.CachingFile, 0)
	Cfg.ServerAgent = "cacheman"
	putMirrorList(Cfg)
}

func putMirrorList(Cfg *shared.Config) {
	//TODO: get mirrorlist from an actual file
	Cfg.FullMirrorList = make([]*url.URL, 4)
	StrMirrorList := make([]string, 4)
	StrMirrorList[0] = "http://mirror.fra10.de.leaseweb.net/archlinux/$repo/os/$arch"
	StrMirrorList[1] = "http://mirror.tarellia.net/distr/archlinux/$repo/os/$arch"
	StrMirrorList[2] = "http://arch.midov.pl/arch/$repo/os/$arch"
	StrMirrorList[3] = "http://mirror.selfnet.de/archlinux/$repo/os/$arch"

	for Index := 0; Index < len(Cfg.FullMirrorList); Index++ { //strips suffix from mirror urls, parses them
		Cfg.FullMirrorList[Index], _ = url.Parse(strings.ReplaceAll(StrMirrorList[Index], Cfg.MirrorSuffix, ""))
	}
	go checkMirrorStatus(Cfg)
}
