package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultListenPort int = 8700
	connLimit             = 15
)

var (
	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
)

type SleepHandler struct {
	sem chan struct{}
}

func (l *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case l.sem <- struct{}{}:
		defer func() {
			<-l.sem
		}()

		ms, waitVal, err := sleepValFromURL(r.URL.Path)
		if err != nil {
			http.Error(w, "💤 invalid sleep value", http.StatusNotFound)
			return
		}

		if ms > (60000 * 15) { // 15 minutes
			http.Error(w, fmt.Sprintf("💤 %s is too long to sleep", waitVal), http.StatusNotFound)
			return
		}

		logger.Info("sleeping", "remoteAddr", r.RemoteAddr, "waitVal", waitVal, "ms", ms)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "💤 for %s\n", waitVal)
	default:
		logger.Error("too many connections!", "remoteAddr", r.RemoteAddr)
		http.Error(w, "💤 too many connections", http.StatusTooManyRequests)
	}
}

type Sleeper struct {
	listenPort int
}

func NewSleeper() *Sleeper {
	return &Sleeper{}
}

func (s *Sleeper) Run() {
	var listenStr string

	if s.listenPort == 0 {
		listenStr = fmt.Sprintf(":%d", DefaultListenPort)
	} else {
		listenStr = fmt.Sprintf(":%d", s.listenPort)
	}

	logger.Info("Server started", "listenStr", listenStr)
	sleepHandler := &SleepHandler{
		make(chan struct{}, connLimit),
	}
	if err := http.ListenAndServe(listenStr, sleepHandler); err != nil {
		logger.Error("Unable to setup listener", "err", err)
		os.Exit(1)
	}
}

func sleepValFromURL(path string) (int, string, error) {
	if len(path) <= 1 {
		return 0, "0s", nil
	}
	waitValue := path[1:]

	var v int
	var err error
	if strings.HasSuffix(waitValue, "ms") {
		v, err = strconv.Atoi(strings.TrimSuffix(waitValue, "ms"))
		if err != nil {
			return 0, "", err
		}
	} else {
		v, err = strconv.Atoi(strings.TrimSuffix(waitValue, "s"))
		if err != nil {
			return 0, "", err
		}
		waitValue = strings.TrimSuffix(waitValue, "s") + "s"
		v = v * 1000
	}
	return v, waitValue, nil
}

func main() {
	NewSleeper().Run()
}
