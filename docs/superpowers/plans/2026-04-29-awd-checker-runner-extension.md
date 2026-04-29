# AWD Checker Runner Extension Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 AWD 增加 `tcp_standard`、安全 sandbox runner 和 `script_checker`，让平台能验证 TCP 服务和教师提交的复杂 checker 脚本。

**Architecture:** 保留 `AWDRoundUpdater` 作为轮次调度入口，但把 checker 执行抽成按类型分派的 runner 接口。`tcp_standard` 在后端进程内执行结构化 TCP 步骤；`script_checker` 必须提交给独立 sandbox runner，在受限容器中执行题目包私有脚本，并返回结构化结果。readiness 继续只读取保存态 validation，不直接执行 checker。

**Tech Stack:** Go, Gin, GORM, Redis, Docker Engine API, PostgreSQL, Vitest/Vue for admin UI, existing AWD checker preview and readiness contracts

---

## Requirement Summary

当前结论：

- 比赛前正式验证走平台 runner，不走本地 `check.py`。
- `http_standard` 只能覆盖 HTTP 请求/响应契约。
- TCP / Binary 服务需要 `tcp_standard` 或 `script_checker`。
- 教师 `check.py` 可以交给平台执行，但必须通过安全 sandbox runner。
- `script_checker` 不能在 API 进程、轮次调度进程或宿主机 shell 中直接执行。

本计划不做：

- 不把 checker 流量当成选手攻击流量。
- 不把 `script_checker` 输出写入 `awd_traffic_events`。
- 不允许赛时联网安装 checker 依赖。
- 不让学生下载 checker 私有脚本。

## Acceptance Criteria

- `AWDCheckerType` 支持 `tcp_standard` 和 `script_checker`。
- `tcp_standard` 支持 TCP connect、send、send_template、expect_contains、expect_regex、flag roundtrip 和 havoc。
- `script_checker` 只能由 sandbox runner 执行。
- sandbox runner 使用一次性容器或等价隔离环境，禁止 privileged、Docker socket、平台配置挂载和非目标网络访问。
- runner 必须限制超时、CPU、内存、进程数、文件描述符、输出大小和展开文件数量。
- runner 对超时、OOM、非 0 退出、无效 JSON 输出、网络拒绝给出稳定错误码。
- checker preview 能为 `tcp_standard` / `script_checker` 生成 preview token。
- 创建或更新赛事服务消费 preview token 后，readiness 能变为 `passed`。
- 赛中轮次能执行新 checker 类型并写入 `awd_team_services`。
- 前端能展示并编辑 `tcp_standard` / `script_checker` 的最小配置。

## Planned File Map

### Contracts and DTO

- Modify: `code/backend/internal/model/awd.go`
  - 扩展 `AWDCheckerType` 常量。
- Modify: `code/backend/internal/dto/awd.go`
  - 允许 preview 请求携带新 checker 类型。
- Modify: `code/backend/internal/dto/contest_awd_service.go`
  - 更新 binding allowlist。
- Modify: `code/frontend/src/api/contracts.ts`
  - 扩展 `AWDCheckerType`。
- Modify: `docs/contracts/api-contract-v1.md`
  - 更新 checker 类型和配置说明。
- Modify: `docs/contracts/openapi-v1.yaml`
  - 同步 OpenAPI enum。

### Backend Checker Runner

- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_config_support.go`
  - 统一 checker type 分派和配置校验入口。
- Create: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_config.go`
  - 解析 `tcp_standard` 配置。
- Create: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
  - 执行 TCP connect/send/expect。
- Create: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner_test.go`
  - 本地 TCP fixture 测试。
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
  - preview 分派到 TCP runner。
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
  - 赛中轮次分派到 TCP runner。

### Sandbox Runner

- Create: `code/backend/internal/module/contest/ports/checker_runner.go`
  - 定义 `CheckerRunner` port。
- Create: `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
  - Docker sandbox runner 实现。
- Create: `code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go`
  - 验证超时、输出限制、网络 allowlist、挂载限制。
- Create: `code/backend/internal/module/contest/application/jobs/awd_script_checker_config.go`
  - 解析 `script_checker` 配置。
- Create: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
  - 将 `script_checker` 请求转换为 runner job。
- Modify: `code/backend/internal/app/composition/contest_module.go`
  - 注入 runner。
- Modify: `code/backend/internal/config/config.go`
  - 新增 runner 配置。
- Modify: `code/backend/configs/config.dev.yaml`
  - 本地默认 runner 限制。
- Modify: `code/backend/configs/config.yaml`
  - 默认安全限制。

### Import and Storage

- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
  - 允许 `tcp_standard` / `script_checker`，并校验 script entry 在包内。
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - 保存 script checker 私有文件摘要，不公开为附件。
- Create or Modify: `code/backend/internal/model/checker_artifact.go`
  - 如现有模型不够，新增 checker artifact 元数据。
- Create: `code/backend/migrations/000002_create_checker_artifacts.up.sql`
  - 视当前 baseline 序号调整。
- Create: `code/backend/migrations/000002_create_checker_artifacts.down.sql`
  - 回滚 checker artifact 表。

### Frontend

- Modify: `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
  - 增加 `tcp_standard` / `script_checker` draft、build、parse。
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
  - 增加新 checker 类型编辑 UI。
- Modify: `code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue`
  - 增加 `tcp_standard` / `script_checker` 示例。
- Modify: `code/frontend/src/api/admin.ts`
  - 类型收敛，不使用 `any`。

### Tests

- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser_test.go`
  - 覆盖新 checker 类型导入。
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
  - 覆盖 preview token 和 readiness。
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
  - 覆盖赛中轮次执行。
- Modify: `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
  - 覆盖 UI 编辑和 preview payload。
- Modify: `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
  - 覆盖文档示例显示。

## Task 1: 扩展 checker type 合同

**Files:**
- Modify: `code/backend/internal/model/awd.go`
- Modify: `code/backend/internal/dto/awd.go`
- Modify: `code/backend/internal/dto/contest_awd_service.go`
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `docs/contracts/api-contract-v1.md`
- Modify: `docs/contracts/openapi-v1.yaml`

- [ ] **Step 1: 写后端失败测试**

在 `code/backend/internal/module/challenge/domain/awd_package_parser_test.go` 增加：

```go
func TestBuildParsedAWDChallengePackageAcceptsTCPStandardChecker(t *testing.T) {
    manifest := validAWDManifestForTest()
    manifest.Extensions.AWD.Checker.Type = "tcp_standard"
    manifest.Extensions.AWD.Checker.Config = map[string]any{
        "connect": map[string]any{"host": "{{TARGET_HOST}}", "port": "{{TARGET_PORT}}"},
        "steps": []any{
            map[string]any{"send": "PING\\n", "expect_contains": "PONG"},
        },
    }

    parsed, err := buildParsedAWDChallengePackage(writeAWDPackageForTest(t), &manifest)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if parsed.CheckerType != "tcp_standard" {
        t.Fatalf("unexpected checker type: %s", parsed.CheckerType)
    }
}
```

- [ ] **Step 2: 运行失败测试**

Run:

```bash
cd code/backend && go test ./internal/module/challenge/domain -run 'AWDChallengePackage.*TCPStandard' -count=1
```

Expected: FAIL，原因是 checker type 不支持 `tcp_standard`。

- [ ] **Step 3: 最小实现 enum 扩展**

扩展后端 model、DTO binding、前端 union type、OpenAPI enum。

- [ ] **Step 4: 验证通过**

Run:

```bash
cd code/backend && go test ./internal/module/challenge/domain -run 'AWDChallengePackage.*TCPStandard' -count=1
cd code/frontend && npm run typecheck
```

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add code/backend/internal/model/awd.go code/backend/internal/dto/awd.go code/backend/internal/dto/contest_awd_service.go code/frontend/src/api/contracts.ts docs/contracts/api-contract-v1.md docs/contracts/openapi-v1.yaml code/backend/internal/module/challenge/domain/awd_package_parser_test.go
git commit -m "feat(awd): 扩展checker类型契约"
```

## Task 2: 实现 `tcp_standard` runner

**Files:**
- Create: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_config.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_config_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`

- [ ] **Step 1: 写 TCP fixture 失败测试**

测试启动本地 TCP listener，协议：

- 收到 `PING\n` 返回 `PONG\n`
- 收到 `SET_FLAG <flag>\n` 返回 `OK\n`
- 收到 `GET_FLAG\n` 返回上一次 flag

断言 runner 输出 `service_status=up`，`status_reason=healthy`。

- [ ] **Step 2: 运行失败测试**

Run:

```bash
cd code/backend && go test ./internal/module/contest/application/jobs -run TCPStandard -count=1
```

Expected: FAIL，runner 不存在。

- [ ] **Step 3: 实现配置解析**

支持字段：

- `timeout_ms`
- `connect.host`
- `connect.port`
- `steps[].send`
- `steps[].send_template`
- `steps[].expect_contains`
- `steps[].expect_regex`
- `havoc[]`

- [ ] **Step 4: 实现 runner**

使用 `net.Dialer` + deadline。每步设置 read/write deadline，总超时由 context 控制。

- [ ] **Step 5: 接入 preview 和 round updater**

在 checker type 分派中新增 `tcp_standard`。

- [ ] **Step 6: 验证**

Run:

```bash
cd code/backend && go test ./internal/module/contest/application/jobs -run 'TCPStandard|AWDRoundUpdater' -count=1
```

Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add code/backend/internal/module/contest/application/jobs
git commit -m "feat(awd): 增加tcp标准checker"
```

## Task 3: 建立 sandbox runner port 与 Docker 实现

**Files:**
- Create: `code/backend/internal/module/contest/ports/checker_runner.go`
- Create: `code/backend/internal/module/contest/infrastructure/docker_checker_runner.go`
- Create: `code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go`
- Modify: `code/backend/internal/config/config.go`
- Modify: `code/backend/configs/config.yaml`
- Modify: `code/backend/configs/config.dev.yaml`

- [ ] **Step 1: 写 sandbox 安全失败测试**

测试项：

- 脚本超时会返回 `checker_timeout`。
- stdout 超限会返回 `checker_output_limit_exceeded`。
- 尝试访问非 allowlist 地址失败。
- 容器配置中没有 privileged、没有 Docker socket mount、rootfs readonly。

- [ ] **Step 2: 运行失败测试**

Run:

```bash
cd code/backend && go test ./internal/module/contest/infrastructure -run DockerCheckerRunner -count=1
```

Expected: FAIL，runner 不存在。

- [ ] **Step 3: 定义 port**

`CheckerRunner` 接口输入：

- runtime
- entry file
- args
- env
- mounted files
- target allowlist
- timeout
- resource limits

输出：

- status
- reason
- exit_code
- stdout/stderr summary
- duration
- resource limit flags

- [ ] **Step 4: 实现 Docker runner**

容器配置必须：

- `ReadonlyRootfs=true`
- `Privileged=false`
- `NetworkDisabled=false` 但使用受控网络或代理 allowlist
- 不挂载 Docker socket
- bind mount 只读 checker 文件
- 设置 memory、nano CPUs、pids limit
- 设置执行超时和清理逻辑

- [ ] **Step 5: 验证**

Run:

```bash
cd code/backend && go test ./internal/module/contest/infrastructure -run DockerCheckerRunner -count=1
```

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add code/backend/internal/module/contest/ports/checker_runner.go code/backend/internal/module/contest/infrastructure/docker_checker_runner.go code/backend/internal/module/contest/infrastructure/docker_checker_runner_test.go code/backend/internal/config/config.go code/backend/configs/config.yaml code/backend/configs/config.dev.yaml
git commit -m "feat(awd): 增加checker安全沙箱runner"
```

## Task 4: 接入 `script_checker`

**Files:**
- Create: `code/backend/internal/module/contest/application/jobs/awd_script_checker_config.go`
- Create: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- Modify: `code/backend/internal/app/composition/contest_module.go`
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`

- [ ] **Step 1: 写导入失败测试**

构造 `script_checker` 题目包：

```yaml
checker:
  type: script_checker
  config:
    runtime: python3
    entry: docker/check/check.py
    timeout_sec: 10
    args:
      - "{{TARGET_URL}}"
    output: json
```

断言 entry 必须在包内，且不会出现在公开附件列表。

- [ ] **Step 2: 写 preview 失败测试**

使用 fake runner，断言 preview 调用 runner，runner 返回 ok 后生成 preview token。

- [ ] **Step 3: 实现配置解析与导入校验**

校验：

- runtime 必须在 allowlist。
- entry 必须是包内相对路径。
- timeout 在配置范围内。
- args/env 只允许模板变量。

- [ ] **Step 4: 实现 preview 和轮次分派**

`script_checker` 不直接执行脚本，只创建 runner job。

- [ ] **Step 5: 验证**

Run:

```bash
cd code/backend && go test ./internal/module/challenge/domain ./internal/module/challenge/application/commands ./internal/module/contest/application/commands ./internal/module/contest/application/jobs -run 'ScriptChecker|CheckerPreview|AWDRoundUpdater' -count=1
```

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add code/backend/internal/module/challenge code/backend/internal/module/contest code/backend/internal/app/composition/contest_module.go
git commit -m "feat(awd): 接入脚本checker"
```

## Task 5: 前端支持新 checker 类型

**Files:**
- Modify: `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- Modify: `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/api/admin.ts`

- [ ] **Step 1: 写 UI 失败测试**

覆盖：

- 选择 `tcp_standard` 后出现 steps 编辑区。
- 选择 `script_checker` 后出现 runtime、entry、timeout、args/env 配置区。
- preview payload 包含正确 checker type 和 config。

- [ ] **Step 2: 运行失败测试**

Run:

```bash
cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts -t "tcp_standard|script_checker"
```

Expected: FAIL。

- [ ] **Step 3: 实现表单支持**

保持现有 UI 风格，不新增大型向导。第一版只提供最小字段。

- [ ] **Step 4: 验证**

Run:

```bash
cd code/frontend && npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts
cd code/frontend && npm run typecheck
```

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts code/frontend/src/api/contracts.ts code/frontend/src/api/admin.ts
git commit -m "feat(awd): 支持新checker配置"
```

## Task 6: 文档与端到端验证

**Files:**
- Modify: `docs/architecture/features/awd-checker-runner-extension-design.md`
- Modify: `docs/architecture/features/awd-http-standard-checker-design.md`
- Modify: `docs/contracts/challenge-pack-v1.md`
- Modify: `challenges/awd/README.md`
- Modify: `code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue`
- Modify: `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`

- [ ] **Step 1: 补题目包契约文档**

写明：

- `http_standard` 用于 HTTP。
- `tcp_standard` 用于简单 TCP。
- `script_checker` 必须走 sandbox runner。
- `docker/check/check.py` 只有声明为 `script_checker.entry` 时才会进入平台执行。

- [ ] **Step 2: 补前端示例**

题目包规范页增加 `tcp_standard` / `script_checker` 简短示例。

- [ ] **Step 3: 运行文档相关测试**

Run:

```bash
cd code/frontend && npm run test:run -- src/views/platform/__tests__/ChallengePackageFormat.test.ts
```

Expected: PASS。

- [ ] **Step 4: 运行后端关键测试**

Run:

```bash
cd code/backend && go test ./internal/module/challenge/... ./internal/module/contest/... -run 'AWD|Checker' -count=1
```

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add docs/architecture/features/awd-checker-runner-extension-design.md docs/architecture/features/awd-http-standard-checker-design.md docs/contracts/challenge-pack-v1.md challenges/awd/README.md code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts
git commit -m "docs(awd): 补充checker runner契约"
```

## Rollout Notes

- 第一阶段可以只在本地开发环境开启 `script_checker`，但 sandbox 安全边界不能关闭。
- `script_checker` 出错时默认判定为 checker failure，不自动降级到 `http_standard`。
- 如果 runner 不可用，`script_checker` preview 必须失败，readiness 不应通过。
- 生产环境开放前，需要额外做镜像供应链、runner 网络策略和审计日志留存评审。
