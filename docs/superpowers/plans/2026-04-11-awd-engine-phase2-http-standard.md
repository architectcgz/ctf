# AWD Engine Phase 2 HTTP Standard Checker Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 AWD 轮次巡检从“路径探活”升级为真正的 `http_standard` checker，支持 `put_flag / get_flag / havoc` 语义，并把 `up / down / compromised` 判定与毕业设计要求对齐。

**Architecture:** 继续复用现有 `AWDRoundUpdater` 作为唯一调度入口，不新起独立 checker 服务。checker 逻辑只落在 `application/jobs` 内，通过读取 `contest_challenges.awd_checker_config`、Redis 轮次 flag 和当前轮次上下文，对每个 `team + challenge + instance` 执行 HTTP checker，并将聚合结果写回 `awd_team_services.check_result / service_status / sla_score / defense_score`。

**Tech Stack:** Go, Gin, GORM, Redis, SQLite test DB, net/http

---

### Task 1: 为 `http_standard` checker 补 RED 测试

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`

- [x] **Step 1: 写失败测试，覆盖 `put_flag + get_flag` 成功路径**

新增用例，构造 `http_standard` 配置：
- `put_flag.path = /api/flag`
- `get_flag.path = /api/flag`
- `get_flag.expected_substring = {{FLAG}}`

测试服务在 `PUT /api/flag` 时保存 body 中的 flag，在 `GET /api/flag` 时返回保存的 flag。

期望：
- `service_status = up`
- `sla_score`、`defense_score` 按 service 配置写入
- `check_result.checker_type = "http_standard"`
- `check_result.put_flag.healthy = true`
- `check_result.get_flag.healthy = true`

- [x] **Step 2: 跑单测确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater.*HTTPStandard' -count=1
```

预期：FAIL，失败原因是当前 updater 还不会执行 `put_flag / get_flag` 语义。

- [x] **Step 3: 写失败测试，覆盖 `get_flag` 取回错误 flag 的失陷路径**

新增用例，服务对 `put_flag` 返回成功，但 `get_flag` 固定返回错误 flag。

期望：
- `service_status = compromised`
- `sla_score = 0`
- `defense_score = 0`
- `check_result.status_reason = "flag_mismatch"`
- `check_result.get_flag.error_code = "flag_mismatch"`

- [x] **Step 4: 写失败测试，覆盖可选 `havoc` 的失败路径**

新增用例，`put_flag / get_flag` 成功，但 `havoc.path` 返回 500。

期望：
- `service_status = down`
- `check_result.havoc.healthy = false`
- `check_result.havoc.error_code = "unexpected_http_status"`

- [x] **Step 5: 重新跑单测确认全部 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater.*HTTPStandard' -count=1
```

预期：FAIL，失败点分别对应 `put_flag / get_flag / havoc` 语义未实现。

### Task 2: 实现 `http_standard` checker 配置解析与 HTTP action 执行器

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checks.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_config_support.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_http_checker_config.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_http_checker_result.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_service_check_outcome_support.go`

- [x] **Step 1: 为 checker 配置建最小结构**

新增配置结构，仅支持设计文档第一版要求的最小字段：
- `put_flag.method`
- `put_flag.path`
- `put_flag.headers`
- `put_flag.body_template`
- `put_flag.expected_status`
- `get_flag.method`
- `get_flag.path`
- `get_flag.headers`
- `get_flag.expected_status`
- `get_flag.expected_substring`
- `havoc.method`
- `havoc.path`
- `havoc.headers`
- `havoc.expected_status`

并提供：
- 配置解析函数
- 默认 method / expected_status 归一化函数
- 判断 `havoc` 是否启用的辅助函数

- [x] **Step 2: 补模板渲染能力**

实现最小模板替换，仅支持：
- `{{FLAG}}`
- `{{ROUND}}`
- `{{TEAM_ID}}`
- `{{CHALLENGE_ID}}`

模板渲染失败时返回明确错误码，避免静默降级。

- [x] **Step 3: 实现单个 HTTP action 执行器**

执行器负责：
- 拼接 `access_url + action.path`
- 构造 method / headers / body
- 发起 HTTP 请求
- 校验 `expected_status`
- 可选校验 `expected_substring`
- 输出结构化 action 结果

- [x] **Step 4: 跑 `jobs` 定向测试，确认编译仍然通过或失败点只剩行为层**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater.*HTTPStandard' -count=1
```

预期：仍然 FAIL，但失败原因应收敛到 updater 尚未把 round flag 和 checker runner 接起来。

### Task 3: 将轮次 flag 与 `http_standard` checker 接入 `AWDRoundUpdater`

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/awd_check_run.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_service_check_result.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_service_check_probe_result.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`

- [x] **Step 1: 先在 jobs 层补 round flag 读取辅助**

在 `AWDRoundUpdater` 内实现：
- 读取当前轮目标队伍 flag
- 在 `PreviousRoundGrace` 窗口内允许上一轮 flag 作为 `get_flag` 合法结果
- Redis 缺失时回退到 `flagSecret + BuildAWDRoundFlag`

复用命令层现有 accepted flags 规则，不引入新存储。

- [x] **Step 2: 接入 `http_standard` runner**

让 `checkTeamChallengeServices` 在 `checker_type = http_standard` 时：
- 解析 checker 配置
- 为每个 instance 依次执行 `put_flag -> get_flag -> havoc`
- 根据 action 结果判定单实例状态

其中：
- `put_flag` 失败 => `down`
- `get_flag` 返回错误 flag => `compromised`
- `havoc` 失败 => `down`

- [x] **Step 3: 聚合实例结果并生成标准化 `check_result`**

聚合结果至少包含：
- `checker_type`
- `status_reason`
- `put_flag`
- `get_flag`
- `havoc`
- `targets`

其中 `targets` 需要展示每个实例的 action 结果，便于后台面板后续直接消费。

- [x] **Step 4: 保留 `legacy_probe` 兼容路径**

保证：
- `checker_type = legacy_probe` 或空值时，仍走现有探活逻辑
- 当前已通过的 phase1 用例不回退

- [x] **Step 5: 跑 `jobs` 定向测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/application/jobs -run 'AWDRoundUpdater.*HTTPStandard' -count=1
go test ./internal/module/contest/application/jobs -run AWDRoundUpdater -count=1
```

预期：PASS，`http_standard` 语义测试与旧的 round updater 用例同时通过。

### Task 4: 做最小回归并提交

**Files:**
- Modify: `docs/superpowers/plans/2026-04-11-awd-engine-phase2-http-standard.md`

- [x] **Step 1: 跑 contest 模块最小充分回归**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration/code/backend
go test ./internal/module/contest/... -count=1
```

预期：PASS，说明命令、查询、基础设施与 jobs 层改动兼容。

- [x] **Step 2: 更新计划文档中的执行状态**

把本计划里已完成的步骤勾掉，保留未做项为空。

- [ ] **Step 3: 提交**

```bash
cd /home/azhi/workspace/projects/ctf/.worktrees/feat-awd-engine-migration
git add docs/superpowers/plans/2026-04-11-awd-engine-phase2-http-standard.md code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go code/backend/internal/module/contest/application/jobs/awd_checks.go code/backend/internal/module/contest/application/jobs/awd_checker_config_support.go code/backend/internal/module/contest/application/jobs/awd_http_checker_config.go code/backend/internal/module/contest/application/jobs/awd_http_checker_template.go code/backend/internal/module/contest/application/jobs/awd_http_checker_request.go code/backend/internal/module/contest/application/jobs/awd_http_checker_result.go code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go code/backend/internal/module/contest/application/jobs/awd_check_run.go code/backend/internal/module/contest/application/jobs/awd_service_check_result.go code/backend/internal/module/contest/application/jobs/awd_service_check_probe_result.go code/backend/internal/module/contest/application/jobs/awd_round_updater.go
git commit -m "feat(awd): 增加http标准checker"
```
