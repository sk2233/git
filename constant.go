/*
@author: sk
@date: 2024/4/27
*/
package main

/*
常用文件
.git/branches/  分支目录
.git/objects/ 对象存储
.git/refs/ 参考存储 存在子文件夹  heads tags
.git/HEAD 对 HEAD的引用
.git/config 配置文件  不经常变动 暂时不支持配置文件
.git/description 描述文件，很少使用 暂不支持
*/

const (
	GitDir = ".git"

	BranchDir = "branches"
	ObjectDir = "objects"

	RefDir  = "refs"
	HeadDir = "heads"
	TagDir  = "tags"

	HeadFile      = "HEAD"
	IndexFile     = "index"
	GitIgnoreFile = ".gitignore"
)

const (
	CommitType = "commit"
	TreeType   = "tree"
	TagType    = "tag"
	BlobType   = "blob"
)

const ( // 没有 tag 类型？
	CommitMode = 0x16
	TreeMode   = 0x04
	BlobMode   = 0x10
)

const (
	ParentKey = "parent"
)
