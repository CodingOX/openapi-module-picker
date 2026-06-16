# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 常用命令

```bash
# 构建
go build -o openapi-trim

# 运行（监听 :8080）
go run main.go

# 运行所有测试
go test ./... -v -count=1 -timeout 60s

# 运行单个包的测试
go test ./openapi/... -v
go test . -v

# 运行单个测试函数
go test ./openapi/ -run TestFilterByTags_SingleTag -v
```

## 架构概览

Go Web 应用，零外部依赖，纯标准库实现。从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出。

### 包结构

- **`main`** — HTTP 服务器入口，提供静态文件服务和两个 API 端点
- **`openapi`** — 核心业务逻辑，文档解析与过滤

### 数据流

```
前端 URL 输入 → POST /api/parse → openapi.ParseOpenAPI(url)
                                  → 解析 JSON，自动检测版本（3.0/2.0）
                                  → 提取所有 tags 返回前端

前端选择 tags → POST /api/filter → OpenAPIDocument.FilterByTags(tags)
                                  → 深拷贝文档，移除不匹配路径
                                  → 返回过滤后的 JSON 供下载
```

### 关键设计决策

1. **全局变量 `currentDocument`** 存储已解析文档，当前不支持并发多用户场景
2. **版本检测**：检查 JSON 中 `openapi` 字段是否以 `"3"` 开头，否则默认为 Swagger 2.0
3. **过滤粒度**：按 path 级别过滤，只要 path 下任一 operation 的 tag 命中即保留整个 path
4. **深拷贝策略**：`FilterByTags` 通过 marshal/unmarshal 实现深拷贝，确保原始文档不被修改

### 前端

纯 HTML/CSS/JS，无框架。CSS 变量驱动亮/暗主题切换，localStorage 持久化偏好。

## OpenAPI 文档地址约定（可选提示）

目标服务通常是 Spring Boot + SpringDoc 项目，API 文档地址模式为：
`http://localhost:{port}/v3/api-docs`

端口一般从 `src/main/resources/application.yml` 的 `server.port` 读取。
若无 SpringDoc，常见替代路径：`/v2/api-docs`（SpringFox）、`/openapi.json`（FastAPI）。

openapi-picker skill 会自动扫描配置文件来发现 URL 并尝试连接。
