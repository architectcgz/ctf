# AWD Runtime Target Identity Phase 3 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改动 `awd_team_services`、攻击日志主键和排行榜聚合口径的前提下，为 AWD 运行态引入 `contest_awd_services.id` 作为显式 service target identity，并让 round flag / checker 取 flag / 攻击提交流程支持 `service_id` 优先、`challenge_id` 回退。

**Architecture:** 这一阶段不把全链路完全改成 `service_id` 持久化，而是先在运行时对象上增加 `ServiceID`，并在 Redis round flag 字段与解析逻辑里建立双键兼容层。轮次生成时写入 `service_id` 字段并保留 legacy `challenge_id` 字段；checker 和攻击提交读取时优先读 `service_id` 字段，缺失时回退到 legacy 字段，同时 flag 内容仍继续沿用 `challenge_id` 参与生成，避免一次切换导致旧赛事 flag 全量失效。

**Tech Stack:** Go, GORM, Redis, existing AWD backend tests

---

### Task 1: Lock The Bridge Behavior In Tests

**Files:**
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`

- [ ] **Step 1: Write the failing round flag bridge test**

```go
func TestAWDRoundUpdaterSyncRoundFlagsWritesServiceAndLegacyFields(t *testing.T) {
	// contest_awd_services 有独立 service_id
	// 断言 Redis 里同时存在 service_id field 和 challenge_id field
	// 且两者值相同
}
```

- [ ] **Step 2: Write the failing attack submit bridge test**

```go
func TestAWDServiceSubmitAttackAcceptsServiceScopedRoundFlagField(t *testing.T) {
	// Redis 只写 service_id field，不写 legacy challenge_id field
	// 断言 SubmitAttack 仍能通过 service_id 映射取到有效 flag
}
```

- [ ] **Step 3: Run targeted tests to verify they fail**

Run: `cd code/backend && go test ./internal/module/contest/application/jobs ./internal/module/contest/application/commands -run 'ServiceScopedRoundFlagField|WritesServiceAndLegacyFields' -count=1`

Expected: FAIL because当前 round flag 仍只写 `team_id:challenge_id`，攻击提交流程也只按 challenge 字段取 flag。

### Task 2: Add `ServiceID` To Runtime Target Objects

**Files:**
- Modify: `code/backend/internal/module/contest/ports/awd.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go`

- [ ] **Step 1: Extend runtime structs**

```go
type AWDServiceDefinition struct {
	ServiceID    int64
	ChallengeID  int64
	FlagPrefix   string
	CheckerType  model.AWDCheckerType
	CheckerConfig string
	SLAScore     int
	DefenseScore int
}
```

```go
type AWDFlagAssignment struct {
	ServiceID    int64
	ChallengeID  int64
	TeamID       int64
	Flag         string
}
```

```go
type AWDServiceInstance struct {
	ServiceID    int64
	ChallengeID  int64
	TeamID       int64
	AccessURL    string
}
```

- [ ] **Step 2: Return service identity from repository**

```go
SELECT
	cas.id AS service_id,
	cas.challenge_id AS challenge_id,
	...
```

```go
SELECT
	COALESCE(cas.id, 0) AS service_id,
	inst.challenge_id AS challenge_id,
	...
```

- [ ] **Step 3: Re-run compile-level tests**

Run: `cd code/backend && go test ./internal/module/contest/application/jobs ./internal/module/contest/application/commands -run '^$' -count=1`

Expected: PASS compilation after struct and repository changes.

### Task 3: Bridge Redis Round Flag Fields And Lookups

**Files:**
- Modify: `code/backend/internal/pkg/redis/keys.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_support.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_sync.go`
- Modify: `code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_flag_support.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go`

- [ ] **Step 1: Add service-aware Redis field helper**

```go
func AWDRoundFlagServiceField(teamID, serviceID int64) string {
	return fmt.Sprintf("%d:s:%d", teamID, serviceID)
}
```

- [ ] **Step 2: Write both field variants during flag sync**

```go
fields[rediskeys.AWDRoundFlagServiceField(item.TeamID, item.ServiceID)] = item.Flag
fields[rediskeys.AWDRoundFlagField(item.TeamID, item.ChallengeID)] = item.Flag
```

- [ ] **Step 3: Read service field first, then legacy field**

```go
flag, err := HGet(... AWDRoundFlagServiceField(teamID, definition.ServiceID))
if miss {
	flag, err = HGet(... AWDRoundFlagField(teamID, definition.ChallengeID))
}
```

- [ ] **Step 4: Let attack submit resolve optional contest service identity**

```go
service, err := s.repo.FindContestAWDServiceByContestAndChallenge(ctx, contestID, challengeID)
// 找到则把 service.ID 带入 resolveAcceptedRoundFlags
```

- [ ] **Step 5: Re-run targeted bridge tests**

Run: `cd code/backend && go test ./internal/module/contest/application/jobs ./internal/module/contest/application/commands -run 'ServiceScopedRoundFlagField|WritesServiceAndLegacyFields' -count=1`

Expected: PASS

### Task 4: Align Flag Injection Target Metadata And Document Boundary

**Files:**
- Modify: `code/backend/internal/module/contest/infrastructure/awd_docker_flag_injector.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_flag_injector_support.go`
- Modify: `code/backend/internal/module/contest/infrastructure/awd_flag_injector_test.go`
- Modify: `docs/architecture/backend/design/awd-engine-migration.md`

- [ ] **Step 1: Thread `ServiceID` through injector metadata**

```go
type pair struct {
	teamID    int64
	serviceID int64
}
```

```go
containerIDs, err := i.findTargetContainers(ctx, contest.ID, item.TeamID, item.ServiceID, item.ChallengeID)
```

- [ ] **Step 2: Keep lookup fallback conservative**

```go
// 当前实例表仍无 service_id，容器筛选仍以 challenge_id 落地；
// 但注入侧已感知 service_id，便于下一阶段继续切 workspace target。
```

- [ ] **Step 3: Update migration doc**

```md
- 已切换：round flag 存储、checker 取 flag、攻击提交支持 `service_id` 优先
- 仍未切换：`awd_team_services` / `awd_attack_logs` 仍以 `challenge_id` 作为持久化聚合键
```

- [ ] **Step 4: Run minimal sufficient verification**

Run:
- `cd code/backend && go test ./internal/module/contest/application/jobs -count=1`
- `cd code/backend && go test ./internal/module/contest/application/commands -count=1`
- `cd code/backend && go test ./internal/module/contest/infrastructure -count=1`
- `cd code/backend && go test ./internal/module/contest/... -count=1`

Expected:
- 新增 bridge 用例通过
- `contest` 模块保持通过

- [ ] **Step 5: Commit**

```bash
git add docs/superpowers/plans/2026-04-18-awd-runtime-target-identity-phase3.md \
  code/backend/internal/module/contest/ports/awd.go \
  code/backend/internal/module/contest/infrastructure/awd_contest_relation_repository.go \
  code/backend/internal/module/contest/infrastructure/awd_service_instance_repository.go \
  code/backend/internal/pkg/redis/keys.go \
  code/backend/internal/module/contest/application/jobs/awd_round_flag_support.go \
  code/backend/internal/module/contest/application/jobs/awd_round_flag_sync.go \
  code/backend/internal/module/contest/application/jobs/awd_round_flag_lookup_support.go \
  code/backend/internal/module/contest/application/commands/awd_flag_support.go \
  code/backend/internal/module/contest/application/commands/awd_attack_submit_support.go \
  code/backend/internal/module/contest/infrastructure/awd_docker_flag_injector.go \
  code/backend/internal/module/contest/infrastructure/awd_flag_injector_support.go \
  code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go \
  code/backend/internal/module/contest/application/commands/awd_service_test.go \
  code/backend/internal/module/contest/infrastructure/awd_flag_injector_test.go \
  docs/architecture/backend/design/awd-engine-migration.md
git commit -m "feat(AWD): 桥接运行态服务目标标识"
```
