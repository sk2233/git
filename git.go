/*
@author: sk
@date: 2024/4/27
*/
package main

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

func Add(gitDir string, paths ...string) { // git add
	index := OpenIndex(gitDir) // 添加文件到暂存区
	items := index.Items
	for _, item := range paths {
		i := strings.LastIndexByte(item, '/')
		bs := ReadFile(gitDir, item)
		blobObj := &BlobObject{}
		blobObj.data = bs
		WriteObject(gitDir, blobObj) // 转移文件，并在 index中标记，注意若是与index中的已有文件重复要覆盖
		items = append(items, &GitIndexItem{
			FileSize: uint64(len(bs)),
			Sha:      Hash(bs),
			Name:     item[i+1:],
		})
	}
	index.Items = items
	index.Count = len(items)
	WriteIndex(gitDir, index)
}

func CatFile(gitDir, sha string) { // git cat-file   因为文件加密了，要用这个查看文件内容是解密后的
	gitObj := OpenObject(gitDir, sha)
	fmt.Printf("sha = %s , gitObj :\n%s", sha, gitObj)
}

// 校验 path0 是否需要被忽略 , git忽略规则与文件path0都是从项目开始的相对路径
func CheckIgnore(gitDir string, path0 string) { // git check -ignore
	gitIgnore := OpenIgnore(gitDir)
	res := gitIgnore.CheckIgnore(path0)
	fmt.Printf("path : %s , %v", path0, res)
}

// 把gitDir下sha的文件应用到path0目录  sha文件可能是单个文件或文件树，若是文件树递归处理 path0必须保证存在且已经包含了文件名

func Checkout(sha, path0 string) { // git checkout
	checkout(GetGitPath(path0), sha, path0) // 切换到对应的分支 把其最终指向的树中的内容应用到本地，也会存在删除操作,整个过程对于没有追踪的文件没有影响
}

func checkout(gitDir string, sha string, path0 string) {
	gitObj := OpenObject(gitDir, sha)
	switch temp := gitObj.(type) {
	case *TreeObject: // 目录需要确保目录存在
		if !PathExist(path0) {
			Mkdir(path0)
		}
		for _, item := range temp.Items { // 再递归目录
			path1 := path.Join(path0, item.Path)
			checkout(gitDir, item.Sha, path1)
		}
	case *BlobObject: // 文件直接写入,没有会创建
		WriteFile(temp.GetData(), path0)
	default:
		panic(fmt.Sprintf("gitObj %s not a tree or blob", gitObj))
	}
}

func Commit(gitDir string) { // git commit   把暂存区的内容应用到分支上
	index := OpenIndex(gitDir)
	treeSha := index2Tree(index) // 暂存区转分支
	commit := &CommitObject{}
	commit.kvs["tree"] = []string{treeSha}
	commit.kvs["parent"] = []string{ParseRef(gitDir, HeadFile)} // 父分支 若是meage 可以指定多个
	WriteObject(gitDir, commit)
	commitSha := "commit"
	WriteFile([]byte(commitSha), gitDir, HeadFile)
}

func index2Tree(index *GitIndex) string {
	return "sha"
}

/*
object文件格式
文件类型(commit/tag/blob) 数据大小\00具体内容
*/

func HashObject(path0, type0 string) { // git hash-object 添加文件写入gitObj
	gitDir := GetGitPath(path0)
	bs := ReadFile(path0)
	var gitObj GitObject
	switch type0 {
	case CommitType:
		gitObj = NewCommitObject(bs)
	case TreeType:
		gitObj = NewTreeObject(bs)
	case TagType:
		gitObj = NewTagObject(bs)
	case BlobType:
		gitObj = NewBlobObject(bs)
	default:
		panic(fmt.Sprintf("not support type %s", type0))
	}
	WriteObject(gitDir, gitObj)
}

/*
常用文件
.git/branches/  分支目录
.git/objects/ 对象存储
.git/refs/ 参考存储 存在子文件夹  heads tags
.git/HEAD 对 HEAD的引用
.git/config 配置文件  不经常变动 暂时不支持配置文件
.git/description 描述文件，很少使用 暂时不支持
*/

func Init(path string) { // git init  创建对应文件
	Mkdir(path, GitDir)

	Mkdir(path, GitDir, BranchDir)
	Mkdir(path, GitDir, ObjectDir)

	Mkdir(path, GitDir, RefDir)
	Mkdir(path, GitDir, RefDir, HeadDir)
	Mkdir(path, GitDir, RefDir, TagDir)

	CreateFile(path, GitDir, HeadFile)
}

func Log(gitDir, sha string) { // git log 查看提交信息，一个提交可能有多个子提交例如 merge
	gitObj, ok := OpenObject(gitDir, sha).(*CommitObject) // 必须是提交类型
	if !ok {
		panic(fmt.Sprintf("gitDir = %s , sha = %s not a commit", gitDir, sha))
	}
	fmt.Printf("%s %s\n", sha, gitObj.GetShortMsg())
	patents := gitObj.GetValue(ParentKey)
	for _, patent := range patents {
		fmt.Printf("%s -> %s\n", sha, patent)
		Log(gitDir, patent)
	}
}

func LsFiles(gitDir string) { // git ls-files  查看暂存区的文件
	index := OpenIndex(gitDir)
	for _, item := range index.Items {
		fmt.Printf("Name = %s , Sha = %v , FileSize = %d , CreateTime = %v , ModifyTime = %v",
			item.Name, item.Sha, item.FileSize, item.CreateTime, item.ModifyTime)
	}
}

func LsTree(gitDir, sha string) { // git ls-tree  查看git中的树文件
	treeObj, ok := OpenObject(gitDir, sha).(*TreeObject)
	if !ok {
		panic(fmt.Sprintf("gitDir = %s , sha = %s not a tree", gitDir, sha))
	}
	for _, item := range treeObj.Items {
		if item.Mode == TreeMode {
			LsTree(gitDir, item.Sha)
		} else {
			fmt.Printf("%s %s %s\n", item.Sha, item.Path, Mode2Type(item.Mode))
		}
	}
}

// 解析 HEAD , tags heads , sha1 sha1前缀 -> sha1

func RevParse(gitDir, name string) { // git rev-parse   通用解析
	items := revParse(gitDir, name)
	if len(items) == 0 {
		fmt.Println("NONE")
	} else {
		for _, item := range items {
			fmt.Println(item)
		}
	}
}

func revParse(gitDir string, name string) []string {
	if len(strings.TrimSpace(name)) == 0 {
		return make([]string, 0)
	}
	if name == HeadFile { // HEAD
		return []string{ParseRef(gitDir, name)}
	}
	dir := path.Join(RefDir, HeadDir) // heads
	items := ReadDir(gitDir, dir)
	for _, item := range items { // 暂时不考虑内嵌文件夹 只有一级
		if item.Name() == name {
			return []string{ParseRef(gitDir, path.Join(dir, name))}
		}
	}
	dir = path.Join(RefDir, TagDir) // tags
	items = ReadDir(gitDir, dir)
	for _, item := range items { // 暂时不考虑内嵌文件夹 只有一级
		if item.Name() == name {
			return []string{ParseRef(gitDir, path.Join(dir, name))}
		}
	}
	items = ReadDir(gitDir, ObjectDir) // sha
	res := make([]string, 0)
	for _, item := range items {
		if strings.HasPrefix(item.Name(), name) {
			res = append(res, item.Name())
		}
	}
	return res
}

func Rm(gitDir string, paths ...string) { // git rm  从暂存区移出，文件不能移出可能还有其他引用
	index := OpenIndex(gitDir)
	set := make(map[string]struct{})
	for _, item := range paths {
		set[item] = struct{}{}
	}
	temp := make([]*GitIndexItem, 0)
	for _, item := range index.Items {
		if _, ok := set[item.Name]; !ok {
			temp = append(temp, item)
		}
	}
	index.Items = temp
	index.Count = len(temp)
	WriteIndex(gitDir, index)
}

func ShowRef(gitDir string) { // git show -ref  查看工作区目录下已有文件的sha1值
	showRef(gitDir, RefDir)
}

// 这里传进来的都保证是目录
func showRef(gitDir string, relPath string) {
	items := ReadDir(gitDir, relPath)
	for _, item := range items {
		path0 := path.Join(relPath, item.Name())
		if item.IsDir() {
			showRef(gitDir, path0)
		} else {
			sha1 := ParseRef(gitDir, path0)
			fmt.Printf("%s <- %s", sha1, path0)
		}
	}
}

func Status(gitDir string) { // git status   对比三个区域的差异
	work := loadWorkStatus(gitDir) // 工作区文件
	index := OpenIndex(gitDir)     // 暂存区文件
	head := loadStash(gitDir)      // 获取git仓库信息,head指向的工作树
	// work -> index 待暂存的   可能出现修改，删除， index使用文件名做唯一标记   不会有新增的，新增的就是没有被跟踪的(没有被add过，只要 add过就会一直存在index中)
	diffWorkAndIndex(work, index)
	// index -> head 待提交的   可能出现新增，删除，修改
	diffIndexAndHead(index, head)
}

func diffIndexAndHead(index *GitIndex, head map[string]*BlobObject) {

}

func diffWorkAndIndex(work map[string]struct{}, index *GitIndex) {

}

func loadStash(gitDir string) map[string]*BlobObject {
	sha := ParseRef(gitDir, HeadFile)
	treeObj := OpenObject(gitDir, sha).(*TreeObject)
	res := make(map[string]*BlobObject)
	flatTree(res, gitDir, treeObj)
	return res
}

func flatTree(res map[string]*BlobObject, gitDir string, treeObj *TreeObject) {
	for _, item := range treeObj.Items {
		if item.Mode == BlobMode {
			res[item.Path] = OpenObject(gitDir, item.Sha).(*BlobObject)
		} else if item.Mode == TreeMode {
			temp := OpenObject(gitDir, item.Sha).(*TreeObject)
			flatTree(res, gitDir, temp)
		}
	}
}

func loadWorkStatus(gitDir string) map[string]struct{} {
	res := make(map[string]struct{})
	index := strings.LastIndexByte(gitDir, '/')

	err := filepath.Walk(gitDir[:index], func(path string, info fs.FileInfo, err error) error {
		res[path] = struct{}{}
		return nil
	})
	HandleErr(err)
	return res
}

/*
git tag 列出所有tag 与 ShowRef 功能重叠，不再支持
git tag tagName sha1 添加tag
*/

func Tag(gitDir, tagName, sha string) { // git tag
	path0 := path.Join(gitDir, RefDir, tagName)
	WriteFile([]byte(sha), path0)
}

func Merge(gitDir, branch string) {
	// 把当前分支(HEAD)与目标分支 branch 进行合并，当前分支指向合并的新提交
	// 两路合并，只有一边有直接取，两边都有，一行行对比，有冲突报出来
	// 三路合并，先找两边的共同分支，若是只有一边对共同分支有修改就应用其修改，否则每个文件与基础文件进行 diff
	// 找不到公共节点就两路合并
	// 冲突样式
	//<<<<<<< HEAD
	//ascnaskndaksndask    当前文件内容
	//=======
	//sdasdasdasd          目标文件内容
	//>>>>>>> test
	// diff采用 by行 diff，实际就是把每行当做一个字段进行找出最长子序列，然后文件使用最长子序列获取差异
}
