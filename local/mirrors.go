package local

import (
	"cacheman/shared"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func checkMirrorStatus(Cfg *shared.Config) {
	NowStr := time.Now().Format(time.Kitchen)
	for {
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

func checkAndAdd(Mirror *url.URL, ValidMirrors *[]*url.URL, Counter *int, Cfg *shared.Config) {
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
