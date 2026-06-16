// Package cmd 提供 CLI 子命令的实现。
// filter 子命令：按标签过滤 OpenAPI 文档并将结果写入文件。
package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"openapi-trim/openapi"
)

// RunFilter 执行 filter 子命令。
// 从 --url 获取 OpenAPI 文档，按 --tags 选中的标签（逗号分隔）过滤路径，
// 将过滤后的 JSON 写入 --output 指定的文件，父目录不存在时自动创建。
// 最终将写入文件的绝对路径输出到 stdout。
// 返回 error 而非直接 os.Exit，便于测试。
func RunFilter(args []string) error {
	fs := flag.NewFlagSet("filter", flag.ContinueOnError)
	url := fs.String("url", "", "OpenAPI 文档 URL（必填）")
	tags := fs.String("tags", "", "过滤标签，逗号分隔（必填）")
	output := fs.String("output", "", "输出文件路径（必填）")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *url == "" {
		return fmt.Errorf("--url 参数为必填项")
	}
	if *tags == "" {
		return fmt.Errorf("--tags 参数为必填项")
	}
	if *output == "" {
		return fmt.Errorf("--output 参数为必填项")
	}

	// 解析逗号分隔的标签列表，去除每个标签的首尾空白
	parts := strings.Split(*tags, ",")
	tagList := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			tagList = append(tagList, trimmed)
		}
	}

	doc, err := openapi.ParseOpenAPI(*url)
	if err != nil {
		return fmt.Errorf("解析 OpenAPI 文档失败: %w", err)
	}

	// 检查文档是否包含标签，无标签时返回哨兵错误
	if len(doc.GetAllTags()) == 0 {
		return ErrNoTagsFound
	}

	data, err := doc.FilterByTags(tagList)
	if err != nil {
		return fmt.Errorf("过滤文档失败: %w", err)
	}

	// 解析为绝对路径
	absPath, err := filepath.Abs(*output)
	if err != nil {
		return fmt.Errorf("解析输出路径失败: %w", err)
	}

	// 确保父目录存在
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 仅输出绝对路径到 stdout
	fmt.Println(absPath)
	return nil
}
