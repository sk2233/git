/*
@author: sk
@date: 2024/5/9
*/
package main

import (
	"bytes"
	"regexp"
)

type GitIgnore struct { // 主要规则包含忽略规则与保留规则，这里只看忽略规则，每行都是正则的忽略规则
	Rules []*regexp.Regexp
}

func (i *GitIgnore) CheckIgnore(path0 string) bool {
	for _, rule := range i.Rules {
		if rule.MatchString(path0) {
			return true
		}
	}
	return false
}

func OpenIgnore(gitDir string) *GitIgnore {
	bs := ReadFile(gitDir, GitIgnoreFile)
	temps := bytes.Split(bs, []byte("\n"))
	rules := make([]*regexp.Regexp, 0)
	for _, item := range temps {
		item = bytes.TrimSpace(item)
		if len(item) == 0 {
			continue
		}
		rules = append(rules, regexp.MustCompile(string(item)))
	}
	return &GitIgnore{
		Rules: rules,
	}
}

func WriteIgnore(gitDir string, ignore *GitIgnore) {
	buff := bytes.Buffer{}
	for i, rule := range ignore.Rules {
		if i > 0 {
			buff.WriteRune('\n')
		}
		buff.WriteString(rule.String())
	}
	WriteFile(buff.Bytes(), gitDir, GitIgnoreFile)
}
