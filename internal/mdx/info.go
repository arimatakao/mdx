package mdx

import (
	"os"

	"github.com/pterm/pterm"
)

type infoParams struct {
	mangaId string
}

func NewInfoParams(mangaId string) infoParams {
	return infoParams{
		mangaId: mangaId,
	}
}

func (p infoParams) GetInfo() {
	spinner, _ := pterm.DefaultSpinner.Start("Fetching info...")
	resp, err := client.GetMangaInfo(p.mangaId)
	if err != nil {
		spinner.Fail("Failed to fetch manga info")
		e.Printfln("While getting manga information: %v\n", err)
		os.Exit(1)
	}
	spinner.Success("Fetched info")
	printMangaInfo(resp.MangaInfo())
}
