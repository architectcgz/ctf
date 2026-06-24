<!-- Managed by code-workflow package (version: 2026-06-06.6) -->
# challenge-package-delivery-gc Implementation Plan

**Goal:** 在 challenge 模块内建立题包交付编排边界，并提供第一版可审计的文件 / registry GC 候选分析能力。

**Architecture:** 本阶段不把所有文件读写重写成通用存储层，而是在 `challenge/application/commands` 增加题包交付 facade 与 artifact GC service。题包上传、commit、镜像构建和 registry 校验继续复用现有 parser、import service、image build service、registry client；GC 只做引用分析、dry-run 报告和本地文件安全删除能力，registry 物理 blob GC 留给后续运维维护窗。

**Tech Stack:** Go 1.26, Gin, GORM, LocalFS, Docker Registry HTTP API, existing challenge module ports/adapters.

---

## Task Metadata

- Task Slug: `2026-06-09-challenge-package-delivery-gc`
- Started At: `2026-06-09T01:59:56Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-09-challenge-package-delivery-gc`
- Branch: `task/2026-06-09-challenge-package-delivery-gc`

## Objective And Non-Goals

- Objective:
  - 建立 `PackageDeliveryService` 作为 Jeopardy / AWD 题包交付编排入口，先委托现有 `ChallengeService` 与 `AWDChallengeImportService`，不改变 HTTP API 行为。
  - 增加 `ArtifactGCService`，从现有 DB 字段和配置目录生成文件 / registry 候选报告。
  - 第一版 CLI 入口默认 dry-run，明确输出删除候选、保护原因和跳过原因。
- Non-Goals:
  - 不新增对外 HTTP API。
  - 不引入 `managed_artifacts` 表。
  - 不在第一阶段执行 registry blob garbage-collect。
  - 不重写题包 parser、镜像构建状态机或 runtime cleaner。

## Inputs

- Source docs:
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/architecture/features/题包Registry交付架构.md`
  - `docs/operations/awd-checker-runner.md`
- Related architecture/contracts:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/contracts/challenge-pack-v1.md`
- Related prior work:
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/image_build_service.go`
  - `code/backend/internal/module/challenge/infrastructure/registry_client.go`

## Task Classification

- Classification: `非琐碎任务`
- Why: 触达 challenge 模块 owner、题包导入、镜像构建、registry、LocalFS 清理和运维入口，属于跨文件结构性后端改动。

## Files

- Create:
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service.go`
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/artifact_gc_service.go`
  - `code/backend/internal/module/challenge/application/commands/artifact_gc_service_test.go`
  - `code/backend/internal/module/challenge/infrastructure/artifact_reference_repository.go`
  - `code/backend/internal/module/challenge/infrastructure/artifact_reference_repository_test.go`
  - `code/backend/cmd/storage-gc/main.go`
- Modify:
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/module/challenge/api/http/awd_challenge_handler.go`
  - `code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go`
  - `code/backend/internal/module/challenge/api/http/awd_challenge_handler_test.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/registry_client.go`
  - `code/backend/Dockerfile`
  - `docs/contracts/api-contract-v1.md`
- Review:
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/architecture/features/题包Registry交付架构.md`
- Test:
  - `go test ./internal/module/challenge/api/http -run 'TestHandler.*Import|TestAWDChallengeHandler.*Import' -count=1`
  - `go test ./internal/module/challenge/application/commands -run 'TestArtifactGC|TestPackageDelivery' -count=1`
  - `go test ./internal/module/challenge/infrastructure -run 'TestArtifactReferenceRepository|TestRegistryClient' -count=1`
  - `go test ./cmd/storage-gc -count=1`

## 复用与 Owner 决策

- Existing patterns searched:
  - challenge command services own business orchestration.
  - registry HTTP detail lives in `challenge/infrastructure/registry_client.go`.
  - runtime cleaner only handles containers/networks/ports/subnets.
- Reuse / extend / split / create-new decision:
  - Create a narrow facade for package delivery and a separate GC service.
  - Extend registry client only for manifest deletion support; do not place registry GC in application code yet.
- Owner boundary:
  - `challenge` owns题包、附件、build source、checker artifact 和 platform build registry manifest lifecycle.
  - `runtime` remains runtime resource cleanup only.
- Why this is the narrowest safe surface:
  - Existing import/build behavior remains delegated to current services.
  - GC starts as dry-run candidate analysis, so behavior risk is constrained and testable.

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
- Why this pass fits:
  - 需求是新生命周期服务设计，关键风险在 owner 边界、清理策略和 registry 破坏性操作。
- grill-with-docs findings:
  - 代码与文档都指向 challenge 模块是题包 / 附件 / 镜像 owner。
  - `docs/architecture/features/题包Registry交付架构.md` 已定义 image build service 与 registry client 的职责，GC 不应把 registry HTTP 细节拉回 application。
  - runtime cleaner 已有独立职责，不能扩展为文件 / registry 管理层。
- Plan adjustments after challenge:
  - 第一阶段不做 registry blob GC execute。
  - 第一阶段不建 `managed_artifacts` 表。
  - 先用现有 DB 字段和目录扫描生成候选报告。

## Validation

- Commands:
  - `go test ./internal/module/challenge/api/http -run 'TestHandler.*Import|TestAWDChallengeHandler.*Import' -count=1`
  - `go test ./internal/module/challenge/application/commands -run 'TestArtifactGC|TestPackageDelivery' -count=1`
  - `go test ./internal/module/challenge/infrastructure -run 'TestArtifactReferenceRepository|TestRegistryClient' -count=1`
  - `go test ./cmd/storage-gc -count=1`
  - `go test ./internal/module/challenge/... -count=1`
  - `bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full`
- Manual checks:
  - Run storage GC command in dry-run mode against an empty temp-like root where practical.
- Review focus:
  - 是否误删仍被 DB 引用的文件或 registry manifest。
  - 是否出现 application 层直接拼 registry HTTP 细节。
  - 是否把 runtime cleaner 和 artifact lifecycle 混成同一个 owner。

## Task Checklist

- [x] Write failing tests for `ArtifactGCService` candidate protection and root-boundary checks.
- [x] Run focused tests and confirm RED.
- [x] Implement `ArtifactGCService` minimal dry-run candidate analysis and safe file delete primitive.
- [x] Run focused tests and confirm GREEN.
- [x] Add `PackageDeliveryService` facade that delegates existing Jeopardy / AWD preview and commit flows.
- [x] Add focused tests for facade delegation.
- [x] Add `storage-gc` CLI entry with dry-run default.
- [x] Extend Dockerfile to build the new CLI binary.
- [x] Run focused challenge command tests.
- [x] Run module-level challenge tests.
- [x] Add configured-root delete guard and focused regression test.
- [x] Fix review blocker: protect active image build source parent candidates.
- [x] Fix review blocker: route HTTP import preview/commit through `PackageDeliveryService`.
- [x] Run workflow completion stage before review fixes.
- [x] Run workflow completion stage after review fixes.
- [x] Run independent review gate.
