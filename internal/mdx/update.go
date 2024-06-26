package mdx

import (
	"os"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/app"
	"github.com/go-resty/resty/v2"
	"github.com/pterm/pterm"
)

func CheckUpdate() {
	result := struct {
		TagName     string `json:"tag_name"`
		Description string `json:"body"`
	}{}

	resp, err := resty.New().R().
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("X-GitHub-Api-Version", "2022-11-28").
		SetHeader("User-Agent", app.USER_AGENT).
		SetResult(&result).
		Get("https://api.github.com/repos/arimatakao/mdx/releases/latest")
	if err != nil {
		e.Printf("While connecting to github api: %v", err)
		os.Exit(1)
	}

	if resp.IsError() {
		e.Println("Wrong response body from github api")
		os.Exit(1)
	}

	isShouldUpdate := false

	parsedLatest := strings.Split(result.TagName, ".")
	if len(parsedLatest) == 3 {
		parsedCurrent := strings.Split(app.VERSION, ".")
		mainCurrent := parsedCurrent[0]
		secondCurrent, _ := strconv.Atoi(parsedCurrent[1])
		thirdCurrent, _ := strconv.Atoi(parsedCurrent[2])

		isNotMainVersion := !strings.Contains(parsedLatest[0], mainCurrent)
		if isNotMainVersion {
			isShouldUpdate = true
		}
		if secondLast, err := strconv.Atoi(parsedLatest[1]); err != nil {
			isShouldUpdate = true
		} else if secondLast > secondCurrent {
			isShouldUpdate = true
		}
		if thirdLast, err := strconv.Atoi(parsedLatest[2]); err != nil {
			isShouldUpdate = true
		} else if thirdLast > thirdCurrent {
			isShouldUpdate = true
		}
	} else if result.TagName != app.VERSION {
		isShouldUpdate = true
	}

	tableData := pterm.TableData{
		{field.Sprint("Your version"), dp.Sprint(app.VERSION)},
		{field.Sprint("Latest version"), dp.Sprint(result.TagName)},
	}
	pterm.DefaultTable.WithData(tableData).Render()
	if isShouldUpdate {
		field.Print("Download new version here: ")
		dp.Println("https://github.com/arimatakao/mdx/releases")
		field.Println("Release description:")
		dp.Println(result.Description)
	} else {
		field.Print("You have latest version.\n")
	}
}
