package mdx

import (
	"os"

	"github.com/arimatakao/mdx/app"
)

func PrintVersion() {
	dp.Printfln(app.VERSION)
	os.Exit(0)
}

func PrintMangaDexAPIVersion() {
	dp.Println(app.API_VERSION)
	os.Exit(0)
}
