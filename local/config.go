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
	StrMirrorList[0] = "http://mirrors.nonexistent.com/archlinux/$repo/os/$arch"
	StrMirrorList[1] = "http://mirrors.prometeusaa.net/archlinux/$repo/os/$arch"
	StrMirrorList[2] = "http://mirrors.nonexistentasda.com/archlinux/$repo/os/$arch"
	StrMirrorList[3] = "http://mirrors.prometeus.net/archlinux/$repo/os/$arch"

	for Index := 0; Index < len(Cfg.FullMirrorList); Index++ { //strips suffix from mirror urls, parses them
		Cfg.FullMirrorList[Index], _ = url.Parse(strings.ReplaceAll(StrMirrorList[Index], Cfg.MirrorSuffix, ""))
	}
	go checkMirrorStatus(Cfg)
}

/*
EXAMPLE CONFIG FILE

# Directory where cacheman is going to store packages
# make sure it's accessible by the user that's going to run cacheman
CacheDir = "/home/mario/cacheman"

# Address and port on which cacheman is going to run
# use ":PORT" to serve all interfaces
# make sure to use an high enough port if you haven't got enough privileges
HostAddr = ":8080"

# Path to the mirrorlist that cacheman should use
MirrorlistPath = "/etc/pacman.d/mirrorlist"

# Buffer size for streaming operations
ChunkSize = 1024

# Mirror URL suffix, please don't touch if you don't know what you're doing
MirrorSuffix = "$repo/os/$arch"

# Time range in seconds every which the mirrorlist will be checked for dead hosts
MirrorRefreshTimeout = 300

# Extensions to exclude from caching, please don't touch if you don't know what you're doing
ExcludedExts = ["db", "sig"]


*/
