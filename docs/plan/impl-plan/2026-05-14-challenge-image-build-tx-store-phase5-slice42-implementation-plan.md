# Challenge Image Build Tx Store Phase5 Slice42 Implementation Plan

**Goal:** 移除 `challenge/application/commands/image_build_service.go` 对 `gorm.io/gorm` 的直接依赖，同时保持 challenge import / AWD import 的事务 owner 不变。

## Objective

- 删除 `challenge/application/commands/image_build_service.go -> gorm.io/gorm`

## Non-goals

- 不在这一刀里移除 `challenge_import_service.go` 或 `awd_challenge_import_service.go` 的事务 concrete
- 不处理 `challenge_package_revision_service.go` 的 `gorm` / `clause` 依赖
- 不改 `docs/design/backend-module-boundary-target.md`

## Inputs

- `.harness/reuse-decisions/challenge-image-build-tx-store-phase5-slice42.md`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/infrastructure/image_command_repository.go`

## Ownership Boundary

- `challenge/application/commands/image_build_service.go`
  - 负责：image build 的业务流程、状态迁移和 not-found 语义分支
  - 不负责：直接持有或操作 `*gorm.DB`
- `challenge/application/commands/challenge_import_service.go`
  - 负责：challenge import 的事务 owner，以及把当前 tx 包装成 image build 需要的窄 tx store
  - 不负责：承载 image build 业务流程
- `challenge/application/commands/awd_challenge_import_service.go`
  - 负责：复用同一 tx store 包装接入 AWD import
  - 不负责：重新实现 image build tx 细节

## Change Surface

- Add: `.harness/reuse-decisions/challenge-image-build-tx-store-phase5-slice42.md`
- Add: `docs/plan/impl-plan/2026-05-14-challenge-image-build-tx-store-phase5-slice42-implementation-plan.md`
- Modify: `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/image_build_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Modify: `code/backend/internal/module/architecture_allowlist_test.go`

## Tasks

1. 在 `image_build_service.go` 中定义 transaction-scoped 窄接口，覆盖：
   - image lookup/create
   - image build job create
   - image unscoped field update
2. 将 `CreatePlatformBuildJobInTx` / `VerifyExternalImageRefInTx` 改成消费该接口，而不是 `*gorm.DB`
3. 非 tx 路径中的 not-found 分支统一改成消费 `challengeports.ErrChallengeImageNotFound`
4. 在 challenge import application 文件中增加局部 tx wrapper，把 `*gorm.DB` 翻译成该窄接口
5. AWD import 复用同一 wrapper
6. 删除 `architecture_allowlist_test.go` 中这 1 条 allowlist，并补测试

## Validation

```bash
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/application/commands -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/challenge/runtime -run '^$' -count=1 -timeout 300s
cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module -run 'TestApplicationConcreteDependencyAllowlistIsCurrent' -count=1 -timeout 300s
```

## Review focus

- image build service 是否已经完全不感知 GORM concrete
- tx wrapper 是否只承载 transaction-scoped persistence 语义，没有把 image build 业务再塞回 import service
- not-found 语义是否统一为 `challengeports.ErrChallengeImageNotFound`
- 这刀是否只解决 image build concrete，不假装已经完成 challenge import / package revision 的 tx seam 迁移
