/*
@author: sk
@date: 2024/5/8
*/
package main

import (
	"encoding/json"
	"time"
)

// index文件存储暂存区信息    HEAD 存储版本区文件    当前工作文件系统为工作区
type GitIndex struct { // 实际是以二进制进行存储，有文件头与文件正文
	Version int
	Count   int
	Items   []*GitIndexItem
}

type GitIndexItem struct {
	CreateTime time.Time
	ModifyTime time.Time
	FileSize   uint64
	Sha        string
	Name       string
}

func OpenIndex(gitDir string) *GitIndex {
	bs := ReadFile(gitDir, IndexFile)
	index := &GitIndex{}
	err := json.Unmarshal(bs, index)
	HandleErr(err)
	return index
}

func WriteIndex(gitDir string, index *GitIndex) {
	bs, err := json.Marshal(index)
	HandleErr(err)
	WriteFile(bs, gitDir, IndexFile)
}
