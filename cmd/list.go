package cmd

import (
	"flag"
	"fmt"
	"strings"

	"openapi-trim/openapi"
)

// RunList 执行 list 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，按 --tags 筛选后输出接口清单 Markdown。
// --tags 为空时列出所有标签的接口。
func RunList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	file := fs.String("file", "", "本地 OpenAPI JSON 文件路径")
	url := fs.String("url", "", "远程 OpenAPI 文档 URL")
	tagsFlag := fs.String("tags", "", "过滤标签，逗号分隔（可选，为空则列出全部）")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := parseDoc(*file, *url)
	if err != nil {
		return err
	}

	var tags []string
	if *tagsFlag != "" {
		parts := strings.Split(*tagsFlag, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

	s := openapi.NewSummarizer(doc)
	output := s.ListByTags(tags)
	fmt.Println(output)
	return nil
}
