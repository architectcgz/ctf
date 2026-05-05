# Goverter Final Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finish the remaining Go response/DTO mapper migration without leaving redundant pass-through wrappers.

**Architecture:** Keep generated mapper ownership local to the package that owns the conversion. Wrappers are allowed only when they add business semantics such as derived fields, visibility rules, parsing, aggregation, or normalization that belongs outside goverter.

**Tech Stack:** Go, goverter v1.9.2, `go generate`, package-scoped mapper interfaces, focused `go test`.

---

## Current Rules

- Do not migrate to an intermediate shape like `mapped := mapper.ToX(...); return &mapped`.
- Prefer direct pointer mapper methods such as `ToXPtr(source *model.X) *dto.X`.
- Delete redundant `ResultToDTO`, `ResultsToDTO`, `RespFromModel`, and `FromRecord` wrappers in the same commit that introduces generated mapper methods.
- Keep wrappers only when they contain real behavior: derived fields, parsing JSON/config, trimming and preview generation, status calculation, permissions/visibility, or map aggregation.
- After each task: generate, format, run focused tests, commit.

## Remaining Scan Snapshot

Commands used:

```bash
rg -n "func\\s+.*(ToDTO|ResultsToDTO|ResultToDTO|RespFromModel|InfoFromModel|FromModel|FromRecord)\\(" code/backend/internal/module --glob '*.go'
rg -n "func\\s+to[A-Z][A-Za-z0-9_]*\\(" code/backend/internal/module --glob '*.go'
```

Remaining candidates:

- `practice/domain/mappers.go`
- `practice/application/commands/service.go`
- `assessment/application/commands/report_service.go`
- `challenge/domain/mappers.go`
- `challenge/domain/topology_codec.go`
- `contest/application/commands/response_mappers.go`
- `contest/application/commands/awd_response_mappers.go`
- `contest/api/http/awd_service_manage_handler.go`
- `contest/application/queries/contest_result.go`
- `contest/application/queries/team_result.go`
- `runtime/application/commands/instance_service.go`
- `runtime/application/queries/instance_service.go`
- `identity/application/commands|queries/support.go`
- `ops/application/commands|queries/notification_service.go`
- `teaching_readmodel/application/queries/service.go`
- `challenge/application/commands/challenge_service.go`
- `challenge/domain/package_topology_parser.go`

---

### Task 1: Fix Challenge Domain Mapper Shape

**Files:**
- Modify: `code/backend/internal/module/challenge/domain/mappers.go`
- Modify: `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- Modify: `code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go`

- [ ] **Step 1: Confirm no redundant local wrappers remain**

Run:

```bash
rg -n "mapped := challengeResponseMapperInst|return &mapped" code/backend/internal/module/challenge/domain
```

Expected: no redundant `mapped := ...; return &mapped` for simple mapper calls.

- [ ] **Step 2: Generate and format**

Run:

```bash
cd code/backend
go generate ./internal/module/challenge/domain
gofmt -w internal/module/challenge/domain/mappers.go internal/module/challenge/domain/response_mapper_goverter*.go
```

- [ ] **Step 3: Verify**

Run:

```bash
go test -timeout=120s ./internal/module/challenge/domain
go test -timeout=120s ./internal/module/challenge/...
```

- [ ] **Step 4: Commit**

```bash
git add code/backend/internal/module/challenge/domain
git commit -m "refactor(challenge): 收敛domain响应映射到goverter直返"
```

### Task 2: Practice Domain Instance Mapper

**Files:**
- Modify: `code/backend/internal/module/practice/domain/mappers.go`
- Create: `code/backend/internal/module/practice/domain/response_mapper_goverter.go`
- Create: `code/backend/internal/module/practice/domain/response_mapper_goverter_assign.go`
- Create: `code/backend/internal/module/practice/domain/response_mapper_goverter_gen.go`

- [ ] **Step 1: Add goverter mapper for pure fields**

Add mapper methods for `model.Instance -> dto.InstanceResp`, ignoring fields with business construction:

```go
// goverter:ignore Access
// goverter:ignore RemainingExtends
ToInstanceRespBase(source model.Instance) dto.InstanceResp
```

- [ ] **Step 2: Keep only semantic wrapper**

Update `InstanceRespFromModel` to:

```go
func InstanceRespFromModel(inst *model.Instance) *dto.InstanceResp {
	resp := practiceResponseMapperInst.ToInstanceRespBasePtr(inst)
	if resp == nil {
		return nil
	}
	resp.Access = dto.BuildInstanceAccessInfo(inst.AccessURL)
	resp.RemainingExtends = RemainingExtends(inst)
	return resp
}
```

This wrapper is retained because `Access` and `RemainingExtends` are derived fields.

- [ ] **Step 3: Generate, format, verify**

Run:

```bash
cd code/backend
go generate ./internal/module/practice/domain
gofmt -w internal/module/practice/domain/*.go
go test -timeout=120s ./internal/module/practice/domain
go test -timeout=120s ./internal/module/practice/...
```

- [ ] **Step 4: Commit**

```bash
git add code/backend/internal/module/practice/domain
git commit -m "refactor(practice): goverter迁移实例响应基础映射"
```

### Task 3: Practice Manual Review Response Mapping

**Files:**
- Modify: `code/backend/internal/module/practice/application/commands/service.go`
- Create or modify: `code/backend/internal/module/practice/application/commands/response_mapper_goverter.go`
- Create or modify: `code/backend/internal/module/practice/application/commands/response_mapper_goverter_assign.go`
- Create or modify: `code/backend/internal/module/practice/application/commands/response_mapper_goverter_gen.go`

- [ ] **Step 1: Split pure copy from semantic fields**

Migrate these functions only where fields are direct copies:

- `manualReviewDetailRespFromRecord`
- `manualReviewListItemRespFromRecord`
- `challengeSubmissionRecordRespFromModel`

Keep semantic code in wrappers:

- `AnswerPreview` truncation
- pending-review status logic
- `Answer` hiding except pending review

- [ ] **Step 2: Avoid intermediate wrappers**

Generated mapper methods should return final pointer forms when no business logic is needed. Any retained wrapper must visibly add one of the semantic fields listed above.

- [ ] **Step 3: Verify**

Run:

```bash
cd code/backend
go generate ./internal/module/practice/application/commands
gofmt -w internal/module/practice/application/commands/*.go
go test -timeout=120s ./internal/module/practice/application/commands
go test -timeout=120s ./internal/module/practice/...
```

- [ ] **Step 4: Commit**

```bash
git add code/backend/internal/module/practice/application/commands
git commit -m "refactor(practice): goverter迁移人工复核响应映射"
```

### Task 4: Challenge Domain Remaining Pure Response Mapping

**Files:**
- Modify: `code/backend/internal/module/challenge/domain/mappers.go`
- Modify: `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- Modify: `code/backend/internal/module/challenge/domain/response_mapper_goverter_gen.go`

- [ ] **Step 1: Add direct mapper methods**

Add base or pointer methods for:

- `ImageRespFromModel`
- pure parts of `ChallengeRespFromModel`
- pure parts of `TeacherSubmissionWriteupItemRespFromRecord`
- pure parts of `TeacherSubmissionWriteupDetailRespFromRecord`

Keep wrappers for:

- `FormatImageSize`
- `Hints` list construction
- `ContentPreview`
- embedded `SubmissionWriteupResp`
- student/challenge metadata from record joins

- [ ] **Step 2: Remove all pass-through wrappers**

Run:

```bash
rg -n "mapped := challengeResponseMapperInst|return &mapped" internal/module/challenge/domain
```

Expected: only wrappers with extra assignments remain.

- [ ] **Step 3: Verify and commit**

Run:

```bash
go generate ./internal/module/challenge/domain
gofmt -w internal/module/challenge/domain/*.go
go test -timeout=120s ./internal/module/challenge/...
git add code/backend/internal/module/challenge/domain
git commit -m "refactor(challenge): goverter迁移剩余纯响应字段映射"
```

### Task 5: Challenge Topology and Package Revision Mapping

**Files:**
- Modify: `code/backend/internal/module/challenge/domain/topology_codec.go`
- Create or modify: `code/backend/internal/module/challenge/domain/response_mapper_goverter.go`
- Modify generated mapper file

- [ ] **Step 1: Evaluate JSON boundary**

Keep manual logic where conversion parses JSON or can return errors:

- `TopologyRespFromModel`
- `TemplateRespFromModel`

Migrate only direct sub-objects and `ChallengePackageRevisionRespFromModel` if it is pure field copy.

- [ ] **Step 2: Verify**

Run:

```bash
cd code/backend
go generate ./internal/module/challenge/domain
gofmt -w internal/module/challenge/domain/*.go
go test -timeout=120s ./internal/module/challenge/domain
go test -timeout=120s ./internal/module/challenge/...
```

- [ ] **Step 3: Commit**

```bash
git add code/backend/internal/module/challenge/domain
git commit -m "refactor(challenge): goverter迁移拓扑响应纯字段映射"
```

### Task 6: Contest Commands Semantic Wrappers Audit

**Files:**
- Modify: `code/backend/internal/module/contest/application/commands/response_mappers.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_response_mappers.go`
- Modify: `code/backend/internal/module/contest/application/commands/response_mapper_goverter.go`
- Modify generated mapper file

- [ ] **Step 1: Classify remaining wrappers**

Keep these wrappers if they still contain semantics:

- `contestRespFromModel`: contest time normalization
- `contestChallengeRespFromModel`: challenge metadata join
- `contestAWDServiceRespFromModel`: runtime config parsing, snapshot metadata, validation state, preview parsing
- `teamRespFromModel`: member count
- `awdTeamServiceRespFromModel`: team/service names, check result parse
- `awdAttackLogRespFromModel`: attacker/victim names, source normalization

- [ ] **Step 2: Migrate only pure pointer/base methods**

Add pointer mapper methods only where they remove a wrapper completely. Do not keep `mapped := ...; return &mapped`.

- [ ] **Step 3: Verify and commit**

Run:

```bash
cd code/backend
go generate ./internal/module/contest/application/commands
gofmt -w internal/module/contest/application/commands/*.go
go test -timeout=120s ./internal/module/contest/application/commands
go test -timeout=120s ./internal/module/contest/...
git add code/backend/internal/module/contest/application/commands
git commit -m "refactor(contest): 收敛命令层响应映射包装"
```

### Task 7: Contest Query Result Mapping

**Files:**
- Modify: `code/backend/internal/module/contest/application/queries/contest_result.go`
- Modify: `code/backend/internal/module/contest/application/queries/team_result.go`
- Create or modify query response mapper files if needed

- [ ] **Step 1: Inspect result structs**

Classify:

- `contestResultFromModel`
- `teamResultFromModel`

Keep wrappers only for derived fields such as `MemberCount`.

- [ ] **Step 2: Generate mapper if pure field surface is large**

Use package-local goverter mapper for `model.Contest -> ContestResult` and `model.Team -> TeamResult` base conversion.

- [ ] **Step 3: Verify and commit**

Run:

```bash
cd code/backend
go generate ./internal/module/contest/application/queries
gofmt -w internal/module/contest/application/queries/*.go
go test -timeout=120s ./internal/module/contest/application/queries
go test -timeout=120s ./internal/module/contest/...
git add code/backend/internal/module/contest/application/queries
git commit -m "refactor(contest): goverter迁移查询结果基础映射"
```

### Task 8: Runtime Response Mapping Review

**Files:**
- Modify: `code/backend/internal/module/runtime/application/commands/instance_service.go`
- Modify: `code/backend/internal/module/runtime/application/queries/instance_service.go`
- Modify existing runtime response mapper files

- [ ] **Step 1: Keep semantic wrappers**

Do not remove wrappers that compute:

- `RemainingExtends`
- visible status from expiration
- `Access` construction
- AWD access URL hiding
- remaining time

- [ ] **Step 2: Remove only redundant base wrappers**

If generated mapper supports pointer return directly, call it directly inside the semantic wrapper. Do not create an extra wrapper that only returns a mapper result.

- [ ] **Step 3: Verify and commit if changed**

Run:

```bash
cd code/backend
go generate ./internal/module/runtime/application/commands ./internal/module/runtime/application/queries
gofmt -w internal/module/runtime/application/commands/*.go internal/module/runtime/application/queries/*.go
go test -timeout=120s ./internal/module/runtime/...
git add code/backend/internal/module/runtime/application
git commit -m "refactor(runtime): 收敛实例响应基础映射"
```

### Task 9: Explicitly Retain Non-Mappable Helpers

**Files:**
- `code/backend/internal/module/identity/application/commands/support.go`
- `code/backend/internal/module/identity/application/queries/support.go`
- `code/backend/internal/module/ops/application/commands/notification_service.go`
- `code/backend/internal/module/ops/application/queries/notification_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/domain/package_topology_parser.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`

- [ ] **Step 1: Audit and document why retained**

Retain these only if they contain business logic:

- optional field normalization
- single-role list construction
- unread calculation
- content normalization
- map aggregation
- request DTO construction
- report export composition

- [ ] **Step 2: Avoid code churn**

Do not convert these just to claim full goverter coverage. If a helper is semantic and already concise, leave it.

- [ ] **Step 3: Commit documentation update only if tracked**

If updating scan documentation under a tracked path, commit separately:

```bash
git add docs/superpowers/plans/2026-05-02-goverter-final-migration.md
git commit -m "docs: 记录goverter剩余迁移计划"
```

### Task 10: Final Verification Sweep

**Files:**
- Whole backend.

- [ ] **Step 1: Scan for forbidden intermediate pattern**

Run:

```bash
cd code/backend
rg -n -U "mapped\\s*:=\\s*.*Mapper.*\\n\\s*return\\s*&mapped" internal/module --glob '*.go'
rg -n "func\\s+.*(ToDTO|ResultsToDTO|ResultToDTO)\\(" internal/module --glob '*.go'
```

Expected: no pure pass-through wrappers. Any hit must have visible business logic and be intentionally retained.

- [ ] **Step 2: Run broad verification**

Run:

```bash
go test -timeout=120s ./internal/module/...
go test -timeout=120s ./internal/app
git diff --check
```

- [ ] **Step 3: Commit final scan doc if needed**

```bash
git add <tracked-docs-if-any>
git commit -m "docs: 更新goverter迁移扫描结论"
```
