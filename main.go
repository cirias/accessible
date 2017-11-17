package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type Attempt struct {
	Timestamp  time.Time `json:"timestamp"`
	URL        string    `json:"url"`
	Err        string    `json:"error,omitempty"`
	StatusCode int       `json:"statusCode,omitempty"`
}

func main() {
	url := flag.String("url", "https://www.google.com", "url to test accessibility")
	attemptInterval := flag.Duration("attempt-interval", 10*time.Minute, "duration between each attempt")
	recycleInterval := flag.Duration("recycle-interval", 10*time.Minute, "duration between recycles")
	keepDuration := flag.Duration("keep-duration", 24*time.Hour, "how long will attempts keep")
	laddr := flag.String("laddr", ":7654", "address that http serve listen to")
	flag.Parse()

	s := NewStore()

	go keepAccessing(*url, *attemptInterval, s)

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
		attempts := make([]*Attempt, len(items))
		for i, item := range items {
			attempts[i] = item.(*Attempt)
		}

		enc := json.NewEncoder(w)
		if err := enc.Encode(attempts); err != nil {
			http.Error(w, fmt.Sprintf("could not encode: %s", err), 500)
		}
	})
	http.ListenAndServe(*laddr, nil)
}

func keepAccessing(url string, wait time.Duration, s *Store) {
	for {
		a := &Attempt{
			Timestamp: time.Now(),
			URL:       url,
		}

		resp, err := http.Get(url)
		if err != nil {
			a.Err = err.Error()
		} else {
			resp.Body.Close()
			a.StatusCode = resp.StatusCode
		}

		s.Append(a)

		time.Sleep(wait)
	}
}
