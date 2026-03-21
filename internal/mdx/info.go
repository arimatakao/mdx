package mdx

import (
	"os"

	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
)

type infoParams struct {
	mangaId  string
	isRandom bool
}

func NewInfoParams(mangaId string, isRandom bool) infoParams {
	return infoParams{
		mangaId:  mangaId,
		isRandom: isRandom,
	}
}

func (p infoParams) GetInfo() {
	spinner, _ := pterm.DefaultSpinner.Start("Fetching info...")

	var (
		resp mangadexapi.MangaInfoResponse
		err  error
	)

	if p.isRandom {
		resp, err = client.GetRandomMangaInfo()
	} else {
		resp, err = client.GetMangaInfo(p.mangaId)
	}

	if err != nil {
		spinner.Fail("Failed to fetch manga info")
		e.Printfln("While getting manga information: %v\n", err)
		os.Exit(1)
	}
	spinner.Success("Fetched info")
	printMangaInfo(resp.MangaInfo())
}
