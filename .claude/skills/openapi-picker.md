---
name: openapi-picker
description: 从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出。当用户需要获取、提取、筛选 API 模块或 OpenAPI 文档时使用。
---

# OpenAPI Module Picker

从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出为文件。

## 使用方式

### 探查标签

```bash
openapi-module-picker fetch --url <openapi-json-url>
```

输出排序后的标签列表（每行一个）。将标签展示给用户确认后再进行过滤。

### 过滤导出

```bash
openapi-module-picker filter --url <url> --tags tag1,tag2,... --output <file-path>
```

输出生成文件的绝对路径。`--output` 必须显式指定。

### 工作流程

1. 从用户消息中获取 OpenAPI 文档 URL
2. 执行 `fetch` 获取可用标签列表
3. 将标签展示给用户，请用户确认需要哪些标签
4. 执行 `filter` 生成过滤后的文档文件
5. 告知用户文件路径

## 错误处理

| 错误场景 | 处理方式 |
|----------|----------|
| 网络不可达 | 确认 URL 是否正确，检查网络连接 |
| JSON 解析失败 | 确认 URL 是否为有效的 OpenAPI 文档 |
| 无标签 | 报告文档中未找到任何标签 |
| 参数缺失 | 检查 `--url`、`--tags`、`--output` 是否都已提供 |

## 输出路径建议

将过滤后的文档放到项目中的 `api-docs/` 目录，文件名使用 `tag1-tag2.json` 格式。
