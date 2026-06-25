// Package cmd 提供 CLI 子命令的实现。
// models 子命令：列出 OpenAPI 文档中所有数据模型。
package cmd

import (
	"flag"
	"fmt"

	"openapi-trim/openapi"
)

// RunModels 执行 models 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，以 Markdown 表格列出所有数据模型。
func RunModels(args []string) error {
	fs := flag.NewFlagSet("models", flag.ContinueOnError)
	file := fs.String("file", "", "本地 OpenAPI JSON 文件路径")
	url := fs.String("url", "", "远程 OpenAPI 文档 URL")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := parseDoc(*file, *url)
	if err != nil {
		return err
	}

	s := openapi.NewSummarizer(doc)
	schemas := s.GetAllSchemas()
	if len(schemas) == 0 {
		fmt.Println("文档中未找到数据模型。")
		return nil
	}

	fmt.Println("# 数据模型")
	fmt.Println()
	fmt.Println("| 模型 | 字段数 | 引用次数 | 描述 |")
	fmt.Println("|------|--------|----------|------|")
	for _, sc := range schemas {
		desc := sc.Description
		if desc == "" {
			desc = "-"
		}
		fmt.Printf("| %s | %d | %d | %s |\n",
			openapi.EscapeMarkdown(sc.Name), sc.FieldCount, sc.RefCount, openapi.EscapeMarkdown(desc))
	}
	return nil
}
