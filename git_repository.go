/*
@author: sk
@date: 2024/4/27
*/
package main

type GitConfig struct {
}

func NewGitConfig(path string) *GitConfig {
	return &GitConfig{}
}

/*
常用文件
.git/branches/  分支目录
.git/objects/ 对象存储
.git/refs/ 参考存储 存在子文件夹  heads tags
.git/HEAD 对 HEAD的引用
.git/config 配置文件  不经常变动 暂时不支持配置文件
.git/description 描述文件，很少使用
*/

type GitRepository struct {
	WorkDir string     // 管理目录 		xxx/
	GitDir  string     // git文件存储目录 xxx/.git
	Config  *GitConfig // git配置 		xxx/.git/config
}

func NewGitRepository(workDir string) *GitRepository {
	gitDir := workDir + "/.git"
	config := NewGitConfig(gitDir + "/config")
	return &GitRepository{WorkDir: workDir, GitDir: gitDir, Config: config}
}
