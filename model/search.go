package model

// SearchResult 搜索结果对象
type SearchResult struct {
	Href    string `json:"href"`
	Title   string `json:"title"`
	IsParse bool   `json:"is_parse"`
	Host    string `json:"host"`
}

// NovelChapter 小说章节对象。每个网址链接对象的小说章节
type NovelChapter struct {
	Name       string                 `json:"name"`
	OriginURL  string                 `json:"origin_url"`
	Chapters   []*NovelChapterElement `json:"chapters"` // 储存每个章节基本元素
	LinkPrefix string                 `json:"link_prefix"`
	Domain     string                 `json:"domain"`
}

// NovelChapterElement 小说章节保存的各个章节数据元素
type NovelChapterElement struct {
	ChapterName string `json:"chapter_name"` // 章节名称
	ChapterHref string `json:"chapter_href"` // 章节链接
}

// NovelContent 小说正文对象
type NovelContent struct {
	NovelName   string `json:"novel_name"`
	Title       string `json:"title"`
	ContentURL  string `json:"content_url"`
	Content     string `json:"content"`
	PreChapter  string `json:"pre_chapter"`
	NextChapter string `json:"next_chapter"`
}
