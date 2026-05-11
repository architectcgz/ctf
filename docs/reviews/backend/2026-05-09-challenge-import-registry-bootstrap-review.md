# Challenge Import Registry Bootstrap Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf-jeopardy-container-verification`
- Branch: `fix/jeopardy-container-verification-boundary`
- Diff source: working tree changes on 2026-05-09
- Files reviewed:
  - `code/backend/scripts/dev-run.sh`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_image_service_support.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_command_facade.go`
  - `code/backend/internal/module/challenge/runtime/module.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`

## Classification Check

- 认同本次变更为 non-trivial backend bugfix。
- 变更同时触达启动脚本、导入服务契约和测试，且直接影响真实平台注册链路。

## Gate Verdict

- Pass

## Findings

- 无 material findings。

## Material Findings

- None.

## Senior Implementation Assessment

- 当前方案把根因和契约边界分开收口，owner 比较清楚：
  - `dev-run.sh` 负责本地启动时把 registry env 补齐。
  - Jeopardy / AWD import service 负责把“缺少 image build service”翻译成 preview warning 和 commit `503`。
- 这种实现比只改 handler 错误文案更稳，因为即使未来有人绕过启动脚本直接起后端，导入链也不会再退化成模糊 `500`。
- AWD logger 通过 facade 定点注入，没有把 logger 依赖扩散到 query 或 core service，改动面控制得住。

## Required Re-validation

- `cd code/backend && go test ./internal/module/challenge/application/commands -run 'Test(PreviewChallengeImportWarnsWhenPlatformBuildServiceUnavailable|CommitChallengeImportReturnsServiceUnavailableWhenPlatformBuildServiceMissing|CommitChallengeImportReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing|AWDChallengeImportPreviewWarnsWhenPlatformBuildServiceUnavailable|AWDChallengeImportCommitReturnsServiceUnavailableWhenPlatformBuildServiceMissing|AWDChallengeImportCommitReturnsServiceUnavailableWhenExternalImageVerificationServiceMissing)' -count=1`
- `bash -n code/backend/scripts/dev-run.sh`
- 受控 PATH stub 启动 `code/backend/scripts/dev-run.sh`，确认 registry env 会进入子进程，且显式导出的 `CTF_CONTAINER_REGISTRY_SERVER` 不会被 env 文件覆盖。

## Residual Risk

- 当前 preview warning 只覆盖导入入口已知的 `platform_build` / `external_ref` 主路径；若后续再扩展更复杂的多镜像拓扑导入语义，应同步检查 warning 是否需要覆盖到节点级镜像来源。
- 本次 review 为同会话切换到 code review 心智完成，未使用独立 subagent。原因是当前 turn 未获得额外代理委派授权。

## Touched Known-debt Status

- 本次 touched surface 没有继续扩大已知结构债。
- 变更集中在已有 import service 和开发启动脚本上，没有把 warning / error contract 再次分散到 handler 或更外层兜底逻辑。
