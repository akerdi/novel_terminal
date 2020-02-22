package searchengine

import (
	"fmt"
	"net/url"
	"novel/conf"
	"novel/fetcher"
	"novel/model"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

// BaiDuSearchEngine 使用百度搜索引擎对象
type BaiDuSearchEngine struct {
	parseRule       string
	searchRule      string
	domain          string
	parseResultFunc func(searchResult *model.SearchResult)
}

// NewBaiduSearchEngine 生成百度搜索引擎对象
func NewBaiduSearchEngine(parseResultFunc func(result *model.SearchResult)) *BaiDuSearchEngine {
	return &BaiDuSearchEngine{
		parseRule:       "#content_left h3.t a",
		searchRule:      "intitle: %s 小说 阅读",
		domain:          "http://www.baidu.com/s?wd=%s&ie=utf-8&rn=15&vf_bl=1",
		parseResultFunc: parseResultFunc,
	}
}

// EngineRun 使用引擎
func (engine *BaiDuSearchEngine) EngineRun(novelName string, group *sync.WaitGroup) {
	defer group.Done()
	searchKey := url.QueryEscape(fmt.Sprintf(engine.searchRule, novelName))
	requestUrl := fmt.Sprintf(engine.domain, searchKey)
	c := fetcher.NewCollector()
	fmt.Println("Search engine start request url: ", requestUrl)
	c.OnHTML(engine.parseRule, func(element *colly.HTMLElement) {
		group.Add(1)
		go engine.extractData(element, group)
	})
	err := c.Visit(requestUrl)
	if err != nil {
		fmt.Println(err)
	}
}
func (engine *BaiDuSearchEngine) extractData(element *colly.HTMLElement, group *sync.WaitGroup) {
	defer group.Done()
	href := element.Attr("href")
	title := element.Text
	// result := &model.SearchResult{Href: href, Title: title, IsParse: 1, Host: "www.baidu"}
	// engine.parseResultFunc(result)
	c := fetcher.NewCollector()
	c.OnResponse(func(response *colly.Response) {
		realURL := response.Request.URL.String()
		isContain := strings.Contains(realURL, "baidu")
		if isContain {
			return
		}
		host := response.Request.URL.Host

		_, ok := conf.RuleConfig.IgnoreDomain[host]
		fmt.Println("host:: ", host, " 是否忽略该网站?: ", ok)
		if ok {
			return
		}
		isParse := engine.CheckIsParse(host)
		fmt.Println("host:: ", host, " 是否已有该模板?: ", isParse)
		if !isParse {
			return
		}
		result := &model.SearchResult{Href: realURL, Title: title, IsParse: isParse, Host: host}
		engine.parseResultFunc(result)
	})
	err := c.Visit(href)
	if err != nil {
		fmt.Println(err)
	}
}

// CheckIsParse 将页面中的链接对比Rule 是否有该模板
func (engine *BaiDuSearchEngine) CheckIsParse(host string) bool {
	isParse := false
	for key := range conf.RuleConfig.Rule {
		if host == key {
			isParse = true
			break
		}
	}
	return isParse
}
