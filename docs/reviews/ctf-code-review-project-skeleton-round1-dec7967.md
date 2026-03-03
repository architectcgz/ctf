# ctf code Review（project-skeleton 第 1 轮）：initialize ctf project skeleton

## Review 信息

| 字段 | 内容 |
| --- | --- |
| 变更主题 | `project-skeleton` |
| 轮次 | `round1` |
| 审查范围 | `ctf` 仓库提交 `dec7967`（`chore: initialize ctf project skeleton`） |
| 变更概述 | 初始化后端（Go + Gin）、前端（Vite + Vue3 + Pinia）、本地依赖（Postgres/Redis）与设计/架构文档 |
| 审查基准 | `ctf/docs/architecture/**`、`ctf/docs/tasks/**`（本次提交内） |
| 审查日期 | 2026-03-03 |

## 变更摘要（按文件）

### 后端
- `ctf/cmd/api/main.go`：进程启动、配置加载、zap logger 初始化、优雅关停（SIGINT/SIGTERM + 10s timeout）。
- `ctf/internal/app/http_server.go`：Gin 引擎初始化，挂载 `RequestID/AccessLog/Recovery` 中间件，路由 `/healthz` 与 `/api/v1/healthz`。
- `ctf/internal/config/config.go`：基于 Viper 的配置读取（`configs/config.{env}.yaml` + `CTF_` 环境变量覆盖）。
- `ctf/internal/middleware/request_id.go`：优先复用 `X-Request-ID`，否则生成随机请求 ID，并回写响应头。
- `ctf/internal/middleware/access_log.go`：请求访问日志（method/path/ip/status/latency/request_id）。
- `ctf/internal/middleware/recovery.go`：panic 捕获并返回统一 500。
- `ctf/pkg/logger/logger.go`：zap logger 构建（json/console）。
- `ctf/pkg/response/response.go` + `ctf/pkg/apperror/error.go`：统一 Envelope 返回与应用错误码定义。

### 基础设施
- `ctf/docker-compose.dev.yml`：本地 Postgres/Redis 依赖编排，带 healthcheck。
- `ctf/configs/config.dev.yaml`、`ctf/configs/config.prod.yaml`：开发/生产默认配置。

### 前端
- `ctf/frontend/src/api/request.ts`：axios 实例、统一 Envelope 解析、401 刷新 token + 队列重放、toast 提示。
- `ctf/frontend/src/router/guards.ts`：鉴权/角色路由守卫、redirect 参数清洗、按需加载 profile。
- `ctf/frontend/src/stores/auth.ts`：token 与用户信息存储/恢复（localStorage）。
- `ctf/frontend/src/composables/useToast.ts`、`ctf/frontend/src/components/common/AppToast.vue`：自定义 Toast 系统。

### 文档与设计稿
- `ctf/docs/**`、`ctf/design-system/**`：大量架构/设计文档与二进制附件（pdf/ppt/docx 等）。

## 问题清单

### 🔴 高优先级

#### [H1] 误提交 Windows Zone.Identifier 文件
- 文件：`ctf/docs/**` 下多处 `*:Zone.Identifier`（例如 `ctf/docs/文献/Experimental Analysis of Security Attacks for Docker Container Communications.pdf:Zone.Identifier`）
- 问题描述：这类文件是 Windows 的“来自互联网”标记，属于环境噪音，不应进入版本库。
- 影响范围/风险：污染 git 历史、合规/安全扫描误报、跨平台协作易冲突。
- 修正建议：
  - 删除所有 `*:Zone.Identifier` 文件。
  - 在 `ctf/.gitignore` 增加 `*:Zone.Identifier`（或更严格地仅忽略 `docs/**` 下此类文件）。
- 可选方案：如果担心历史污染，合并前清理；若已合并则单独补一个“清理提交”集中删除，便于回滚。

#### [H2] 刷新 token 的重放流程存在 headers 类型假设，可能运行时崩溃
- 文件：`ctf/frontend/src/api/request.ts:45`、`ctf/frontend/src/api/request.ts:127`
- 问题描述：`pendingRequests` 保存的是 `AxiosRequestConfig`，重放时强转成 `InternalAxiosRequestConfig` 并假定 `headers` 有 `.set()`；当 `headers` 是普通对象时，`(config.headers as AxiosHeaders).set(...)` 可能触发 `set is not a function`。
- 影响范围/风险：401 刷新链路抖动时，队列中的请求重放可能整体失败，表现为随机登出/白屏或请求静默失败。
- 修正建议：
  - 在 `attachAuth` 内统一归一化 headers（例如用 `AxiosHeaders.from(config.headers)` 生成可 `.set()` 的对象）。
  - 或者让 `pendingRequests` 存 `InternalAxiosRequestConfig`（避免强转）。

#### [H3] refresh token 落 localStorage（前端）是高风险默认值
- 文件：`ctf/frontend/src/stores/auth.ts:20`
- 问题描述：`refreshToken` 存 localStorage，一旦发生 XSS，将导致长期会话凭证泄露。
- 影响范围/风险：账号被接管、越权、横向移动；对 CTF 平台这类高对抗环境尤为敏感。
- 修正建议：
  - 生产环境改为 HttpOnly + SameSite Cookie 存 refresh token（前端不落盘 refresh token）。
  - 至少在配置/文档中明确“仅 dev 可用”，并在 prod 构建/运行时禁用 localStorage refresh token。

### 🟡 中优先级

#### [M1] 业务错误码（`code != 0`）的 toast 行为不一致
- 文件：`ctf/frontend/src/api/request.ts:92`
- 问题描述：在响应成功分支里 `throw new ApiError(...)`，不保证走到统一的错误提示逻辑；网络错误与业务错误的用户提示可能不一致。
- 影响范围/风险：前端表现不稳定，排查困难（部分业务失败无明显提示）。
- 修正建议：统一错误处理策略（在同一个拦截器路径里做 toast，或把 toast 下沉到调用方并在 request 层只抛错）。

#### [M2] 刷新分支使用错误码 `11002`，但错误码映射表未定义
- 文件：`ctf/frontend/src/api/request.ts:114`、`ctf/frontend/src/utils/errorMap.ts:1`
- 问题描述：刷新逻辑硬编码 `code === 11002`，但 `errorMap` 没有该码的文案，且其他地方也未集中声明错误码常量。
- 影响范围/风险：后端错误码调整后前端可能进入错误分支（频繁登出或不刷新）。
- 修正建议：将鉴权相关错误码收敛为常量/枚举并在单处维护，避免散落魔法数字。

#### [M3] dev compose 将 Postgres/Redis 端口暴露到所有网卡
- 文件：`ctf/docker-compose.dev.yml:11`
- 问题描述：`"5432:5432"`、`"6379:6379"` 默认绑定到 0.0.0.0。
- 影响范围/风险：在共享网络或被端口扫描时暴露弱口令；与本机已有服务冲突概率更高。
- 修正建议：改为 `127.0.0.1:5432:5432`、`127.0.0.1:6379:6379`，并在 README 标注仅本机使用。

### 🟢 低优先级

#### [L1] RequestID 随机源失败时返回固定值，导致链路串联失效
- 文件：`ctf/internal/middleware/request_id.go:25`
- 问题描述：`rand.Read` 失败时返回常量 `req_fallback`，会导致多个请求共享同一 request_id。
- 影响范围/风险：极端情况下（系统熵源异常）日志关联严重退化。
- 修正建议：fallback 加入时间戳/递增计数，避免全局相同。

#### [L2] Redis key 直接拼接 username，建议规范化以减少 key-space 污染
- 文件：`ctf/internal/pkg/redis/keys.go:39`
- 问题描述：用户名包含 `:` 等特殊字符时，可读性差且可能造成 key-space 污染。
- 影响范围/风险：可观测性与排障体验变差；部分运维脚本可能误判 key 结构。
- 修正建议：对用户名做编码（url/base64）或使用哈希（例如 `sha256(username)`）。

## 运行/测试结果（命令 + 结果）
- `cd ctf && go test ./...`：失败（`go: command not found`，当前环境未安装 Go）
- `cd ctf/frontend && npm run typecheck`：通过（`vue-tsc --noEmit`）
- `cd ctf/frontend && npm run build`：通过（Vite 提示 chunk > 500k 的告警，且出现 `Generated an empty chunk: "echarts"` 的提示）

## 风险点与回滚点
- 风险点：`ctf/docs/**` 中存在大量二进制附件与 `_generated/unpacked` 目录，再加上 `*:Zone.Identifier`，会显著增加仓库体积、合规与跨平台协作风险。
- 回滚点：
  - 若该提交尚未合并主分支：建议在合并前剔除 `*:Zone.Identifier` 与非必要生成物，避免污染历史。
  - 若已合并：建议新增一个独立的清理提交，仅做“删除误入文件 + 更新 ignore”，便于单独回滚。

