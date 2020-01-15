package model

type SearchResult struct {
	Href    string `json:"href"`
	Title   string `json:"title"`
	IsParse bool   `json:"is_parse"`
	Host    string `json:"host"`
}

type NovelChapter struct {
	Name       string                 `json:"name"`
	OriginUrl  string                 `json:"origin_url"`
	Chapters   []*NovelChapterElement `json:"chapters"` // 储存每个章节基本元素
	LinkPrefix string                 `json:"link_prefix"`
	Domain     string                 `json:"domain"`
}
type NovelChapterElement struct {
	ChapterName string `json:"chapter_name"` // 章节名称
	ChapterHref string `json:""chapter_href` // 章节链接
}

type NovelContent struct {
	NovelName   string `json:"novel_name"`
	Title       string `json:"title"`
	ContentURL  string `json:"content_url"`
	Content     string `json:"content"`
	PreChapter  string `json:"pre_chapter"`
	NextChapter string `json:"next_chapter"`
}
