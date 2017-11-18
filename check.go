package accessible

import (
	"net/http"
	"time"
)

type Result struct {
	URL         string        `json:"url"`
	ElapsedTime time.Duration `json:"elapsedTime"`
	Err         string        `json:"error,omitempty"`
	StatusCode  int           `json:"statusCode,omitempty"`
}

func (r *Result) IsFailed() bool {
	return r.StatusCode != http.StatusOK || r.ElapsedTime > 5*time.Second
}

func Check(url string) *Result {
	r := &Result{
		URL: url,
	}

	start := time.Now()
	resp, err := http.Get(url)
	r.ElapsedTime = time.Now().Sub(start)
	if err != nil {
		r.Err = err.Error()
	} else {
		resp.Body.Close()
		r.StatusCode = resp.StatusCode
	}

	return r
}
