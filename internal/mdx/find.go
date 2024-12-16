package mdx

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pterm/pterm"
)

type findParams struct {
	title            string
	isDoujinshiAllow bool
	printedCount     int
	offset           int
	outputFile       bool
}

func NewFindParams(title string, isDoujinshiAllow bool, outputFile bool) findParams {
	return findParams{
		title:            title,
		isDoujinshiAllow: isDoujinshiAllow,
		printedCount:     25,
		offset:           0,
		outputFile:       outputFile,
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

	// If output file is specified, fetch all results and save to JSON
	if p.outputFile {
		// If there are more results, fetch them all
		allResults := response
		currentOffset := p.printedCount

		for currentOffset < response.Total {
			spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("Fetching more results (%d/%d)...", currentOffset, response.Total))
			moreResults, err := client.Find(p.title, p.printedCount, currentOffset, p.isDoujinshiAllow)
			if err != nil {
				spinner.Fail("Failed to fetch additional results")
				e.Printf("error while fetching additional results: %v\n", err)
				os.Exit(1)
			}
			allResults.Data = append(allResults.Data, moreResults.Data...)
			currentOffset += p.printedCount
		}

		jsonData, err := json.MarshalIndent(allResults.Data, "", "    ")
		if err != nil {
			e.Printf("error while marshaling JSON: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile("output.json", jsonData, 0644)
		if err != nil {
			e.Printf("error while writing JSON file: %v\n", err)
			os.Exit(1)
		}

		spinner.Success(fmt.Sprintf("All %d results saved to output.json", response.Total))
		// Don't output to console since it was saved to "output.json"
		return
	}

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
