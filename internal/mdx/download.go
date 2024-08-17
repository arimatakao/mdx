package mdx

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/app"
	"github.com/arimatakao/mdx/filekit"
	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
)

var (
	ErrEmptyChapters = errors.New("empty chapters")
)

type dlParam struct {
	mangaInfo      mangadexapi.MangaInfo
	chapters       []mangadexapi.ChapterFullInfo
	lowestChapter  int
	highestChapter int
	chaptersRange  string
	language       string
	translateGroup string
	outputDir      string
	outputExt      string
	isJpg          bool
	isMerge        bool
}

func NewDownloadParam(chaptersRange string, lowestChapter, highestChapter int,
	language, translateGroup, outputDir, outputExt string, isJpg, isMerge bool) dlParam {

	return dlParam{
		mangaInfo:      mangadexapi.MangaInfo{},
		chapters:       []mangadexapi.ChapterFullInfo{},
		lowestChapter:  lowestChapter,
		highestChapter: highestChapter,
		chaptersRange:  chaptersRange,
		language:       language,
		translateGroup: translateGroup,
		outputDir:      outputDir,
		outputExt:      outputExt,
		isJpg:          isJpg,
		isMerge:        isMerge,
	}
}

func (p dlParam) printDlInteractiveParams() {
	printMangaInfo(p.mangaInfo)
	field.Println("---")
	dlChapterList := ""
	for _, c := range p.chapters {
		dlChapterList += " " + c.Number()
	}
	dp.Println(field.Sprint("Chapters:"), dlChapterList)
	dp.Println(field.Sprint("Output directory: "), p.outputDir)
	dp.Println(field.Sprint("Fileformat: "), p.outputExt)
	isMerging := "no"
	if p.isMerge {
		isMerging = "yes"
	}
	dp.Println(field.Sprint("Merging chapters: "), isMerging)
}

func (p dlParam) getMangaInfo(mangaId string) (mangadexapi.MangaInfo, error) {
	resp, err := client.GetMangaInfo(mangaId)
	if err != nil {
		return mangadexapi.MangaInfo{}, err
	}
	return resp.MangaInfo(), nil
}

func (p dlParam) downloadMergeChapters() {
	containerFile, err := filekit.NewContainer(p.outputExt)
	if err != nil {
		e.Printf("While creating output file: %v\n", err)
		os.Exit(1)
	}

	for _, chapter := range p.chapters {
		printChapterInfo(chapter)

		err = p.downloadProcess(containerFile, chapter)
		if err != nil {
			e.Printf("While downloading chapter: %v\n", err)
			os.Exit(1)
		}
	}

	chaptersRange := p.chapters[0].Info.Number()
	if len(p.chapters) > 1 {
		chaptersRange += "-" + p.chapters[len(p.chapters)-1].Number()
	}

	filename := ""
	if p.chapters[0].Translator() == "" {
		filename = fmt.Sprintf("[%s] %s ch. %s",
			p.language, p.mangaInfo.Title("en"), chaptersRange)
	} else {
		filename = fmt.Sprintf("[%s] %s ch. %s | %s",
			p.language, p.mangaInfo.Title("en"), chaptersRange, p.chapters[0].Translator())
	}

	metaInfo := metadata.NewMetadata(app.USER_AGENT, p.mangaInfo, p.chapters[0])
	err = containerFile.WriteOnDiskAndClose(p.outputDir, filename, metaInfo, p.chaptersRange)
	if err != nil {
		e.Printf("While saving %s on disk: %v\n", filename, err)
		os.Exit(1)
	}
}

func (p dlParam) downloadChapters() {
	for _, chapter := range p.chapters {
		printChapterInfo(chapter)

		containerFile, err := filekit.NewContainer(p.outputExt)
		if err != nil {
			e.Printf("While creating output file: %v\n", err)
			os.Exit(1)
		}

		err = p.downloadProcess(containerFile, chapter)
		if err != nil {
			e.Printf("While downloading chapter: %v\n", err)
			os.Exit(1)
		}

		filename := ""
		if chapter.Translator() == "" {
			filename = fmt.Sprintf("[%s] %s vol. %s ch. %s",
				p.language, p.mangaInfo.Title("en"), chapter.Volume(),
				chapter.Number())
		} else {
			filename = fmt.Sprintf("[%s] %s vol. %s ch. %s | %s",
				p.language, p.mangaInfo.Title("en"), chapter.Volume(),
				chapter.Number(),
				chapter.Translator())
		}
		metaInfo := metadata.NewMetadata(app.USER_AGENT, p.mangaInfo, chapter)
		err = containerFile.WriteOnDiskAndClose(p.outputDir, filename, metaInfo, "")
		if err != nil {
			e.Printf("While saving %s on disk: %v\n", filename, err)
			os.Exit(1)
		}
	}
}

func (p dlParam) downloadProcess(outputFile filekit.Container,
	chapter mangadexapi.ChapterFullInfo) error {
	if len(p.chapters) == 0 {
		return ErrEmptyChapters
	}

	if p.language == "ru" {
		printUaNotification()
	}

	files := chapter.PngFiles
	imgExt := "png"
	if p.isJpg {
		files = chapter.JpgFiles
		imgExt = "jpg"
	}

	dlbar, _ := pterm.DefaultProgressbar.WithTotal(len(files)).
		WithTitle("Downloading pages...").
		WithBarStyle(pterm.NewStyle(pterm.FgGreen)).Start()
	defer dlbar.Stop()

	for _, imageFile := range files {
		outputImage, err := client.DownloadImage(chapter.DownloadBaseURL,
			chapter.HashId, imageFile, p.isJpg)
		if err != nil {
			dlbar.WithBarStyle(pterm.NewStyle(pterm.FgRed)).
				UpdateTitle("Failed downloading").Stop()
			return err
		}

		if err := outputFile.AddFile(imgExt, outputImage); err != nil {
			dlbar.WithBarStyle(pterm.NewStyle(pterm.FgRed)).
				UpdateTitle("Failed downloading").Stop()
			return err
		}
		dlbar.Increment()
	}
	dp.Println("")
	return nil
}

func (p dlParam) DownloadSpecificChapter(chapterId string) {
	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapter info...")
	resp, err := client.GetChapterInfo(chapterId)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		os.Exit(1)
	}
	chapterInfo := resp.GetChapterInfo()
	chapterFullInfo, err := client.GetChapterImagesInFullInfo(chapterInfo)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapter info")

	mangaId := chapterInfo.GetMangaId()
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.chapters = []mangadexapi.ChapterFullInfo{chapterFullInfo}
	printShortMangaInfo(mangaInfo)
	p.downloadChapters()
}

func (p dlParam) DownloadLastChapter(mangaId string) {
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.mangaInfo = mangaInfo

	printShortMangaInfo(mangaInfo)
	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapter info...")
	chapterFullInfo, err := client.
		GetLastChapterFullInfo(mangaId, p.language, p.translateGroup)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapter info")
		e.Printf("While getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapter info")

	p.chapters = []mangadexapi.ChapterFullInfo{chapterFullInfo}

	p.downloadChapters()
}

func (p dlParam) DownloadChapters(mangaId string) {
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.mangaInfo = mangaInfo

	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapters info...")
	p.chapters, err = client.GetFullChaptersInfo(mangaId, p.language, p.translateGroup,
		p.lowestChapter, p.highestChapter)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapters info")
		e.Printf("While getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapters info")

	if len(p.chapters) == 0 {
		e.Printf("Chapters %s not found, try another "+
			"range, language, translation group etc.\n", p.chaptersRange)
		os.Exit(0)
	}

	printShortMangaInfo(mangaInfo)
	if p.isMerge {
		p.downloadMergeChapters()
	} else {
		p.downloadChapters()
	}
}

func (p dlParam) DownloadAllChapters(mangaId string) {
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	p.mangaInfo = mangaInfo

	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapters info...")
	p.chapters, err = client.GetAllFullChaptersInfo(mangaId, p.language, p.translateGroup)
	if err != nil {
		spinnerChapInfo.Fail("Failed to get chapters info")
		e.Printf("While getting manga chapters: %v\n", err)
		os.Exit(1)
	}
	spinnerChapInfo.Success("Fetched chapters info")

	if len(p.chapters) == 0 {
		e.Println("Chapters not found, try another language or translation group")
		os.Exit(0)
	}

	printShortMangaInfo(mangaInfo)
	if p.isMerge {
		p.downloadMergeChapters()
	} else {
		p.downloadChapters()
	}
}

const OPTION_MANGA_TEMPLATE = "%d | %s | %s"                 // numnber | authors | title
const OPTION_CHAPTER_TEMPLATE = "%d | vol. %s | ch. %s | %s" // number | volume | chapter | chapter title
const OPTION_SAVING_TEMPLATE = "%d | %s"

func toMangaInfoOptions(m []mangadexapi.MangaInfo, maxOptionSize int) ([]string, map[string]string) {
	printOptions := []string{}
	associationNums := make(map[string]string)
	for i, manga := range m {
		option := fmt.Sprintf(OPTION_MANGA_TEMPLATE, i+1, manga.Authors(), manga.Title("en"))
		if len(option)+2 >= maxOptionSize {
			option = option[:maxOptionSize-2]
		}
		printOptions = append(printOptions, option)
		associationNums[strconv.Itoa(i+1)] = manga.ID
	}
	return printOptions, associationNums
}

func getMangaNumOption(option string) string {
	return strings.Split(option, " | ")[0]
}

func toChaptersOptions(c []mangadexapi.Chapter, maxOptionSize int) ([]string, map[string]string) {
	options := []string{}
	associationNums := make(map[string]string)
	for i, chapter := range c {
		option := fmt.Sprintf(OPTION_CHAPTER_TEMPLATE,
			i+1, chapter.Volume(), chapter.Number(), chapter.Title())
		if len(option)+6 >= maxOptionSize {
			option = option[:maxOptionSize-6]
		}
		options = append(options, option)
		associationNums[strconv.Itoa(i+1)] = chapter.ID
	}
	return options, associationNums
}

func getChapterNumsFromOptions(options []string) []string {
	i := []string{}
	for _, o := range options {
		i = append(i, strings.Split(o, " | ")[0])
	}
	return i
}

func toSavingOptions() []string {
	options := []string{}
	options = append(options, fmt.Sprintf(OPTION_SAVING_TEMPLATE, 1, filekit.CBZ_EXT))
	options = append(options, fmt.Sprintf(OPTION_SAVING_TEMPLATE, 2, filekit.PDF_EXT))
	options = append(options, fmt.Sprintf(OPTION_SAVING_TEMPLATE, 3, filekit.EPUB_EXT))
	options = append(options, fmt.Sprintf(OPTION_SAVING_TEMPLATE, 4,
		filekit.CBZ_EXT+" + merge chapters in one file"))
	options = append(options, fmt.Sprintf(OPTION_SAVING_TEMPLATE, 5,
		filekit.PDF_EXT+" + merge chapters in one file"))
	options = append(options, fmt.Sprintf(OPTION_SAVING_TEMPLATE, 6,
		filekit.EPUB_EXT+" + merge chapters in one file"))
	return options
}

func getSavingOption(option string) (string, bool) {
	parts := strings.Split(option, " | ")
	if len(parts) != 2 {
		return "", false
	}
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		dp.Printfln("er %v", err)
		return "", false
	}
	dp.Println(num)
	switch num {
	case 1:
		return filekit.CBZ_EXT, false
	case 2:
		return filekit.PDF_EXT, false
	case 3:
		return filekit.EPUB_EXT, false
	case 4:
		return filekit.CBZ_EXT, true
	case 5:
		return filekit.PDF_EXT, true
	case 6:
		return filekit.EPUB_EXT, true
	default:
		return filekit.CBZ_EXT, false
	}
}

func (p dlParam) RunInteractiveDownload() {
	cols, rows := getTerminalSize()

	foundManga := []string{}
	associationMangaIdNums := make(map[string]string)
	for isSearching := true; isSearching; {
		clearOutput()
		searchTitle, _ := pterm.DefaultInteractiveTextInput.
			WithTextStyle(field).Show("Search manga")

		searchResult := []mangadexapi.MangaInfo{}

		for offset := 0; ; offset += 50 {
			mangaList, err := client.Find(searchTitle, 50, offset, true)
			if err != nil {
				e.Printfln("%v", err)
				os.Exit(1)
			}

			if len(mangaList.Data) == 0 {
				break
			}

			searchResult = append(searchResult, mangaList.List()...)
		}

		if len(searchResult) == 0 {
			isContinue, _ := pterm.DefaultInteractiveConfirm.
				Show("Manga not found, try again?")
			isSearching = isContinue
			continue
		}

		isSearching = false
		printOptions, associationNums := toMangaInfoOptions(searchResult, cols)
		maps.Copy(associationMangaIdNums, associationNums)
		foundManga = append(foundManga, printOptions...)
	}

	mangaInfo := mangadexapi.MangaInfo{}
	for isSelected := false; !isSelected; {
		clearOutput()
		mangaOption, _ := pterm.DefaultInteractiveSelect.WithOptions(foundManga).
			WithMaxHeight(rows - 2).Show("Select manga from list")
		mangaId := associationMangaIdNums[getMangaNumOption(mangaOption)]

		respMangaInfo, err := client.GetMangaInfo(mangaId)
		if err != nil {
			e.Printfln("%v", err)
			os.Exit(1)
		}

		printMangaInfo(respMangaInfo.Data)

		isSelected, _ = pterm.DefaultInteractiveConfirm.Show("Is correct manga?")
		if isSelected {
			mangaInfo = respMangaInfo.Data
		}
	}
	p.mangaInfo = mangaInfo

	clearOutput()
	translatedLanguage, _ := pterm.DefaultInteractiveSelect.
		WithOptions(mangaInfo.TranslatedLanguages()).WithFilter(false).
		WithMaxHeight(rows - 2).Show("Select language")
	p.language = translatedLanguage

	foundChapters := []mangadexapi.Chapter{}
	for offset := 0; ; offset += 50 {
		clearOutput()
		chapterlist, err := client.GetChaptersList(96, offset, mangaInfo.ID, p.language)
		if err != nil {
			e.Printfln("%v", err)
			os.Exit(1)
		}

		if len(chapterlist.Data) == 0 {
			break
		}

		foundChapters = append(foundChapters, chapterlist.Data...)
	}

	if len(foundChapters) == 0 {
		e.Println("Chapters not found, try another language or translation group")
		return
	}

	associationChapterIdNums := make(map[string]string)
	selectedChapterNums := []string{}
	for isSelected := false; !isSelected; {
		clearOutput()
		printChapterOptions, associationIdNums := toChaptersOptions(foundChapters, cols)
		selectedChapters, _ := pterm.DefaultInteractiveMultiselect.
			WithOptions(printChapterOptions).
			WithMaxHeight(rows - 3).Show("Select chapters from list")

		if len(selectedChapters) == 0 {
			isContinue, _ := pterm.DefaultInteractiveConfirm.
				Show("Chapters not selected, try again?")
			if !isContinue {
				return
			}
		}

		isSelected, _ = pterm.DefaultInteractiveConfirm.Show("Is correct chapters?")
		if isSelected {
			selectedChapterNums = append(selectedChapterNums,
				getChapterNumsFromOptions(selectedChapters)...)
			maps.Copy(associationChapterIdNums, associationIdNums)
		}
	}

	clearOutput()
	chaptersFullInfo := []mangadexapi.ChapterFullInfo{}
	for _, num := range selectedChapterNums {
		chapterFullInfo := mangadexapi.ChapterFullInfo{}
		for _, chapter := range foundChapters {
			if chapter.ID == associationChapterIdNums[num] {
				chapterFullInfo.Info = chapter
			}
		}

		imageInfo, err := client.GetChapterImageList(associationChapterIdNums[num])
		if err != nil {
			e.Printf("%v", err)
			os.Exit(1)
		}
		chapterFullInfo.DownloadBaseURL = imageInfo.BaseURL
		chapterFullInfo.HashId = imageInfo.ChapterMetaInfo.Hash
		chapterFullInfo.PngFiles = imageInfo.ChapterMetaInfo.Data
		chapterFullInfo.JpgFiles = imageInfo.ChapterMetaInfo.DataSaver

		chaptersFullInfo = append(chaptersFullInfo, chapterFullInfo)
	}
	p.chapters = chaptersFullInfo

	savingOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions(toSavingOptions()).
		WithMaxHeight(rows - 2).
		Show("Select saving options")

	outputExt, isMerge := getSavingOption(savingOption)
	p.outputExt = outputExt
	p.isMerge = isMerge

	clearOutput()
	outputDir, _ := pterm.DefaultInteractiveTextInput.
		WithTextStyle(field).
		Show("Save path (press Enter for save in current folder)")

	if outputDir == "" {
		outputDir = "."
	}
	p.outputDir = outputDir

	clearOutput()
	p.printDlInteractiveParams()
	isCorrectDlParams, _ := pterm.DefaultInteractiveConfirm.
		Show("Is correct downloading parameters?")
	if !isCorrectDlParams {
		field.Println("Run interactive mode again!")
		return
	}

	field.Println("Downloading chapters...")
	if p.isMerge {
		p.downloadMergeChapters()
	} else {
		p.downloadChapters()
	}
}
