// Package cmd 提供 CLI 子命令的实现。
// describe-model 子命令：查看单个数据模型的字段定义。
package cmd

import (
	"flag"
	"fmt"

	"openapi-trim/openapi"
)

// RunDescribeModel 执行 describe-model 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，展开指定数据模型的所有字段（含嵌套 $ref）。
func RunDescribeModel(args []string) error {
	fs := flag.NewFlagSet("describe-model", flag.ContinueOnError)
	file := fs.String("file", "", "本地 OpenAPI JSON 文件路径")
	url := fs.String("url", "", "远程 OpenAPI 文档 URL")
	name := fs.String("name", "", "数据模型名称（必填）")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *name == "" {
		return fmt.Errorf("--name 参数为必填项")
	}

	doc, err := parseDoc(*file, *url)
	if err != nil {
		return err
	}

	s := openapi.NewSummarizer(doc)
	fields, err := s.DescribeSchema(*name)
	if err != nil {
		return err
	}

	fmt.Printf("# 数据模型: %s\n\n", *name)
	fmt.Println("| 字段 | 类型 | 必填 | 说明 |")
	fmt.Println("|------|------|------|------|")
	for _, f := range fields {
		req := "否"
		if f.Required {
			req = "是"
		}
		desc := f.Title
		if f.Description != "" {
			desc = f.Description
		}
		if desc == "" {
			desc = "-"
		}
		fmt.Printf("| `%s` | %s | %s | %s |\n",
			openapi.EscapeMarkdown(f.Name), f.Type, req, openapi.EscapeMarkdown(desc))
	}
	return nil
}
