package local

import (
	. "cacheman/shared"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/sparrc/go-ping"
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
	ConfigData, _ := ioutil.ReadFile("/etc/cacheman/cacheman.conf")
	_ = toml.Unmarshal(ConfigData, &Intermediary)

	Cfg.CacheDir = Intermediary.CacheDir
	Cfg.HostAddr = Intermediary.HostAddr
	Cfg.ChunkSize = Intermediary.ChunkSize
	Cfg.MirrorSuffix = Intermediary.MirrorSuffix
	Cfg.MirrorRefreshTimeout = time.Duration(Intermediary.MirrorRefreshTimeout) * time.Second
	Cfg.ExcludedExts = Intermediary.ExcludedExts
	Cfg.CachingFiles = make([]*CachingFile, 0)
}

func GetMirrorList(Cfg *Config) {
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

func checkMirrorStatus(Cfg *Config) {
	//TODO: parallelize this
	for {
		fmt.Println("[MIRROR] Refreshing valid mirror list...")
		var ValidMirrors []*url.URL
		for _, Mirror := range Cfg.FullMirrorList {
			if isAlive(*Mirror) {
				ValidMirrors = append(ValidMirrors, Mirror)
				fmt.Printf("[MIRROR] %s is alive!\n", Mirror.Host)
			} else {
				fmt.Printf("[MIRROR] %s is dead!\n", Mirror.Host)
			}
		}
		Cfg.MirrorList = ValidMirrors
		MirTimeout := 10 * time.Millisecond //if we have no mirrors, don't wait for refresh timeout
		if len(ValidMirrors) > 0 {
			MirTimeout = Cfg.MirrorRefreshTimeout
		}
		time.Sleep(MirTimeout)
	}
}

func isAlive(url url.URL) bool {
	//TODO: sudo sysctl -w net.ipv4.ping_group_range="0   2147483647"
	Address := url.Host
	HostPinger, Err := ping.NewPinger(Address)
	if Err != nil {
		return false
	} //if the name is malformed/host doesn't exist

	HostPinger.Count = 2
	HostPinger.Timeout = 1200 * time.Millisecond

	HostPinger.Run()
	return HostPinger.Statistics().PacketsRecv > 0
}
