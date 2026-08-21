//go:build debug
// +build debug

package debug

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
)

const (
	defaultPprofAddr       = "127.0.0.1:6060"
	defaultProfileFraction = 100
)

func init() {
	var (
		addr          = defaultPprofAddr
		blockRate     = defaultProfileFraction
		mutexFraction = defaultProfileFraction
	)

	if v := os.Getenv("DEBUG_BLOCK_RATE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			blockRate = n
		}
	}

	if v := os.Getenv("DEBUG_MUTEX_FRACTION"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			mutexFraction = n
		}
	}

	if v := os.Getenv("DEBUG_PPROF_ADDR"); v != "" {
		addr = v
	}

	runtime.SetBlockProfileRate(blockRate)
	runtime.SetMutexProfileFraction(mutexFraction)

	go func() {
		log.Printf("debug: pprof server on http://%s/debug/pprof/", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("debug: pprof server stopped: %v", err)
		}
	}()
}
