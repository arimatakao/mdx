package mangadexapi

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

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
	Username    string `json:"username"`
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

func (mi MangaInfo) Title(language string) string {
	return mi.Attributes.Title[language]
}

func (mi MangaInfo) AltTitles() string {
	altTitles := []string{}
	for _, m := range mi.Attributes.AltTitles {
		for language, title := range m {
			altTitles = append(altTitles, fmt.Sprintf("%s (%s)", title, language))
		}
	}
	return strings.Join(altTitles, " | ")
}

func (mi MangaInfo) Authors() string {
	authors := []string{}
	for _, relation := range mi.Relationships {
		if relation.Type == "author" {
			authors = append(authors, relation.Attributes.Name)
		}
	}
	return strings.Join(authors, ", ")
}

func (mi MangaInfo) AuthorsArr() []string {
	authors := []string{}
	for _, relation := range mi.Relationships {
		if relation.Type == "author" {
			authors = append(authors, relation.Attributes.Name)
		}
	}
	return authors
}

func (mi MangaInfo) Artists() string {
	artists := []string{}
	for _, relation := range mi.Relationships {
		if relation.Type == "artist" {
			artists = append(artists, relation.Attributes.Name)
		}
	}
	return strings.Join(artists, ", ")
}

func (mi MangaInfo) ArtistsArr() []string {
	artists := []string{}
	for _, relation := range mi.Relationships {
		if relation.Type == "artist" {
			artists = append(artists, relation.Attributes.Name)
		}
	}
	return artists
}

func (mi MangaInfo) Publisher() string {
	return mi.Attributes.PublicationDemographic
}

func (mi MangaInfo) Year() int {
	return mi.Attributes.Year
}

func (mi MangaInfo) Status() string {
	return mi.Attributes.Status
}

func (mi MangaInfo) OriginalLanguage() string {
	return mi.Attributes.OriginalLanguage
}

func (mi MangaInfo) TranslatedLanguages() []string {
	return mi.Attributes.AvailableTranslatedLanguages
}

func (mi MangaInfo) Description(language string) string {
	return mi.Attributes.Description[language]
}

func (mi MangaInfo) Tags() string {
	tags := []string{}
	for _, tagEntity := range mi.Attributes.Tags {
		if tagEntity.Type == "tag" {
			tags = append(tags, tagEntity.Attributes.Name["en"])
		}
	}
	return strings.Join(tags, ", ")
}

func (mi MangaInfo) TagsArr() []string {
	tags := []string{}
	for _, tagEntity := range mi.Attributes.Tags {
		if tagEntity.Type == "tag" {
			tags = append(tags, tagEntity.Attributes.Name["en"])
		}
	}
	return tags
}

func (mi MangaInfo) Links() string {
	joinedLinks := strings.Join(mi.LinksArr(), "\n")
	return joinedLinks
}

func (mi MangaInfo) LinksArr() []string {
	links := []string{}
	for val, link := range mi.Attributes.Links {
		switch val {
		case "al":
			links = append(links, "https://anilist.co/manga/"+link)
			continue
		case "ap":
			links = append(links, "https://www.anime-planet.com/manga/"+link)
			continue
		case "bw":
			links = append(links, "https://bookwalker.jp/"+link)
			continue
		case "nu":
			links = append(links, "https://www.novelupdates.com/series/"+link)
			continue
		case "kt":
			links = append(links, "https://kitsu.io/api/edge/manga/"+link)
			continue
		case "mal":
			links = append(links, "https://myanimelist.net/manga/"+link)
			continue
		}

		u, err := url.Parse(link)
		if err == nil && u.Scheme == "https" {
			links = append(links, link)
		}
	}
	return links
}

type MangaInfoResponse struct {
	Result   string    `json:"result"`
	Response string    `json:"response"`
	Data     MangaInfo `json:"data"`
}

func (m MangaInfoResponse) MangaInfo() MangaInfo {
	return m.Data
}

type ResponseMangaList struct {
	Result   string      `json:"result"`
	Response string      `json:"response"`
	Data     []MangaInfo `json:"data"`
	Limit    int         `json:"limit"`
	Offset   int         `json:"offset"`
	Total    int         `json:"total"`
}

func (ml ResponseMangaList) List() []MangaInfo {
	return ml.Data
}
