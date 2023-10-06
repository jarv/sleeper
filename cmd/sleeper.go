package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultListenPort int = 8700
	ConnLimit             = 300
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

		ms, waitVal, err := msFromURL(r.URL.Path)
		if err != nil {
			http.Error(w, "💤 invalid sleep value", http.StatusNotFound)
			return
		}

		if ms > (60000 * 15) { // 15 minutes
			http.Error(w, fmt.Sprintf("💤 %s is too long to sleep", waitVal), http.StatusNotFound)
			return
		}

		time.Sleep(time.Duration(ms) * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "💤 for %s\n", waitVal)
	default:
		http.Error(w, "💤 too many requests", http.StatusTooManyRequests)
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

	slog.Info("Server started", "listenStr", listenStr)
	sleepHandler := &SleepHandler{
		make(chan struct{}, ConnLimit),
	}
	log.Fatal(http.ListenAndServe(listenStr, sleepHandler))
}

func msFromURL(path string) (int, string, error) {
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
