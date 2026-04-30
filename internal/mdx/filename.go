package mdx

import (
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
)

func (p dlParam) chapterFileName(chapter mangadexapi.ChapterFullInfo) string {
	if p.fileNameTemplate == "" {
		return pterm.Sprintf("[%s %s] %s vol. %s ch. %s",
			p.language, chapter.Translator(), p.mangaInfo.Title("en"), chapter.Volume(),
			chapter.Number())
	}

	return formatFileNameTemplate(p.fileNameTemplate, []string{
		p.language,
		chapter.Translator(),
		p.mangaInfo.Title("en"),
		chapter.Volume(),
		chapter.Number(),
		chapter.Title(),
	})
}

func (p dlParam) mergeChaptersFileName(chaptersRange string) string {
	if p.fileNameTemplate == "" {
		return pterm.Sprintf("[%s %s] %s ch. %s",
			p.language, p.chapters[0].Translator(), p.mangaInfo.Title("en"), chaptersRange)
	}

	return formatFileNameTemplate(p.fileNameTemplate, []string{
		p.language,
		p.chapters[0].Translator(),
		p.mangaInfo.Title("en"),
		"",
		chaptersRange,
		p.chapters[0].Title(),
	})
}

func (p dlParam) volumeFileName(volume, chaptersRange string) string {
	if p.fileNameTemplate == "" {
		return pterm.Sprintf("[%s] %s | vol. %s | ch. %s",
			p.language, p.mangaInfo.Title("en"), volume, chaptersRange)
	}

	chapter := selectedVolumeChapterMap[volume][0]
	return formatFileNameTemplate(p.fileNameTemplate, []string{
		p.language,
		chapter.GetTranslator(),
		p.mangaInfo.Title("en"),
		volume,
		chaptersRange,
		chapter.Title(),
	})
}

func formatFileNameTemplate(template string, fields []string) string {
	fileName := template
	for i := len(fields); i > 0; i-- {
		fileName = strings.ReplaceAll(fileName, "%"+strconv.Itoa(i), fields[i-1])
	}
	return strings.TrimSpace(fileName)
}
