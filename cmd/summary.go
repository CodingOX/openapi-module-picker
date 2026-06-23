package cmd

import (
	"flag"
	"fmt"

	"openapi-trim/openapi"
)

// RunSummary 执行 summary 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，生成 Markdown 概览输出到 stdout。
func RunSummary(args []string) error {
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
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
	output := s.GenerateSummary()
	fmt.Println(output)
	return nil
}
