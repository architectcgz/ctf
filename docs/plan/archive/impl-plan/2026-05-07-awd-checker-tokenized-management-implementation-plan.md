# AWD Checker Tokenized Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 收口 AWD TCP 题包中“公网直接暴露 flag 管理命令”的设计错误，并补齐平台 `CHECKER_TOKEN` 注入能力。

**Architecture:** 平台以 `container.flag_global_secret` 为根 secret，为声明了 `runtime_config.checker_token_env` 的 AWD 服务派生稳定的 checker token。赛中 runtime 按 `contest + team + service + awd_challenge` 维度注入 token；preview 链路使用独立 preview scope 派生 token，并同时注入 preview runtime 与 checker runner。`tcp_standard` 与 `script_checker` 共享 `{{CHECKER_TOKEN}}` 模板变量，样例题 `awd-tcp-length-gate` 把 `SET_FLAG/GET_FLAG` 改成 token 鉴权的私有管理命令，公网仅保留学生可修的业务攻击面。

**Tech Stack:** Go、Gin、GORM、Python、YAML、Markdown。

---

## 目标与非目标

- 目标：
  - 平台 runtime 链路真正消费 `checker_token_env`
  - `tcp_standard` / `script_checker` 都能渲染 `{{CHECKER_TOKEN}}`
  - `awd-tcp-length-gate` 不再把 flag 管理命令裸露到公网
  - 契约文档、架构文档、题包题解保持一致
- 非目标：
  - 不引入新的 checker 类型
  - 不实现通用 secret 管理中心
  - 不改动其他 AWD 题的业务逻辑；仅修正受 token 注入影响的 checker 配置与本地校验脚本

## 输入文档

- `docs/plan/impl-plan/2026-05-06-awd-defense-workspace-boundary-implementation-plan.md`
- `docs/architecture/features/AWD防守工作区与边界设计.md`
- `docs/architecture/features/AWD检查器运行器扩展设计.md`
- `docs/contracts/challenge-pack-v1.md`

## 需要收口的计划缺口

- 现有 defense workspace 方案已经要求学生不能接触 `CHECKER_TOKEN`，但 `tcp_standard` 仍没有 `{{CHECKER_TOKEN}}` 模板变量。
- `checker_token_env` 目前停留在题包导入与 YAML 契约层，没有贯通到实例启动和赛中 checker 执行。
- `awd-tcp-length-gate` 当前把 `SET_FLAG/GET_FLAG` 暴露在公网，而且泄露点不在学生可写工作区内。

## Token Source Of Truth

- 赛中 runtime token：
  - source of truth：平台后端基于 `container.flag_global_secret` 派生
  - 维度：`contest_id + team_id + service_id + awd_challenge_id`
  - 生命周期：`restart` 不变；`reseed` 不变；只要 service 身份不变就保持稳定
- preview token：
  - source of truth：同一根 secret，但使用独立 preview scope
  - 维度：`contest_id + service_id + awd_challenge_id`
  - 生命周期：只用于当前 preview runtime / preview checker，不写回赛中实例
- 学生边界：
  - token 只进入 runtime 容器与 checker runner
  - 不进入 defense workspace、学生态 API 或题面附件

## 当前题包目录约定

- 本次实现按当前正式契约执行：`docker/runtime`、`docker/workspace`、`docker/check`
- 不在这次 token 修复里同时引入新的根级目录布局，避免和现有 parser / contract 再次分叉
- 若后续需要调整题包根布局，应单独走契约迁移，而不是在本次 token 修复中顺手混做

## 文件边界

- 文档与契约：
  - Modify: `docs/architecture/features/AWD检查器运行器扩展设计.md`
  - Modify: `docs/contracts/challenge-pack-v1.md`
- 后端 token 注入与 checker 渲染：
  - Modify: `code/backend/internal/module/contest/ports/awd.go`
  - Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
  - Modify: `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
  - Modify: `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
  - Modify: `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
  - Modify: `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner.go`
  - Modify: `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner.go`
  - Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
  - Modify related tests:
    - `code/backend/internal/module/contest/application/jobs/awd_tcp_checker_runner_test.go`
    - `code/backend/internal/module/contest/application/jobs/awd_script_checker_runner_test.go`
    - `code/backend/internal/module/contest/application/commands/awd_service_test.go`
    - `code/backend/internal/module/contest/application/commands/contest_awd_service_service_test.go`
    - `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- 题包与题解：
  - Modify: `challenges/awd/ctf-1/awd-tcp-length-gate/challenge.yml`
  - Modify: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/app.py`
  - Modify: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/runtime/ctf_runtime.py`
  - Modify: `challenges/awd/ctf-1/awd-tcp-length-gate/docker/check/check.py`
  - Modify: `challenges/awd/ctf-1/awd-tcp-length-gate/statement.md`
  - Create: `challenges/awd/ctf-1/awd-tcp-length-gate/writeup/attack.md`
  - Create: `challenges/awd/ctf-1/awd-tcp-length-gate/writeup/defense.md`
  - Modify compatibility-only checker configs:
    - `challenges/awd/ctf-1/awd-campus-drive/challenge.yml`
    - `challenges/awd/ctf-1/awd-campus-drive/docker/check/check.py`
    - `challenges/awd/ctf-1/awd-iot-hub/challenge.yml`
    - `challenges/awd/ctf-1/awd-iot-hub/docker/check/check.py`
    - `challenges/awd/ctf-1/awd-supply-ticket/challenge.yml`
    - `challenges/awd/ctf-1/awd-supply-ticket/docker/check/check.py`

## 任务切片

### Task 1: 补齐契约与计划说明

**Review focus:** 文档是否明确区分“checker 私有管理通道”和“学生可防守业务面”。

- [x] 更新 `AWD检查器运行器扩展设计.md`，明确 `CHECKER_TOKEN` 的注入边界与 `tcp_standard` 用法。
- [x] 更新 `challenge-pack-v1.md`，补 `{{CHECKER_TOKEN}}` 模板变量与 `checker_token_env` 语义。

### Task 2: 贯通平台 checker token 链路

**Review focus:** token 是否只在 runtime / checker 内部流转，且不会进入学生态 API 或 workspace。

- [x] 让 `AWDServiceDefinition` 带出题目声明的 `checker_token_env`。
- [x] 在 contest repository 中从 `challenge_runtime` 解析该字段，并交给 checker runner。
- [x] 为 preview request、preview runtime、preview checker 补同一套 token 语义，避免 preview 与赛中链路分叉。
- [x] 为 `tcp_standard` / `script_checker` 共用模板渲染补 `{{CHECKER_TOKEN}}`。
- [x] 在实例启动链路中，根据 `checker_token_env` 给 AWD runtime 容器注入稳定 token。
- [x] 把 checker preview freshness 判断扩到会改变 checker 语义的 runtime contract，至少覆盖 `checker_token_env`。
- [x] 为 token 渲染和 runtime env 注入补定向测试。

### Task 3: 改造题包并收口兼容面

**Review focus:** 学生是否仍可在 workspace 内修题，同时现有样例题不会因平台开始注入真实 token 而失效。

- [x] 更新 `challenge.yml`，让 TCP checker 使用 token 化的 `SET_FLAG/GET_FLAG`。
- [x] 修改 runtime 协议实现，要求管理命令携带 token；未带 token 时返回明确错误。
- [x] 更新本地 `check.py`，用环境变量 token 完成 roundtrip。
- [x] 补 `statement.md`、`writeup/attack.md`、`writeup/defense.md`，明确可修补业务点与受保护边界。
- [x] 把 `awd-campus-drive`、`awd-iot-hub`、`awd-supply-ticket` 的 checker 配置与本地 `check.py` 改成消费注入的 `CHECKER_TOKEN`。

### Task 4: 验证与评审

**Review focus:** 验证是否覆盖了平台链路和题包行为，且 review 明确记录残余风险。

- [x] 运行 Go 定向测试：

```bash
cd code/backend
go test ./internal/module/contest/application/commands ./internal/module/contest/application/jobs ./internal/module/practice/application/commands -run 'AWD|Checker|TCP|Script' -count=1
```

- [x] 运行题包最小验证：

```bash
cd challenges/awd/ctf-1/awd-tcp-length-gate
CHECKER_TOKEN=local-preview-token python3 docker/check/check.py 127.0.0.1 18080
```

- [x] 归档独立 review 到 `docs/reviews/{architecture|backend|general}/...`

## 回退与兼容

- 未声明 `runtime_config.checker_token_env` 的题目继续按现有行为运行，不强制注入新的环境变量。
- 新增 `{{CHECKER_TOKEN}}` 模板变量只扩能力，不破坏旧 checker 配置。
- `awd-tcp-length-gate` 作为样例题直接切到新契约，不保留旧公网管理命令兼容。
