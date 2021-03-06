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
	Username   string
	Password   string
	Entrypoint string
	Httpc      *http.Client
}

func (c *Client) Check(target string) (*Result, error) {
	u := fmt.Sprintf("%s/check?url=%s", c.Entrypoint, url.PathEscape(target))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("could not new request: %s", err)
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.Httpc.Do(req)
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

func (c *Client) Poll(ctx context.Context, h func(*Result, error) error, url string, interval time.Duration) error {
	for {
		err := h(c.Check(url))
		if err != nil {
			return fmt.Errorf("could not handle: %s", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		time.Sleep(interval)
	}
}
