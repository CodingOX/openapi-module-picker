package cmd

import (
	"flag"
	"fmt"
	"strings"

	"openapi-trim/openapi"
)

// RunDescribe 执行 describe 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，输出指定接口的详细契约 Markdown。
func RunDescribe(args []string) error {
	fs := flag.NewFlagSet("describe", flag.ContinueOnError)
	file := fs.String("file", "", "本地 OpenAPI JSON 文件路径")
	url := fs.String("url", "", "远程 OpenAPI 文档 URL")
	path := fs.String("path", "", "API 路径（必填）")
	method := fs.String("method", "get", "HTTP 方法（默认 get）")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *path == "" {
		return fmt.Errorf("--path 参数为必填项")
	}

	doc, err := parseDoc(*file, *url)
	if err != nil {
		return err
	}

	// 统一 method 为小写
	m := strings.ToLower(*method)

	s := openapi.NewSummarizer(doc)
	output := s.DescribeEndpoint(*path, m)
	fmt.Println(output)
	return nil
}
