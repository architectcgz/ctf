# Challenge Command Service Decomposition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 challenge command 写侧从一个过宽 `ChallengeService` 收口为 core command、题包导入、题包导出、自检、发布检查五个清晰 application service，并把 LocalFS / zip 边界移到 ports + infrastructure。

**Architecture:** 继续保持 `challenge` 模块 owner，不拆 bounded context，不改变 HTTP API / contracts。`api/http` 只依赖明确的 application service interface；`application/commands` 只编排用例和依赖 `ports/domain/contracts`；LocalFS、zip 解包、preview record、附件复制、题包 revision source/export archive 全部由 infrastructure adapter 实现端口。

**Tech Stack:** Go 1.26, Gin handler, GORM-backed repositories, LocalFS storage adapters, existing challenge domain parser, existing module architecture guards.

---

## Task Metadata

- Task Slug: `2026-06-11-challenge-command-service-decomposition`
- Started At: `2026-06-11T00:00:00Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-challenge-command-service-decomposition`
- Branch: `task/2026-06-11-challenge-command-service-decomposition`

## Objective And Non-Goals

- Objective:
  - `ChallengeService` 只保留普通题目的核心写路径：`CreateChallenge`、`UpdateChallenge`、`DeleteChallenge`、必要的 `PublishChallenge`，以及发布目录变更事件的最小支持。
  - 新增 `ChallengeImportService`，负责 Jeopardy 题包 `Preview/List/Get/Commit`，对齐现有 `AWDChallengeImportService` 的独立 service 模式。
  - 新增 `ChallengeSelfCheckService`，负责预检、Flag 校验、镜像可用性、拓扑 runtime request 构建、runtime probe 拉起和清理。
  - 新增 `ChallengePublishCheckService`，负责发布检查 job lifecycle、后台轮询、调用 self-check、通过后发布题目和发布通知事件。
  - 新增 `ChallengePackageExportService`，负责 `ExportChallengePackage` 与 `GetChallengePackageExport`。
  - 新增或拆细 LocalFS / zip ports，并由 `challenge/infrastructure` 实现 preview、附件、revision source、export archive 的文件读写。
- Non-Goals:
  - 不改变对外 HTTP 路由、请求/响应 JSON、OpenAPI 或用户可见行为。
  - 不迁移数据库 schema，不改题包 v1 contract。
  - 不引入对象存储、异步导入队列或新的部署组件。
  - 不保留“过渡期宽 facade”作为完成状态；允许测试过程中短暂存在，但最终 handler/runtime 不应通过一个宽 command interface 调所有 command use case。

## Inputs

- Source docs:
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/architecture/features/题包Registry交付架构.md`
  - `docs/plan/archive/impl-plan/2026-06/2026-06-09-challenge-package-delivery-gc-implementation-plan.md`
- Source code:
  - `code/backend/internal/module/challenge/application/commands/challenge_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - `code/backend/internal/module/challenge/runtime/wiring.go`
  - `code/backend/internal/module/challenge/runtime/module.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
- Guardrails:
  - `code/backend/internal/module/architecture_test.go`
  - `code/backend/internal/module/challenge/architecture_test.go`

## Task Classification

- Classification: `结构性改动`
- Why:
  - 触达 challenge command owner、runtime wiring、handler dependency surface、ports / infrastructure 边界、后台任务注册和多组测试。
  - 若只挪文件不拆依赖，会继续保留同一个 `ChallengeService` 过宽 owner，因此必须用结构 guard 和 focused tests 证明依赖面真正变窄。

## Intake Analysis Gate

- Analysis skills applied:
  - `brainstorming`：确认拆分目标不是单纯降行数，而是收口普通题 core/import/self-check/publish-check/package-export 的 use-case owner。
  - `grill-with-docs`：对照 `docs/architecture/backend/07-modular-monolith-refactor.md`、`docs/architecture/backend/06-file-storage.md` 和 `docs/architecture/features/题包Registry交付架构.md` 检查旧路径、文件存储边界和后台 job owner。
- Key intake conclusion:
  - 普通题导入、自检、发布检查和题包导出不应继续挂在 core `ChallengeService`。
  - LocalFS / zip / preview JSON / 附件复制属于 infrastructure storage adapter，不应留在 application service。
  - 后续包级优化必须避免在 `application/commands` 做生产 facade；runtime 和 handler 应直接依赖 owning use-case package。
- Gate result: `PASS`。实现按结构性任务进入 task worktree、implementation plan、startup gate 和完成验证链路。

## Clean Completion Criteria

- `ChallengeService` 不再有以下方法：`PreviewChallengeImport`、`ListChallengeImports`、`GetChallengeImport`、`CommitChallengeImport`、`SelfCheckChallenge`、`RequestPublishCheck`、`GetLatestPublishCheck`、`RunPublishCheckLoop`、`ExportChallengePackage`、`GetChallengePackageExport`。
- `challenge_import_service.go` 内的 service 类型为 `ChallengeImportService`，不再定义 `ChallengeService` struct / constructor。
- `api/http.Handler` 不再持有一个包含 core/import/self-check/publish-check/export 全部方法的宽 `challengeCommandService` interface。
- `api/http.Handler` 不再在 `NewHandler` 内自行构造 `PackageDeliveryService`；runtime wiring 显式传入已装配的 package delivery service。
- `application/commands` 中不再直接执行 zip 解包、preview JSON 持久化、附件复制、revision source 复制、export zip 生成等 LocalFS 细节；这些操作通过 `ports` 完成。
- 后台任务 `challenge_publish_check_worker` 由 `ChallengePublishCheckService.RunPublishCheckLoop` 注册。
- 现有 HTTP 行为、题包导入/导出、自检、发布检查生命周期回归测试保持通过。

## Files

- Create:
  - `code/backend/internal/module/challenge/application/commands/challenge_self_check_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_publish_check_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_package_export_service.go`
  - `code/backend/internal/module/challenge/application/commands/published_catalog_event.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_self_check_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_publish_check_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_package_export_service_test.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store_test.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store_test.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_package_storage.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_package_storage_test.go`
- Modify:
  - `code/backend/internal/module/challenge/application/commands/challenge_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_image_service_support.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
  - `code/backend/internal/module/challenge/runtime/wiring.go`
  - `code/backend/internal/module/challenge/runtime/module.go`
  - `code/backend/internal/module/challenge/runtime/module_import_test.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go`
  - `code/backend/internal/module/challenge/architecture_test.go`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/architecture/features/题包Registry交付架构.md`
- Test:
  - `go test ./internal/module/challenge/application/commands -run 'TestChallengeService|TestChallengeImport|TestChallengeSelfCheck|TestChallengePublishCheck|TestChallengePackageExport|TestPackageDelivery' -count=1`
  - `go test ./internal/module/challenge/infrastructure -run 'TestChallengeImportPreviewStore|TestChallengeAttachmentStore|TestChallengePackageStorage|TestChallengeCommandRepository' -count=1`
  - `go test ./internal/module/challenge/api/http -run 'TestHandler.*Import|TestHandler.*PublishCheck|TestHandler.*SelfCheck|TestHandler.*PackageExport' -count=1`
  - `go test ./internal/module/challenge/runtime -run 'TestBuildWiresChallengeImportImageBuildService|Test.*PublishCheck' -count=1`
  - `go test ./internal/module/challenge/... -count=1`
  - `go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1`
  - `go test ./internal/app -run 'TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime|TestFullRouter_AdminChallengePublishRequestLifecycle|TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges' -count=1`
  - `bash scripts/check-backend-architecture.sh --full`

## 复用与 Owner 决策

- Existing patterns searched:
  - AWD side already uses `AWDChallengeService + AWDChallengeImportService + AWDChallengeCommandFacade` instead of making core AWD command service own import.
  - `ports.go` already has small capability interfaces: `ChallengeWriteRepository`、`ChallengeInstanceUsageRepository`、`ChallengePublishCheckRepository`、`ChallengeImportTxRunner`、`ChallengePackageExportTxRunner`、`ChallengeRuntimeProbe`。
  - challenge architecture tests already allow backend boundary string guards for package ownership and forbidden imports.
- Reuse / extend / split / create-new decision:
  - Reuse existing repository and tx runner contracts where they already express the right use-case boundary.
  - Create new storage ports only for LocalFS / zip concerns that are currently inside application code.
  - Split services by command use case, not by table.
- Owner boundary:
  - `ChallengeService` owns core challenge mutation and catalog-change event generation.
  - `ChallengeImportService` owns Jeopardy import orchestration and import commit transaction.
  - `ChallengeSelfCheckService` owns runtime self-check orchestration and cleanup.
  - `ChallengePublishCheckService` owns publish-check job lifecycle and background polling.
  - `ChallengePackageExportService` owns package export and export revision lookup.
  - `challenge/infrastructure` owns LocalFS, zip, preview record, attachment and package storage implementation details.
- Why this is the narrowest safe clean surface:
  - It keeps all behavior inside the existing `challenge` bounded context and preserves API contracts.
  - It removes the real wide dependency owner instead of only moving functions across files.
  - It brings the file/zip boundary into ports now, matching the user's "不要过渡" constraint.

## Target Dependency Shape

```text
api/http.Handler
  -> ChallengeService              # core Create/Update/Delete/Publish
  -> ChallengeImportService         # preview/list/get/commit
  -> ChallengeSelfCheckService      # self-check endpoint
  -> ChallengePublishCheckService   # publish-check endpoint + worker
  -> ChallengePackageExportService  # export/download
  -> PackageDeliveryService         # mode switch, injected by runtime
  -> query ChallengeService

application/{challengecore,challengeimport,challengeselfcheck,challengepublishcheck,challengepackageexport}
  -> domain/contracts/ports only

application/commands
  -> existing image/AWD/package-delivery write use cases

infrastructure
  -> implements repository + LocalFS/zip storage ports

runtime
  -> constructs all concrete services and registers publish-check worker
```

## Task Checklist

### Task 1: Add structural guard for final service ownership

**Files:**
- Modify: `code/backend/internal/module/challenge/architecture_test.go`

- [x] **Step 1: Add a failing architecture test for clean service split**
  - Assert `challenge_import_service.go` contains `type ChallengeImportService struct`.
  - Assert `challenge_import_service.go` does not contain `type ChallengeService struct`.
  - Assert `challenge_service.go` does not declare import, self-check, publish-check or package-export methods on `*ChallengeService`.
  - Assert `api/http/handler.go` does not instantiate `NewPackageDeliveryService(commands, nil)`.

- [x] **Step 2: Run the guard and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge -run TestChallengeCommandServicesAreSeparated -count=1`
  - Expected: FAIL on current wide `ChallengeService` ownership.

### Task 2: Add LocalFS / zip storage ports and infrastructure adapters

**Files:**
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Create: `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store.go`
- Create: `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store_test.go`
- Create: `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
- Create: `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store_test.go`
- Create: `code/backend/internal/module/challenge/infrastructure/challenge_package_storage.go`
- Create: `code/backend/internal/module/challenge/infrastructure/challenge_package_storage_test.go`

- [x] **Step 1: Define narrow ports**
  - `ChallengeImportPreviewStore`: create upload workspace, save/load/list preview records, delete preview workspace.
  - `ChallengeAttachmentStore`: persist parsed import attachments and return the API-visible attachment URL.
  - `ChallengePackageStorage`: persist imported package source/archive and build exported package archives.
  - Keep app-facing values typed and boring; do not expose `zip.File`, `os.File`, `*gorm.DB`, Redis, Gin or raw HTTP types.

- [x] **Step 2: Write failing infrastructure tests**
  - Cover zip-slip rejection, symlink rejection, file count/size limits, root resolution, preview JSON round trip, attachment URL path cleanup, imported revision source copy, export archive creation.

- [x] **Step 3: Run focused store tests and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'TestChallengeImportPreviewStore|TestChallengeAttachmentStore|TestChallengePackageStorage' -count=1`
  - Expected: FAIL because adapters are not implemented yet.

- [x] **Step 4: Implement infrastructure adapters for LocalFS / zip semantics**
  - Implement current zip extraction, preview record JSON, attachment copy/zip, root env resolution, revision source/archive copy and export zip code in infrastructure adapters.
  - Application call sites switch to these ports in the import/export service split tasks; final cleanup still removes the old application helpers before completion.
  - Preserve current environment variable names and default roots.
  - Keep all path traversal and symlink checks at the infrastructure boundary.

- [x] **Step 5: Run focused store tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'TestChallengeImportPreviewStore|TestChallengeAttachmentStore|TestChallengePackageStorage' -count=1`
  - Expected: PASS.

### Task 3: Shrink core `ChallengeService`

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Create: `code/backend/internal/module/challenge/application/commands/published_catalog_event.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`

- [x] **Step 1: Write/adjust tests for core-only service**
  - Core service tests cover create/update/delete/publish and catalog-change events only.
  - Context tests for import/self-check/publish-check/export are moved to their owning service tests, not left on `ChallengeService`.

- [x] **Step 2: Run focused core tests and confirm current failures**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeService(Create|Update|Delete|Publish)|TestChallengeService.*Context|TestChallengeService.*ErrorContract' -count=1`
  - Expected: FAIL until constructor and moved methods are reconciled.

- [x] **Step 3: Move `ChallengeService` struct/constructor to core service and reduce dependencies**
  - Keep only `ChallengeWriteRepository`、`ChallengeInstanceUsageRepository`、`ImageQueryRepository`、`ChallengeTopologyReadRepository`、event publisher and logger dependencies required by core writes.
  - Remove import tx runner, package export tx runner, runtime probe, publish-check batch config and image build service fields from `ChallengeService`.

- [x] **Step 4: Extract published catalog event helper**
  - Put `publishedChallengeCatalogState` and weak event publication in `published_catalog_event.go`.
  - Let core service and import service both use this helper without making import depend on core service.

- [x] **Step 5: Run focused core tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeService(Create|Update|Delete|Publish)|TestChallengeService.*Context|TestChallengeService.*ErrorContract' -count=1`
  - Expected: PASS.

### Task 4: Create `ChallengeImportService`

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_image_service_support.go`
- Modify: `code/backend/internal/module/challenge/application/commands/package_delivery_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/package_delivery_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/tx_runner_test.go`

- [x] **Step 1: Rewrite import tests around `NewChallengeImportService`**
  - Service constructor receives preview store, attachment store, package storage, import tx runner, image build service, event publisher and logger.
  - Tests assert preview/list/get/commit behavior without reaching through `ChallengeService`.

- [x] **Step 2: Run import tests and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeImport|TestPackageDelivery' -count=1`
  - Expected: FAIL because service type and dependencies are not yet split.

- [x] **Step 3: Implement `ChallengeImportService`**
  - Move `PreviewChallengeImport`、`ListChallengeImports`、`GetChallengeImport`、`CommitChallengeImport` to `ChallengeImportService`.
  - Replace direct LocalFS/zip helpers with storage ports.
  - Keep import transaction logic in application service; keep transaction opening in existing tx runner bridge.
  - Keep image source resolution semantics unchanged.

- [x] **Step 4: Update `PackageDeliveryService` to depend on `ChallengeImportService` directly**
  - Jeopardy package delivery delegates to `ChallengeImportService`.
  - AWD package delivery continues to delegate to `AWDChallengeImportService`.
  - No package delivery path delegates through core `ChallengeService`.

- [x] **Step 5: Run import and package delivery tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeImport|TestPackageDelivery' -count=1`
  - Expected: PASS.

### Task 5: Create `ChallengePackageExportService`

**Files:**
- Create: `code/backend/internal/module/challenge/application/commands/challenge_package_export_service.go`
- Create: `code/backend/internal/module/challenge/application/commands/challenge_package_export_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`

- [x] **Step 1: Move package export tests to the new service**
  - Tests cover export revision creation, topology baseline checks, challenge mismatch forbidden result, and latest export fallback.

- [x] **Step 2: Run package export tests and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengePackageExport' -count=1`
  - Expected: FAIL until service exists and filesystem work moves through `ChallengePackageStorage`.

- [x] **Step 3: Implement `ChallengePackageExportService`**
  - Move `ExportChallengePackage` and `GetChallengePackageExport` off `ChallengeService`.
  - Depend on `ChallengePackageExportTxRunner`、`ChallengePackageRevisionRepository`、`ChallengeTopologyReadRepository`、core challenge lookup and `ChallengePackageStorage`.
  - Keep manifest/topology rewrite business shaping in application if it uses domain/contracts only; keep file copy/zip in infrastructure.

- [x] **Step 4: Run package export tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengePackageExport' -count=1`
  - Expected: PASS.

### Task 6: Create `ChallengeSelfCheckService`

**Files:**
- Create: `code/backend/internal/module/challenge/application/commands/challenge_self_check_service.go`
- Create: `code/backend/internal/module/challenge/application/commands/challenge_self_check_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`

- [x] **Step 1: Move self-check tests to the new service**
  - Tests cover failed precheck skips runtime, attachment-only skip, single container success, startup failure, invalid regex, manual review flag behavior and context propagation.

- [x] **Step 2: Run self-check tests and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeSelfCheck|TestChallengeSelfCheckService' -count=1`
  - Expected: FAIL until service exists.

- [x] **Step 3: Implement `ChallengeSelfCheckService`**
  - Move `SelfCheckChallenge`、`runPrecheck`、`validateFlagConfig`、`buildRuntimeFlag`、`buildTopologyRuntimeRequest`、`resolveAvailableImageRef` to the new service.
  - Dependencies: challenge read/write lookup, image query, topology read, runtime probe, self-check config, logger.
  - Do not import or call publish-check job repository.

- [x] **Step 4: Run self-check tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeSelfCheck|TestChallengeSelfCheckService' -count=1`
  - Expected: PASS.

### Task 7: Create `ChallengePublishCheckService`

**Files:**
- Create: `code/backend/internal/module/challenge/application/commands/challenge_publish_check_service.go`
- Create: `code/backend/internal/module/challenge/application/commands/challenge_publish_check_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_context_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_error_contract_test.go`

- [x] **Step 1: Move publish-check tests to the new service**
  - Tests cover request dedupe, stale latest job, dispatch success publish, dispatch failure keeps draft, attachment-only publish success, event payload and context propagation.

- [x] **Step 2: Run publish-check tests and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengePublishCheck|TestServiceDispatchPublishCheck|TestGetLatestPublishCheck' -count=1`
  - Expected: FAIL until service exists.

- [x] **Step 3: Implement `ChallengePublishCheckService`**
  - Move `RequestPublishCheck`、`GetLatestPublishCheck`、`RunPublishCheckLoop`、`dispatchPublishCheckJobs`、`processPublishCheckJob` and job response helpers.
  - Dependencies: `ChallengePublishCheckRepository`、challenge lookup, `ChallengeSelfCheckService` interface, `ChallengePublisher` interface backed by core `ChallengeService.PublishChallenge`, event publisher, config and logger.
  - Do not depend on runtime probe or topology repo directly.

- [x] **Step 4: Run publish-check tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengePublishCheck|TestServiceDispatchPublishCheck|TestGetLatestPublishCheck' -count=1`
  - Expected: PASS.

### Task 8: Wire explicit services through runtime and HTTP handler

**Files:**
- Modify: `code/backend/internal/module/challenge/runtime/wiring.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/challenge/runtime/module_import_test.go`
- Modify: `code/backend/internal/module/challenge/api/http/handler.go`
- Modify: `code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go`

- [x] **Step 1: Write failing handler/runtime wiring tests**
  - Handler import tests prove preview/commit route through injected `PackageDeliveryService`.
  - Publish-check and self-check handler tests prove calls go to separate service interfaces.
  - Runtime tests prove publish worker uses `ChallengePublishCheckService.RunPublishCheckLoop`.

- [x] **Step 2: Run handler/runtime tests and confirm RED**
  - Run: `cd code/backend && go test ./internal/module/challenge/api/http ./internal/module/challenge/runtime -run 'TestHandler.*Import|TestHandler.*PublishCheck|TestHandler.*SelfCheck|TestBuildWiresChallengeImportImageBuildService|Test.*PublishCheck' -count=1`
  - Expected: FAIL until wiring is explicit.

- [x] **Step 3: Replace handler wide command interface with explicit dependency struct**
  - `Handler` receives core command, query service, import service, self-check service, publish-check service, package export service and package delivery service.
  - `NewHandler` stops constructing application services internally.

- [x] **Step 4: Update runtime builders**
  - Build core service, import service, self-check service, publish-check service, package export service and package delivery service in `runtime/wiring.go`.
  - Construct LocalFS/zip storage adapters in runtime/infrastructure wiring.
  - Register background job from publish-check service.

- [x] **Step 5: Run handler/runtime tests and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge/api/http ./internal/module/challenge/runtime -run 'TestHandler.*Import|TestHandler.*PublishCheck|TestHandler.*SelfCheck|TestBuildWiresChallengeImportImageBuildService|Test.*PublishCheck' -count=1`
  - Expected: PASS.

### Task 9: Remove wide dependencies and stale helpers

**Files:**
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/challenge_command_repository.go`
- Modify: `code/backend/internal/module/challenge/architecture_test.go`

- [x] **Step 1: Remove old setters and fields**
  - Delete `SetChallengeImportTxRunner`、`SetChallengePackageExportTxRunner`、`SetImageBuildService` from `ChallengeService`.
  - Remove package export, import, self-check and publish-check fields from core service.

- [x] **Step 2: Narrow repository adapters where useful**
  - Keep `ChallengeCommandRepository` only if it now represents core writes plus running-instance guard.
  - If publish-check methods no longer travel with core writes, introduce a smaller infrastructure adapter or typed source dependency for `ChallengePublishCheckRepository`.

- [x] **Step 3: Run architecture guard and confirm GREEN**
  - Run: `cd code/backend && go test ./internal/module/challenge -run TestChallengeCommandServicesAreSeparated -count=1`
  - Expected: PASS.

### Task 10: Sync architecture facts

**Files:**
- Modify: `docs/architecture/backend/07-modular-monolith-refactor.md`
- Modify: `docs/architecture/backend/06-file-storage.md`
- Modify: `docs/architecture/features/题包Registry交付架构.md`

- [x] **Step 1: Update backend modular-monolith architecture facts**
  - Reflect final service ownership: core command, import, self-check, publish-check and package export are separate application services.
  - State that runtime wiring constructs explicit services and registers publish-check worker from `ChallengePublishCheckService`.

- [x] **Step 2: Update file storage facts**
  - Replace the old statement that `application/commands/challenge_import_service.go` owns preview directory and LocalFS implementation details.
  - State that application services own orchestration, while `challenge/infrastructure` LocalFS/zip adapters own upload extraction, preview records, attachments, package source and export archives.

- [x] **Step 3: Update registry delivery facts**
  - Keep image delivery semantics unchanged.
  - Adjust code落点 to point at `ChallengeImportService` and storage adapters after the split.

- [x] **Step 4: Run docs consistency check**
  - Run: `python3 scripts/check-docs-consistency.py`
  - Expected: PASS.

### Task 11: Full validation and review pass

**Files:**
- Review all touched challenge command/runtime/api/infrastructure files.

- [x] **Step 1: Run focused command validation**
  - Run: `cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeService|TestChallengeImport|TestChallengeSelfCheck|TestChallengePublishCheck|TestChallengePackageExport|TestPackageDelivery' -count=1`
  - Expected: PASS.
  - Result: PASS. Fresh run also included `TestAWDChallengeImport` and preview workspace cleanup regression tests.

- [x] **Step 2: Run focused infrastructure validation**
  - Run: `cd code/backend && go test ./internal/module/challenge/infrastructure -run 'TestChallengeImportPreviewStore|TestChallengeAttachmentStore|TestChallengePackageStorage|TestChallengeCommandRepository' -count=1`
  - Expected: PASS.
  - Result: PASS. Fresh run also included `TestAWDChallengeImportPreviewStore`.

- [x] **Step 3: Run focused handler/runtime validation**
  - Run: `cd code/backend && go test ./internal/module/challenge/api/http ./internal/module/challenge/runtime -run 'TestHandler.*Import|TestHandler.*PublishCheck|TestHandler.*SelfCheck|TestBuildWiresChallengeImportImageBuildService|Test.*PublishCheck' -count=1`
  - Expected: PASS.
  - Result: PASS. Fresh run also included `TestHandler.*PackageExport`.

- [x] **Step 4: Run challenge module validation**
  - Run: `cd code/backend && go test ./internal/module/challenge/... -count=1`
  - Expected: PASS.
  - Result: PASS.

- [x] **Step 5: Run architecture validation**
  - Run: `cd code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1`
  - Expected: PASS.
  - Run: `bash scripts/check-backend-architecture.sh --full`
  - Expected: PASS.
  - Result: PASS.

- [x] **Step 6: Run app integration smoke validation**
  - Run: `cd code/backend && go test ./internal/app -run 'TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime|TestFullRouter_AdminChallengePublishRequestLifecycle|TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges' -count=1`
  - Expected: PASS.
  - Result: PASS.

- [ ] **Step 7: Independent review gate**
  - Review focus:
    - No `ChallengeService` wide owner remains.
    - No LocalFS/zip implementation leaks back into application services.
    - Handler dependency surface is explicit and not a renamed wide command service.
    - Publish-check service does not directly know runtime probe/topology internals.
    - Import service owns commit orchestration but not file extraction mechanics.
    - Existing API behavior is preserved.
  - Result: same-context review record archived at `docs/reviews/backend/2026-06-11-gate-review-challenge-command-service-decomposition.md`; true independent reviewer gate remains unmet because no subagent/independent review tool is available in this session.

## Architecture-Fit Evaluation

- Target architecture boundary is explicit: all changes stay in `challenge` module and preserve `api -> application -> ports/domain` plus `infrastructure -> ports`.
- Shared layers and owners are named: core command, import, self-check, publish-check, package export, package delivery, LocalFS/zip adapters.
- This plan does not defer structural convergence: LocalFS/zip ports and handler/runtime dependency split are included in the same completion criteria.
- The main risk is scope size. To keep review manageable, each task creates a passing intermediate state with focused tests, but the final task is not complete until the wide service ownership and direct filesystem leakage are both removed.

### Task 12: Split challenge command services into use-case packages

**Why:** 文件级拆分已经去掉了宽 `ChallengeService`，但 `application/commands` 仍然是一个过宽 Go package。普通题 core/import/self-check/publish-check/export 之间还没有 package 边界，包内未导出 helper 仍可被任意 service 复用。这个后续切片把普通题 command 相关 service 拆成 use-case package，`commands` 只保留其他既有写侧 use case、package delivery 和 image build 等当前共享能力。

**Files:**
- Create/Move under `code/backend/internal/module/challenge/application/`:
  - `challengecore/`
  - `challengeimport/`
  - `challengeselfcheck/`
  - `challengepublishcheck/`
  - `challengepackageexport/`
  - `challengecatalog/` for published catalog event helper
- Modify runtime, handler imports and focused tests.

- [x] **Step 1: Add architecture guard for package split**
  - Assert core/import/self-check/publish-check/export service definitions no longer live in `application/commands`.
  - Assert `runtime/wiring.go` imports the new use-case packages directly.
  - Assert `application/commands` does not import those new use-case packages, avoiding facade-style back references.

- [x] **Step 2: Move shared catalog event helper**
  - Move `published_catalog_event.go` into `application/challengecatalog`.
  - Export only the state builder and weak event publication helpers needed by core/import.

- [x] **Step 3: Move core command service**
  - Move `ChallengeService`, `CreateChallengeInput`, `UpdateChallengeInput`, `ChallengeHintInput` into `application/challengecore`.
  - Update handler request mapper and tests to depend on `challengecore` input types.

- [x] **Step 4: Move import service**
  - Move `ChallengeImportService` and import-specific helpers into `application/challengeimport`.
  - Depend on a narrow local `ImageBuildService` capability interface that the existing image build service satisfies; do not create a production facade in `commands`.

- [x] **Step 5: Move self-check, publish-check and package export services**
  - Move self-check service and config into `application/challengeselfcheck`.
  - Move publish-check service and polling config into `application/challengepublishcheck`.
  - Move package export service and revision methods into `application/challengepackageexport`.

- [x] **Step 6: Update runtime, handlers, tests and docs**
  - Runtime constructs concrete services from their owning packages.
  - Handler interfaces use input/request/result types from the owning packages.
  - Architecture docs mention package-level owner split.

- [x] **Step 7: Re-run focused and completion validation**
  - Run focused command, handler/runtime, challenge module, architecture, docs, diff and `completion-full` checks.
  - Result: PASS. During `go test ./internal/module/challenge/... -count=1`, the service split exposed a typed nil `ImageBuildService` mismatch in ordinary import service construction; fixed by normalizing typed nil image build dependencies in `application/challengeimport`.

## Validation

- Planning-time validation:
  - `bash scripts/check-task-intake.sh`
  - `git diff --check`
- Implementation-time validation:
  - Commands listed in Task 11.
