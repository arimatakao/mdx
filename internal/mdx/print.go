package mdx

import (
	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
)

var (
	// for error print
	e = pterm.Error
	// default print
	dp = pterm.NewStyle(pterm.FgDefault, pterm.BgDefault)
	// for field print
	field = pterm.NewStyle(pterm.FgGreen, pterm.BgDefault, pterm.Bold)
)

func printMangaInfo(i mangadexapi.MangaInfo) {
	dp.Println(field.Sprint("Link: "), dp.Sprintf("https://mangadex.org/title/%s", i.ID))
	dp.Println(field.Sprint("Title: "), i.Title("en"))
	dp.Println(field.Sprint("Alternative titles: "), i.Title("en"))
	dp.Println(field.Sprint("Type: "), i.Type)
	dp.Println(field.Sprint("Authors: "), i.Authors())
	dp.Println(field.Sprint("Artists: "), i.Artists())
	dp.Println(field.Sprint("Year: "), i.Year())
	dp.Println(field.Sprint("Status: "), i.Status())
	dp.Println(field.Sprint("Original language: "), i.OriginalLanguage())
	dp.Println(field.Sprint("Translated: "), i.TranslatedLanguages())
	dp.Println(field.Sprint("Tags: "), i.Tags())
	dp.Println(field.Sprint("Description:\n"), i.Description("en"))
	dp.Println(field.Sprint("Read or Buy here:\n"), i.Links())
}

func printShortMangaInfo(i mangadexapi.MangaInfo) {
	dp.Println(field.Sprint("Manga title: "), i.Title("en"))
	dp.Println(field.Sprint("Alt titles: "), i.AltTitles())
	field.Println("Read or Buy here:")
	dp.Println(i.Links())
	dp.Printf("==============\n\n")
}

func printChapterInfo(c mangadexapi.ChapterFullInfo) {
	tableData := pterm.TableData{
		{field.Sprint("Chapter"), dp.Sprint(c.Number())},
		{field.Sprint("Chapter title"), dp.Sprint(c.Title())},
		{field.Sprint("Volume"), dp.Sprint(c.Volume())},
		{field.Sprint("Language"), dp.Sprint(c.Language())},
		{field.Sprint("Translated by"), dp.Sprint(c.Translator())},
		{field.Sprint("Uploaded by"), dp.Sprint(c.UploadedBy())},
	}
	pterm.DefaultTable.WithData(tableData).Render()
}

func printUaNotification() {
	y := pterm.NewStyle(pterm.FgYellow)
	b := pterm.NewStyle(pterm.FgBlue)

	b.Println("ПОМОГИ УКРАИНЕ В БОРЬБЕ")
	y.Println("ПРОТИВ РОССИЙСКОЙ АГРЕССИИ")

	field.Println("\n===ПОЛЕЗНЫЕ ССЫЛКИ===")
	field.Println("Как война касается тебя лично?:")
	dp.Println("https://war.ukraine.ua/ru/kak-vojna-kasaetsya-tebya-lychno")
	field.Println("(СМИ) BBC Русская служба:")
	dp.Println("https://www.bbc.com/russian\n" +
		"https://t.me/bbcrussian")
	field.Println("(СМИ) Радио Свобода:")
	dp.Println("https://www.svoboda.org\n" +
		"https://www.svoboda.org/block\n" +
		"https://t.me/radiosvoboda")
	field.Println("(СМИ) Голос Америки:")
	dp.Println("https://www.golosameriki.com\n" +
		"https://t.me/GolosAmeriki")
	field.Print("Используй VPN для своей безопасности!\n\n")

	b.Println("ПОМОГИ УКРАИНЕ В БОРЬБЕ")
	y.Println("ПРОТИВ РОССИЙСКОЙ АГРЕССИИ")
}
