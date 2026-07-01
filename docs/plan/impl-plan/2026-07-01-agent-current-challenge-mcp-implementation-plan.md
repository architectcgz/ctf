<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# agent-current-challenge-mcp Implementation Plan

**Goal:** 在 CTF 后端新增正式 MCP 入口，让外部 agent 通过用户签发的 Bearer token 读取当前用户正在做的题目信息。

**Architecture:** 按 `docs/design/ctf-tutor-agent-and-mcp.md` 的边界，MCP 只作为后端协议适配层，不承担 tutor-agent 编排、教学策略或数据库直查。外部 agent 先提示用户登录 CTF 平台，再由用户调用 `POST /api/v1/auth/mcp-token` 签发 MCP token；`POST /mcp` 使用 `Authorization: Bearer <token>` 解析用户上下文。当前题目信息由 `instance` 查询服务的用户实例列表和 `challenge` 查询服务的已发布题目详情组合得到。

**Tech Stack:** Go、Gin、JSON-RPC 2.0、MCP tools/list + tools/call、`go test`

---

## Task Metadata

- Task Slug: `2026-07-01-agent-current-challenge-mcp`
- Started At: `2026-07-01T03:27:16Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-07-01-agent-current-challenge-mcp`
- Branch: `task/2026-07-01-agent-current-challenge-mcp`
- Plan Type: `slice`

## Plan Status

- Status: `implemented`
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 新增一个正式 `/mcp` HTTP 入口，支持 MCP client 使用用户级 Bearer token 调用 `get_current_challenge` 读取当前用户最近的活动实例和对应题目详情；同时提供登录态下签发 MCP token 的 REST 接口。
- Non-Goals: 不新增 `tutor-agent/` Python 服务；不开放提交 Flag、解锁 Hint、下载附件等副作用工具；不新增数据库表；不把 MCP 暴露到未鉴权公网入口。

## Problem Statement

- Current behavior / structure: 后端已有 `GET /api/v1/instances` 和 `GET /api/v1/challenges/:id`，但 agent 不能通过统一 MCP 工具读取“当前用户正在做哪道题”。
- Target behavior / structure: MCP endpoint 暴露 `tools/list` 和 `tools/call`，`get_current_challenge` 返回 `has_current_challenge`、实例信息和题目详情；未提供有效 Bearer token 时返回 MCP JSON-RPC auth-required 错误并携带登录/token URL。
- Why this task is needed now: 伴学 agent 的第一步需要稳定读取当前题目事实，避免 prompt 或 agent 自己猜测题目上下文。

## Inputs

- Source docs: `AGENTS.md`、`code/backend/tests/README.md`、`docs/文档规范.md`、`docs/design/ctf-tutor-agent-and-mcp.md`
- Related architecture/contracts: `docs/architecture/backend/modules/practice.md`、`docs/architecture/backend/modules/challenge.md`、`docs/architecture/backend/04-api-design.md`
- Related prior work: 当前设计草案确认 MCP 放在后端协议适配层，并复用 application service 与用户上下文。

## Task Classification

- Classification: `非琐碎任务`
- Why: 新增后端协议入口、路由接线和测试，触达受保护实现面，需要 task gate、计划和最小验证。

## Files

- Create:
  - `code/backend/internal/interfaces/mcp/handler.go`
  - `code/backend/internal/interfaces/mcp/handler_test.go`
- Modify:
  - `code/backend/internal/app/router.go`
  - `code/backend/internal/module/auth/contracts/token_service.go`
  - `code/backend/internal/module/auth/contracts/public_errors.go`
  - `code/backend/internal/module/auth/infrastructure/token_service.go`
  - `code/backend/internal/module/auth/api/http/handler.go`
  - `code/backend/internal/module/auth/api/http/auth_types.go`
  - `code/backend/internal/app/composition/instance_module.go`
  - `code/backend/internal/app/composition/challenge_module.go`
  - `docs/contracts/api-contract-v1.md`
  - `docs/contracts/openapi-v1/paths/auth.yaml`
  - `docs/contracts/openapi-v1/components/schemas/auth.yaml`
  - `docs/contracts/openapi-v1.yaml`
  - `docs/architecture/backend/modules/auth.md`
  - `docs/plan/impl-plan/2026-07-01-agent-current-challenge-mcp-implementation-plan.md`
- Review:
  - `code/backend/internal/module/instance/application/queries/instance_service.go`
  - `code/backend/internal/module/challenge/application/queries/challenge_service.go`
  - `code/backend/internal/app/router_user_practice_routes.go`
- Test:
  - `go test ./internal/interfaces/mcp`
  - `go test ./internal/app ./internal/module/instance/application/queries ./internal/module/challenge/application/queries`

## 复用与 Owner 决策

- Existing patterns searched: 使用 `rg` 搜索 `GetUserInstances`、`GetPublishedChallenge`、router 接线、`authctx.MustCurrentUser` 和实例状态常量。
- Reuse / extend / split / create-new decision: 新建 `internal/interfaces/mcp` 作为协议适配层；复用现有 `InstanceQueryService` 和 challenge query service，不新增 repository。
- Owner boundary: `instance` owner 提供当前用户实例事实，`challenge` owner 提供题目详情和可见性判断，MCP handler 只做 JSON-RPC/MCP 协议适配与结果组装。
- Auth owner boundary: `auth` owner 负责 MCP token 签发、Redis 存储、session version 失效判断和 REST 签发接口；MCP handler 只解析 Bearer token 并把有效用户上下文用于只读工具调用。
- Why this is the narrowest safe surface: 不改现有 HTTP 契约、不改数据库、不引入 tutor-agent，只新增一个只读工具和路由。

## Intake Analysis Gate

- Relevant superpowers analysis pass: `brainstorming`
- Why this pass fits: 需要从设计草案、现有模块和用户目标中收敛最小可实现切片。
- grill-with-docs findings: practice/challenge 文档确认训练实例和题目事实分别由 `instance` / `challenge` owner 承接；MCP 不应绕过这些 owner。
- Plan adjustments after challenge: 第一版只暴露 `get_current_challenge`，不提前实现文档中列出的完整 MCP 工具矩阵。

## Execution Slices

### Slice 1: MCP 当前题目读取

- Goal: 完成 `/mcp` endpoint 和 `get_current_challenge` 工具。
- Dependencies: 现有登录态中间件、`InstanceQueryService.GetUserInstances`、`ChallengeService.GetPublishedChallenge`。
- Files:
  - Create: `code/backend/internal/interfaces/mcp/handler.go`、`handler_test.go`
  - Modify: `code/backend/internal/app/router.go`、composition module 暴露 query service
  - Review: 实例状态选择规则和题目详情权限规则
  - Test: MCP handler 单元测试和相关 app 编译测试
- 步骤：
  - [x] 步骤 1：写 `tools/list`、`tools/call` 和无活动实例的失败测试。
  - [x] 步骤 2：运行 `go test ./internal/interfaces/mcp`，确认测试因未实现而失败。
  - [x] 步骤 3：实现 MCP handler、当前题目选择规则和只读 DTO。
  - [x] 步骤 4：接入 `/mcp` protected route，并从 composition module 暴露所需 query service。
  - [x] 步骤 5：运行最小相关 Go 测试。
- Validation: `go test ./internal/interfaces/mcp`、`go test ./internal/app ./internal/module/instance/application/queries ./internal/module/challenge/application/queries`
- Review focus: MCP 是否绕过 application service、是否错误暴露未鉴权数据、是否把“当前题目”定义成不可维护的猜测。
- Done criteria: MCP 工具可被登录用户调用，返回当前活动实例和题目详情；无活动实例时返回空结果；相关测试通过。

### Slice 2: 正式外部 agent MCP 认证

- Goal: 让外部 agent 不依赖浏览器 Cookie，而是使用用户签发的 MCP Bearer token 调用 `/mcp`。
- Dependencies: `authcontracts.TokenService`、Redis-backed token service、现有登录态保护路由。
- Files:
  - Modify: `auth/contracts`、`auth/infrastructure/token_service.go`、`auth/api/http`、`internal/app/router.go`、`interfaces/mcp/handler.go`
  - Test: MCP handler auth error / Bearer token 单元测试、auth token service 测试、auth HTTP 签发接口测试
- 步骤：
  - [x] 步骤 1：补 `tools/call` Bearer token 和 auth-required JSON-RPC 错误红测。
  - [x] 步骤 2：补 token service 签发/解析/过期红测，以及 `/api/v1/auth/mcp-token` HTTP 红测。
  - [x] 步骤 3：扩展 `TokenService` 契约和 Redis 实现，MCP token 绑定用户 session version。
  - [x] 步骤 4：新增登录态保护的 `POST /api/v1/auth/mcp-token`，并把 `/mcp` 改为 Bearer token 解析。
  - [x] 步骤 5：更新 OpenAPI、API contract 和架构事实文档。
- Validation: `go test ./internal/interfaces/mcp -count=1`、`go test ./internal/module/auth/infrastructure -count=1`、`go test ./internal/module/auth/api/http -run TestHTTP_IssueMCPTokenRequiresSessionAndReturnsBearerToken -count=1`
- Review focus: MCP token 是否在主解析链路随用户撤销失效；未认证返回是否是 MCP JSON-RPC 而不是 REST envelope；外部 agent 是否不再依赖 Cookie。
- Done criteria: 外部 agent 可用 Bearer token 调用 `/mcp`；无 token 或无效 token 能收到可提示用户登录的 MCP 错误；相关契约文档同步。

### Slice 3: MCP 防爬增强

- Goal: 降低 MCP token 泄露或自动化轮询时的爬取风险。
- Dependencies: `config.AuthConfig`、`config.RateLimitConfig`、Redis rate checker、`auditlog.Recorder`。
- Files:
  - Modify: `config`、`interfaces/mcp`、`auth/api/http`、`auth/infrastructure/token_service.go`、`internal/app/router.go`
  - Test: MCP handler 限流/审计测试、auth HTTP 签发审计测试、token service 独立 TTL 测试、config validation 测试
- 步骤：
  - [x] 步骤 1：补 MCP token 独立 TTL 红测。
  - [x] 步骤 2：补 `/mcp` 用户级限流和成功工具调用审计红测。
  - [x] 步骤 3：补 `POST /api/v1/auth/mcp-token` 签发审计红测。
  - [x] 步骤 4：新增 `auth.mcp_token_ttl` 和 `rate_limit.mcp` 配置，并接入 Redis token TTL / MCP rate checker。
  - [x] 步骤 5：更新架构和 API 契约说明。
- Validation: `go test ./internal/interfaces/mcp -count=1`、`go test ./internal/module/auth/... ./internal/config ./internal/middleware -count=1`、`go test ./internal/app -run '^$' -count=1`
- Review focus: 超限是否仍返回 MCP JSON-RPC；审计是否避免记录 token 明文；独立 TTL 是否不会继续复用网页登录 session TTL。
- Done criteria: MCP token 默认 6h 过期；`/mcp` 默认每用户 120/min；签发和成功工具调用均可审计。

## Impact And Compatibility

- API / DTO: 新增 `/mcp` JSON-RPC endpoint；新增 `POST /api/v1/auth/mcp-token`，返回 `{token, expires_at}`，用于外部 agent 的 `Authorization: Bearer <token>`。
- Data / migration: `none`
- State / cache / queue / event: Redis 新增 `auth.SessionKeyPrefix + ":mcp:" + token` 临时键，TTL 使用 `auth.mcp_token_ttl`，默认 6h；解析时校验用户 session version。`/mcp` 工具调用新增 `rate_limit.mcp` 用户级限流，默认 120/min。审计日志新增 `mcp_token/create` 和 `mcp_tool/read` 记录。
- Runtime / config: `none`
- Frontend route / state / UX: `none`
- Docs / contracts: 更新 `docs/contracts/api-contract-v1.md`、OpenAPI v1 auth path/schema、`docs/architecture/backend/04-api-design.md` 和 auth 模块文档。

## Plan Review / Architecture Fit

- Target owner boundary: MCP handler 是协议适配 owner，实例和题目事实仍由各自模块 owner 提供。
- Reuse points / landing zones: `InstanceQueryService.GetUserInstances`、`ChallengeService.GetPublishedChallenge`、Gin protected route 和 `authctx.CurrentUser`。
- Known structural debt touched: `none`
- How this plan avoids behavior-only convergence: 不直接拼 SQL 或复制现有 handler 逻辑，而是在 composition 层注入稳定 query service。
- Hidden second-redesign risk: 未来如果完整 MCP 工具矩阵落地，可能需要把工具注册拆成多个文件；当前单工具不提前抽象。
- Decision after review: 计划符合前置设计草案和当前模块边界，可以开始 TDD 实现。

## Documentation Owner

- Current fact sources to read: `docs/design/ctf-tutor-agent-and-mcp.md`、`docs/architecture/backend/modules/practice.md`、`docs/architecture/backend/modules/challenge.md`
- Fact sources to update after implementation: 暂不更新架构事实源；本次是第一版窄切片，设计草案仍处于 Draft。
- Plan-only notes that must not become architecture source: 当前题目选择规则属于本次实现方案，正式扩展 MCP 工具矩阵前再回收进事实源。
- Archive condition: 代码、测试和 review 完成后通过 `archive_task_artifacts.sh` 归档。

## Validation

- 计划验证范围: handler 单元测试、app 路由编译、instance/challenge 查询依赖编译。
- 命名 / 契约检查范围: 搜索 `/mcp`、`get_current_challenge`、`tools/list`、`tools/call` 的实现和测试覆盖。
- 完成判定: 相关 `go test` 通过，并能说明独立 review gate 是否满足。

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/interfaces/mcp`
  - `cd code/backend && go test ./internal/app ./internal/module/instance/application/queries ./internal/module/challenge/application/queries`
- Integration commands: `cd code/backend && go test ./internal/app -run Test -count=1`
- Manual checks: 确认 `/mcp` 不依赖 Cookie auth middleware；未登录/无 Bearer token 的 `tools/call` 返回 JSON-RPC auth-required data。
- Commands intentionally skipped and why: 全量 `go test ./...` 可能耗时较长且依赖运行时环境；本切片优先跑最小相关范围。

## Validation Evidence

- Command: `cd code/backend && go test ./internal/interfaces/mcp`
  - Result: 初次 RED 失败，报 `undefined: NewHandler`、`undefined: Deps`、`undefined: Handler`。
  - Notes: 证明测试先于生产实现生效。
- Command: `cd code/backend && go test ./internal/interfaces/mcp`
  - Result: PASS
  - Notes: 证明 `tools/list`、`tools/call get_current_challenge`、`notifications/initialized` 和无活动实例返回空结果的行为通过。
- Command: `cd code/backend && go test ./internal/app ./internal/module/instance/application/queries ./internal/module/challenge/application/queries`
  - Result: FAIL
  - Notes: `internal/module/challenge/application/queries` 通过，`internal/module/instance/application/queries` 无测试文件；`internal/app` 存在既有失败，包括 sqlite fixture 缺 `contest_realtime_outbox` / `platform_event_outbox` 表，以及未触碰文件 `assessment_module.go`、`router_user_self_routes.go` 的文本 guardrail 不一致。本次 MCP 相关接线需另用编译和 workflow gate 验证。
- Command: `cd code/backend && go test ./internal/app -run '^$' -count=1 && go test ./internal/module/challenge/runtime -count=1`
  - Result: PASS
  - Notes: 证明 `internal/app` 路由接线可编译，challenge runtime 暴露 query service 后仍通过测试。
- Command: `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - Result: PASS
  - Notes: 通过 startup gate、后端 module/shared architecture、前端架构和测试架构轻量门禁。
- Command: `python3 scripts/check-docs-consistency.py`
  - Result: FAIL
  - Notes: 失败点为既有 `harness/prompts/AGENTS.md:13` 引用缺失 `/home/azhi/.agents/skills/code-reviewer/frontend/architecture-review.md`，与本次新增 `/mcp` 文档内容无关。
- Command: `bash scripts/run-workflow-stage.sh completion-full`
  - Result: 初次 FAIL，补充 `docs/architecture/backend/04-api-design.md` 后 PASS
  - Notes: 证明 API surface 文档登记、后端架构、前端架构和测试架构完成阶段门禁通过。
- Command: `cd code/backend && go test ./internal/app -run '^$' -count=1 && go test ./internal/module/challenge/runtime -count=1 && cd ../.. && bash scripts/run-workflow-stage.sh completion-full`
  - Result: PASS
  - Notes: 补充 `notifications/initialized` 兼容处理后，最终相关编译、challenge runtime 和 completion-full 仍通过。
- Command: `cd code/backend && go test ./internal/interfaces/mcp -count=1`
  - Result: PASS
  - Notes: 正式版 Bearer token 调用、未认证 auth-required JSON-RPC 错误、`tools/list`、`tools/call`、`notifications/initialized` 通过。
- Command: `cd code/backend && go test ./internal/module/auth/... ./internal/middleware -count=1`
  - Result: PASS
  - Notes: MCP token 签发/解析、过期拒绝、HTTP 签发接口和 TokenService 接口替身均通过。
- Command: `cd code/backend && go test ./internal/app -run '^$' -count=1 && go test ./internal/module/challenge/runtime -count=1`
  - Result: PASS
  - Notes: 证明 `/api/v1/auth/mcp-token` 与 `/mcp` 路由接线可编译，challenge runtime 暴露 query service 后仍通过测试。
- Command: `python3 tools/sync_openapi_from_contract.py`
  - Result: PASS
  - Notes: 已从 `docs/contracts/openapi-v1/index.yaml` 同步生成 `docs/contracts/openapi-v1.yaml`。
- Command: `python3 scripts/check-docs-consistency.py`
  - Result: FAIL
  - Notes: 失败点仍为既有 `harness/prompts/AGENTS.md:13` 引用缺失 `/home/azhi/.agents/skills/code-reviewer/frontend/architecture-review.md`，与本次 MCP/auth 文档改动无关。
- Command: `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - Result: PASS
  - Notes: 通过 startup gate、后端 module/shared architecture、前端架构和测试架构轻量门禁。
- Command: `bash scripts/run-workflow-stage.sh completion-full`
  - Result: PASS
  - Notes: API surface contract、后端架构、前端架构和测试架构完成门禁通过。
- Command: `cd code/backend && go test ./internal/interfaces/mcp -count=1`
  - Result: PASS
  - Notes: MCP auth-required、Bearer token 工具调用、用户级限流和成功工具调用审计通过。
- Command: `cd code/backend && go test ./internal/module/auth/... ./internal/config ./internal/middleware -count=1`
  - Result: PASS
  - Notes: MCP token 独立 TTL、签发审计、配置默认/校验和既有 auth/middleware 测试通过。
- Command: `cd code/backend && go test ./internal/app -run '^$' -count=1 && go test ./internal/module/challenge/runtime -count=1`
  - Result: PASS
  - Notes: 防爬增强后的 router 接线和 challenge runtime 编译验证通过。
- Command: `python3 scripts/check-docs-consistency.py`
  - Result: FAIL
  - Notes: 仍失败在既有 `harness/prompts/AGENTS.md:13` 引用缺失 `/home/azhi/.agents/skills/code-reviewer/frontend/architecture-review.md`，与本轮 MCP 防爬文档改动无关。
- Command: `bash scripts/run-workflow-stage.sh pre-commit-quick`
  - Result: PASS
  - Notes: 通过 startup gate、后端 module/shared architecture、前端架构和测试架构轻量门禁。
- Command: `bash scripts/run-workflow-stage.sh completion-full`
  - Result: PASS
  - Notes: API surface contract、后端架构、前端架构和测试架构完成门禁通过。
- Command: `git diff --check`
  - Result: PASS
  - Notes: 未发现空白或格式差异问题。

## Independent Review Handoff

- Review target: 新增 MCP endpoint、协议 handler、当前题目选择规则和路由接线。
- Validation evidence summary: 待实现后填写。
- Architecture / contract inputs: `docs/design/ctf-tutor-agent-and-mcp.md`、`docs/architecture/backend/modules/practice.md`、`docs/architecture/backend/modules/challenge.md`
- Known risks / review focus: MCP 不应绕过权限；不应把停止/过期实例误判为当前正在做的题。
- Project-local checks to consider: `bash scripts/run-workflow-stage.sh completion-full`

## Rollback / Recovery

- Safe revert boundary: 回滚新增 `internal/interfaces/mcp` 包和 `/mcp` 路由接线即可。
- Data / config / runtime recovery notes: `none`
- Irreversible operations: `none`

## Residual Risks

- Risk: “正在做的题目”目前按活动实例推断；纯静态题如果用户只打开详情但没有实例，后端没有现成状态可读。
- Why acceptable: 用户请求是第一版 MCP；现有事实源中活动实例是唯一稳定、权限受控、可复用的当前做题信号。
- Follow-up owner, if any: 后续如需覆盖静态题详情页停留状态，应由前端或后端新增显式 current workspace 状态 owner。
