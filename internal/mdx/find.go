package mdx

import (
	"os"

	"github.com/pterm/pterm"
)

type findParams struct {
	title            string
	isDoujinshiAllow bool
	printedCount     int
	offset           int
}

func NewFindParams(title string, isDoujinshiAllow bool) findParams {
	return findParams{
		title:            title,
		isDoujinshiAllow: isDoujinshiAllow,
		printedCount:     25,
		offset:           0,
	}
}

func (p findParams) Find() {
	spinner, _ := pterm.DefaultSpinner.Start("Searching manga...")
	response, err := client.Find(p.title, p.printedCount, p.offset, p.isDoujinshiAllow)
	if err != nil {
		spinner.Fail("Failed to search manga")
		e.Printf("error while search manga: %v\n", err)
		os.Exit(1)
	}

	if response.Total == 0 {
		spinner.Warning("Nothing found...")
		os.Exit(0)
	}
	spinner.Success("Manga found!")

	for _, m := range response.List() {
		dp.Println("------------------------------")
		printMangaInfo(m)
	}

	if response.Total > p.printedCount {
		dp.Println("==============================")
		field.Printf("Full results: ")
		dp.Printfln(" https://mangadex.org/search?q=%s", p.title)
		field.Print("Total found: ")
		dp.Println(response.Total)
	}
}
