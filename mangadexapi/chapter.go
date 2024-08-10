package mangadexapi

import (
	"strconv"
	"strings"
	"time"
)

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

func (c Chapter) Volume() string {
	return c.Attributes.Volume
}

func (c Chapter) Number() string {
	return c.Attributes.Chapter
}

func (c Chapter) Title() string {
	return c.Attributes.Title
}

func (c Chapter) PagesCount() int {
	return c.Attributes.Pages
}

func (c Chapter) UploadedBy() string {
	for _, rel := range c.Relationships {
		if rel.Type == "user" {
			return rel.Attributes.Username
		}
	}
	return ""
}

func (c Chapter) Language() string {
	return c.Attributes.TranslatedLanguage
}

func (c Chapter) isTranslatedByGroup(translateGroup string) bool {
	return strings.Contains(c.getTranslator(), translateGroup)
}

func (c Chapter) getTranslator() string {
	for _, rel := range c.Relationships {
		if rel.Type == "scanlation_group" {
			return rel.Attributes.Name
		}
	}
	return ""
}

func (c Chapter) GetMangaId() string {
	for _, rel := range c.Relationships {
		if rel.Type == "manga" {
			return rel.ID
		}
	}
	return ""
}

type ResponseChapterList struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     []Chapter `json:"data"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
	Total    int       `json:"total"`
}

type ResponseChapter struct {
	Result   string  `json:"result"`
	Response string  `json:"response"`
	Data     Chapter `json:"data"`
}

func (r ResponseChapter) GetChapterInfo() Chapter {
	return r.Data
}

func (l ResponseChapterList) GetAllChapters(transgp string) []Chapter {
	found := []Chapter{}
	for _, c := range l.Data {
		if len(found) != 0 {
			if found[len(found)-1].Number() == c.Number() {
				continue
			}
		}

		if c.isTranslatedByGroup(transgp) {
			found = append(found, c)
		}
	}
	return found
}

func (l ResponseChapterList) GetChapters(lowest, highest int, transgp string) ([]Chapter, int) {
	if len(l.Data) == 0 {
		return []Chapter{}, 0
	}

	found := []Chapter{}
	countExtraChapters := 0

	for _, chapter := range l.Data {
		if chapter.Number() == "" {
			continue
		}

		num, err := strconv.Atoi(chapter.Number())
		if err == nil {
			if num >= lowest &&
				num <= highest &&
				chapter.isTranslatedByGroup(transgp) {
				if len(found) != 0 {
					if found[len(found)-1].Number() == chapter.Number() {
						continue
					}
				}

				found = append(found, chapter)
			}
			continue
		}

		countExtraChapters += 1

		nums := strings.Split(chapter.Attributes.Chapter, ".")
		if len(nums) != 2 {
			continue
		}

		num, err = strconv.Atoi(nums[0])
		if err != nil {
			continue
		}

		if num >= lowest && num <= highest && chapter.isTranslatedByGroup(transgp) {
			if len(found) != 0 {
				if found[len(found)-1].Number() == chapter.Number() {
					continue
				}
			}
			found = append(found, chapter)
		}
	}

	return found, countExtraChapters
}

type ChapterMetaInfo struct {
	Hash      string   `json:"hash"`
	Data      []string `json:"data"`
	DataSaver []string `json:"dataSaver"`
}

type ResponseChapterImages struct {
	Result          string          `json:"result"`
	BaseURL         string          `json:"baseUrl"`
	ChapterMetaInfo ChapterMetaInfo `json:"chapter"`
}

type ChapterFullInfo struct {
	Info            Chapter
	DownloadBaseURL string
	HashId          string
	PngFiles        []string
	JpgFiles        []string
}

func (c ChapterFullInfo) Title() string {
	return c.Info.Title()
}

func (c ChapterFullInfo) Number() string {
	return c.Info.Number()
}

func (c ChapterFullInfo) Volume() string {
	return c.Info.Volume()
}

func (c ChapterFullInfo) Language() string {
	return c.Info.Language()
}

func (c ChapterFullInfo) Translator() string {
	return c.Info.getTranslator()
}

func (c ChapterFullInfo) UploadedBy() string {
	return c.Info.UploadedBy()
}

func (c ChapterFullInfo) PagesCount() int {
	return c.Info.PagesCount()
}

func (c ChapterFullInfo) ImagesBaseUrl() string {
	return c.DownloadBaseURL
}

func (c ChapterFullInfo) ImagesFiles() []string {
	return c.PngFiles
}

func (c ChapterFullInfo) ImagesCompressedFiles() []string {
	return c.JpgFiles
}
