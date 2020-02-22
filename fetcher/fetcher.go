package fetcher

import (
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// NewCollector 生成colly 对象
func NewCollector() *colly.Collector {
	uaStr := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36"
	c := colly.NewCollector(
		colly.DetectCharset(),
	)
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
		r.Headers.Set("Accept-Encoding", "gzip, deflate")
		r.Headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
		r.Headers.Set("X-UA-Compatible", uaStr)
	})
	extensions.RandomUserAgent(c)
	extensions.Referer(c)
	return c
}
