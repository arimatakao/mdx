package mangadexapi

import (
	"errors"
	"fmt"
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
	chapter_info_path          = "/chapter/{id}"
	chapter_images_path        = "/at-home/server/{id}"
	download_high_quility_path = "/data/{chapterHash}/{imageFilename}"
	download_low_quility_path  = "/data-saver/{chapterHash}/{imageFilename}"
)

var (
	ErrUnknown          = errors.New("unknown error")
	ErrBadInput         = errors.New("bad input")
	ErrConnection       = errors.New("request is failed")
	ErrUnexpectedHeader = errors.New("unexpected response header value")
)

func getMangaDexPaths(link string) []string {
	if !strings.HasPrefix(link, "https://") && !strings.HasPrefix(link, "http://") {
		link = "https://" + link
	}

	parsedUrl, err := url.Parse(link)
	if err != nil {
		return []string{}
	}

	if parsedUrl.Host != "mangadex.org" {
		return []string{}
	}

	return strings.Split(parsedUrl.Path, "/")
}

func GetMangaIdFromUrl(link string) string {
	paths := getMangaDexPaths(link)
	if len(paths) < 3 {
		return ""
	}
	if paths[1] != "title" {
		return ""
	}
	return paths[2]
}

func GetMangaIdFromArgs(args []string) string {
	for _, arg := range args {
		if u := GetMangaIdFromUrl(arg); u != "" {
			return u
		}
	}
	return ""
}

func GetChapterIdFromUrl(link string) string {
	paths := getMangaDexPaths(link)
	if len(paths) < 3 {
		return ""
	}
	if paths[1] != "chapter" {
		return ""
	}
	return paths[2]
}

func GetChapterIdFromArgs(args []string) string {
	for _, arg := range args {
		if u := GetChapterIdFromUrl(arg); u != "" {
			return u
		}
	}
	return ""
}

type Clientapi struct {
	c *resty.Client
}

type silentLogger struct{}

func (l silentLogger) Errorf(format string, v ...interface{}) {}
func (l silentLogger) Warnf(format string, v ...interface{})  {}
func (l silentLogger) Debugf(format string, v ...interface{}) {}

func NewClient(userAgent string) Clientapi {
	if userAgent == "" {
		userAgent = default_useragent
	}

	c := resty.New().
		SetLogger(silentLogger{}).
		SetRetryCount(5).
		SetRetryWaitTime(time.Millisecond*200).
		SetBaseURL(base_url).
		SetHeader("User-Agent", userAgent)

	return Clientapi{
		c: c,
	}
}

func (a Clientapi) Ping() bool {
	resp, err := a.c.R().Get(health_path)
	if err != nil {
		return false
	}

	if resp.StatusCode() != http.StatusOK {
		return false
	}

	return true
}

func (a Clientapi) Find(title string, limit, offset int, isDoujinshiAllow bool) (ResponseMangaList, error) {
	if title == "" || limit == 0 || offset < 0 {
		return ResponseMangaList{}, ErrBadInput
	}

	mangaList := ResponseMangaList{}
	respErr := ErrorResponse{}

	query := fmt.Sprintf("title=%s&limit=%d&offset=%d&order[relevance]=asc"+
		"&includes[]=author&includes[]=artist",
		title, limit, offset)

	if !isDoujinshiAllow {
		query += "&excludedTags[]=b13b2a48-c720-44a9-9c77-39c9979373fb&excludedTagsMode=OR"
	}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&mangaList).
		SetQueryString(query).
		Get(manga_path)
	if err != nil {
		return ResponseMangaList{}, ErrConnection
	}

	if resp.IsError() {
		return ResponseMangaList{}, &respErr
	}

	return mangaList, nil
}

func (a Clientapi) GetMangaInfo(mangaId string) (MangaInfoResponse, error) {
	if mangaId == "" {
		return MangaInfoResponse{}, ErrBadInput
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
		return MangaInfoResponse{}, ErrConnection
	}

	if resp.IsError() {
		return MangaInfoResponse{}, &respErr
	}

	return info, nil
}

func (a Clientapi) GetChapterInfo(chapterId string) (ResponseChapter, error) {
	if chapterId == "" {
		return ResponseChapter{}, ErrBadInput
	}

	chapterInfo := ResponseChapter{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&chapterInfo).
		SetPathParam("id", chapterId).
		SetQueryString("includes[]=scanlation_group&includes[]=user").
		Get(chapter_info_path)
	if err != nil {
		return ResponseChapter{}, ErrConnection
	}

	if resp.IsError() {
		return ResponseChapter{}, &respErr
	}

	return chapterInfo, nil
}

func (a Clientapi) GetChaptersList(limit, offset int, mangaId, language string) (ResponseChapterList, error) {

	if mangaId == "" {
		return ResponseChapterList{}, ErrBadInput
	}

	list := ResponseChapterList{}
	respErr := ErrorResponse{}

	query := fmt.Sprintf(
		"limit=%d&offset=%d&translatedLanguage[]=%s"+
			"&includes[]=scanlation_group&order[volume]=asc&order[chapter]=asc",
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

func (a Clientapi) GetChapterImageList(chapterId string) (ResponseChapterImages, error) {
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

func (a Clientapi) DownloadImage(baseUrl, chapterHash, imageFilename string,
	isJpg bool) ([]byte, error) {
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

	h := resp.Header().Get("Content-Type")
	if h != "image/jpeg" && h != "image/png" {
		return nil, ErrUnexpectedHeader
	}

	return resp.Body(), nil
}

func (a Clientapi) GetChapterImagesInFullInfo(chap Chapter) (ChapterFullInfo, error) {
	chapImages := ResponseChapterImages{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&chapImages).
		SetPathParam("id", chap.ID).
		Get(chapter_images_path)
	if err != nil {
		return ChapterFullInfo{}, ErrConnection
	}

	if resp.IsError() {
		return ChapterFullInfo{}, &respErr
	}

	fullInfo := ChapterFullInfo{
		info:            chap,
		DownloadBaseURL: chapImages.BaseURL,
		HashId:          chapImages.ChapterMetaInfo.Hash,
		PngFiles:        chapImages.ChapterMetaInfo.Data,
		JpgFiles:        chapImages.ChapterMetaInfo.DataSaver,
	}

	return fullInfo, nil
}

func (a Clientapi) GetFullChaptersInfo(mangaId, language, translationGroup string,
	lowestChapter, highestChapter int) ([]ChapterFullInfo, error) {

	if mangaId == "" ||
		language == "" ||
		lowestChapter > highestChapter ||
		lowestChapter < 0 ||
		highestChapter < 0 ||
		highestChapter < lowestChapter {
		return []ChapterFullInfo{}, ErrBadInput
	}

	chapters := []Chapter{}
	chaptersInfo := []ChapterFullInfo{}

	lowBound := (lowestChapter / 10) * 10
	highBound := ((highestChapter + 11) / 10) * 10

	for lowBound <= highBound {
		query := fmt.Sprintf(
			"limit=%d&offset=%d&translatedLanguage[]=%s"+
				"&includes[]=scanlation_group&includes[]=user"+
				"&order[volume]=asc&order[chapter]=asc",
			10, lowBound, language)

		list := ResponseChapterList{}
		respErr := ErrorResponse{}

		resp, err := a.c.R().
			SetError(&respErr).
			SetResult(&list).
			SetPathParam("id", mangaId).
			SetQueryString(query).
			Get(manga_feed_path)
		if err != nil {
			return []ChapterFullInfo{}, ErrConnection
		}
		if resp.IsError() {
			return []ChapterFullInfo{}, &respErr
		}

		found, extra := list.GetChapters(lowestChapter, highestChapter, translationGroup)

		chapters = append(chapters, found...)
		lowBound += 10
		if extra != 0 {
			highBound += ((extra + 11) / 10) * 10
		}
	}

	for _, chapter := range chapters {
		chapImages := ResponseChapterImages{}
		respErr := ErrorResponse{}

		resp, err := a.c.R().
			SetError(&respErr).
			SetResult(&chapImages).
			SetPathParam("id", chapter.ID).
			Get(chapter_images_path)
		if err != nil {
			return []ChapterFullInfo{}, ErrConnection
		}

		if resp.IsError() {
			return []ChapterFullInfo{}, &respErr
		}

		fullInfo := ChapterFullInfo{
			info:            chapter,
			DownloadBaseURL: chapImages.BaseURL,
			HashId:          chapImages.ChapterMetaInfo.Hash,
			PngFiles:        chapImages.ChapterMetaInfo.Data,
			JpgFiles:        chapImages.ChapterMetaInfo.DataSaver,
		}

		chaptersInfo = append(chaptersInfo, fullInfo)
	}

	return chaptersInfo, nil
}

func (a Clientapi) GetLastChapterFullInfo(mangaId, language,
	translationGroup string) (ChapterFullInfo, error) {
	if mangaId == "" || language == "" {
		return ChapterFullInfo{}, ErrBadInput
	}

	query := fmt.Sprintf(
		"limit=%d&&translatedLanguage[]=%s"+
			"&includes[]=scanlation_group&includes[]=user"+
			"&order[volume]=desc&order[chapter]=desc",
		1, language)

	list := ResponseChapterList{}
	respErr := ErrorResponse{}

	resp, err := a.c.R().
		SetError(&respErr).
		SetResult(&list).
		SetPathParam("id", mangaId).
		SetQueryString(query).
		Get(manga_feed_path)
	if err != nil {
		return ChapterFullInfo{}, ErrConnection
	}
	if resp.IsError() {
		return ChapterFullInfo{}, &respErr
	}

	if len(list.Data) == 0 {
		return ChapterFullInfo{}, nil
	}

	chapImages := ResponseChapterImages{}
	respErr = ErrorResponse{}

	respChap, err := a.c.R().
		SetError(&respErr).
		SetResult(&chapImages).
		SetPathParam("id", list.Data[0].ID).
		Get(chapter_images_path)
	if err != nil {
		return ChapterFullInfo{}, ErrConnection
	}

	if respChap.IsError() {
		return ChapterFullInfo{}, &respErr
	}

	fullInfo := ChapterFullInfo{
		info:            list.Data[0],
		DownloadBaseURL: chapImages.BaseURL,
		HashId:          chapImages.ChapterMetaInfo.Hash,
		PngFiles:        chapImages.ChapterMetaInfo.Data,
		JpgFiles:        chapImages.ChapterMetaInfo.DataSaver,
	}

	return fullInfo, nil
}
