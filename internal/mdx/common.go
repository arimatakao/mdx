package mdx

import (
	"github.com/arimatakao/mdx/app"
	"github.com/arimatakao/mdx/mangadexapi"
)

var client = mangadexapi.NewClient(app.USER_AGENT)
