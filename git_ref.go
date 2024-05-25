/*
@author: sk
@date: 2024/4/30
*/
package main

import "strings"

// 解析引用为 sha1 relPath是相对 .git 的路径
// refs/heads 引用 不过存储各个分支
// refs/tags 引用 不过存储各个tag
// HEAD 引用且是间接引用存储当前分支

func ParseRef(gitDir string, relPath string) string {
	temp := ReadFile(gitDir, relPath)
	data := string(temp)
	if strings.HasPrefix(data, "ref:") { // 相对引用
		return ParseRef(gitDir, data[5:])
	} else {
		return data // 绝对引用
	}
}
