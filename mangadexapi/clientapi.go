package mangadexapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	default_useragent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.3"

	base_url                   = "https://api.mangadex.org"
	health_path                = "/ping"
	manga_path                 = "/manga"
	specific_manga_path        = "/manga/{id}"
	manga_feed_path            = "/manga/{id}/feed"
	chapter_images_path        = "/at-home/server/{id}"
	download_high_quility_path = "/data/{chapterHash}/{imageFilename}"
	download_low_quility_path  = "/data-saver/{chapterHash}/{imageFilename}"
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
		errorMsg += fmt.Sprintf("{id: %s, status: %d, title: %s, detail: %s, context: %s}",
			err.ID, err.Status, err.Title, err.Detail, err.Context)
		if i < len(e.Errors)-1 {
			errorMsg += ", "
		}
	}
	errorMsg += "]"

	return errorMsg
}

func GetMangaIdFromUrl(link string) string {
	parsedUrl, err := url.Parse(link)
	if err != nil {
		return ""
	}

	if parsedUrl.Host != "mangadex.org" {
		return ""
	}

	paths := strings.Split(parsedUrl.Path, "/")
	if len(paths) < 3 {
		return ""
	}
	return paths[2]
}

func GetMangaIdFromArg(args []string) string {
	for _, arg := range args {
		if u := GetMangaIdFromUrl(arg); u != "" {
			return u
		}
	}
	return ""
}

type clientapi struct {
	c *resty.Client
}

func NewClient(userAgent string) clientapi {
	if userAgent == "" {
		userAgent = default_useragent
	}

	c := resty.New().
		SetRetryCount(5).
		SetRetryWaitTime(time.Millisecond*200).
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
	Name        string `json:"name"`
	Description string `json:"description"`
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

func (mi MangaInfo) GetAuthors() string {
	authors := []string{}
	for _, relation := range mi.Relationships {
		if relation.Type == "author" {
			authors = append(authors, relation.Attributes.Name)
		}
	}
	return strings.Join(authors, ", ")
}

func (mi MangaInfo) GetArtists() string {
	artists := []string{}
	for _, relation := range mi.Relationships {
		if relation.Type == "artist" {
			artists = append(artists, relation.Attributes.Name)
		}
	}
	return strings.Join(artists, ", ")
}

func (mi MangaInfo) GetTags() string {
	tags := []string{}
	for _, tagEntity := range mi.Attributes.Tags {
		if tagEntity.Type == "tag" {
			tags = append(tags, tagEntity.Attributes.Name["en"])
		}
	}
	return strings.Join(tags, ", ")
}

func (mi MangaInfo) GetAltTitles() string {
	altTitles := []string{}
	for _, m := range mi.Attributes.AltTitles {
		for language, title := range m {
			altTitles = append(altTitles, fmt.Sprintf("%s (%s)", title, language))
		}
	}
	return strings.Join(altTitles, " | ")
}

func (mi MangaInfo) GetLinks() []string {
	links := []string{}
	for _, link := range mi.Attributes.Links {
		u, err := url.Parse(link)
		if err == nil && u.Scheme == "https" {
			links = append(links, link)
		}
	}
	return links
}

type ResponseMangaList struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []MangaInfo `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

func (a clientapi) Find(title, limit, offset string) (ResponseMangaList, error) {
	if title == "" || limit == "" || offset == "" {
		return ResponseMangaList{}, ErrBadInput
	}

	mangaList := ResponseMangaList{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&mangaList).
		SetQueryParams(map[string]string{
			"title":            title,
			"limit":            limit,
			"offset":           offset,
			"order[relevance]": "asc",
		}).
		Get(manga_path)
	if err != nil {
		return ResponseMangaList{}, ErrConnection
	}

	if resp.IsError() {
		return ResponseMangaList{}, &respErr
	}

	return mangaList, nil
}

type MangaInfoResponse struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     MangaInfo `json:"data"`
}

func (a clientapi) GetMangaInfo(mangaId string) (MangaInfo, error) {
	if mangaId == "" {
		return MangaInfo{}, ErrBadInput
	}

	info := MangaInfoResponse{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&info).
		SetPathParam("id", mangaId).
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

type Chapter struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Attributes    ChapterAttr    `json:"attributes"`
	Relationships []Relationship `json:"relationships"`
}

type ChapterAttr struct {
	Volume             string    `json:"volume"`
	Chapter            string    `json:"chapter"`
	Title              string    `json:"title"`
	TranslatedLanguage string    `json:"translatedLanguage"`
	ExternalUrl        string    `json:"externalUrl"`
	PublishAt          time.Time `json:"publishAt"`
	ReadableAt         time.Time `json:"readableAt"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	Pages              int       `json:"pages"`
	Version            int       `json:"version"`
}

type ResponseChapterList struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     []Chapter `json:"data"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
	Total    int       `json:"total"`
}

func (l ResponseChapterList) FirstID() string {
	if len(l.Data) == 0 {
		return ""
	}
	return l.Data[0].ID
}

func (l ResponseChapterList) FirstChapter() string {
	if len(l.Data) == 0 {
		return ""
	}
	return l.Data[0].Attributes.Chapter
}

func (l ResponseChapterList) FirstChapterTitle() string {
	if len(l.Data) == 0 {
		return ""
	}
	return l.Data[0].Attributes.Title
}

func (l ResponseChapterList) FirstVolume() string {
	if len(l.Data) == 0 {
		return ""
	}
	return l.Data[0].Attributes.Volume
}

func (l ResponseChapterList) IsTranslateGroup(translateGroup string) bool {
	if len(l.Data) == 0 {
		return false
	}

	for _, rel := range l.Data[0].Relationships {
		if rel.Type == "scanlation_group" &&
			strings.Contains(rel.Attributes.Name, translateGroup) {
			return true
		}
	}
	return false
}

func (l ResponseChapterList) FirstTranslateGroup() string {
	if len(l.Data) == 0 {
		return ""
	}

	for _, rel := range l.Data[0].Relationships {
		if rel.Type == "scanlation_group" {
			return rel.Attributes.Name
		}
	}
	return ""
}

func (l ResponseChapterList) FirstTranslateGroupDescription() string {
	if len(l.Data) == 0 {
		return ""
	}

	for _, rel := range l.Data[0].Relationships {
		if rel.Type == "scanlation_group" {
			return rel.Attributes.Description
		}
	}
	return ""
}

func (l ResponseChapterList) FirstTranslationLanguage() string {
	if len(l.Data) == 0 {
		return ""
	}
	return l.Data[0].Attributes.TranslatedLanguage
}

func (l ResponseChapterList) FirstChapterPages() int {
	if len(l.Data) == 0 {
		return 0
	}
	return l.Data[0].Attributes.Pages
}

func (a clientapi) GetChaptersList(limit, offset, mangaId,
	language string) (ResponseChapterList, error) {

	if mangaId == "" {
		return ResponseChapterList{}, ErrBadInput
	}

	list := ResponseChapterList{}
	respErr := ErrorResponse{}

	query := fmt.Sprintf(
		"limit=%s&offset=%s&translatedLanguage[]=%s&includes[]=scanlation_group&order[volume]=asc&order[chapter]=asc",
		limit, offset, language)

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&list).
		SetPathParam("id", mangaId).
		SetQueryString(query).
		Get(manga_feed_path)
	if err != nil {
		return ResponseChapterList{}, ErrConnection
	}

	if resp.IsError() {
		return ResponseChapterList{}, &respErr
	}

	return list, nil
}

type ChapterData struct {
	Hash      string   `json:"hash"`
	Data      []string `json:"data"`
	DataSaver []string `json:"dataSaver"`
}

type ResponseChapterImages struct {
	Result  string      `json:"result"`
	BaseURL string      `json:"baseUrl"`
	Chapter ChapterData `json:"chapter"`
}

func (a clientapi) GetChapterImageList(chapterId string) (ResponseChapterImages, error) {
	if chapterId == "" {
		return ResponseChapterImages{}, ErrBadInput
	}

	list := ResponseChapterImages{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&list).
		SetPathParam("id", chapterId).
		Get(chapter_images_path)
	if err != nil {
		return ResponseChapterImages{}, ErrConnection
	}

	if resp.IsError() {
		return ResponseChapterImages{}, &respErr
	}

	return list, nil
}

func (a clientapi) DownloadImage(baseUrl, chapterHash, imageFilename string,
	isJpg bool) (io.Reader, error) {
	if baseUrl == "" || chapterHash == "" || imageFilename == "" {
		return nil, ErrBadInput
	}

	path := download_high_quility_path
	if isJpg {
		path = download_low_quility_path
	}

	respErr := ErrorResponse{}

	resp, err := a.c.SetBaseURL(baseUrl).
		R().
		SetError(respErr).
		SetPathParams(map[string]string{
			"chapterHash":   chapterHash,
			"imageFilename": imageFilename,
		}).
		Get(path)
	if err != nil {
		return nil, ErrConnection
	}

	if resp.IsError() {
		return nil, &respErr
	}

	return bytes.NewBuffer(resp.Body()), nil
}
