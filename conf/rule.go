package conf

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config 基本配置
type Config struct {
	Engines      []string
	Rule         map[string]map[string]interface{} `json:"rules"`
	IgnoreDomain map[string]int                    `json:"ignores"`
}

/*
Config.Rule
	link_prefix: ||| -1 代表取chapter_url ||| 1 代表直接取URL ||| 0 代表使用域名加拼接
	chapter_selector: 用于寻找chapter 目录章节的元素
	content_selector: 用于寻找content 的元素
	chapter_tail: 如果存在，则作为chapter 附加添加到link 的后缀
*/

// RuleConfig 规则基本配置
var RuleConfig *Config

func init() {
	RuleConfig = &Config{}
	RuleConfig.Engines = []string{"baidu"}

	file, err := os.Open("rule.json")
	if err != nil {
		log.Fatal("read rule.json meet error: ", err)
	}
	b, _ := ioutil.ReadAll(file)

	// link_prefix ||| -1 代表取chapter_url ||| 1 代表直接取URL ||| 0 代表使用域名加拼接 ||| 2 代表需要拼接后缀.html（如/dir.html 等, 如果已存在.html 则不再添加）
	var rule Config
	err = json.Unmarshal(b, &rule)
	if err != nil {
		log.Fatal("unmarshal rule meet error :", err)
	}
	RuleConfig.Rule = rule.Rule
	RuleConfig.IgnoreDomain = rule.IgnoreDomain
}
