<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# backend-config-package-split Implementation Plan

**Goal:** 将 `code/backend/internal/config` 从单个超长文件拆成同包内的多个职责文件，保持对外类型、函数名和调用方式不变。

**Architecture:** 保留 `package config` 作为唯一 owner，不做 package 级迁移，也不改变 `Config`、`Load()`、`Validate()`、`PostgresConfig.DSN()`、`AuthConfig.CookieSameSite()` 等现有导出入口。拆分只发生在 `internal/config` 包内部，按“类型定义 / 加载 / 默认值 / 校验 / container flag secret”五类职责分文件，并用结构护栏测试锁住新的文件布局。

**Tech Stack:** Go, Viper, Go test, apply_patch, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-backend-config-package-split`
- Started At: `2026-06-12T04:23:45Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-12-backend-config-package-split`
- Branch: `task/2026-06-12-backend-config-package-split`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `review-pending` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [x] Intake analysis gate completed
  - [x] Plan review / architecture-fit check completed
  - [x] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective:
  - 把 `code/backend/internal/config/config.go` 的混合职责拆成同包内多个文件，降低后续新增配置继续堆到单文件的风险。
  - 保持 `config` 包对外 API、导出类型名、调用路径和 YAML / env 配置契约不变。
  - 同步把当前事实文档中对单文件事实源的引用收口到 `code/backend/internal/config/` 目录 owner。
- Non-Goals:
  - 不新增或删除任何配置字段。
  - 不把 `internal/config` 再拆成多个 package。
  - 不顺手重构调用方对 `*config.Config` 的依赖方式。
  - 不改 `code/backend/configs/*.yaml` 的配置内容，除非为了保持测试或文档一致性必须做最小修正。

## Problem Statement

- Current behavior / structure:
  - `code/backend/internal/config/config.go` 当前 1189 行，混放了类型定义、Viper 加载、默认值、校验、CIDR 辅助函数、container flag secret 持久化与 keyring 解析。
  - `code/backend/internal/config/config_test.go` 当前 723 行，已经反映出 `config.go` 是高频演进热点。
  - 近期多项运行时 / HA / registry / shared storage 改动都持续触碰这一单文件 surface，继续叠加会让 review 和后续增量修改成本越来越高。
- Target behavior / structure:
  - `code/backend/internal/config/` 目录按职责拆成多个同包文件：`types.go`、`load.go`、`defaults.go`、`validate.go`、`container_flag_secret.go`，并保留现有测试入口。
  - 现有外部调用仍通过 `config.Load()`、`config.Config`、`config.RedisConfig` 等入口工作，无需调用方修改 import 或类型名。
- Why this task is needed now:
  - 用户明确质疑 config 过长；当前结构已经是已知热点，继续在单文件追加配置只会放大结构债。
  - 这次拆分不改变行为，适合作为独立、可审查的结构性收口切片。

## Inputs

- Source docs:
  - `AGENTS.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/architecture/features/题包Registry交付架构.md`
- Related prior work:
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/redis-sentinel-and-postgres-ha-connectivity.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-12-true-ha-group/shared-storage-owner-convergence.md`
  - `.harness/reuse-decisions/runtime-subnet-pool-split.md`
  - `.harness/reuse-decisions/registry-access-endpoint.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 涉及多文件后端重构、测试调整和架构文档事实源更新。
  - 触达 `internal/config` 这一高频共享 surface，需要 code-workflow 的 plan / validation / review gate。

## Files

- Create:
  - `code/backend/internal/config/types.go`
  - `code/backend/internal/config/load.go`
  - `code/backend/internal/config/defaults.go`
  - `code/backend/internal/config/validate.go`
  - `code/backend/internal/config/container_flag_secret.go`
  - `code/backend/internal/config/config_structure_test.go`
- Modify:
  - `code/backend/internal/config/config.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/architecture/features/题包Registry交付架构.md`
- Review:
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/config/air_config_test.go`
- Test:
  - `code/backend/internal/config/config_structure_test.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/config/air_config_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - 现有 `internal/config` 目录只有 `config.go`、`config_test.go`、`air_config_test.go`，没有现成的分文件布局可复用。
  - 文档里已有多处把 `code/backend/internal/config/config.go` 作为运行时配置事实源，说明当前 owner 是整个包而不是某个业务模块。
- Reuse / extend / split / create-new decision:
  - 复用 `package config` 与现有导出 API。
  - 只做目录内分文件拆分，不做 package 级迁移。
  - 新增一个结构护栏测试，避免后续再退回单文件聚集。
- Owner boundary:
  - `internal/config` 继续 owner 平台级配置契约、默认值、加载与启动校验。
  - 调用方只消费导出的类型和方法，不感知文件布局。
- Why this is the narrowest safe surface:
  - 避开了广泛的 import / package 迁移风险。
  - 只动 `internal/config` 包和少量把单文件写成事实源的文档，能直接收口当前结构债。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这次是结构性重构，不是线上故障；关键在于先确认“拆到什么层级最小且安全”，而不是直接编码。
- grill-with-docs findings:
  - 本仓库文档把 `internal/config/config.go` 多次当作事实源，拆分后必须把活动事实源改成 `internal/config/` 目录 owner，而不是继续指向单个文件。
  - 现有对外使用方式广泛依赖 `config.Config` 与子配置结构体；不适合在本切片里做 package 级拆分。
  - 当前 touched structural debt 就在 `internal/config/config.go` 自身，必须在本次切片内收口，而不是只给 follow-up。
- Plan adjustments after challenge:
  - 明确把文档引用同步纳入本切片。
  - 明确先写结构护栏测试，再做分文件实现。

## Execution Slices

### Slice 1: internal/config 同包分文件拆分

- Goal:
  - 用最小改动把 `internal/config` 从单文件拆成多文件，同时保持行为与对外 API 不变。
- Dependencies:
  - 依赖现有 `config` 包测试可稳定运行。
- Files:
  - Create:
    - `code/backend/internal/config/config_structure_test.go`
    - `code/backend/internal/config/types.go`
    - `code/backend/internal/config/load.go`
    - `code/backend/internal/config/defaults.go`
    - `code/backend/internal/config/validate.go`
    - `code/backend/internal/config/container_flag_secret.go`
  - Modify:
    - `code/backend/internal/config/config.go`
    - `docs/architecture/backend/03-container-architecture.md`
    - `docs/operations/awd-host-reboot-recovery-drill.md`
    - `docs/architecture/features/题包Registry交付架构.md`
  - Review:
    - `code/backend/internal/config/config_test.go`
  - Test:
    - `code/backend/internal/config/config_structure_test.go`
    - `code/backend/internal/config/config_test.go`
- Steps:
  - [x] Step 1: 新增 `config_structure_test.go`，要求 `internal/config` 至少包含 `types.go`、`load.go`、`defaults.go`、`validate.go`、`container_flag_secret.go` 五个职责文件。
  - [x] Step 2: 运行 `go test ./internal/config -run TestConfigPackageSplitByResponsibility -count=1`，确认新测试先失败。
  - [x] Step 3: 按职责把 `config.go` 中的类型定义迁到 `types.go`，把 `Load()` 迁到 `load.go`，把 `setDefaults()` 迁到 `defaults.go`，把 `Validate()` 与校验辅助函数迁到 `validate.go`，把 flag secret 相关逻辑迁到 `container_flag_secret.go`。
  - [x] Step 4: 把 `config.go` 收口为仅保留仍未迁出的公共辅助逻辑；若没有剩余 owner，则删除空壳并确保包内编译通过。
  - [x] Step 5: 更新活动事实文档，将 `code/backend/internal/config/config.go` 的活动事实源引用调整为 `code/backend/internal/config/`。
  - [x] Step 6: 运行最小充分验证并做同上下文自检 review。
- Validation:
  - `cd code/backend && go test ./internal/config -run TestConfigPackageSplitByResponsibility -count=1`
  - `cd code/backend && go test ./internal/config -count=1`
  - `git diff --check -- code/backend/internal/config docs/architecture/backend/03-container-architecture.md docs/operations/awd-host-reboot-recovery-drill.md docs/architecture/features/题包Registry交付架构.md`
- Review focus:
  - 对外 API 是否完全不变。
  - 校验、默认值和 secret 持久化逻辑是否被完整迁移，无遗漏。
  - 文档事实源是否不再误指向单个 `config.go`。
- Done criteria:
  - `internal/config` 包通过测试。
  - 结构护栏测试转绿。
  - 活动文档完成事实源更新。

## Impact And Compatibility

- API / DTO:
  - 无对外 API / DTO 变化。
- Data / migration:
  - 无数据库或 migration 变化。
- State / cache / queue / event:
  - 无状态模型变化。
- Runtime / config:
  - 运行时配置字段、默认值、环境变量覆盖与启动校验应保持不变。
- Frontend route / state / UX:
  - 无影响。
- Docs / contracts:
  - 更新活动架构 / 运维文档中的事实源引用。

## Plan Review / Architecture Fit

- Target owner boundary:
  - `internal/config` 目录继续作为平台配置 owner；文件内职责拆开，但 package boundary 不变。
- Reuse points / landing zones:
  - 所有现有调用方继续复用 `package config` 的导出类型和函数；新 landing zone 只体现在目录内文件分工。
- Known structural debt touched:
  - `code/backend/internal/config/config.go` 的过长与职责混放。
- How this plan avoids behavior-only convergence:
  - 这次不是只增加注释或局部整理，而是直接把混放职责拆到稳定文件 owner，并加测试防回退。
- Hidden second-redesign risk:
  - 若本次只拆文件但仍保留“新增配置统一往一个文件加”的习惯，债务会回流；因此需要结构护栏测试锁住新的布局。
- Decision after review:
  - 可以实施；当前最小正确切片就是同包分文件，而不是 package 级抽象或调用方迁移。

## Documentation Owner

- Current fact sources to read:
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/architecture/features/题包Registry交付架构.md`
- Fact sources to update after implementation:
  - 上述 3 份活动文档
- Plan-only notes that must not become architecture source:
  - “按五类职责分文件”的实施过程属于本 plan，不自动写成架构事实；事实文档只说明 `internal/config/` 目录是 owner。
- Archive condition:
  - 代码和文档完成、验证完成后，按 code-workflow 归档当前 plan。

## Validation Plan

- Per-slice commands:
  - `cd code/backend && go test ./internal/config -run TestConfigPackageSplitByResponsibility -count=1`
  - `cd code/backend && go test ./internal/config -count=1`
- Integration commands:
  - `git diff --check -- code/backend/internal/config docs/architecture/backend/03-container-architecture.md docs/operations/awd-host-reboot-recovery-drill.md docs/architecture/features/题包Registry交付架构.md`
- Manual checks:
  - 检查 `rg -n "事实源：.*internal/config/config.go|internal/config/config.go"` 在活动文档上是否已清空或只剩历史 / archive 引用。
- Commands intentionally skipped and why:
  - `bash scripts/check-workflow-governance.sh`：本次不新增文档路径或 guardrail 规则，先以更窄的 `git diff --check` 和包级测试为主；若后续文档入口或治理脚本被触达，再补跑。

## Validation Evidence

- Command:
  - Result: `PASS`
  - Notes: `bash scripts/check-task-intake.sh`
- Command:
  - Result: `PASS`
  - Notes: `bash scripts/check-startup-gate.sh`
- Command:
  - Result: `PASS`
  - Notes: `cd code/backend && go test ./internal/config -count=1`（改动前基线）
- Command:
  - Result: `FAIL（预期）`
  - Notes: `cd code/backend && go test ./internal/config -run TestConfigPackageSplitByResponsibility -count=1`，失败点为缺少 `types.go`
- Command:
  - Result: `PASS`
  - Notes: `cd code/backend && gofmt -w internal/config/*.go`
- Command:
  - Result: `PASS`
  - Notes: `cd code/backend && go test ./internal/config -run TestConfigPackageSplitByResponsibility -count=1`
- Command:
  - Result: `PASS`
  - Notes: `cd code/backend && go test ./internal/config -count=1`
- Command:
  - Result: `PASS`
  - Notes: `bash scripts/run-workflow-stage.sh completion-full`
- Command:
  - Result: `PASS`
  - Notes: `git diff --check -- code/backend/internal/config docs/architecture/backend/03-container-architecture.md docs/operations/awd-host-reboot-recovery-drill.md docs/architecture/features/题包Registry交付架构.md`
- Command:
  - Result: `PASS`
  - Notes: `bash scripts/run-workflow-stage.sh workflow-governance`

## Independent Review Handoff

- Review target:
  - `code/backend/internal/config/` 分文件拆分与活动文档事实源更新
- Validation evidence summary:
  - 结构护栏测试已完成红绿验证，`./internal/config` 包测试通过，`completion-full` 与 `workflow-governance` 已通过。
- Architecture / contract inputs:
  - `AGENTS.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
  - `docs/architecture/features/题包Registry交付架构.md`
- Known risks / review focus:
  - 是否遗漏了包内辅助函数迁移。
  - 是否意外改变了 `Load()` / `Validate()` 的行为顺序。
  - 是否仍有活动事实源指向 `config.go` 单文件。
- Project-local checks to consider:
  - `cd code/backend && go test ./internal/config -count=1`
  - `git diff --check -- <touched-files>`
  - `bash scripts/run-workflow-stage.sh completion-full`
  - `bash scripts/run-workflow-stage.sh workflow-governance`

## Rollback / Recovery

- Safe revert boundary:
  - 整个切片可通过回退本次提交恢复到单文件布局。
- Data / config / runtime recovery notes:
  - 无数据层恢复动作；运行时配置契约应保持兼容。
- Irreversible operations:
  - 无。

## Residual Risks

- Risk:
  - 仍然保留一个较大的 `Config` 聚合结构体，后续如果配置继续增长，可能需要进一步引入更细的测试分层。
- Why acceptable:
  - 当前用户诉求是收口文件级过长问题；不改变 package boundary 是这次最小安全改动。
- Follow-up owner, if any:
  - `internal/config` 后续新增大块配置时，在对应任务里继续按当前分文件 owner 落位，不再回堆单文件。
