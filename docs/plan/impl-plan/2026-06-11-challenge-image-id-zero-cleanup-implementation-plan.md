<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# challenge-image-id-zero-cleanup Implementation Plan

**Goal:** 清理 `challenges.image_id = 0` 的历史 no-image 哨兵值，把普通题 `image_id` 收口成真正的 nullable 语义，并消除 challenge/practice runtime 投影对 `0` 的兼容依赖。

**Architecture:** CTF 后端模块化单体；owner 仍在 `challenge` 模块，`practice` 只消费 challenge runtime contract。

**Tech Stack:** Go, Gin, GORM, PostgreSQL, goverter, backend package/runtime tests

---

## Task Metadata

- Task Slug: `2026-06-11-challenge-image-id-zero-cleanup`
- Started At: `2026-06-11T11:19:44Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-challenge-image-id-zero-cleanup`
- Branch: `task/2026-06-11-challenge-image-id-zero-cleanup`

## Objective And Non-Goals

- Objective:
  - `challenges.image_id` 数据库存储只允许 `NULL` 或真实镜像 ID，不再写入 `0`
  - 普通题 authoring create/update/read contract 把无镜像表达为 `null`，不再对外暴露 `0`
  - `challenge -> practice` runtime subject projection 同步改成 nullable `image_id`
  - 追加 migration，把历史 `image_id = 0` 清理成 `NULL`
- Non-Goals:
  - 不处理 topology node `image_id = 0` 这条“继承默认镜像”的既有语义
  - 不在本任务里补 `challenges.image_id` 的物理 FK
  - 不改 AWD service/runtime_config 的 `image_id` 语义

## Inputs

- Source docs:
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
  - `docs/文档规范.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/contracts/openapi-v1/components/schemas/challenges.yaml`
  - `docs/contracts/api-contract-v1.md`
- Related prior work:
  - `docs/operations/2026-05-28-awd-service-orphan-backfill.md`
  - `docs/reviews/backend/archive/2026-05/2026-05-28-awd-service-orphan-backfill-review.md`

## Task Classification

- Classification: `非琐碎任务`
- Why:
  - 同时触达 schema、HTTP contract、challenge owner、practice runtime contract 和 migration
  - 需要明确 nullable owner，不能继续保留 `0` / `NULL` 双语义

## Files

- Create:
  - `code/backend/migrations/000017_cleanup_challenge_image_id_zero.up.sql`
  - `code/backend/migrations/000017_cleanup_challenge_image_id_zero.down.sql`
- Modify:
  - `code/backend/internal/module/challenge/entity/challenge.go`
  - `code/backend/internal/module/challenge/application/challengecore/*`
  - `code/backend/internal/module/challenge/contracts/challenge_core.go`
  - `code/backend/internal/module/challenge/contracts/contracts.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/*challenge*repository*.go`
  - `code/backend/internal/module/challenge/api/http/challenge_request_types.go`
  - `code/backend/internal/module/challenge/api/http/challenge_response_types.go`
  - `code/backend/internal/module/practice/entity/challenge.go`
  - `code/backend/internal/module/practice/infrastructure/runtime_subject_repository.go`
  - `code/backend/internal/module/practice/application/commands/*`
  - `code/backend/cmd/import-challenge-packs/main.go`
  - `docs/todos/2026-05-17-project-tech-debt-from-migrations.md`
- Review:
  - `code/backend/internal/module/challenge/application/challengeselfcheck/service.go`
  - `code/backend/internal/module/challenge/application/challengepackageexport/revision_service.go`
- Test:
  - `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
  - `code/backend/internal/module/challenge/api/http/*test.go`
  - `code/backend/internal/module/practice/infrastructure/runtime_subject_repository_test.go`
  - `code/backend/internal/module/practice/application/commands/*`

## 复用与 Owner 决策

- Existing patterns searched:
  - 现有 authoring challenge CRUD / query / runtime subject / import CLI 全链路 `ImageID` 使用
  - 现有 request mapper / goverter 生成面
- Reuse / extend / split / create-new decision:
  - 复用现有 challenge owner 和 practice runtime subject owner，不新增兼容 facade
  - 在 challenge update request 层补本地 tri-state nullable 解析，避免继续把 `0` 当清空协议
- Owner boundary:
  - `challenge` owner：authoring CRUD contract、持久化、import/export 读写
  - `practice` owner：消费 nullable runtime challenge，并在启动/runtime create 路径按 `nil` 判断是否存在默认镜像
  - migration owner：历史 `0 -> NULL` 数据清理
- Why this is the narrowest safe surface:
  - `challenge.image_id` 的 owner 本来就横跨 CRUD、query、runtime subject；如果只改一层会继续保留双语义
  - topology node `image_id` 与 AWD runtime_config 不在本次 no-image 哨兵清理范围内，保持不动

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 这是 schema / contract owner 清理，不是单点 bug；关键在于先分清普通题 image owner 与 topology/AWD image owner
- grill-with-docs findings:
  - `000001_init_schema.up.sql` 已允许 `challenges.image_id` 为 `NULL`，真实债在代码 contract
  - 普通题 create/update/read、query response、practice runtime subject 仍把 `0` 当“无镜像”
  - `UpdateChallenge` 不能直接用普通 `*int64` 承担 nullable patch，因为 JSON `null` 与字段缺省都映射成 `nil`
  - topology node `image_id = 0` 表示继承默认镜像，不能误伤
- Plan adjustments after challenge:
  - 不保留 `nil <-> 0` 兼容映射
  - runtime/practice contract 一并改成 nullable
  - update request 采用 tri-state nullable 字段解析，显式支持 `null` 清空

## Validation

- Commands:
  - `go test ./internal/module/challenge/application/commands -run 'TestService(CreateChallengeWithoutImageSuccess|Update.*Image)' -count=1`
  - `go test ./internal/module/challenge/api/http -run 'Test.*Challenge.*' -count=1`
  - `go test ./internal/module/practice/infrastructure -run TestRuntimeSubjectRepository -count=1`
  - `go test ./internal/module/practice/application/commands -run 'Test(StartChallenge|CreateSingleContainer|ContestInstance).*' -count=1`
  - `go generate ./internal/module/challenge/api/http ./internal/module/challenge/domain ./internal/module/challenge/application/queries`
  - `go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1`
- Manual checks:
  - 核对 create/update challenge 响应中的 `image_id` 对无镜像场景输出 `null`
  - 核对 migration 只清理 `challenges.image_id = 0`，不触碰 topology / AWD runtime_config
- Review focus:
  - `null` / 缺省 / 正整数三种输入语义是否只有一个 owner
  - practice runtime 无默认镜像时，单容器和拓扑分支是否都按 `nil` 正确判定
  - 是否还存在把 `0` 写回 `challenges.image_id` 的残留路径
