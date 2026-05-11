# Challenge Import Registry Bootstrap & Error Handling Implementation Plan

## Objective

修复本地后端题包导入链里两处已经影响真实平台注册的边界问题：

- `code/backend/scripts/dev-run.sh` 启动本地后端时应自动加载并透传 registry 环境，避免本地 `air` / `go run` 进程遗漏 `CTF_CONTAINER_REGISTRY_*` 配置。
- Jeopardy / AWD 题包导入在需要镜像构建或外部镜像校验、但后端未装配 `ImageBuildService` 时，不再返回模糊 `500`，而是：
  - preview 给出明确 warning；
  - commit 返回 `503 Service Unavailable`；
  - 服务端记录 warning 日志，明确指出 registry 配置缺失。

## Non-goals

- 不修改题包 parser、registry build worker 或镜像状态机主流程。
- 不调整 `container.registry` 配置结构，也不改 Compose registry 编排。
- 不在本轮里补前端展示逻辑；本轮只保证后端 preview JSON 已返回 warning。
- 不顺手处理其他题包导入问题，例如 duplicate slug、拓扑同步或镜像构建成功后的轮询体验。

## Inputs

- `docs/plan/impl-plan/2026-05-05-challenge-package-registry-delivery-implementation-plan.md`
- `docs/reviews/general/2026-05-08-ctf-infra-registry-reconciliation-review.md`
- `code/backend/scripts/dev-run.sh`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/module/challenge/runtime/module.go`

## Current Facts

- 本地通过 `scripts/dev-run.sh --hot` / `--background` 启动时，只显式传入 `APP_ENV`、`CTF_CONTAINER_FLAG_GLOBAL_SECRET`、`CTF_FLAG_SECRET`，没有加载 `docker/ctf/infra/registry/ctf-platform-registry.env`，也没有透传现有 `CTF_CONTAINER_REGISTRY_*` 变量。
- `buildImageBuildService` 只有在 `container.registry.enabled` 或 `container.registry.build_enabled` 为真时才会装配 `ImageBuildService`。
- Jeopardy preview 只有在 `s.imageBuild != nil` 时才补 `target_image_ref/build_status`；否则静默退化。
- Jeopardy commit 在 `platform_build` / `external_ref` 路径下若 `s.imageBuild == nil`，直接返回普通 `fmt.Errorf("image build service is not configured")`。
- AWD preview / commit 也有同样问题。
- 当前真实平台注册里剩余失败题包都落在这条边界上：preview 成功，但 commit 因缺失 image build service 返回 `500`。

## Chosen Direction

本轮收口“registry 能力缺失”的 owner 为两层：

1. `dev-run.sh`
   - 作为本地启动入口 owner，负责在启动前加载 registry env 文件，并把 `CTF_CONTAINER_REGISTRY_*` 全量暴露给前台、热重载和后台子进程。
2. 导入服务
   - 作为导入契约 owner，负责把“当前后端未启用镜像构建/校验服务”的事实显式反映到 preview、commit 和日志，而不是让 handler 最后兜成通用 `500`。

这样做的原因：

- 启动脚本修的是根因，避免本地开发进程天然丢配置。
- 导入服务修的是契约边界，避免以后再次出现“配置缺了却只返回内部错误”的问题。
- 即使未来有人绕过脚本直接启动后端，导入链也还能给出可诊断的 warning/503。

## Ownership Boundary

- `code/backend/scripts/dev-run.sh`
  - 唯一负责本地后端开发启动前的 registry env 加载与透传。
- `ChallengeService`
  - 负责 Jeopardy preview/commit 对 image build service 缺失的 warning / error contract。
- `AWDChallengeImportService`
  - 负责 AWD preview/commit 对 image build service 缺失的 warning / error contract。
- `runtime/module.go`
  - 继续只负责按配置决定是否装配 `ImageBuildService`，不额外承担导入时的错误翻译。

## Change Surface

- Modify: `code/backend/scripts/dev-run.sh`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- Create: `docs/reviews/backend/2026-05-09-challenge-import-registry-bootstrap-review.md`

## Task Slices

### Slice 1: 脚本补齐 registry env 注入

目标：

- 自动读取 `docker/ctf/infra/registry/ctf-platform-registry.env`
- 已存在环境变量优先，env 文件只补空缺默认值
- 前台、`--hot`、`--background` 三条路径都能拿到 `CTF_CONTAINER_REGISTRY_*`

Validation:

- `bash -n code/backend/scripts/dev-run.sh`
- 受控 shell 下 `source` 脚本辅助逻辑或执行最小启动路径，确认 registry 变量已进入子进程环境

Review focus:

- 是否只在脚本内收口一次，避免三条启动分支重复拼装 env
- env 文件不存在时是否安全降级，不影响普通本地启动

### Slice 2: Jeopardy 导入缺服务 warning / 503

目标：

- preview 对 `platform_build` / `external_ref` 路径追加 warning
- commit 返回 `errcode.ErrServiceUnavailable`
- 日志写出缺失 service、题包 slug、镜像来源

Validation:

- `go test ./internal/module/challenge/application/commands -run 'TestPreviewChallengeImportWarnsWhenPlatformBuildServiceUnavailable|TestCommitChallengeImportReturnsServiceUnavailableWhenPlatformBuildServiceMissing|TestCommitChallengeImportReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing' -count=1`

Review focus:

- warning 与 commit 错误消息是否一致、可诊断
- 是否只在真正依赖 image build service 的导入路径上触发

### Slice 3: AWD 导入缺服务 warning / 503

目标：

- 与 Jeopardy 保持同一契约
- 如有需要，给 AWD service 注入 logger，避免静默失败

Validation:

- `go test ./internal/module/challenge/application/commands -run 'TestAWDChallengeImportPreviewWarnsWhenPlatformBuildServiceUnavailable|TestAWDChallengeImportCommitReturnsServiceUnavailableWhenPlatformBuildServiceMissing|TestAWDChallengeImportCommitReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing' -count=1`

Review focus:

- AWD preview / commit 是否与 Jeopardy 契约对齐
- logger 注入是否保持最小改动，不影响现有调用方

### Slice 4: 集成验证与 review

目标：

- 跑最小充分验证
- 归档独立 review，确认 touched surface 上不再残留“缺服务 -> 模糊 500”边界

Validation:

- `go test ./internal/module/challenge/application/commands -run 'Test(PreviewChallengeImportWarnsWhenPlatformBuildServiceUnavailable|CommitChallengeImportReturnsServiceUnavailableWhenPlatformBuildServiceMissing|CommitChallengeImportReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing|AWDChallengeImportPreviewWarnsWhenPlatformBuildServiceUnavailable|AWDChallengeImportCommitReturnsServiceUnavailableWhenPlatformBuildServiceMissing|AWDChallengeImportCommitReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing)' -count=1`
- `bash -n code/backend/scripts/dev-run.sh`
- `bash scripts/check-workflow-complete.sh`

Review focus:

- warning / error contract 是否已稳定
- 是否还有其他导入入口继续把缺失 service 暴露成裸 `500`

## Risks

- `dev-run.sh` 若直接 `source` env 文件，需避免覆盖用户手工导出的变量。
- AWD service 当前没有 logger 字段；若补 logger，构造函数和 facade 需要同步更新，但应避免扩大调用面。
- preview warning 只会在需要 image build service 的题包上出现，不能污染纯手工镜像题包。

## Verification Plan

1. `cd code/backend && go test ./internal/module/challenge/application/commands -run 'Test(PreviewChallengeImportWarnsWhenPlatformBuildServiceUnavailable|CommitChallengeImportReturnsServiceUnavailableWhenPlatformBuildServiceMissing|CommitChallengeImportReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing|AWDChallengeImportPreviewWarnsWhenPlatformBuildServiceUnavailable|AWDChallengeImportCommitReturnsServiceUnavailableWhenPlatformBuildServiceMissing|AWDChallengeImportCommitReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing)' -count=1`
2. `bash -n code/backend/scripts/dev-run.sh`
3. `bash scripts/check-workflow-complete.sh`

## Architecture-Fit Evaluation

- 目标 owner 明确：脚本负责配置注入，导入服务负责错误契约与日志，不把两者混在 runtime 装配层。
- 这次不是只把报错文案改漂亮，而是同时收口根因和导入边界，避免“脚本少环境变量 + 后端通用 500”再次组合出现。
- touched surface 上已知的 Jeopardy / AWD 双路径问题一并纳入，不留单侧 follow-up。
- 实现完成后不应再需要第二轮“把 AWD 也补上同样处理”的重复修补。
