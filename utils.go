/*
@author: sk
@date: 2024/4/27
*/
package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Mkdir(paths ...string) {
	err := os.Mkdir(path.Join(paths...), os.ModePerm)
	HandleErr(err)
}

func CreateFile(paths ...string) {
	_, err := os.Create(path.Join(paths...))
	HandleErr(err)
}

func ReadFile(paths ...string) []byte {
	bs, err := os.ReadFile(path.Join(paths...))
	HandleErr(err)
	return bs
}

func ReadDir(paths ...string) []os.DirEntry {
	entries, err := os.ReadDir(path.Join(paths...))
	HandleErr(err)
	return entries
}

func WriteFile(bs []byte, paths ...string) {
	err := os.WriteFile(path.Join(paths...), bs, os.ModePerm)
	HandleErr(err)
}

func PathExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func GetGitPath(path0 string) string {
	gitPath := path.Join(path0, GitDir)
	if PathExist(gitPath) {
		return gitPath
	}
	index := strings.LastIndex(path0, "/")
	return GetGitPath(path0[:index])
}

func ParseInt(str string) int {
	res, err := strconv.ParseInt(str, 10, 64)
	HandleErr(err)
	return int(res)
}

func FormatInt(val int) string {
	return strconv.FormatInt(int64(val), 10)
}

func Hash(bs []byte) string {
	temp := sha256.Sum256(bs)
	res := make([]byte, 0, len(temp))
	for i := 0; i < len(temp); i++ {
		res = append(res, temp[i])
	}
	return fmt.Sprintf("%x", res)
}

func UniqueList(items []string) []string {
	temp := make(map[string]struct{}, len(items))
	for _, item := range items {
		temp[item] = struct{}{}
	}
	res := make([]string, 0, len(temp))
	for key := range temp {
		res = append(res, key)
	}
	return res
}

func Mode2Type(mode uint8) string {
	switch mode {
	case TreeMode:
		return TreeType
	case BlobMode:
		return BlobType
	case CommitMode:
		return CommitType
	default:
		panic(fmt.Sprintf("invalid mode %v", mode))
	}
}
