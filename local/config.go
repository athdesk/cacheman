package local

import (
	"io/ioutil"
	"net/url"
	"regexp"
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

const HttpMirrorRegex = "^http://[A-Za-z0-9.]*/[A-Za-z0-9./$]*$" // Exclude https to avoid duplicates TODO make it a choice

//PutConfig populates a Cfg struct with settings from the config files
func PutConfig(Cfg *Config) {
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
	Cfg.ServerAgent = "cacheman"
	putMirrorList(Cfg, Intermediary.MirrorlistPath)
}

func putMirrorList(Cfg *Config, MirrorlistPath string) {

	if !FileExists(MirrorlistPath) {
		panic("Error reading mirrorlist file")
	}
	MirrorData, Err := ioutil.ReadFile(MirrorlistPath)
	if Err != nil {
		panic(Err)
	}

	MirrorLines := strings.Split(string(MirrorData), "\n")

	RE := regexp.MustCompile(HttpMirrorRegex)
	ValidCounter := 0
	StrMirrorList := make([]string, len(MirrorLines))
	for Index := 0; Index < len(MirrorLines); Index++ {
		Replaced := strings.ReplaceAll(MirrorLines[Index], "Server = ", "")
		if RE.MatchString(Replaced) {
			StrMirrorList[ValidCounter] = Replaced
			ValidCounter++
		}
	}
	if ValidCounter > 0 {
		StrMirrorList = StrMirrorList[0:ValidCounter]
	} else {
		panic("Error parsing mirrorlist file")
	}

	Cfg.FullMirrorList = make([]*url.URL, len(StrMirrorList))

	for Index := 0; Index < len(Cfg.FullMirrorList); Index++ { //strips suffix from mirror urls, parses them
		Cfg.FullMirrorList[Index], Err = url.Parse(strings.ReplaceAll(StrMirrorList[Index], Cfg.MirrorSuffix, ""))
		if Err != nil {
			panic("Error parsing url " + StrMirrorList[Index])
		}
	}
	go checkMirrorStatus(Cfg)

}
