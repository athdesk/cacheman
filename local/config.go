package local

import (
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
)

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
	putMirrorList(Cfg, Intermediary.MirrorlistPath, Intermediary.MirrorMaxAmount)
}
