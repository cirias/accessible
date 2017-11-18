package accessible

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	username   string
	password   string
	entrypoint string
	httpc      *http.Client
}

func (c *Client) Check(target string) (*Result, error) {
	u := fmt.Sprintf("%s/check?url=%s", c.entrypoint, url.PathEscape(target))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("could not new request: %s", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get: %s", err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var r Result
	err = dec.Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("could not decode: %s", err)
	}

	return &r, nil
}

func (c *Client) Poll(ctx context.Context, out chan<- *Result, url string, interval time.Duration) error {
	for {
		r, err := c.Check(url)
		if err != nil {
			return fmt.Errorf("could not check: %s", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- r:
		}

		time.Sleep(interval)
	}
}
