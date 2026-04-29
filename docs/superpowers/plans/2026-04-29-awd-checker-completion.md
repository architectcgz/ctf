# AWD Checker Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 AWD TCP / script checker 从题目包、前端编辑、平台预检、文档到运维审计的完整可用闭环。

**Architecture:** 保留当前 `legacy_probe` / `http_standard` / `tcp_standard` / `script_checker` 分派模型。先用一个真实 AWD TCP 题目包验证后端 runner，再补管理端配置编辑和题目包契约文档；随后把 `script_checker` 从单 entry artifact 扩展为受控多文件 artifact，并补齐审计、清理和端到端验证。

**Tech Stack:** Go, Gin, GORM, Redis, Docker Engine API, PostgreSQL, Vue 3, Vite, Vitest, Playwright/manual browser smoke, existing AWD checker preview and readiness contracts

---

## Requirement Summary

当前已经具备：

- `tcp_standard` 后端 runner：connect、`send` / `send_template` / `send_hex`、`expect_contains` / `expect_regex`、flag roundtrip、本地 TCP fixture。
- `script_checker` sandbox runner 接入：preview / 赛中分派、私有 entry artifact 存储和只读注入。
- 题目包导入允许 `tcp_standard` / `script_checker`。
- 后端枚举、DTO、OpenAPI 基础枚举已支持新 checker 类型。

仍需补齐：

- 一个真实可导入、可预检的 AWD TCP 题目包。
- 管理端 `tcp_standard` 配置编辑 UI。
- 题目包规范和前端格式指南里的 TCP / script checker 示例。
- 真实平台端到端上传、创建赛事服务、预检、readiness、赛中轮次验证。
- `script_checker` 多文件 artifact。
- checker 审计、flag 脱敏、artifact 清理和运维验证。

## Acceptance Criteria

- 仓库存在至少一个 `service_type: binary_tcp` 且 `checker.type: tcp_standard` 的 AWD 题目包。
- 该 AWD TCP 题目包可以通过题目包导入、赛事服务创建、checker preview，并让 readiness 变为 `passed`。
- 管理端可以选择并编辑 `tcp_standard`，保存后 preview payload 与后端契约一致。
- 题目包规范和前端格式指南包含 `tcp_standard`、`script_checker.entry`、`script_checker.files` 的明确示例和限制。
- `script_checker` 可以导入 entry 依赖的同包多文件，runner 只读注入全部声明文件，仍不公开给学生。
- checker 执行记录能定位 contest/service/team/round/checker 类型、artifact digest、耗时、错误码和截断输出；持久化前脱敏 flag。
- artifact 孤儿文件有可重复执行的清理路径。
- 所有新增能力有最小充分自动化测试，真实平台预检有可复现手工或脚本化验证记录。

## File Map

### AWD TCP 题目包

- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/challenge.yml`
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/statement.md`
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/Dockerfile`
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/service.py`
- Optional Create: `challenges/awd/ctf-1/awd-tcp-length-gate/README.md`

### Backend Verification

- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner_test.go`

### Frontend TCP Editor

- Modify: `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- Modify: `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
- Modify: `code/frontend/src/composables/useAwdCheckResultPresentation.ts`

### Package Contract Docs

- Modify: `docs/contracts/challenge-pack-v1.md`
- Modify: `docs/contracts/api-contract-v1.md`
- Modify: `docs/contracts/openapi-v1.yaml`
- Modify: `code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue`
- Modify: `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- Modify: `docs/architecture/features/awd-checker-runner-extension-design.md`

### Script Checker Multi-File Artifact

- Modify: `code/backend/internal/module/challenge/domain/package_manifest.go`
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_config.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner_test.go`

### Audit, Cleanup, and Ops

- Create: `code/backend/internal/module/contest/application/jobs/checker_audit.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Optional Create: `code/backend/cmd/cleanup-awd-checker-artifacts/main.go`
- Modify: `docs/operations/awd-checker-runner.md` or create it if absent.

---

## Task 1: 增加真实 AWD TCP 题目包

**Files:**
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/challenge.yml`
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/statement.md`
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/Dockerfile`
- Create: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/service.py`

- [ ] **Step 1: 写 TCP 服务**

`service.py` 实现一个长期运行 TCP 服务：

- `PING\n` -> `PONG\n`
- `SET_FLAG <flag>\n` -> 保存 flag，返回 `OK\n`
- `GET_FLAG\n` -> 返回当前 flag
- 漏洞逻辑保持在业务功能里，不影响 checker 的稳定读写。

- [ ] **Step 2: 写 AWD 题目包 manifest**

`challenge.yml` 使用：

```yaml
extensions:
  awd:
    service_type: binary_tcp
    checker:
      type: tcp_standard
      config:
        timeout_ms: 3000
        steps:
          - send: "PING\n"
            expect_contains: "PONG"
          - send_template: "SET_FLAG {{FLAG}}\n"
            expect_contains: "OK"
          - send: "GET_FLAG\n"
            expect_contains: "{{FLAG}}"
```

- [ ] **Step 3: 本地构建镜像**

Run:

```bash
cd challenges/awd/ctf-1/awd-tcp-length-gate
docker build -t ctf/awd-tcp-length-gate:latest docker
```

Expected: PASS。

- [ ] **Step 4: 本地 TCP smoke**

Run:

```bash
container_id="$(docker run -d --rm -p 18081:8080 ctf/awd-tcp-length-gate:latest)"
trap 'docker stop "$container_id" >/dev/null 2>&1 || true' EXIT
printf 'PING\n' | nc 127.0.0.1 18081
docker stop "$container_id"
trap - EXIT
```

Expected: 输出包含 `PONG`。测试后关闭容器。

- [ ] **Step 5: 提交**

```bash
git add challenges/awd/ctf-1/awd-tcp-length-gate
git commit -m "feat(awd): 增加tcp演示题目包"
```

## Task 2: 补 AWD TCP 题目包导入与 preview 测试

**Files:**
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner_test.go`

- [ ] **Step 1: 写 manifest 解析测试**

覆盖 `service_type=binary_tcp`、`checker.type=tcp_standard`、`runtime_config.service_port=8080`。

- [ ] **Step 2: 写导入测试**

构造 zip 包导入，断言：

- `CheckerType == tcp_standard`
- `ServiceType == binary_tcp`
- `checker_config` 保留 steps。

- [ ] **Step 3: 写 preview token / readiness 测试**

使用本地 TCP fixture 调用现有 preview flow，断言 preview 成功后保存赛事服务能消费 token，readiness 变为 `passed`。

- [ ] **Step 4: 运行测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/... ./internal/module/contest/... -run 'TCPStandard|AWDChallengeImport|PreviewChecker|Readiness' -count=1
```

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add code/backend/internal/module/challenge code/backend/internal/module/contest
git commit -m "test(awd): 覆盖tcp题目包导入预检"
```

## Task 3: 前端支持 `tcp_standard` 编辑

**Files:**
- Modify: `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- Modify: `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
- Modify: `code/frontend/src/composables/useAwdCheckResultPresentation.ts`
- Modify: `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`

- [ ] **Step 1: 写 UI 失败测试**

测试选择 `tcp_standard` 后出现：

- timeout 输入
- steps 列表
- `send` / `send_template` / `send_hex`
- `expect_contains` / `expect_regex`

保存 payload 包含 `checker_type=tcp_standard` 和合法 `checker_config.steps`。

- [ ] **Step 2: 扩展 draft/build/parse**

在 `awdCheckerConfigSupport.ts` 新增：

- `AWDTCPStandardDraft`
- `createTCPStandardDraft`
- `parseTCPStandardCheckerConfig`
- `buildTCPStandardCheckerConfig`

验证规则：

- `timeout_ms` 范围 `1-60000`
- 至少一个 step
- 每个 step 至少有一个 send 字段或一个 expect 字段
- `send_hex` 不能与 `send` / `send_template` 同时出现
- `expect_regex` 必须可编译。

- [ ] **Step 3: 更新 Dialog UI**

`AWDChallengeConfigDialog.vue` 的 checker type 下拉加入：

```html
<option value="tcp_standard">TCP 标准 Checker</option>
```

新增 TCP step 编辑区。不要把实现说明或架构说明渲染到 UI。

- [ ] **Step 4: 更新只读展示**

`AWDChallengeConfigPanel.vue` 和 `useAwdCheckResultPresentation.ts` 显示 TCP checker 摘要和错误码。

- [ ] **Step 5: 运行前端测试**

Run:

```bash
cd code/frontend
npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts -t 'tcp_standard|script_checker'
npm run typecheck
```

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add code/frontend/src/components/platform/contest code/frontend/src/composables code/frontend/src/components/platform/__tests__
git commit -m "feat(awd): 前端支持tcp checker配置"
```

## Task 4: 补题目包契约和格式指南

**Files:**
- Modify: `docs/contracts/challenge-pack-v1.md`
- Modify: `docs/contracts/api-contract-v1.md`
- Modify: `docs/contracts/openapi-v1.yaml`
- Modify: `code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue`
- Modify: `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
- Modify: `docs/architecture/features/awd-checker-runner-extension-design.md`

- [ ] **Step 1: 补 `tcp_standard` 文档**

说明适用场景、字段、模板变量、错误边界，并加入完整 YAML 示例。

- [ ] **Step 2: 补 `script_checker.files` 文档**

先写目标契约：

```yaml
checker:
  type: script_checker
  config:
    runtime: python3
    entry: docker/check/check.py
    files:
      - docker/check/check.py
      - docker/check/protocol.py
    timeout_sec: 10
```

约束：

- `entry` 必须在 `files` 内；未声明 `files` 时默认只包含 `entry`。
- `files` 只允许包内相对文件路径。
- 不支持目录通配符。
- 不进入选手附件。

- [ ] **Step 3: 更新前端格式指南**

格式指南展示 TCP 和 script 多文件 checker 示例。

- [ ] **Step 4: 运行文档相关测试**

Run:

```bash
cd code/frontend
npm run test:run -- src/views/platform/__tests__/ChallengePackageFormat.test.ts
```

Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add docs/contracts docs/architecture/features/awd-checker-runner-extension-design.md code/frontend/src/components/platform/challenge code/frontend/src/views/platform/__tests__
git commit -m "docs(awd): 补充tcp和脚本checker题目包契约"
```

## Task 5: 实现 `script_checker` 多文件 artifact

**Files:**
- Modify: `code/backend/internal/module/challenge/domain/package_manifest.go`
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_config.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner_test.go`

- [ ] **Step 1: 写导入失败测试**

构造 `script_checker.files` 包含：

- 绝对路径
- `..`
- 目录
- 不存在文件
- entry 不在 files

Expected: 全部导入失败，错误码为 invalid params。

- [ ] **Step 2: 写导入成功测试**

包内包含：

```text
docker/check/check.py
docker/check/protocol.py
```

配置：

```yaml
entry: docker/check/check.py
files:
  - docker/check/check.py
  - docker/check/protocol.py
```

断言 `checker_config.artifact.files` 保存两个文件的 path、storage_path、sha256、size。

- [ ] **Step 3: 实现解析和存储**

规则：

- 未配置 `files` 时兼容当前行为，只存 entry。
- 配置 `files` 时去重、排序、逐个校验 safe package path。
- artifact digest 基于 `path + sha256 + size` 的稳定排序结果生成。
- 写入 `{artifactRoot}/{slug}/{digest}/{relative_path}`。

- [ ] **Step 4: 写 runner 注入测试**

断言 `CheckerRunJob.Files` 包含所有 artifact files，entry 文件路径保持 `cfg.Entry`。

- [ ] **Step 5: 实现 runner 注入**

读取 `artifact.files`；如果没有 `files`，兼容读取旧 `artifact.storage_path` 单文件结构。

- [ ] **Step 6: 运行测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/application/commands ./internal/module/challenge/domain ./internal/module/contest/application/jobs -run 'ScriptChecker|AWDChallengeImport' -count=1
```

Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add code/backend/internal/module/challenge code/backend/internal/module/contest/application/jobs
git commit -m "feat(awd): 支持脚本checker多文件artifact"
```

## Task 6: 补 checker 审计、脱敏和 artifact 清理

**Files:**
- Create: `code/backend/internal/module/contest/application/jobs/checker_audit.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Optional Create: `code/backend/cmd/cleanup-awd-checker-artifacts/main.go`
- Create or Modify: `docs/operations/awd-checker-runner.md`

- [ ] **Step 1: 写脱敏测试**

输入 stdout/stderr 或 error 中包含当前 flag，断言持久化结果不包含原始 flag。

- [ ] **Step 2: 实现脱敏 helper**

新增 helper：

```go
func redactAWDCheckerSecret(value string, secrets ...string) string
```

只替换非空 secret，替换为 `[REDACTED]`。

- [ ] **Step 3: 补审计字段**

至少记录到 check result：

- checker type
- artifact digest
- runtime
- duration_ms
- error_code
- target allowlist 或 target address 摘要

- [ ] **Step 4: artifact 清理设计**

第一版实现可重复执行的清理命令或服务方法：

- 扫描 artifact root。
- 读取当前 AWD challenge `checker_config` 中仍被引用的 storage path。
- 未引用且超过保留时间的 artifact 才删除。

删除 Linux 路径时使用 `gio trash` 或等价安全回收方式；自动化测试只验证候选计算，不真实删除用户数据。

- [ ] **Step 5: 运行测试**

Run:

```bash
cd code/backend
go test ./internal/module/contest/application/jobs ./internal/module/challenge/application/commands -run 'CheckerAudit|ScriptChecker|TCPStandard|Artifact' -count=1
```

Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add code/backend/internal/module/contest code/backend/internal/module/challenge docs/operations
git commit -m "feat(awd): 补充checker审计和artifact清理"
```

## Task 7: 真实平台端到端验证

**Files:**
- Create: `docs/verification/2026-04-29-awd-checker-completion.md`
- Optional Create: `scripts/verification/awd-checker-e2e.sh`

- [ ] **Step 1: 启动本地依赖**

按项目现有方式启动 backend、frontend、数据库、Redis、Docker runner。记录实际命令和端口。

- [ ] **Step 2: 上传 AWD TCP 题目包**

使用管理端或 API 上传 `awd-tcp-length-gate`，记录：

- upload response
- preview import response
- commit import response

- [ ] **Step 3: 创建赛事服务并 preview**

创建/编辑赛事服务，选择导入的 TCP AWD 题，执行 checker preview。

Expected:

- `service_status=up`
- `checker_type=tcp_standard`
- `status_reason=healthy`
- preview token 被消费
- readiness item 为 `passed`

- [ ] **Step 4: 验证 `script_checker` 多文件题目包**

上传一个含 `check.py + protocol.py` 的 script checker 包。

Expected:

- 导入后 artifact files 有两个文件。
- preview 进入 sandbox runner。
- readiness 通过。

- [ ] **Step 5: 触发一轮赛中检查**

运行 round updater 或缩短 round interval，确认 `awd_team_services` 写入新 checker 类型结果。

- [ ] **Step 6: 记录验证文档**

在 `docs/verification/2026-04-29-awd-checker-completion.md` 写入：

- 环境信息
- 命令
- API 响应摘要
- 截图或关键日志路径
- 失败和修复记录
- 未覆盖风险

- [ ] **Step 7: 提交**

```bash
git add docs/verification scripts/verification
git commit -m "test(awd): 记录checker端到端验证"
```

## Final Verification

完成所有任务后运行：

```bash
cd code/backend
go test ./internal/module/challenge/... ./internal/module/contest/... ./internal/app/... ./internal/config -run 'TCPStandard|ScriptChecker|PreviewChecker|Readiness|AWDRoundUpdater|DockerCheckerRunner|Contest|Config|Package|Artifact|CheckerAudit|^$' -count=1
```

```bash
cd code/frontend
npm run typecheck
npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/views/platform/__tests__/ChallengePackageFormat.test.ts
```

```bash
git diff --check
git status --short
```

Expected:

- 所有测试通过。
- 工作区干净。
- 端到端验证文档存在并包含 TCP 与 script checker 两条链路。
