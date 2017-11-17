package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Attempt struct {
	Timestamp time.Time       `json:"timestamp"`
	Sites     []*Availability `json:"sites"`
}

type Availability struct {
	URL        string `json:"url"`
	Err        string `json:"error,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
}

func main() {
	rawURLs := flag.String("urls", "https://www.google.com,https://www.baidu.com", "urls to test accessibility")
	attemptInterval := flag.Duration("attempt-interval", 10*time.Minute, "duration between each attempt")
	recycleInterval := flag.Duration("recycle-interval", 10*time.Minute, "duration between recycles")
	keepDuration := flag.Duration("keep-duration", 24*time.Hour, "how long will attempts keep")
	laddr := flag.String("laddr", ":7654", "address that http serve listen to")
	flag.Parse()

	s := NewStore()

	urls := parseURLs(*rawURLs)

	go keepAccessing(urls, *attemptInterval, s)

	go func() {
		for now := range time.Tick(*recycleInterval) {
			t := now.Add(-*keepDuration)
			items := s.Load()

			before := 0
			for i, a := range items {
				if a.(*Attempt).Timestamp.After(t) {
					break
				}
				before = i
			}

			s.Drop(before)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		items := s.Load()
		length := len(items)
		attempts := make([]*Attempt, len(items))
		for i, item := range items {
			// return in reversed order
			// recent one first
			attempts[length-i-1] = item.(*Attempt)
		}

		enc := json.NewEncoder(w)
		if err := enc.Encode(attempts); err != nil {
			http.Error(w, fmt.Sprintf("could not encode: %s", err), 500)
		}
	})
	http.ListenAndServe(*laddr, nil)
}

func parseURLs(rawURLs string) []string {
	urls := make([]string, 0)
	for _, raw := range strings.Split(rawURLs, ",") {
		trimed := strings.Trim(raw, " ")

		if trimed != "" {
			urls = append(urls, trimed)
		}
	}
	return urls
}

func keepAccessing(urls []string, wait time.Duration, s *Store) {
	var wg sync.WaitGroup

	for {
		attempt := &Attempt{
			Timestamp: time.Now(),
			Sites:     make([]*Availability, len(urls)),
		}

		wg.Add(len(urls))
		for i, url := range urls {
			go func(i int, url string) {
				defer wg.Done()

				a := access(url)
				attempt.Sites[i] = a
			}(i, url)
		}
		wg.Wait()

		s.Append(attempt)

		time.Sleep(wait)
	}
}

func access(url string) *Availability {
	a := &Availability{
		URL: url,
	}

	resp, err := http.Get(url)
	if err != nil {
		a.Err = err.Error()
	} else {
		resp.Body.Close()
		a.StatusCode = resp.StatusCode
	}

	return a
}
