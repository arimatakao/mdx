package metadata

import (
	"encoding/xml"
	"time"

	"github.com/pterm/pterm"
)

type Metadata struct {
	// ComicBookInfo format
	CBI ComicBookMetadata
	// ComicRack old metadata
	CI ComicInfoMetadata
	// Plain metadata
	P PlainMetadata
}

type PlainMetadata struct {
	Authors string
	Artists string
	Tags    string
}

type ComicInfoMetadata struct {
	XMLName     xml.Name `xml:"ComicInfo"`
	Title       string   `xml:"Title"`
	Number      string   `xml:"Number,omitempty"`
	Volume      string   `xml:"Volume,omitempty"`
	Year        int      `xml:"Year"`
	Writer      string   `xml:"Writer"`
	Penciller   string   `xml:"Penciller"`
	Inker       string   `xml:"Inker"`
	Publisher   string   `xml:"Publisher"`
	PageCount   int      `xml:"PageCount"`
	LanguageISO string   `xml:"LanguageISO"`
	Format      string   `xml:"Format"`
	Manga       string   `xml:"Manga"`
	Summary     string   `xml:"Summary"`
}

type ComicBookMetadata struct {
	AppID             string        `json:"appID"`
	LastModified      string        `json:"lastModified"`
	ComicBookInfoData ComicBookInfo `json:"ComicBookInfo/1.0"`
}

type ComicBookInfo struct {
	Series    string   `json:"series"`
	Title     string   `json:"title"`
	Publisher string   `json:"publisher"`
	Issue     string   `json:"issue"`
	Volume    string   `json:"volume"`
	Language  string   `json:"language"`
	Credits   []Credit `json:"credits"`
	Tags      []string `json:"tags"`
}

type Credit struct {
	Person string `json:"person"`
	Role   string `json:"role"`
}

type MangaProvider interface {
	Title(language string) string
	Description(language string) string
	Publisher() string
	Year() int
	AuthorsArr() []string
	Authors() string
	ArtistsArr() []string
	Artists() string
	TagsArr() []string
	Tags() string
	LinksArr() []string
	Links() string
}

type ChapterProvider interface {
	Title() string
	Number() string
	Volume() string
	Language() string
	PagesCount() int
}

func NewMetadata(appId string, m MangaProvider, c ChapterProvider) Metadata {

	credits := []Credit{}
	for _, au := range m.AuthorsArr() {
		credit := Credit{
			Person: au,
			Role:   "Writer",
		}
		credits = append(credits, credit)
	}
	for _, ar := range m.ArtistsArr() {
		credit := Credit{
			Person: ar,
			Role:   "Artist",
		}
		credits = append(credits, credit)
	}

	mangaTitle := pterm.Sprintf("%s | %s vol%s ch%s",
		c.Language(), m.Title("en"), c.Volume(), c.Number())

	mangaDescription := m.Description("en") + "<br>Read or Buy here:<br>"
	for _, l := range m.LinksArr() {
		mangaDescription += l + "<br>"
	}

	metadata := Metadata{
		CBI: ComicBookMetadata{
			AppID:        appId,
			LastModified: time.Now().UTC().String(),
			ComicBookInfoData: ComicBookInfo{
				Series:    mangaTitle,
				Title:     c.Title(),
				Publisher: m.Publisher(),
				Issue:     c.Number(),
				Volume:    c.Volume(),
				Language:  c.Language(),
				Credits:   credits,
				Tags:      m.TagsArr(),
			},
		},
		CI: ComicInfoMetadata{
			XMLName:     xml.Name{Local: "ComicInfo"},
			Title:       mangaTitle,
			Number:      c.Number(),
			Volume:      c.Volume(),
			Year:        m.Year(),
			Writer:      m.Authors(),
			Penciller:   m.Artists(),
			Inker:       m.Artists(),
			Publisher:   m.Publisher(),
			PageCount:   c.PagesCount(),
			LanguageISO: c.Language(),
			Format:      "Comic Book",
			Manga:       "No",
			Summary:     mangaDescription,
		},
		P: PlainMetadata{
			Authors: m.Authors(),
			Artists: m.Artists(),
			Tags:    m.Tags(),
		},
	}

	return metadata
}
