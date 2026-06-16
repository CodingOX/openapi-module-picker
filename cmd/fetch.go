// Package cmd 提供 CLI 子命令的实现。
// fetch 子命令：从远程 OpenAPI 文档中提取所有 tag 并输出。
package cmd

import (
	"errors"
	"flag"
	"fmt"

	"openapi-trim/openapi"
)

// ErrNoTagsFound 表示文档中未找到任何 tag，调用方应据此以 exit code 2 退出。
var ErrNoTagsFound = errors.New("文档中未找到任何 tag")

// RunFetch 执行 fetch 子命令。
// 从 --url 指定的地址拉取 OpenAPI 文档，提取所有 tag 并逐行输出到 stdout。
// 返回 error 而非直接 os.Exit，便于测试。
func RunFetch(args []string) error {
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	url := fs.String("url", "", "OpenAPI 文档 URL（必填）")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *url == "" {
		return fmt.Errorf("--url 参数为必填项")
	}

	doc, err := openapi.ParseOpenAPI(*url)
	if err != nil {
		return fmt.Errorf("解析 OpenAPI 文档失败: %w", err)
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
