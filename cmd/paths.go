// Package cmd 提供 CLI 子命令的实现。
// paths 子命令：扁平列出所有 API 路径。
package cmd

import (
	"flag"
	"fmt"
	"strings"

	"openapi-trim/openapi"
)

// RunPaths 执行 paths 子命令。
// 从 --file 或 --url 读取 OpenAPI 文档，以对齐表格列出所有 API 端点。
// 支持 --method 可选过滤。
func RunPaths(args []string) error {
	fs := flag.NewFlagSet("paths", flag.ContinueOnError)
	file := fs.String("file", "", "本地 OpenAPI JSON 文件路径")
	url := fs.String("url", "", "远程 OpenAPI 文档 URL")
	method := fs.String("method", "", "按 HTTP 方法过滤（可选，如 get, post）")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := parseDoc(*file, *url)
	if err != nil {
		return err
	}

	endpoints := doc.GetAllEndpoints()

	// 可选的 method 过滤
	if *method != "" {
		targetMethod := strings.ToLower(strings.TrimSpace(*method))
		var filtered []openapi.EndpointInfo
		for _, ep := range endpoints {
			if ep.Method == targetMethod {
				filtered = append(filtered, ep)
			}
		}
		endpoints = filtered
	}

	if len(endpoints) == 0 {
		fmt.Println("无匹配的接口。")
		return nil
	}

	// 动态计算列宽
	methodWidth := 6
	pathWidth := 4
	summaryWidth := 7

	for _, ep := range endpoints {
		if len(strings.ToUpper(ep.Method)) > methodWidth {
			methodWidth = len(strings.ToUpper(ep.Method))
		}
		if len(ep.Path) > pathWidth {
			pathWidth = len(ep.Path)
		}
		if len(ep.Summary) > summaryWidth {
			summaryWidth = len(ep.Summary)
		}
	}

	// 对齐表格输出
	fmt.Printf("%-*s  %-*s  %-*s  %s\n",
		methodWidth, "METHOD", pathWidth, "PATH", summaryWidth, "SUMMARY", "TAGS")
	fmt.Printf("%s  %s  %s  %s\n",
		strings.Repeat("-", methodWidth), strings.Repeat("-", pathWidth),
		strings.Repeat("-", summaryWidth), strings.Repeat("-", 20))

	for _, ep := range endpoints {
		tags := strings.Join(ep.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		fmt.Printf("%-*s  %-*s  %-*s  %s\n",
			methodWidth, strings.ToUpper(ep.Method),
			pathWidth, ep.Path,
			summaryWidth, ep.Summary,
			tags)
	}
	return nil
}
