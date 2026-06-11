# Challenge command service 拆分 Gate Review

日期：2026-06-11

## Review Target

- Repository: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-11-challenge-command-service-decomposition`
- Branch: `task/2026-06-11-challenge-command-service-decomposition`
- Task: `2026-06-11-challenge-command-service-decomposition`
- Plan: `docs/plan/impl-plan/2026-06-11-challenge-command-service-decomposition-implementation-plan.md`
- Diff source: 当前 worktree 未提交 diff
- Files reviewed:
  - `code/backend/internal/module/challenge/application/challengecatalog/published_catalog_event.go`
  - `code/backend/internal/module/challenge/application/challengecore/{service.go,input.go}`
  - `code/backend/internal/module/challenge/application/challengeimport/{service.go,package_revision.go}`
  - `code/backend/internal/module/challenge/application/challengeselfcheck/service.go`
  - `code/backend/internal/module/challenge/application/challengepublishcheck/service.go`
  - `code/backend/internal/module/challenge/application/challengepackageexport/{service.go,revision_service.go}`
  - `code/backend/internal/module/challenge/application/commands/package_delivery_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/module/challenge/runtime/wiring.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_package_storage.go`
  - `code/backend/internal/module/challenge/architecture_test.go`
  - `docs/architecture/backend/04-api-design.md`
  - `docs/architecture/backend/06-file-storage.md`
  - `docs/architecture/backend/07-modular-monolith-refactor.md`
  - `docs/architecture/features/题包Registry交付架构.md`

## Classification Check

同意当前任务按非琐碎结构性改动处理。该 diff 同时触达 challenge command application service 拆分、HTTP handler dependency surface、runtime wiring、后台 publish-check worker、LocalFS / zip storage ports、infrastructure adapters、架构 guard 和事实源文档。

## Gate Verdict

Pass with review-process limitation.

未发现仍需阻塞完成的 material finding。限制是：当前工具集没有可用的独立 `code-reviewer` subagent，本记录是同一实现上下文内按 `code-reviewer` skill 做的 self-check，不能等同真正独立 gate review。`code-workflow` 要求的独立 review gate 仍未满足。

## Findings

无剩余 material findings。

本轮 review 过程中已修正以下问题：

- 普通导入 service 拆到 `application/challengeimport` 后，`(*commands.ImageBuildService)(nil)` 作为接口传入会变成 typed nil interface，导致 preview 不产生“镜像构建/校验服务未启用” warning，commit 又从 tx store 泄出裸 `image build service is not configured`。已在 `ChallengeImportService` constructor / setter / 缺失判断中归一 typed nil，并用失败用例回归。
- 普通导入和 AWD 导入的 preview workspace 创建成功后，如果后续包解析或 preview record 保存失败，会留下本地 preview 目录。已在 `PreviewChallengeImport` 与 `PreviewImport` 中补充失败清理，并新增 `challenge_import_preview_cleanup_test.go` 覆盖普通/AWD 解析失败路径。
- 普通/AWD commit 的必要依赖检查已前移到文件副作用之前，避免配置错误时留下 imported image build source。
- `ChallengePackageStorage.PersistImportedImageBuildSource` 已先校验 Dockerfile/context 相对路径，再复制目录，避免先复制再发现路径越界。
- preview store 创建 workspace 后，如果写入 archive 或解包失败，会清理 workspace；unsafe archive 拒绝路径已有 infrastructure 测试覆盖。

## Material Findings

无。

## Senior Implementation Assessment

当前实现把 `ChallengeService` 缩回 `application/challengecore` 的普通题核心写路径，导入、自检、发布检查和题包导出分别由 `application/challengeimport`、`application/challengeselfcheck`、`application/challengepublishcheck`、`application/challengepackageexport` 拥有。发布目录变更事件的共享构造收口到 `application/challengecatalog`。HTTP handler 通过 `HandlerDeps` 接收明确依赖，runtime wiring 显式构造 core/import/self-check/publish-check/export/package-delivery 服务，publish-check worker 从 `ChallengePublishCheckService.RunPublishCheckLoop` 注册。

`application/commands` 不再作为普通题 core/import/self-check/publish-check/export 的生产 facade；它保留 image build、AWD command/import、package delivery 等既有写侧能力。`challenge_service_split_test_aliases_test.go` 只服务存量 commands 包测试编译，不作为生产依赖边界。

普通题包导入和题包导出的 LocalFS / zip 细节已移动到 `challenge/infrastructure` adapter，并通过 `ports` 暴露给 application service。`application/commands` 中仍有 AWD checker artifact 的既有本地文件持久化逻辑；它不属于本轮普通题包 preview/package storage 边界，但后续若继续收口 AWD checker artifact storage，应拆成单独端口处理，不能混回 core challenge command service。

## Required Re-validation

已执行并通过：

```bash
cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestPreviewChallengeImportDeletesWorkspaceWhenParseFails|TestAWDChallengeImportPreviewDeletesWorkspaceWhenParseFails|TestChallengeImport|TestAWDChallengeImport' -count=1
cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestPreviewChallengeImportWarnsWhenPlatformBuildServiceUnavailable|TestCommitChallengeImportReturnsServiceUnavailableWhenPlatformBuildServiceMissing|TestCommitChallengeImportReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing' -count=1
cd code/backend && go generate ./internal/module/challenge/api/http
cd code/backend && go test ./internal/module/challenge/application/commands -run 'TestChallengeService|TestChallengeImport|TestChallengeSelfCheck|TestChallengePublishCheck|TestChallengePackageExport|TestPackageDelivery|TestAWDChallengeImport' -count=1
cd code/backend && go test ./internal/module/challenge/infrastructure -run 'TestChallengeImportPreviewStore|TestAWDChallengeImportPreviewStore|TestChallengePackageStorage|TestChallengeAttachmentStore|TestChallengeCommandRepository' -count=1
cd code/backend && go test ./internal/module/challenge/api/http ./internal/module/challenge/runtime -run 'TestHandler.*Import|TestHandler.*PublishCheck|TestHandler.*SelfCheck|TestHandler.*PackageExport|TestBuildWiresChallengeImportImageBuildService|Test.*PublishCheck' -count=1
cd code/backend && go test ./internal/module/challenge/... -count=1
cd code/backend && go test ./internal/module -run 'TestModuleArchitectureBoundaries|TestBoundaryPackagesDoNotDependOnOuterLayers|TestModuleDependencyBaselineIsCurrent' -count=1
cd code/backend && go test ./internal/module -run TestMapperWrappersFollowGlobalDelegationPolicy -count=1
bash scripts/check-backend-architecture.sh --full
cd code/backend && go test ./internal/app -run 'TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime|TestFullRouter_AdminChallengePublishRequestLifecycle|TestFullRouter_TeacherCanBrowseArchivedAndDraftChallengesButOnlyManageOwnChallenges' -count=1
python3 scripts/check-docs-consistency.py
git diff --check
rg --files code/backend | rg 'service.*\.go$|.*_service\.go$' | rg -v '_test\.go$' | xargs wc -l | awk '$1 > 1000 && $2 != "total" {print}'
bash harness/workflow-plugins/code-workflow/run_workflow_stage.sh completion-full
```

## Residual Risk

- 真正独立 reviewer gate 未执行，原因是当前工具没有 subagent/独立 review agent 可用。该限制不影响本记录中的 self-check 结论，但不能把它计为 `code-workflow` 的独立完成门禁。
- `AWDChallengeImportService` 仍直接持久化 checker artifact 到 LocalFS。它是本轮之外的既有 AWD checker artifact 边界；当前 diff 没有继续扩大该边界，也没有让普通题包导入/导出回到 application filesystem helper。

## Touched Known-debt Status

- `ChallengeService` 过宽 owner：本轮已在 touched surface 收口，core service 不再拥有 import/self-check/publish-check/package-export 方法，普通题相关 service 也不再挤在同一个 `application/commands` Go package。
- 普通题包导入与题包导出的 application LocalFS / zip helper：本轮已迁移到 `ports + infrastructure`，并新增架构 guard 防止旧 helper 名称回流。
