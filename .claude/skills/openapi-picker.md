---
name: openapi-picker
description: 从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出为文件。当用户提到 OpenAPI、Swagger、API 文档、接口文档、API 模块筛选、按 tag 提取接口、或需要获取/过滤/拆分远程 API 规范时，务必使用此 skill——即使用户没有明确说"用 skill"或"用 CLI"，只要涉及从 OpenAPI 文档中提取子集，都应该触发。
---

# openapi-trim

从远程 OpenAPI/Swagger 文档中按标签筛选 API 模块并导出为文件。CLI 工具 `openapi-trim` 提供 `fetch`（探查标签）和 `filter`（过滤导出）两个子命令。

## 核心原则

- **不在对话中消费文档内容** —— CLI 输出极简（fetch 输出标签列表，filter 输出文件路径），大模型无需读取或展示原始 JSON
- **先探查再操作** —— 永远先 `fetch` 让用户看到有哪些标签，再让用户选择后执行 `filter`
- **用户确认优先** —— 涉及网络请求和文件写入，必须让用户确认后再执行

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

2. **查找已记录约定** —— 检查 `CLAUDE.md`、`README.md` 中是否记录了项目的 OpenAPI 文档地址

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
- 项目的 `CLAUDE.md`、`README.md` 中是否有记录？
- `api-docs/` 目录中是否有历史导出文件（可从中反推 URL）？

**0.3 主动推荐**

- 发现了可用 URL → 主动询问是否使用
- `api-docs/` 中有历史导出 → 建议：「上次导出了 `<tags>`，这次基于相同 URL 更新？」
- 用户打开的文件中包含 API 路径 → 在后续 fetch 结果中关联对应标签

**0.4 兜底引导**

以上全部失败时，引导用户提供：
> 「请提供 OpenAPI 文档的 URL 地址，我将帮你探查可用的标签。」
>
> 提示：可以在项目 `CLAUDE.md` 中记录 API 文档地址，下次就能自动发现。

### 阶段 1：探查标签

```bash
openapi-trim fetch --url <openapi-json-url>
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

## 主动推荐策略

基于上下文减少选择负担：

| 上下文信号 | 推荐行为 |
|-----------|----------|
| 当前打开的文件引用了特定 API 路径 | fetch 结果中标注：「你正在查看的代码可能关联了 `user` 标签」|
| `api-docs/` 中有历史导出文件 | 提示：「上次导出了 `user`、`order`，这次是否一样？」|
| 用户提到业务模块名 | 在标签列表中高亮匹配项 |
| 用户直接说出了标签名 | 跳过 fetch，直接 filter（仍需确认 URL）|
| `CLAUDE.md` 中记录了 API 文档地址 | 直接使用，无需询问 |

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
