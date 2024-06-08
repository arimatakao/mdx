package mangadexapi

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	default_useragent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.3"

	base_url    = "https://api.mangadex.org"
	health_path = "/ping"
	manga_path  = "/manga"
)

type clientapi struct {
	c *resty.Client
}

func NewClient(userAgent string) clientapi {
	if userAgent == "" {
		userAgent = default_useragent
	}
	c := resty.New().
		SetBaseURL(base_url).
		SetHeader("User-Agent", userAgent)
	return clientapi{
		c: c,
	}
}

func (a clientapi) Ping() bool {
	resp, err := a.c.R().Get(health_path)
	if err != nil {
		return false
	}

	if resp.StatusCode() != http.StatusOK {
		return false
	}

	return true
}
