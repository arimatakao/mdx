package mangadexapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	default_useragent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.3"

	base_url            = "https://api.mangadex.org"
	health_path         = "/ping"
	manga_path          = "/manga"
	specific_manga_path = "/manga/{id}"
)

var (
	ErrUnknown    = errors.New("unknown error")
	ErrBadInput   = errors.New("bad input")
	ErrConnection = errors.New("request is failed")
)

type ErrorDetail struct {
	ID      string `json:"id"`
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	Context string `json:"context"`
}

type ErrorResponse struct {
	Result string        `json:"result"`
	Errors []ErrorDetail `json:"errors"`
}

func (e *ErrorResponse) Error() string {
	errorMsg := fmt.Sprintf("result: %s ; errors: [", e.Result)

	for i, err := range e.Errors {
		errorMsg += fmt.Sprintf("{id: %s, status: %d, title: %s, detail: %s, context: %s}", err.ID, err.Status, err.Title, err.Detail, err.Context)
		if i < len(e.Errors)-1 {
			errorMsg += ", "
		}
	}

	errorMsg += "]"

	return errorMsg
}

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

type MangaTag struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Name        map[string]string `json:"name"`
		Description struct{}          `json:"description"`
		Group       string            `json:"group"`
		Version     int               `json:"version"`
	} `json:"attributes"`
	Relationships []interface{} `json:"relationships"`
}

type MangaAttrib struct {
	Title                          map[string]string   `json:"title"`
	AltTitles                      []map[string]string `json:"altTitles"`
	Description                    map[string]string   `json:"description"`
	IsLocked                       bool                `json:"isLocked"`
	Links                          map[string]string   `json:"links"`
	OriginalLanguage               string              `json:"originalLanguage"`
	LastVolume                     string              `json:"lastVolume"`
	LastChapter                    string              `json:"lastChapter"`
	PublicationDemographic         string              `json:"publicationDemographic"`
	Status                         string              `json:"status"`
	Year                           int                 `json:"year"`
	ContentRating                  string              `json:"contentRating"`
	Tags                           []MangaTag          `json:"tags"`
	State                          string              `json:"state"`
	ChapterNumbersResetOnNewVolume bool                `json:"chapterNumbersResetOnNewVolume"`
	CreatedAt                      time.Time           `json:"createdAt"`
	UpdatedAt                      time.Time           `json:"updatedAt"`
	Version                        int                 `json:"version"`
	AvailableTranslatedLanguages   []string            `json:"availableTranslatedLanguages"`
	LatestUploadedChapter          string              `json:"latestUploadedChapter"`
}

type RelAttribute struct {
	Name string `json:"name"`
}

type Relationship struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"`
	Related    string       `json:"related,omitempty"`
	Attributes RelAttribute `json:"attributes"`
}

type MangaInfo struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    MangaAttrib    `json:"attributes"`
	Relationships []Relationship `json:"relationships"`
}

type MangaList struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []MangaInfo `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

func (a clientapi) Find(title, limit, offset string) (MangaList, error) {
	if title == "" || limit == "" || offset == "" {
		return MangaList{}, ErrBadInput
	}

	mangaList := MangaList{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&mangaList).
		SetQueryParams(map[string]string{
			"title":  title,
			"limit":  limit,
			"offset": offset,
		}).
		Get(manga_path)
	if err != nil {
		return MangaList{}, ErrConnection
	}

	if resp.IsError() {
		return MangaList{}, &respErr
	}

	return mangaList, nil
}

type MangaInfoResponse struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     MangaInfo `json:"data"`
}

func (a clientapi) GetMangaInfo(id string) (MangaInfo, error) {
	if id == "" {
		return MangaInfo{}, ErrBadInput
	}

	info := MangaInfoResponse{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&info).
		SetPathParam("id", id).
		SetQueryString("includes[]=author&includes[]=artist").
		Get(specific_manga_path)
	if err != nil {
		return MangaInfo{}, ErrConnection
	}

	if resp.IsError() {
		return MangaInfo{}, &respErr
	}

	return info.Data, nil
}
