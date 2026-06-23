// Package cmd 提供 CLI 子命令的实现。
package cmd

import (
	"fmt"

	"openapi-trim/openapi"
)

// parseDoc 根据 --file 或 --url 参数解析 OpenAPI 文档。
// file 和 url 二选一，同时为空或同时非空均返回错误。
func parseDoc(file, url string) (*openapi.OpenAPIDocument, error) {
	if file != "" && url != "" {
		return nil, fmt.Errorf("--file 和 --url 不能同时使用")
	}
	if file != "" {
		return openapi.ParseOpenAPIFromFile(file)
	}
	if url != "" {
		return openapi.ParseOpenAPI(url)
	}
	return nil, fmt.Errorf("必须指定 --file 或 --url 参数")
}
