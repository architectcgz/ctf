# AWD Runtime Cutover Phase 2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改动 AWD 轮次调度主体和攻击提交流程的前提下，把运行态的 service definition / readiness 读取源优先切到 `contest_awd_services`，并保留对 legacy `contest_challenges.awd_*` 的兼容 fallback。

**Architecture:** 这一阶段不改变现有 runtime 仍以 `challenge_id` 作为 service target key 的事实，也不新增新的 worker 或 checker 调度器。改动只收敛在 `AWDRepository` 的两个查询入口：`ListServiceDefinitionsByContest` 和 `ListReadinessChallengesByContest`，让它们优先从 `contest_awd_services.runtime_config + score_config` 解析运行配置，再在配置缺失或赛事尚未迁移时回退到 `contest_challenges.awd_*`。

**Tech Stack:** Go, GORM, PostgreSQL, existing AWD backend tests

---

### Task 1: Lock Phase Boundary In Tests

**Files:**
- Modify: `code/backend/internal/module/contest/application/queries/awd_service_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`

- [ ] **Step 1: Write the failing readiness test**

```go
func TestAWDQueryServiceGetReadinessPrefersContestAWDServiceRuntimeConfig(t *testing.T) {
	// contest_challenges 保留旧配置，contest_awd_services 写入新配置
	// 断言 readiness 使用 service 层 checker_type / checker_config
}
```

- [ ] **Step 2: Write the failing round runtime test**

```go
func TestAWDServiceRunCurrentRoundChecksPrefersContestAWDServiceDefinitions(t *testing.T) {
	// service 层 runtime_config / score_config 与 legacy 字段故意不一致
	// 断言 checker_type、sla_score、defense_score 都来自 contest_awd_services
}
```

- [ ] **Step 3: Verify both tests fail for the expected reason**

Run: `cd code/backend && go test ./internal/module/contest/application/queries ./internal/module/contest/application/commands -run 'ContestAWDService|ReadinessPrefers|RunCurrentRoundChecksPrefers' -count=1`

Expected: FAIL because仓储仍然只读取 `contest_challenges.awd_*`。

### Task 2: Cut Runtime Read Paths Over To `contest_awd_services`

**Files:**
- Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
- Modify: `code/backend/internal/module/contest/ports/awd.go`

- [ ] **Step 1: Add repository-side parsing helpers**

```go
type awdContestServiceRuntimeRow struct {
	ChallengeID         int64
	FlagPrefix          string
	DisplayName         string
	RuntimeConfig       string
	ScoreConfig         string
	LegacyCheckerType   model.AWDCheckerType
	LegacyCheckerConfig string
	LegacySLAScore      int
	LegacyDefenseScore  int
}
```

- [ ] **Step 2: Implement service-first query path**

```go
func (r *AWDRepository) ListServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	rows, err := r.listContestAWDServiceRuntimeRows(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return r.listLegacyServiceDefinitionsByContest(ctx, contestID)
	}
	return mapContestAWDServiceDefinitions(rows), nil
}
```

- [ ] **Step 3: Keep partial fallback semantics**

```go
// runtime_config / score_config 缺字段时，回退到 cc.awd_*，避免旧赛事因缺省配置直接失效。
```

- [ ] **Step 4: Re-run targeted tests**

Run: `cd code/backend && go test ./internal/module/contest/application/queries ./internal/module/contest/application/commands -run 'ContestAWDService|ReadinessPrefers|RunCurrentRoundChecksPrefers' -count=1`

Expected: PASS

### Task 3: Record The New Runtime Boundary And Verify Module Scope

**Files:**
- Modify: `docs/architecture/backend/design/awd-engine-migration.md`

- [ ] **Step 1: Update migration doc**

```md
- 已切换：readiness / round updater / checker runner 的 service definition 读取入口优先读 `contest_awd_services`
- 仍未切换：workspace 编排、flag 注入目标标识、攻击提交流程的 challenge 兼容键
```

- [ ] **Step 2: Run minimal sufficient verification**

Run:
- `cd code/backend && go test ./internal/module/contest/application/queries -count=1`
- `cd code/backend && go test ./internal/module/contest/application/commands -count=1`
- `cd code/backend && go test ./internal/module/contest/... -count=1`

Expected:
- 新增 runtime cutover 测试通过
- 既有 `contest` 模块测试保持通过

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/plans/2026-04-17-awd-runtime-cutover-phase2.md \
  code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go \
  code/backend/internal/module/contest/application/queries/awd_service_test.go \
  code/backend/internal/module/contest/application/commands/awd_service_test.go \
  docs/architecture/backend/design/awd-engine-migration.md
git commit -m "feat(AWD): 切换运行态服务读取入口"
```
