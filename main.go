// Package main 提供 OpenAPI Module Picker 的 CLI 入口。
// 支持三个子命令：fetch（探查标签）、filter（过滤导出）、serve（Web 服务）。
package main

import (
	"errors"
	"fmt"
	"os"
	"openapi-module-picker/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "fetch":
		err = cmd.RunFetch(os.Args[2:])
	case "filter":
		err = cmd.RunFilter(os.Args[2:])
	case "serve":
		err = cmd.RunServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		// ErrNoTagsFound 对应 exit code 2（文档中无标签）
		if errors.Is(err, cmd.ErrNoTagsFound) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// printUsage 输出帮助信息到 stderr。
func printUsage() {
	fmt.Fprintf(os.Stderr, "用法: openapi-module-picker <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "命令:\n")
	fmt.Fprintf(os.Stderr, "  fetch   从远程 OpenAPI 文档提取所有标签\n")
	fmt.Fprintf(os.Stderr, "  filter  按标签过滤 OpenAPI 文档并导出文件\n")
	fmt.Fprintf(os.Stderr, "  serve   启动 Web 服务\n\n")
	fmt.Fprintf(os.Stderr, "示例:\n")
	fmt.Fprintf(os.Stderr, "  openapi-module-picker fetch --url https://api.example.com/openapi.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-module-picker filter --url <url> --tags user,order --output api-docs/result.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-module-picker serve --port 8326\n")
}
