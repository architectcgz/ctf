# challenge / contest / instance / awd DTO 内化下一批切片方案

## Objective

在不改变外部 API 契约的前提下，继续推进剩余高占比全局 DTO 的模块内化，覆盖：

- `challenge`
- `contest`
- `instance`
- `awd`（当前主要落在 `contest` 与 `instance`）

## Scope Snapshot

当前 `internal/module` 对 `internal/dto` 的 import 数量（2026-05-17）：

- `contest`: 75
- `challenge`: 53
- `instance`: 7

以上是下一批的主战场；`auth` 与 `notification` 在本轮独立切片收口。

## Slice Strategy

1. `contest` 先切 request DTO
   - owner：`contest/api/http`
   - 收口对象：create/update contest、team、registration、challenge 管理、AWD round/check/attack/traffic 请求体与查询参数
   - 目标：request mapper 与 handler 不再依赖 `internal/dto`

2. `contest` 再切 response/output DTO
   - owner：`contest/application/{commands,queries}` + `contest/api/http`
   - 收口对象：contest/team/challenge/registration/scoreboard/awd 响应
   - 目标：application 输出与 HTTP DTO 分层，移除 `contest.go`、`contest_awd_*.go` 对 `contest` 的遗留依赖

3. `challenge` request/query DTO 切片
   - owner：`challenge/api/http` + `challenge/application/{commands,queries}`
   - 收口对象：challenge list/query、tag/image、topology、writeup、challenge import 输入
   - 目标：query/command 入参不再使用全局 `dto.ChallengeQuery` 等类型

4. `challenge` response/topology DTO 切片
   - owner：`challenge/domain` + `challenge/api/http`
   - 收口对象：challenge/image/tag/topology/import 结果映射
   - 目标：去除 `challenge/domain/*mapper*` 对全局 DTO 的依赖

5. `instance` + `awd` 防守工作台 DTO 切片
   - owner：`instance/application` + `instance/contracts` + `instance/api/http`（如存在）
   - 收口对象：instance 基础响应、AWD defense directory/file/command 输入输出
   - 目标：`instance/contracts` 不再暴露全局 DTO 类型

## Order And Risk Control

- 执行顺序：`contest request` -> `contest response` -> `challenge request/query` -> `challenge response/topology` -> `instance/awd`
- 每个切片完成后立即跑模块级测试，避免跨模块累计回归
- `contest` 与 `challenge` 都含生成 mapper，必须每切片后同步 `go generate`

## Expected Follow-up Plans

按上面 5 个切片分别落单独 implementation plan，并在每个计划里明确：

- owner 边界
- 受影响 DTO 文件
- 验证命令
- 回退面
