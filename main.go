package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"time"
)

type Attempt struct {
	timestamp time.Time
	url       string
	err       error
	resp      *http.Response
}

func (a *Attempt) MarshalText() ([]byte, error) {
	// TODO
	return nil, nil
}

func main() {
	recycleInterval := flag.Duration("recycle-interval", 10*time.Minute, "duration between recycles")
	keepDuration := flag.Duration("keep-duration", 24*time.Hour, "how long will attempts keep")
	laddr := flag.String("laddr", ":7654", "address that http serve listen to")
	flag.Parse()

	s := NewStore()

	go keepAccessing(s, "https://www.google.com") // TODO

	go func() {
		for now := range time.Tick(*recycleInterval) {
			t := now.Add(-*keepDuration)
			items := s.Load()

			before := 0
			for i, a := range items {
				if a.(*Attempt).timestamp.After(t) {
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

func keepAccessing(s *Store, url string) {
	for {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
		}

		a := &Attempt{
			timestamp: time.Now(),
			url:       url,
			err:       err,
			resp:      resp,
		}
		s.Append(a)

		time.Sleep(10 * time.Minute)
	}
}
