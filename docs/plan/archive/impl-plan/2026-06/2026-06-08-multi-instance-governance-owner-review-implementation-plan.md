<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# Multi-instance governance owner review and smoke script split Implementation Plan

**Goal:** 收口多实例部署剩余工程项：拆分 `tools/multi-instance-nginx-proxy-smoke.sh` 解除 workflow governance 脚本长度失败，并复核 / 修正 `contest` 旧 keepalive、`runtime_cleaner`、`assessment_cleaner` 的多副本 owner 风险。

**Architecture:** 保持 `tools/multi-instance-nginx-proxy-smoke.sh` 作为操作者入口，把实现细节下沉到 `scripts/lib/` 内部 helper。后台 owner 侧复用 `internal/shared/lockkeepalive` 的 lease keepalive deadline 策略；`contest` 本地旧 keepalive 改为 shared wrapper，`runtime_cleaner` 在持有 Redis cleanup lock 时续租，`assessment_cleaner` 基于现有 per-user full profile lock 做风险分级和 review 记录，不在本轮引入新的全局 assessment cleaner lock。

**Tech Stack:** Bash, Python script guard, Go, Redis lease, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-08-multi-instance-governance-owner-review`
- Started At: `2026-06-08T12:06:23Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-08-multi-instance-governance-owner-review`
- Branch: `task/2026-06-08-multi-instance-governance-owner-review`

## Objective And Non-Goals

- Objective:
  - 把 `tools/multi-instance-nginx-proxy-smoke.sh` 拆成薄入口和内部 helper，保证 `workflow-governance` 不再因为该脚本超过 260 行失败。
  - 复核 `contest` 旧 keepalive、`runtime_cleaner`、`assessment_cleaner` 在多 API 副本下的 owner 语义。
  - 对已确认的 correctness 风险做最小代码修复：`contest` keepalive deadline 与 `runtime_cleaner` 长任务续租。
  - 归档一份 backend review 证据，说明三条 owner 复核结论、剩余风险和验证结果。
- Non-Goals:
  - 不执行真实多实例 Docker / Nginx smoke；本轮只拆分脚本并保留行为入口。
  - 不做真实宿主重启演练。
  - 不把 `assessment_cleaner` 改造成全局单 owner，除非复核发现 per-user lock 不能保护 correctness。
  - 不重构 Docker socket、安全凭据或生产配置待办。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-distributed-lock-hardening-implementation-plan.md`
- Related architecture/contracts:
  - `harness/policies/script-guard.json`
  - `harness/policies/script-layer-manifest.json`
  - `tools/AGENTS.md`
  - `code/backend/internal/shared/lockkeepalive/lockkeepalive.go`
  - `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
  - `code/backend/internal/module/runtime/infrastructure/cleaner.go`
  - `code/backend/internal/module/assessment/application/commands/cleaner.go`
- Related prior work:
  - `docs/reviews/backend/2026-06-08-gate-review-multi-instance-startup-recovery-gate-fix.md`
  - `docs/reviews/backend/2026-06-08-gate-review-multi-instance-distributed-lock-hardening.md`
  - `docs/reviews/backend/2026-06-08-multi-instance-distributed-lock-hardening-review.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达 workflow governance、脚本层规则、后台分布式锁 owner 和 review 证据；需要隔离 worktree、实施计划、验证和独立 review。

## Files

- Create:
  - `scripts/lib/multi-instance-nginx-proxy-smoke/run.sh`
  - `docs/reviews/backend/2026-06-08-multi-instance-governance-owner-review.md`
- Modify:
  - `tools/multi-instance-nginx-proxy-smoke.sh`
  - `code/backend/internal/module/contest/application/jobs/lock_keepalive.go`
  - `code/backend/internal/module/contest/application/jobs/lock_keepalive_test.go`
  - `code/backend/internal/module/runtime/infrastructure/cleaner.go`
  - `code/backend/internal/module/runtime/infrastructure/cleaner_test.go`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-08-multi-instance-governance-owner-review-implementation-plan.md`
- Review:
  - `code/backend/internal/module/contest/application/jobs/status_update_runner.go`
  - `code/backend/internal/module/contest/application/jobs/awd_round_scheduler_runtime.go`
  - `code/backend/internal/module/assessment/application/commands/profile_service.go`
  - `code/backend/internal/module/assessment/infrastructure/state_store.go`
  - `code/backend/internal/module/assessment/runtime/module.go`
- Test:
  - `bash scripts/check-script-guard.sh`
  - `bash scripts/check-script-layer.sh`
  - `go test ./internal/module/contest/application/jobs -run 'RedisLockKeepalive|Lock|Scheduler' -count=1`
  - `go test ./internal/module/runtime/infrastructure -run Cleaner -count=1`
  - `go test ./internal/shared/lockkeepalive -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`

## 复用与 Owner 决策

- Existing patterns searched:
  - `scripts/lib` 已是内部 helper 命名空间，并被 script guard 排除。
  - `internal/shared/lockkeepalive` 已实现 refresh deadline、TTL 超时 fail-closed 和测试。
  - `assessment` profile rebuild 已有 per-user `AcquireFullProfileRebuildLock`。
- Reuse / extend / split / create-new decision:
  - 脚本拆分复用 `scripts/lib`，不新增 `tools/lib`，避免触发 script-layer guard。
  - `contest` 旧 keepalive 改为调用 shared `lockkeepalive.Start/RefreshInterval`，不维护第二套租约续租逻辑。
  - `runtime_cleaner` 复用 shared `lockkeepalive`，在单轮清理期间保持 cleanup lock。
  - `assessment_cleaner` 本轮不新增全局 lock；将其评估为“重复全量扫描负载风险，但 per-user lock 保护单用户画像写入 correctness”。
- Owner boundary:
  - `tools/multi-instance-nginx-proxy-smoke.sh` 只做入口和 source guard。
  - `scripts/lib/multi-instance-nginx-proxy-smoke/run.sh` 承接 smoke 细节。
  - `shared/lockkeepalive` 是 Redis lease keepalive policy owner。
  - `runtime_cleaner` 是过期实例 / orphan cleanup 单 owner。
  - `assessment` profile service 是单用户画像写入 owner。
- Why this is the narrowest safe surface:
  - 不改变脚本外部调用方式，不新增配置项，不改 DB schema；只把已有长脚本和已有锁语义收口到现有 owner。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这是用户指定的工程收口项，需要先判定三条后台 owner 哪些是 correctness、哪些是 residual / capacity 风险。
- grill-with-docs findings:
  - `tools/` 目录不允许新增未登记子目录，脚本 helper 应放到已登记的 `scripts/lib`。
  - `contest` 旧 keepalive 与 shared keepalive 语义重复，且缺少单次 refresh deadline；应复用 shared owner。
  - `runtime_cleaner` 只有 acquire/release，没有 keepalive；清理超过 TTL 时可能出现第二副本接管同一轮 cleanup。
  - `assessment_cleaner` 没有全局 owner，但 `CalculateSkillProfile()` 对每个 user 使用 `AcquireFullProfileRebuildLock`；多副本重复 cron 不会双写同一 user，但会重复扫描学生列表。
- Plan adjustments after challenge:
  - 代码修复只覆盖 `contest` keepalive 和 `runtime_cleaner` lease keepalive。
  - `assessment_cleaner` 不混入新锁，避免在缺少明确 correctness 风险时扩大配置和测试面；通过 review 记录风险分级。

## Validation

- Commands:
  - `bash scripts/check-script-guard.sh`
  - `bash scripts/check-script-layer.sh`
  - `go test ./internal/shared/lockkeepalive -count=1`
  - `go test ./internal/module/contest/application/jobs -run 'RedisLockKeepalive|Lock|Scheduler' -count=1`
  - `go test ./internal/module/runtime/infrastructure -run Cleaner -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh pre-commit-quick`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh workflow-governance`
- Manual checks:
  - `wc -l tools/multi-instance-nginx-proxy-smoke.sh` 小于 260。
  - smoke 脚本入口路径和环境变量兼容。
  - owner review 结论覆盖 `contest`、`runtime_cleaner`、`assessment_cleaner`。
- Review focus:
  - 脚本拆分是否保持原命令行为。
  - shared keepalive 是否正确传播 lock loss 到 run context。
  - `runtime_cleaner` release 是否不被 canceled run context 打断。
  - `assessment_cleaner` 风险分级是否有代码证据支撑。
