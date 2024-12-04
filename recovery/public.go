package recovery

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func restoreFilesFromGitDir(gitDir string, outputDir string) error {
	// 打开本地Git仓库
	r, err := git.PlainOpen(gitDir)
	if err != nil {
		return err
	}

	// 当前引用获取
	ref, err := r.Head()
	if err != nil {
		return err
	}

	// 当前提交获取
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return err
	}

	// 遍历对象
	tree, err := commit.Tree()
	if err != nil {
		return err
	}

	// 开始恢复
	return tree.Files().ForEach(func(file *object.File) error {
		// 构建输出文件的完整路径
		outputFilePath := filepath.Join(outputDir, file.Name)

		// 获取文件内容
		content, err := file.Contents()
		if err != nil {
			return err
		}

		// 创建输出文件所在的目录（如果不存在）
		dir := filepath.Dir(outputFilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// 将文件内容写入到输出文件中
		return os.WriteFile(outputFilePath, []byte(content), 0644)
	})
}
