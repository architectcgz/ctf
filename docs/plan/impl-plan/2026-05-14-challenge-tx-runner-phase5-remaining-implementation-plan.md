# Challenge Tx Runner Phase5 Remaining Implementation Plan

**Goal:** 用 use-case-oriented tx runner / tx store 收口 challenge 模块剩余 4 条 application concrete import allowlist，一次移除 `challenge_import_service.go`、`awd_challenge_import_service.go`、`challenge_package_revision_service.go` 对 `gorm` / `clause` 的直接依赖。

## Objective

- 删除：
  - `challenge/application/commands/awd_challenge_import_service.go -> gorm.io/gorm`
  - `challenge/application/commands/challenge_import_service.go -> gorm.io/gorm`
  - `challenge/application/commands/challenge_package_revision_service.go -> gorm.io/gorm`
  - `challenge/application/commands/challenge_package_revision_service.go -> gorm.io/gorm/clause`

## Non-goals

- 不在这一轮里重做 challenge handler / facade 结构
- 不顺手改 `docs/design/backend-module-boundary-target.md`
- 不在没有必要的前提下扩大 raw repository 的全局错误语义
- 不承诺本轮解决长事务里的所有文件副作用中间态

## Inputs

- `.harness/reuse-decisions/challenge-tx-runner-phase5-remaining.md`
- `.harness/reuse-decisions/challenge-image-build-tx-store-phase5-slice42.md`
- `code/backend/internal/module/challenge/application/commands/{challenge_import_service.go,awd_challenge_import_service.go,challenge_package_revision_service.go}`
- `code/backend/internal/module/challenge/infrastructure/{repository.go,image_repository.go,topology_service_repository.go}`
- `code/backend/internal/module/challenge/runtime/module.go`

## Target Boundary

- application service
  - 负责：题包解析、文件编排、errcode 映射、流程控制
  - 不负责：直接持有 `*gorm.DB`、分支 `gorm.ErrRecordNotFound`、使用 `clause.OnConflict`
- runtime tx bridge
  - 负责：打开事务、组装 tx-only store、复用 raw repo / image build 能力
  - 不负责：业务判断、errcode 映射、题包解析
- infrastructure raw repo
  - 负责：底层 GORM source
  - 不负责：application use case 的 tx owner

## Planned Ports

1. `ChallengeImportedImageTxStore`
   - `ResolvePlatformBuildImage(...)`
   - `ResolveExternalImage(...)`
   - `ResolveExistingImageRef(...)`
2. `ChallengeImportTxRunner`
   - `WithinChallengeImportTransaction(ctx, func(store ChallengeImportTxStore) error) error`
3. `ChallengeImportTxStore`
   - `RejectImportedChallengeSlugConflict`
   - `FindLegacyChallengeForImportedPackageCreate`
   - `CreateImportedChallenge`
   - `UpdateImportedChallenge`
   - `ClearPublishCheckJobs`
   - `ReplaceImportedHints`
   - `ApplyImportedFlag`
   - `CreateImportedPackageRevision`
   - `UpsertImportedTopology`
   - 嵌入 `ChallengeImportedImageTxStore`
4. `AWDChallengeImportTxRunner`
   - `WithinAWDChallengeImportTransaction(ctx, func(store AWDChallengeImportTxStore) error) error`
5. `AWDChallengeImportTxStore`
   - `RejectImportedAWDChallengeSlugConflict`
   - `CreateImportedAWDChallenge`
   - 嵌入 `ChallengeImportedImageTxStore`
6. `ChallengePackageExportTxRunner`
   - `WithinChallengePackageExportTransaction(ctx, func(store ChallengePackageExportTxStore) error) error`
7. `ChallengePackageExportTxStore`
   - `FindChallenge`
   - `FindTopology`
   - `FindPackageRevisionByID`
   - `NextPackageRevisionNo`
   - `ListChallengeHints`
   - `FindImageRefByID`
   - `CreateExportRevision`
   - `MarkTopologyExported`

## Change Surface

- Add: `.harness/reuse-decisions/challenge-tx-runner-phase5-remaining.md`
- Add: `docs/plan/impl-plan/2026-05-14-challenge-tx-runner-phase5-remaining-implementation-plan.md`
- Add: `code/backend/internal/module/challenge/runtime/import_tx_bridge.go`
- Add: `code/backend/internal/module/challenge/runtime/awd_import_tx_bridge.go`
- Add: `code/backend/internal/module/challenge/runtime/package_export_tx_bridge.go`
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/application/commands/{challenge_import_service.go,awd_challenge_import_service.go,challenge_package_revision_service.go}`
- Modify: `code/backend/internal/module/challenge/runtime/module.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

## Ordered Tasks

1. 定义共享 imported image tx contract，以及三组 use-case tx runner / tx store。
2. 在 runtime 落三组 tx bridge，底层复用：
   - `challenge/infrastructure.Repository`
   - `challenge/infrastructure.ImageRepository`
   - `ImageBuildService`
3. 迁 `challenge_import_service.go` 到 `ChallengeImportTxRunner`。
4. 迁 `awd_challenge_import_service.go` 到 `AWDChallengeImportTxRunner`，复用 shared imported image contract。
5. 迁 `challenge_package_revision_service.go` 到 `ChallengePackageExportTxRunner`，顺手移除 `clause` 直接依赖。
6. 更新 runtime wiring、allowlist、测试与 review 归档。

## Validation

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/infrastructure -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf && bash scripts/check-consistency.sh
```

## Review focus

- tx bridge 是否只承载事务适配，没有重新变成宽业务 service
- application 是否已经完全不感知 `gorm` / `clause`
- shared imported image contract 是否足够窄，没有重新泄漏 raw repo
- `challenge_package_revision_service` 的 export baseline 更新是否仍保持当前行为
- 是否还残留长事务里明显可避免的业务分支或 ORM sentinel
