# Challenge Import Duplicate Slug Review

## Review Target

- Repository: `ctf`
- Worktree: `/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug`
- Branch: `feat/import-duplicate-slug`
- Review mode: independent subagent static review
- Diff source: current uncommitted worktree diff
- Files reviewed:
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
  - `code/backend/internal/app/challenge_import_integration_test.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
  - `code/backend/cmd/import-challenge-packs/main.go`
  - `code/backend/cmd/import-challenge-packs/main_test.go`
  - `code/frontend/src/features/challenge-package-import/model/useChallengePackageImport.ts`
  - `code/frontend/src/features/challenge-package-import/model/useChallengePackageImport.test.ts`
  - `code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportFlow.test.ts`
  - `code/docs/tasks/2026-03-31-challenge-package-import-design.md`
  - `docs/plan/impl-plan/2026-05-03-challenge-import-duplicate-slug-implementation-plan.md`

## Findings

1. `Major` [challenge_import_service.go](/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug/code/backend/internal/module/challenge/application/commands/challenge_import_service.go:151)
   - 普通题在线导入在进入事务前先执行 `persistImportedAttachmentBundle(parsed)`，而该函数会把附件直接写到 `imports/{parsed.Slug}` 目录：[challenge_import_service.go](/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug/code/backend/internal/module/challenge/application/commands/challenge_import_service.go:577)。
   - 随后的重复 slug 拒绝发生在事务内：[challenge_import_service.go](/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug/code/backend/internal/module/challenge/application/commands/challenge_import_service.go:159)。这意味着第二次导入同 slug 时，即使接口最终返回 409，已有题目正在引用的同名附件也可能已被新包覆盖。
   - 当前新增测试只验证数据库记录、发布检查任务和 HTTP 冲突响应未变化，没有覆盖“冲突后附件文件内容保持不变”这个关键副作用路径：[challenge_import_integration_test.go](/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug/code/backend/internal/app/challenge_import_integration_test.go:122)、[challenge_import_service_test.go](/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug/code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go:287)。
   - 这和本次“duplicate slug 直接拒绝、不再覆盖更新”的目标冲突，属于 correctness / data integrity 问题。

## Classification Check

- 结论：agree with leader/pipeline
- 说明：本次改动同时调整普通题导入、AWD 导入、CLI 导入语义，以及前端错误透传和测试预期，`non-trivial` 分类成立。

## Gate Verdict

- `blocked`
- 说明：只有一个 material finding，但它会让被拒绝的普通题重复导入仍然修改已有附件资产，不能按当前状态放行。

## Material Findings

- 普通题在线导入的附件持久化顺序错误：冲突检查前先落盘，导致返回 409 时文件系统仍可能被改写。

## Senior Implementation Assessment

- 总体方向是对的：把 duplicate slug 从“隐式覆盖更新”改成“明确拒绝”，并保留无 `package_slug` 的 legacy challenge 认领路径，边界更清晰。
- AWD 路径这次删除旧的 upsert 分支也合理，能减少未来维护噪音。
- 风险点只出现在普通题在线导入的副作用顺序。更低风险的实现应先完成 slug 冲突判定，再决定是否落附件；或者把附件写入临时位置，并在事务成功后再原子替换。

## Required Re-Validation

- 修复后至少重跑：
  - `go test ./internal/module/challenge/application/commands -run 'ChallengeImport'`
- 若 `code/backend/internal/app` 的既有 runtime 基线问题解除，再补：
  - `go test ./internal/app -run 'ChallengeImportCommitRejectsDuplicatePackageSlug'`
- 行为回归要点：
  - 重复 slug 且附件同名时，commit 返回 409 后原附件文件内容不变。
  - 现有“DB 不新增 / 不改标题分值 / 不清空 legacy publish check jobs”的断言仍成立。

## Residual Risk

- 本次归档基于独立 subagent 的静态 review，没有把测试结果作为当前结论的一部分。
- 已知无关基线问题：`code/backend/internal/app` 因 runtime 模块另一个未定义符号整体无法通过；该问题不计入本次 finding。
- 静态检查下，AWD 导入没有看到与普通题附件覆盖相同的副作用路径；CLI 路径也没有出现相同的在线附件持久化顺序问题。
