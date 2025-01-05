package mdx

import (
	"errors"
	"maps"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/arimatakao/mdx/app"
	"github.com/arimatakao/mdx/filekit"
	"github.com/arimatakao/mdx/filekit/metadata"
	"github.com/arimatakao/mdx/mangadexapi"
	"github.com/pterm/pterm"
)

var (
	ErrEmptyChapters         = errors.New("empty chapters")
	selectedVolumeChapterMap map[string][]mangadexapi.Chapter
)

type dlParam struct {
	mangaInfo      mangadexapi.MangaInfo
	chapters       []mangadexapi.ChapterFullInfo
	lowestChapter  int
	highestChapter int
	lowestVolume   int
	highestVolume  int
	chaptersRange  string
	volumesRange   string
	language       string
	translateGroup string
	outputDir      string
	outputExt      string
	isJpg          bool
	isMerge        bool
	isVolume       bool
}

func NewDownloadParam(chaptersRange, volumesRange string, lowestChapter, highestChapter, lowestVolume, highestVolume int,
	language, translateGroup, outputDir, outputExt string, isJpg, isMerge, isVolume bool) dlParam {

	return dlParam{
		mangaInfo:      mangadexapi.MangaInfo{},
		chapters:       []mangadexapi.ChapterFullInfo{},
		lowestChapter:  lowestChapter,
		highestChapter: highestChapter,
		lowestVolume:   lowestVolume,
		highestVolume:  highestVolume,
		chaptersRange:  chaptersRange,
		volumesRange:   volumesRange,
		language:       language,
		translateGroup: translateGroup,
		outputDir:      outputDir,
		outputExt:      outputExt,
		isJpg:          isJpg,
		isMerge:        isMerge,
		isVolume:       isVolume,
	}
}

func (p dlParam) printDlInteractiveParams() {
	printMangaInfo(p.mangaInfo)
	field.Println("---")
	dlChapterList, dlVolumeList := "", ""
	for _, c := range p.chapters {
		dlChapterList += " " + c.Number()
		if !strings.Contains(dlVolumeList, c.Volume()) {
			dlVolumeList += " " + c.Volume()
		}
	}
	dp.Println(field.Sprint("Chapters:"), dlChapterList)
	dp.Println(field.Sprint("Volumes:"), dlVolumeList)
	dp.Println(field.Sprint("Output directory: "), p.outputDir)
	dp.Println(field.Sprint("Fileformat: "), p.outputExt)
	isMerging := "no"
	if p.isMerge {
		isMerging = "yes"
	}
	if p.isVolume {
		dp.Println(field.Sprint("Merging volumes: "), isMerging)
	} else {
		dp.Println(field.Sprint("Merging chapters: "), isMerging)
	}
}

func (p dlParam) getMangaInfo(mangaId string) (mangadexapi.MangaInfo, error) {
	resp, err := client.GetMangaInfo(mangaId)
	if err != nil {
		return mangadexapi.MangaInfo{}, err
	}
	return resp.MangaInfo(), nil
}

func (p dlParam) downloadMergeVolumes() {
	for volumeId, volume := range selectedVolumeChapterMap {
		containerFile, err := filekit.NewContainer(p.outputExt)
		if err != nil {
			e.Printf("While creating output file: %v\n", err)
			os.Exit(1)
		}

		volumeChaptersRange := []string{}
		for _, chapter := range volume {
			for _, chapterFullInfo := range p.chapters {
				if chapterFullInfo.Info.ID == chapter.ID && !contains(volumeChaptersRange, chapterFullInfo.Info.Number()) {
					volumeChaptersRange = append(volumeChaptersRange, chapterFullInfo.Info.Number())
					volumeId = chapterFullInfo.Volume()
					err = p.downloadProcess(containerFile, chapterFullInfo)
					if err != nil {
						e.Printf("While downloading chapter: %v\n", err)
						os.Exit(1)
					}
					break
				}
			}
		}
		startChapter := minChapter(volumeChaptersRange)
		endChapter := maxChapter(volumeChaptersRange)
		chaptersRange := startChapter + "-" + endChapter
		filename := pterm.Sprintf("[%s] %s | vol. %s | ch. %s",
			p.language, p.mangaInfo.Title("en"), volumeId, chaptersRange)
		metaInfo := metadata.NewMetadata(app.USER_AGENT, p.mangaInfo, selectedVolumeChapterMap[volumeId][0])
		err = containerFile.WriteOnDiskAndClose(p.outputDir, filename, metaInfo, "")
		if err != nil {
			e.Printf("While saving %s on disk: %v\n", filename, err)
			os.Exit(1)
		}
	}
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
		filename = pterm.Sprintf("[%s] %s ch. %s",
			p.language, p.mangaInfo.Title("en"), chaptersRange)
	} else {
		filename = pterm.Sprintf("[%s] %s ch. %s | %s",
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
			filename = pterm.Sprintf("[%s] %s vol. %s ch. %s",
				p.language, p.mangaInfo.Title("en"), chapter.Volume(),
				chapter.Number())
		} else {
			filename = pterm.Sprintf("[%s] %s vol. %s ch. %s | %s",
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

func (p dlParam) DownloadVolumes(mangaId string) {
	spinnerMangaInfo, _ := pterm.DefaultSpinner.Start("Fetching manga info...")
	mangaInfo, err := p.getMangaInfo(mangaId)
	if err != nil {
		spinnerMangaInfo.Fail("Failed to get manga info")
		os.Exit(1)
	}
	spinnerMangaInfo.Success("Fetched manga info")
	printShortMangaInfo(mangaInfo)
	p.mangaInfo = mangaInfo
	spinnerVolChapInfo, _ := pterm.DefaultSpinner.Start("Fetching volume and chapter info...")
	selectedVolumeChapterMap = make(map[string][]mangadexapi.Chapter)
	foundChapters := []mangadexapi.Chapter{}
	for offset := 0; ; offset += 50 {
		chapterlist, err := client.GetChaptersList(96, offset, mangaId, p.language)
		if err != nil {
			spinnerVolChapInfo.Fail("Failed to get volume info")
			os.Exit(1)
		}

		if len(chapterlist.Data) == 0 {
			break
		}

		foundChapters = append(foundChapters, chapterlist.Data...)
	}

	if len(foundChapters) == 0 {
		spinnerVolChapInfo.Fail("Volume not found, try another language or translation group")
		return
	}
	spinnerVolChapInfo.Success("Fetched volume info")
	spinnerVolInfo, _ := pterm.DefaultSpinner.Start("Creating volume and chapter map...")
	for _, chapter := range foundChapters {
		volume := chapter.Volume()
		// convert volume to int
		volumeInt, err := strconv.Atoi(volume)
		if err != nil {
			spinnerVolInfo.Fail("Failed to convert volume to int")
			os.Exit(1)
		}
		if volumeInt >= p.lowestVolume && volumeInt <= p.highestVolume {
			selectedVolumeChapterMap[volume] = append(selectedVolumeChapterMap[volume], chapter)
		}
	}
	spinnerVolInfo.Success("Created volume and chapter map")
	spinnerChapInfo, _ := pterm.DefaultSpinner.Start("Fetching chapter info...")
	chaptersFullInfo := []mangadexapi.ChapterFullInfo{}
	for _, volume := range selectedVolumeChapterMap {
		for _, chapter := range volume {
			chapterFullInfo := mangadexapi.ChapterFullInfo{}
			chapterFullInfo.Info = chapter

			imageInfo, err := client.GetChapterImageList(chapter.ID)
			if err != nil {
				spinnerChapInfo.Fail("Failed to get chapter info")
				os.Exit(1)
			}
			chapterFullInfo.DownloadBaseURL = imageInfo.BaseURL
			chapterFullInfo.HashId = imageInfo.ChapterMetaInfo.Hash
			chapterFullInfo.PngFiles = imageInfo.ChapterMetaInfo.Data
			chapterFullInfo.JpgFiles = imageInfo.ChapterMetaInfo.DataSaver

			chaptersFullInfo = append(chaptersFullInfo, chapterFullInfo)
		}
	}
	spinnerChapInfo.Success("Fetched chapter info")
	p.chapters = chaptersFullInfo
	if p.isMerge {
		p.downloadMergeVolumes()
	} else {
		p.downloadChapters()
	}
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

func (p dlParam) DownloadAllChapters(mangaId string, isVolume bool) {
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

	if isVolume {
		// Create volume mapping for all chapters
		selectedVolumeChapterMap = make(map[string][]mangadexapi.Chapter)
		for _, chapter := range p.chapters {
			volume := chapter.Volume()
			selectedVolumeChapterMap[volume] = append(selectedVolumeChapterMap[volume], chapter.Info)
		}
		p.downloadMergeVolumes()
	} else if p.isMerge {
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
		option := pterm.Sprintf(OPTION_MANGA_TEMPLATE, i+1, manga.Authors(), manga.Title("en"))
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
		option := pterm.Sprintf(OPTION_CHAPTER_TEMPLATE,
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

func toSavingOptions(isVolume bool) []string {
	options := []string{}
	options = append(options, pterm.Sprintf(OPTION_SAVING_TEMPLATE, 1, filekit.CBZ_EXT))
	options = append(options, pterm.Sprintf(OPTION_SAVING_TEMPLATE, 2, filekit.PDF_EXT))
	options = append(options, pterm.Sprintf(OPTION_SAVING_TEMPLATE, 3, filekit.EPUB_EXT))
	if !isVolume {
		options = append(options, pterm.Sprintf(OPTION_SAVING_TEMPLATE, 4,
			filekit.CBZ_EXT+" + merge chapters in one file"))
		options = append(options, pterm.Sprintf(OPTION_SAVING_TEMPLATE, 5,
			filekit.PDF_EXT+" + merge chapters in one file"))
		options = append(options, pterm.Sprintf(OPTION_SAVING_TEMPLATE, 6,
			filekit.EPUB_EXT+" + merge chapters in one file"))
	}
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
	p.isVolume = false

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
		e.Println("Chapters and/or volumes not found, try another language or translation group")
		return
	}

	downloadOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"Download by Volume", "Download by Chapter"}).
		WithMaxHeight(rows - 2).Show("Select download option")

	selectedChapterNums := []string{}
	associationChapterIdNums := make(map[string]string)

	if downloadOption == "Download by Volume" {
		p.isVolume = true
		volumeChapterMap := make(map[string][]mangadexapi.Chapter)
		for _, chapter := range foundChapters {
			volume := chapter.Volume()
			volumeChapterMap[volume] = append(volumeChapterMap[volume], chapter)
		}

		selectedVolumes := []string{}
		for isSelected := false; !isSelected; {
			clearOutput()

			// Prepare options with volume and chapter range
			printVolumeOptions := []string{}
			volumes := make([]int, 0, len(volumeChapterMap))
			for volume := range volumeChapterMap {
				vol, _ := strconv.Atoi(volume)
				volumes = append(volumes, vol)
			}
			sort.Ints(volumes)

			for _, vol := range volumes {
				volume := strconv.Itoa(vol)
				chapters := volumeChapterMap[volume]
				if len(chapters) > 0 {
					startChapter := chapters[0].Number()
					endChapter := chapters[len(chapters)-1].Number()
					option := pterm.Sprintf(
						"%s | Volume %s | Chapters %s-%s",
						p.mangaInfo.Title("en"), volume, startChapter, endChapter,
					)
					printVolumeOptions = append(printVolumeOptions, option)
				}
			}

			selectedVolumes, _ = pterm.DefaultInteractiveMultiselect.
				WithOptions(printVolumeOptions).
				WithMaxHeight(rows - 3).Show("Select volumes from list")

			if len(selectedVolumes) == 0 {
				isContinue, _ := pterm.DefaultInteractiveConfirm.
					Show("Volumes not selected, try again?")
				if !isContinue {
					return
				}
			}

			isSelected, _ = pterm.DefaultInteractiveConfirm.Show("Is correct volumes?")
			if isSelected {
				selectedVolumeNumbers := []int{}
				// Build the actual "UI index -> chapter ID" map (just like we do for chapters)
				// We'll keep track of a "virtual" selection index (i) for each chapter found inside each volume.
				i := 1
				selectedVolumeChapterMap = make(map[string][]mangadexapi.Chapter)
				for _, selectedVolume := range selectedVolumes {
					// Extract volume number from "xxx | Volume NN |..."
					volumeStr := strings.TrimSpace(strings.Split(selectedVolume, "|")[1][7:])
					volumeNum, _ := strconv.Atoi(volumeStr)
					selectedVolumeNumbers = append(selectedVolumeNumbers, volumeNum)
					selectedVolumeChapterMap[volumeStr] = volumeChapterMap[volumeStr]
					// For each chapter in that volume
					for _, ch := range volumeChapterMap[volumeStr] {
						// We store "i -> chapter.ID" into associationChapterIdNums
						idx := strconv.Itoa(i)
						associationChapterIdNums[idx] = ch.ID
						selectedChapterNums = append(selectedChapterNums, idx)
						i++
					}
				}
				sort.Ints(selectedVolumeNumbers)
				p.lowestVolume = selectedVolumeNumbers[0]
				p.highestVolume = selectedVolumeNumbers[len(selectedVolumeNumbers)-1]
				p.volumesRange = strconv.Itoa(p.lowestVolume) + "-" + strconv.Itoa(p.highestVolume)
			}
		}
	} else {
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
				selectedChapterNums = append(selectedChapterNums, getChapterNumsFromOptions(selectedChapters)...)
				maps.Copy(associationChapterIdNums, associationIdNums)
			}
		}
	}
	clearOutput()

	savingOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions(toSavingOptions(p.isVolume)).
		WithMaxHeight(rows - 2).
		Show("Select saving options")

	outputExt, isMerge := getSavingOption(savingOption)
	p.outputExt = outputExt
	p.isMerge = isMerge
	if p.isVolume {
		p.isMerge = true
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

	field.Println("Downloading selections...")
	if p.isMerge && p.isVolume {
		p.downloadMergeVolumes()
	} else if p.isMerge {
		p.downloadMergeChapters()
	} else {
		p.downloadChapters()
	}
}

func maxChapter(chapters []string) string {
	if len(chapters) == 0 {
		return ""
	}
	max := chapters[0]
	for _, ch := range chapters {
		if ch > max {
			max = ch
		}
	}
	return max
}

func minChapter(chapters []string) string {
	if len(chapters) == 0 {
		return ""
	}
	min := chapters[0]
	for _, ch := range chapters {
		if ch < min {
			min = ch
		}
	}
	return min
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
