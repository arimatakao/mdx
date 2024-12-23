package mangadexapi

// ResponseVolumes represents the top-level response for volume data
type ResponseVolumes struct {
	Result  string            `json:"result"`
	Volumes map[string]Volume `json:"volumes"`
}

// Volume represents a manga volume and its chapters
type Volume struct {
	Volume   string                   `json:"volume"`
	Count    int                      `json:"count"`
	Chapters map[string]VolumeChapter `json:"chapters"`
}

// VolumeChapter represents a chapter within a volume
type VolumeChapter struct {
	Chapter string   `json:"chapter"`
	ID      string   `json:"id"`
	Others  []string `json:"others"`
	Count   int      `json:"count"`
}

// Helper methods for Volume
func (v Volume) GetChapterIDs() []string {
	ids := make([]string, 0, len(v.Chapters))
	for _, chapter := range v.Chapters {
		ids = append(ids, chapter.ID)
		ids = append(ids, chapter.Others...)
	}
	return ids
}

func (v Volume) GetChapterCount() int {
	return v.Count
}

func (v Volume) GetVolumeNumber() string {
	return v.Volume
}

// Helper methods for VolumeChapter
func (c VolumeChapter) GetChapterNumber() string {
	return c.Chapter
}

func (c VolumeChapter) GetMainID() string {
	return c.ID
}

func (c VolumeChapter) GetAllIDs() []string {
	ids := make([]string, 0, c.Count)
	ids = append(ids, c.ID)
	ids = append(ids, c.Others...)
	return ids
}
