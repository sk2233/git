/*
@author: sk
@date: 2024/4/27
*/
package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

/*
object文件格式
文件类型(commit/tag/blob) 数据大小\00具体内容

默认所有文件都存储在objects下面，并截取前两个字符作为目录，防止单个目录下文件太多，这里不再分割文件名作为目录
*/

type GitObject interface {
	fmt.Stringer

	GetType() string
	GetData() []byte
}

/*
文件格式
key val
key val
gpgsig ----BEGIN xxx     这里忽略gpg的begin与end  使用空行分割

asdasasdasdasd
----END xxx

commit msg
*/

type CommitObject struct {
	kvs  map[string][]string
	keys []string // 必须保持顺序，防止读取与写入因为顺序不同导致文件hash不同，但实际是同一个文件
	gpg  []byte   // 校验信息
	msg  string   // 提交的git消息
}

func (c *CommitObject) String() string {
	buff := strings.Builder{}
	for _, key := range c.keys {
		for _, val := range c.kvs[key] {
			buff.WriteString(key)
			buff.WriteRune(' ')
			buff.WriteString(val)
			buff.WriteRune('\n')
		}
	}
	buff.WriteString(c.msg)
	buff.WriteRune('\n')
	return buff.String()
}

func (c *CommitObject) GetType() string {
	return CommitType
}

func (c *CommitObject) GetData() []byte {
	buff := bytes.Buffer{}
	sort.Slice(c.keys, func(i, j int) bool { // 写入必须保证其一致性
		return c.keys[i] < c.keys[j]
	})
	for _, key := range c.keys {
		vals := c.kvs[key]
		sort.Slice(vals, func(i, j int) bool {
			return vals[i] < vals[j]
		})
		for _, val := range vals {
			buff.WriteString(key)
			buff.WriteRune(' ')
			buff.WriteString(val)
			buff.WriteRune('\n')
		}
	}
	buff.WriteRune('\n')
	buff.Write(c.gpg)
	buff.WriteString("\n\n")
	buff.WriteString(c.msg)
	return buff.Bytes()
}

func (c *CommitObject) GetShortMsg() string { // 只获取第一行
	return strings.Split(c.msg, "\n")[0]
}

func (c *CommitObject) GetValue(key string) []string {
	return c.kvs[key]
}

func NewCommitObject(data []byte) *CommitObject {
	datas := bytes.Split(data, []byte("\n"))
	index := 0
	kvs := make(map[string][]string)
	keys := make([]string, 0)
	for index < len(datas) && len(datas[index]) > 0 {
		items := bytes.Split(datas[index], []byte(" "))
		key := string(items[0])
		keys = append(keys, key)
		kvs[key] = append(kvs[key], string(items[1]))
		index++
	}
	keys = UniqueList(keys)
	index++
	gpg := make([]byte, 0)
	for index < len(datas) && len(datas[index]) > 0 {
		gpg = append(gpg, datas[index]...)
		index++
	}
	index++
	msg := make([]byte, 0)
	for index < len(datas) {
		msg = append(msg, datas[index]...)
		index++
	}
	return &CommitObject{
		kvs:  kvs,
		keys: keys,
		gpg:  gpg,
		msg:  string(msg),
	}
}

// Mode Path0x00Sha\n

type TreeItem struct {
	Mode uint8  // 这里只保存文件类型，实际文件类型采用 2byte保存，还有4byte保存文件读写权限
	Path string // 只保存一段路径，若有多个进行嵌套存储
	Sha  string // 对应内容文件hash值
}

func ComTreeItem(item1, item2 *TreeItem) bool {
	if item1.Mode != item2.Mode { // 名称相同的话，mode肯定不同
		return item1.Mode < item2.Mode
	}
	return item1.Path < item2.Path
}

type TreeObject struct {
	Items []*TreeItem
}

func (t *TreeObject) String() string {
	buff := strings.Builder{}
	for _, item := range t.Items {
		buff.WriteString(item.Sha)
		buff.WriteRune(' ')
		buff.WriteString(item.Path)
		buff.WriteRune(' ')
		buff.WriteString(Mode2Type(item.Mode))
		buff.WriteRune('\n')
	}
	return buff.String()
}

func (t *TreeObject) GetType() string {
	return TreeType
}

func (t *TreeObject) GetData() []byte {
	buff := bytes.Buffer{}
	sort.Slice(t.Items, func(i, j int) bool {
		return ComTreeItem(t.Items[i], t.Items[j])
	})
	for i, item := range t.Items {
		if i > 0 {
			buff.WriteRune('\n')
		}
		buff.WriteByte(item.Mode)
		buff.WriteRune(' ')
		buff.WriteString(item.Path)
		buff.WriteByte(0x00)
		buff.WriteString(item.Sha)
	}
	return buff.Bytes()
}

func NewTreeObject(data []byte) *TreeObject {
	datas := bytes.Split(data, []byte("\n"))
	items := make([]*TreeItem, 0, len(data))
	for i := 0; i < len(datas); i++ {
		mode := datas[i][0]
		index := bytes.IndexByte(datas[i], 0x00)
		path := string(datas[i][2:index])
		sha := string(datas[i][index+1:])
		items = append(items, &TreeItem{
			Mode: mode,
			Path: path,
			Sha:  sha,
		})
	}
	return &TreeObject{
		Items: items,
	}
}

/*
tag分为轻量tag与tag对象
轻量tag实际就是一个引用 ref
tag结构与CommitObject对象一致
*/

type TagObject struct {
	*CommitObject
}

func (t *TagObject) GetType() string {
	return TagType
}

func NewTagObject(data []byte) *TagObject {
	return &TagObject{NewCommitObject(data)}
}

type BlobObject struct {
	data []byte
}

func (b *BlobObject) String() string {
	return fmt.Sprintf("blob content data len = %d", len(b.data))
}

func (b *BlobObject) GetType() string {
	return BlobType
}

func (b *BlobObject) GetData() []byte {
	return b.data
}

func NewBlobObject(data []byte) *BlobObject {
	return &BlobObject{data: data}
}

// 这里 sha应该支持更多的类型  HEAD sha1 sha1前缀 tags branchs

func OpenObject(gitDir, sha string) GitObject {
	bs := ReadFile(gitDir, ObjectDir, sha) // 默认一次全部读取出来，大文件不友好
	index := bytes.IndexByte(bs, ' ')
	fileType := string(bs[:index])
	bs = bs[index+1:]
	index = bytes.IndexByte(bs, 0x00)
	fileSize := ParseInt(string(bs[:index]))
	bs = bs[index+1:]
	if len(bs) != fileSize {
		panic(fmt.Sprintf("filesize err len(bs) = %d , fileSize = %d", len(bs), fileSize))
	}

	switch fileType {
	case CommitType:
		return NewCommitObject(bs)
	case TreeType:
		return NewTreeObject(bs)
	case TagType:
		return NewTagObject(bs)
	case BlobType:
		return NewBlobObject(bs)
	default:
		panic(fmt.Sprintf("not support type %s", fileType))
	}
}

func WriteObject(gitDir string, object GitObject) {
	bs := object.GetData()

	buff := bytes.Buffer{}
	buff.WriteString(object.GetType())
	buff.WriteRune(' ')
	buff.WriteString(FormatInt(len(bs)))
	buff.WriteByte(0x00)
	buff.Write(bs)

	fileName := Hash(buff.Bytes())
	WriteFile(buff.Bytes(), gitDir, ObjectDir, fileName)
}
