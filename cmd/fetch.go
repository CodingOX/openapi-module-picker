// Package cmd 提供 CLI 子命令的实现。
// fetch 子命令：从 OpenAPI 文档中提取所有 tag 并输出。
package cmd

import (
	"errors"
	"flag"
	"fmt"
)

// ErrNoTagsFound 表示文档中未找到任何 tag，调用方应据此以 exit code 2 退出。
var ErrNoTagsFound = errors.New("文档中未找到任何 tag")

// RunFetch 执行 fetch 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，提取所有 tag 并逐行输出到 stdout。
func RunFetch(args []string) error {
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	file := fs.String("file", "", "本地 OpenAPI JSON 文件路径")
	url := fs.String("url", "", "远程 OpenAPI 文档 URL")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := parseDoc(*file, *url)
	if err != nil {
		return err
	}

	tags := doc.GetAllTags()
	if len(tags) == 0 {
		return ErrNoTagsFound
	}

	for _, tag := range tags {
		fmt.Println(tag)
	}
	return nil
}
