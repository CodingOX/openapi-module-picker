---
name: openapi-picker
description: 从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出为文件，或对已有文档进行 LLM 友好的结构化查询（概述、按标签列接口、查看单个接口契约、列数据模型、查看模型详情、列路径）。当用户提到 OpenAPI、Swagger、API 文档、接口文档、API 模块筛选、按 tag 提取接口、查看 API 概览、查看接口详情、查看数据模型、列出路径、或需要获取/过滤/查询远程或本地 API 规范时，务必使用此 skill——即使用户没有明确说"用 skill"或"用 CLI"，只要涉及 OpenAPI 文档的处理和查询，都应该触发。
---

# openapi-trim

从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出为文件。也支持对本地或远程的 OpenAPI 文档进行 LLM 友好的结构化查询。CLI 工具 `openapi-trim` 提供十个子命令：

| 命令 | 用途 |
|------|------|
| `fetch` | 提取文档中的所有标签 |
| `filter` | 按标签过滤并导出 JSON 文件 |
| `summary` | 生成 API 全局概览 Markdown（LLM 友好） |
| `list` | 按标签列出接口清单 Markdown |
| `describe` | 查看单个接口的完整契约 Markdown |
| `models` | 列出所有数据模型（字段数、引用次数、描述） |
| `describe-model` | 查看单个数据模型的完整字段定义（展开 `$ref`） |
| `paths` | 列出所有 API 路径（支持 `--method` 过滤） |
| `serve` | 启动 Web 服务 |

除 `filter` 仅支持 `--url` 外，其余所有命令均支持 `--file` 和 `--url` 双源。

## 核心原则

- **输出即消费** —— fetch/models/paths 输出极简（标签列表/模型表格/路径表格），大模型无需读取原始 JSON；summary/list/describe/describe-model 直接输出结构化 Markdown，大模型可直接理解
- **先探查再操作** —— 涉及筛选导出时永远先 `fetch` 让用户看到有哪些标签，再让用户选择后执行 `filter`；涉及查询时直接用 `--file` 定位已有文档
- **用户确认优先** —— 涉及网络请求和文件写入，必须让用户确认后再执行
- **业务标签优先，排除公共标签** —— SpringDoc 项目中常见 `通用权限`、`管理员权限`、`admin` 等"万能标签"，它们覆盖几乎全部接口，用作 filter 条件等于没过滤。始终选择具体业务标签（如 `exam`、`user`、`face-profile`），将公共标签排除在 filter 条件之外

## 工作流程

### 阶段 0：信息收集（引导流程）

目标：自动获取 URL，减少手动输入。如果用户已明确给出 URL，跳过此阶段。

**0.1 自动发现 URL（优先尝试）**

尝试从项目中自动构造 URL：

1. **查找服务端口** —— 扫描以下文件提取端口号：
   - `application.yml` / `application.properties` → `server.port`
   - `application-{profile}.yml` / `application-{profile}.properties` → `server.port`
   - `.env` / `.env.local` → `PORT`、`SERVER_PORT`
   - `src/main/resources/` 下的配置文件
   - 其他常见配置文件：`config.yaml`、`settings.toml`、`config.json`

2. **查找已记录约定** —— 检查 `CLAUDE.md`、`AGENTS.md`、`README.md` 中是否记录了项目的 OpenAPI 文档地址

3. **拼接候选 URL** —— 将提取到的端口与常见路径组合（以下仅为尝试性拼接，不是必然存在的路径）：
   - `http://localhost:{port}/v3/api-docs`
   - `http://localhost:{port}/v2/api-docs`
   - `http://localhost:{port}/openapi.json`
   - `http://localhost:{port}/swagger.json`
   - `http://localhost:{port}/api/openapi.json`

4. **快速验证** —— `curl -s -o /dev/null -w "%{http_code}" --connect-timeout 3 <url>`：
   - 返回 200 → 可用，进入步骤 5
   - 返回其他或超时 → 不可用，尝试下一个候选
   - 全部失败 → 进入 0.2

5. **提议而非假设** —— 验证通过后向用户提议：
   > 「检测到服务端口 `<port>`，尝试 `<url>` 返回了 200。使用这个地址获取 OpenAPI 文档？」

**0.2 检查已有上下文**

自动发现失败时，从上下文中查找：
- 当前打开的文件中是否有 OpenAPI/Swagger URL？
- 项目的 `CLAUDE.md`、`AGENTS.md`、`README.md` 中是否有记录？
- `api-docs/` 目录中是否有历史导出文件（可从中反推 URL）？

**0.3 主动推荐**

- 发现了可用 URL → 主动询问是否使用
- `api-docs/` 中有历史导出 → 建议：「上次导出了 `<tags>`，这次基于相同 URL 更新？」
- 用户打开的文件中包含 API 路径 → 在后续 fetch 结果中关联对应标签

**0.4 兜底引导**

以上全部失败时，引导用户提供：
> 「请提供 OpenAPI 文档的 URL 地址，我将帮你探查可用的标签。」
>
> 提示：可以在项目 `CLAUDE.md`、`AGENTS.md` 中记录 API 文档地址，下次就能自动发现。

### 阶段 1：探查标签

```bash
# 从远程 URL
openapi-trim fetch --url <openapi-json-url>

# 从本地已有文件（如重新过滤时提取之前使用的标签）
openapi-trim fetch --file <local-json-path>
```

stdout 输出排序后的标签列表（每行一个）。失败时参考「错误处理」章节。

### 阶段 2：用户选择

将标签列表展示给用户，必须等用户选择后再继续——切勿自行决定。

> 以下是从文档中获取到的标签（共 N 个）：
>
>  1. user
>  2. order
>  3. product
>  ...
>
> 请选择需要导出的标签（输入标签名或序号，多个用逗号分隔）：

**交互要点：**
- 列出标签时标注序号，方便用户按序号选择
- 标签超过 15 个时，可主动询问缩小范围（如「是否有特定业务方向？比如用户、订单相关？」）
- 确认选择后，输出文件名建议格式：`tag1-tag2.json`
- **关键：帮助用户识别并排除公共标签**。`fetch` 结果中如果出现 `通用权限`、`管理员权限`、`admin`、`common`、`public` 等标签，务必提醒用户：这些标签通常覆盖几乎所有接口，用于 filter 会失去筛选意义。引导用户只选择业务标签

### 阶段 3：过滤导出

```bash
openapi-trim filter --url <url> --tags tag1,tag2,... --output <file-path>
```

- `--url` 与 fetch 阶段保持相同
- `--tags` 逗号分隔，值为用户选中的标签名（非序号）
- `--output` 必须显式指定，CLI 不设默认值
- 执行前用 `mkdir -p` 确保输出目录存在

### 阶段 4：完成确认

`filter` 在 stdout 输出生成文件的绝对路径（一行），告知用户即可。

### 阶段 5：结构化查询（LLM 友好输出）

当用户已有本地 OpenAPI JSON 文件（或可通过 URL 访问远程文档），且需要大模型理解 API 内容时，使用以下查询命令。输出为结构化 Markdown，大模型无需读取原始 JSON 即可理解。

**5.1 生成全局概览（summary）**

```bash
openapi-trim summary --file output.json
# 或从远程文档
openapi-trim summary --url https://api.example.com/openapi.json
```

输出 Markdown，包含：
- API 标题与版本号
- 业务模块（标签）列表及接口数
- 被引用最多的 Top 10 数据模型（含字段数和引用次数）

**5.2 按标签列出接口（list）**

```bash
# 查看指定标签下的接口清单
openapi-trim list --file output.json --tags exam,user

# 不加 --tags 则列出全部标签
openapi-trim list --file output.json
```

输出 Markdown 表格，按标签分组展示：
- HTTP 方法、路径、摘要
- 该模块涉及的数据模型列表

**5.3 查看单个接口的完整契约（describe）**

```bash
openapi-trim describe --file output.json --path /exams --method get
# 查看 POST 接口
openapi-trim describe --file output.json --path /exams --method post
```

输出 Markdown，包含：
- 接口摘要和说明
- 请求参数（按 path/query/header 分组，含类型、必填、说明）
- 请求体（展开 `$ref` 为扁平字段列表，含中文标题）
- 响应（按状态码分组，字段同样自动展开 `$ref`）

**核心特性：`$ref` 自动展开**

`describe` 命令会自动递归解析 JSON 指针引用，将 `$ref` 引用的嵌套 schema 拍平为带前缀的字段名。例如：

| 原始 schema | 展开后 |
|------------|--------|
| `result` → `$ref: ExamResultDetail` | `result.id`、`result.name`、`result.status` |
| `teacher` → `$ref: TeacherInfo` | `teacher.id`、`teacher.name` |

大模型无需自己追查 `$ref` 即可理解数据结构。

**5.4 列出所有数据模型（models）**

```bash
# 从本地文件
openapi-trim models --file output.json

# 从远程文档
openapi-trim models --url https://api.example.com/openapi.json
```

输出 Markdown 表格：
- 模型名称、字段数、被接口引用次数、模型描述
- 按引用次数降序排列，未引用的模型 RefCount 为 0（方便排查冗余模型）
- 同时支持 OpenAPI 3.x（`components/schemas`）和 Swagger 2.0（`definitions`）

**5.5 查看单个数据模型的字段定义（describe-model）**

```bash
openapi-trim describe-model --file output.json --name AbnormalDetail
```

输出 Markdown 表格，包含模型的所有字段（名称、类型、必填、说明），自动递归展开嵌套 `$ref`——能力与 `describe` 对请求体/响应的展开一致。

**5.6 列出所有 API 路径（paths）**

```bash
# 列出全部接口
openapi-trim paths --file output.json

# 只看 GET 方法
openapi-trim paths --file output.json --method get

# 从远程文档
openapi-trim paths --url https://api.example.com/openapi.json
```

输出对齐的纯文本表格（METHOD、PATH、SUMMARY、TAGS），`--method` 为可选过滤器。

## 主动推荐策略

基于上下文减少选择负担：

| 上下文信号 | 推荐行为 |
|-----------|----------|
| 当前打开的文件引用了特定 API 路径 | fetch 结果中标注：「你正在查看的代码可能关联了 `user` 标签」|
| `api-docs/` 中有历史导出文件 | 提示：「上次导出了 `user`、`order`，这次是否一样？」|
| 用户提到业务模块名 | 在标签列表中高亮匹配项 |
| 用户直接说出了标签名 | 跳过 fetch，直接 filter（仍需确认 URL）|
| `CLAUDE.md`、`AGENTS.md` 中记录了 API 文档地址 | 直接使用，无需询问 |
| 用户想知道 API 整体结构 | 推荐 summary 命令，展示全局概览 |
| 用户只想看某个模块的接口 | 推荐 list 命令配合 --tags 筛选 |
| 用户需要理解某个接口的请求/响应 | 推荐 describe 命令查看完整契约 |
| 用户需要了解有哪些数据模型 | 推荐 models 命令查看所有模型及统计 |
| 用户需要看某个数据模型的字段细节 | 推荐 describe-model 命令展开完整字段定义 |
| 用户想快速浏览全部可用接口 | 推荐 paths 命令查看扁平路径列表 |
| 本地已有 `.json` 文档 | 所有命令均可用 --file，无需 fetch |

## 错误处理

exit code: 0 成功，1 通用错误，2 文档中无 tag。

| 错误场景 | 处理方式 |
|----------|----------|
| 网络不可达 | 1. 确认 URL 可公网访问；2. 建议 `curl -I <url>` 验证；3. 是否需 VPN/内网 |
| JSON/YAML 解析失败 | 确认 URL 返回的是合法的 OpenAPI 规范文档 |
| 文档中无 tag | 报告用户：该文档未定义任何 tag，是否仍需导出全量文档？|
| 用户未提供 URL | 执行阶段 0 自动发现流程，全部失败则引导用户提供 |
| 用户未选择标签 | fetch 后必须等待用户确认，不可自行决定 |
| 输出目录不存在 | filter 前执行 `mkdir -p` 创建目标目录 |
| describe 的路径或方法不存在 | 命令会在输出首行提示「错误：路径不存在」或「路径下无此方法」，引导用户用 list 确认可用路径 |
| describe-model 的数据模型不存在 | 命令提示「数据模型不存在」，引导用户用 models 确认可用模型名 |
| summary/list/describe/models/describe-model/paths 未提供 --file 或 --url | 提示用户指定来源：「请指定 --file 本地路径或 --url 远程地址」|

## 输出路径规范

- 默认目录：`<项目根目录>/api-docs/`
- 命名格式：`<tag1>-<tag2>.json`（tag 按字母序排列，`-` 连接）
- 用户有命名偏好时优先用户要求

## 使用示例

**1. 用户提供了完整信息 —— 直接 filter**

```
用户：帮我从 https://api.example.com/openapi.json 导出 user 和 order
助手：直接执行 filter，生成 api-docs/order-user.json
```

**2. 用户只给了 URL —— fetch → 选择 → filter**

```
用户：看看 https://api.example.com/openapi.json 有哪些模块
助手：执行 fetch，展示标签列表，等用户选择后 filter
```

**3. 用户没有任何信息 —— 自动发现**

```
用户：帮我导出 API 文档
助手：
  1. 扫描项目配置文件找到端口 8382
  2. curl 验证 http://localhost:8382/v3/api-docs 返回 200
  3. 提议：「检测到端口 8382 的 API 文档可用，是否使用？」
  4. 用户确认后执行 fetch → 选择 → filter
```

**4. 自动发现失败 —— 兜底引导**

```
用户：帮我筛选 API 文档
助手：（扫描无果）「请提供 OpenAPI 文档的 URL 地址。提示：可以在 CLAUDE.md 中记录地址方便下次自动发现。」
```

**5. 公共标签陷阱 —— 反模式与正确做法**

这是一个高频错误：

```
❌ 错误做法：
  助手：fetch 到标签：通用权限、管理员权限、exam、user、face-profile ...
  用户：选 通用权限、管理员权限
  结果：生成的文件包含几乎所有接口，等于没过滤
  原因：SpringDoc 项目常用 通用权限/管理员权限 作为接口的权限标记，
        几乎所有接口都带这两个标签，按其过滤失去了筛选意义

✅ 正确做法：
  fetch 后发现 通用权限、管理员权限 等公共标签，主动提醒：
  「⚠️ 通用权限 和 管理员权限 是公共标签，覆盖了几乎所有接口。
   用它们过滤等于没过滤。业务标签是：exam、user、face-profile、...」
  引导用户只选择业务标签，生成 4 个聚焦的业务文件
```

**5. 已有本地文件 —— 生成概览供 LLM 理解**

```
用户：帮我看看这个 API 的整体结构
助手：检测到当前文件 output.json 为 OpenAPI 文档。执行 summary：
  openapi-trim summary --file output.json
  输出：API 名称、版本、标签列表（含接口数）、Top 10 数据模型
  大模型据此理解 API 的全貌
```

**6. 已有本地文件 —— 只看某个业务模块**

```
用户：我想看 exam 模块有哪些接口
助手：执行 list：
  openapi-trim list --file output.json --tags exam
  输出：exam 标签下的所有接口（方法、路径、摘要）和涉及的数据模型
  大模型据此定位到具体接口
```

**7. 已有本地文件 —— 查看单个接口细节**

```
用户：帮我解释 POST /exams 这个接口要怎么调
助手：执行 describe：
  openapi-trim describe --file output.json --path /exams --method post
  输出：请求体字段（类型、必填、中文说明）、响应结构（含展开后的嵌套字段）
  大模型据此给出精确的调用方式，无需阅读原始 JSON
```

**8. 已有本地文件 —— 查看全部数据模型**

```
用户：这个 API 定义了哪些数据结构？
助手：执行 models：
  openapi-trim models --file output.json
  输出：所有模型名称、字段数、引用次数、描述
  发现 UnusedModel 引用次数为 0——可能是已废弃的模型
```

**9. 已有本地文件 —— 查看单个模型字段**

```
用户：AbnormalDetail 这个模型有哪些字段？
助手：执行 describe-model：
  openapi-trim describe-model --file output.json --name AbnormalDetail
  输出：模型所有字段（类型、必填、说明），嵌套 $ref 已自动展开
```

**10. 已有本地文件 —— 浏览全部接口路径**

```
用户：告诉我这个 API 有哪些接口
助手：执行 paths：
  openapi-trim paths --file output.json
  输出：对齐表格（METHOD、PATH、SUMMARY、TAGS），按路径排序
```

**11. 已有本地文件 —— 直接提取标签重新过滤**

```
用户：这个 output.json 我之前过滤过了，现在想重新生成，但忘了当时选了哪些标签
助手：直接从已有文件提取标签即可：
  openapi-trim fetch --file output.json
  输出：orders、users
  然后引导用户重新选择并执行 filter
```
