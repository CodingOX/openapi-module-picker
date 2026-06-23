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

Go CLI 应用，零外部依赖，纯标准库实现。从 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出，同时提供 LLM 友好的结构化查询能力。

### 包结构

- **`main`** — CLI 入口，分发子命令
- **`cmd`** — 子命令实现（fetch/filter/summary/list/describe/serve）
- **`openapi`** — 核心业务逻辑，文档解析、过滤、摘要生成

### 数据流

```
# 原有：JSON 导出
远程 URL → ParseOpenAPI(url) → FilterByTags(tags) → 过滤后的 JSON

# 新增：LLM 友好查询
本地文件/远程 URL → NewSummarizer(doc) → GenerateSummary()   → Markdown 概览
                                        → ListByTags(tags)    → Markdown 接口清单
                                        → DescribeEndpoint()  → Markdown 完整契约
```

### CLI 命令

| 命令 | 用途 | 典型用法 |
|------|------|---------|
| `fetch` | 列出所有标签 | `--url <url>` |
| `filter` | 按标签过滤导出 JSON | `--url <url> --tags a,b --output out.json` |
| `summary` | 全局概览 Markdown | `--file output.json` |
| `list` | 按标签列出接口 | `--file output.json --tags exam,user` |
| `describe` | 单个接口完整契约 | `--file output.json --path /exams --method get` |
| `serve` | 启动 Web 服务 | `--port 8326` |

所有查询类命令均支持 `--file` 和 `--url` 双源。

### 关键设计决策

1. **全局变量 `currentDocument`** 存储已解析文档，当前不支持并发多用户场景
2. **版本检测**：检查 JSON 中 `openapi` 字段是否以 `"3"` 开头，否则默认为 Swagger 2.0
3. **过滤粒度**：按 path 级别过滤，只要 path 下任一 operation 的 tag 命中即保留整个 path
4. **深拷贝策略**：`FilterByTags` 通过 marshal/unmarshal 实现深拷贝，确保原始文档不被修改
5. **$ref 解析**：`Summarizer` 内置 `$ref` 解析器和 `schemaCache`，支持递归展开嵌套引用（由 `visited` 集合防止循环）
6. **字段渲染**：`DescribeEndpoint` 自动展开属性中的 `$ref` 为带前缀的子字段（如 `teacher.id`、`teacher.name`），LLM 无需再追查引用

### 前端

纯 HTML/CSS/JS，无框架。CSS 变量驱动亮/暗主题切换，localStorage 持久化偏好。

## OpenAPI 文档地址约定（可选提示）

目标服务通常是 Spring Boot + SpringDoc 项目，API 文档地址模式为：
`http://localhost:{port}/v3/api-docs`

端口一般从 `src/main/resources/application.yml` 的 `server.port` 读取。
若无 SpringDoc，常见替代路径：`/v2/api-docs`（SpringFox）、`/openapi.json`（FastAPI）。

openapi-picker skill 会自动扫描配置文件来发现 URL 并尝试连接。
