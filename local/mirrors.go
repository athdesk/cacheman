package local

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

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

func checkMirrorStatus(Cfg *Config) {
	for {
		NowStr := time.Now().Format(time.Kitchen)
		fmt.Printf("[MIRROR %s] Refreshing valid mirror list...\n", NowStr)
		var CompletedJobs = 0
		var StartedJobs = 0
		var ValidMirrors []*url.URL

		for _, Mirror := range Cfg.FullMirrorList {
			StartedJobs++
			go checkAndAdd(Mirror, &ValidMirrors, &CompletedJobs, Cfg)
		}

		for StartedJobs > CompletedJobs {
			time.Sleep(10 * time.Millisecond)
		}

		MirTimeout := 1000 * time.Millisecond //if we have no mirrors, don't wait for refresh timeout
		if len(ValidMirrors) > 0 {
			MirTimeout = Cfg.MirrorRefreshTimeout
		}
		fmt.Printf("[MIRROR %s] %d out of %d mirrors are valid\n", NowStr, len(ValidMirrors), len(Cfg.FullMirrorList))
		time.Sleep(MirTimeout)
	}
}

func checkAndAdd(Mirror *url.URL, ValidMirrors *[]*url.URL, Counter *int, Cfg *Config) {
	NowStr := time.Now().Format(time.Kitchen)
	if isAlive(*Mirror) {
		*ValidMirrors = append(*ValidMirrors, Mirror)
		fmt.Printf("[MIRROR %s] %s is alive!\n", NowStr, Mirror.Host)
		Cfg.MirrorList = *ValidMirrors
	} else {
		fmt.Printf("[MIRROR %s] %s is dead!\n", NowStr, Mirror.Host)
	}
	*Counter++
}

func isAlive(url url.URL) bool {
	_, Err := http.Get(url.String())
	if Err != nil {
		return false
	}
	return true
}
