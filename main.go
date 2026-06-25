// Package main 提供 openapi-trim 的 CLI 入口。
// 支持三个子命令：fetch（探查标签）、filter（过滤导出）、serve（Web 服务）。
package main

import (
	"errors"
	"fmt"
	"os"
	"openapi-trim/cmd"
)

// version 在构建时通过 -ldflags 注入，如：
//
//	go build -ldflags="-X main.version=1.0.0" -o openapi-trim
var version = "dev"

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		os.Exit(1)
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("openapi-trim version %s\n", version)
		return
	}

	var err error
	switch os.Args[1] {
	case "fetch":
		err = cmd.RunFetch(os.Args[2:])
	case "filter":
		err = cmd.RunFilter(os.Args[2:])
	case "summary":
		err = cmd.RunSummary(os.Args[2:])
	case "list":
		err = cmd.RunList(os.Args[2:])
	case "describe":
		err = cmd.RunDescribe(os.Args[2:])
	case "models":
		err = cmd.RunModels(os.Args[2:])
	case "describe-model":
		err = cmd.RunDescribeModel(os.Args[2:])
	case "paths":
		err = cmd.RunPaths(os.Args[2:])
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
	fmt.Fprintf(os.Stderr, "用法: openapi-trim <command> [options]\n\n")
	fmt.Fprintf(os.Stderr, "命令:\n")
	fmt.Fprintf(os.Stderr, "  fetch           提取 OpenAPI 文档中所有标签\n")
	fmt.Fprintf(os.Stderr, "  filter          按标签过滤 OpenAPI 文档并导出 JSON 文件\n")
	fmt.Fprintf(os.Stderr, "  summary         生成 API 全局概览 Markdown\n")
	fmt.Fprintf(os.Stderr, "  list            按标签列出 API 接口清单 Markdown\n")
	fmt.Fprintf(os.Stderr, "  describe        查看单个接口的完整契约 Markdown\n")
	fmt.Fprintf(os.Stderr, "  models          列出所有数据模型\n")
	fmt.Fprintf(os.Stderr, "  describe-model  查看单个数据模型的字段定义\n")
	fmt.Fprintf(os.Stderr, "  paths           列出所有 API 路径\n")
	fmt.Fprintf(os.Stderr, "  serve           启动 Web 服务\n\n")
	fmt.Fprintf(os.Stderr, "选项:\n")
	fmt.Fprintf(os.Stderr, "  -v, --version   显示版本号\n")
	fmt.Fprintf(os.Stderr, "  -h, --help      显示帮助信息\n\n")
	fmt.Fprintf(os.Stderr, "示例:\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim fetch --url https://api.example.com/openapi.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim fetch --file output.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim filter --url <url> --tags user,order --output api-docs/result.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim summary --file output.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim list --file output.json --tags exam,user\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim describe --file output.json --path /exams --method get\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim models --file output.json\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim describe-model --file output.json --name AbnormalDetail\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim paths --file output.json --method get\n")
	fmt.Fprintf(os.Stderr, "  openapi-trim serve --port 8326\n")
}
