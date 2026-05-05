# Challenge Import Duplicate Slug Rejection Implementation Plan

## Plan Summary

- Objective
  - 将普通题题目包导入、AWD 题目包导入，以及离线 `import-challenge-packs` CLI 的重复 slug 行为从“按 slug 覆盖更新”收紧为“明确拒绝导入并返回可读错误”。
- Non-goals
  - 不改题目编辑/更新链路。
  - 不引入新的导入记录表或新的导入状态机。
  - 不处理与 slug 无关的导入校验规则。
- Source architecture or design docs
  - [code/docs/tasks/2026-03-31-challenge-package-import-design.md](/home/azhi/workspace/projects/ctf/.worktrees/feat-import-duplicate-slug/code/docs/tasks/2026-03-31-challenge-package-import-design.md)
- Dependency order
  - 先补后端冲突判定与错误契约，再补前端错误透传与提示，最后做验证与 review。
- Expected specialist skills
  - `backend-engineer`
  - `test-engineer`
  - `doc-admin-agent`

## Task 1

- Goal
  - 收紧普通题在线导入与 CLI 导入的重复 slug 规则，slug 已存在时拒绝导入。
- Touched modules or boundaries
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `code/backend/cmd/import-challenge-packs/main.go`
  - 对应测试
- Dependencies
  - 依赖现有 `package_slug` 唯一约束和当前导入解析逻辑。
- Validation
  - `go test ./internal/module/challenge/application/commands -run 'ChallengeImport|ImportOnePack'`
- Review focus
  - 在线导入预览是否仍允许成功，冲突是否在 commit 阶段被明确拒绝。
  - CLI 是否不再静默覆盖已有题目。
- Risk notes
  - 现有部分测试建立在“slug upsert”旧语义上，需要同步改写。

## Task 2

- Goal
  - 收紧 AWD 题目包导入的重复 slug 规则，slug 已存在时拒绝导入。
- Touched modules or boundaries
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - 对应测试
- Dependencies
  - 复用 Task 1 的冲突语义，保持普通题和 AWD 导入口径一致。
- Validation
  - `go test ./internal/module/challenge/application/commands -run 'AWDChallengeImport'`
- Review focus
  - AWD 导入在冲突时是否返回清晰错误，而不是更新已存在题目。
- Risk notes
  - 不能影响 script checker artifact 等现有导入副作用路径。

## Task 3

- Goal
  - 前端导入页与 AWD 导入页在 commit 冲突时透传后端错误，给出“slug 已被占用”这类可读提示。
- Touched modules or boundaries
  - `code/frontend/src/features/challenge-package-import/model/useChallengePackageImport.ts`
  - `code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportFlow.ts`
  - 对应测试
- Dependencies
  - 依赖后端返回稳定错误消息。
- Validation
  - `npm run test -- --run src/features/challenge-package-import/model/*.test.ts src/features/platform-awd-challenges/model/*.test.ts src/api/__tests__/admin.test.ts`
- Review focus
  - 不要把已有可读错误吞掉后退化成“导入失败”。
- Risk notes
  - worktree 需先具备前端依赖，否则无法跑 `vitest`。

## Integration Checks

- 普通题在线导入：预览成功，commit 冲突时返回 409 与明确错误。
- AWD 在线导入：预览成功，commit 冲突时返回 409 与明确错误。
- CLI 导入：重复 slug 时返回错误并停止覆盖更新。
- 前端 toast / 上传结果文案能直接透传后端冲突原因。

## Review Gate Notes

- 本任务的独立 review 必须由单独 subagent 执行，不能直接在当前实现上下文里自审后给出“无问题”结论。
- 发起 review 时，只向 review agent 提供当前 diff、该 implementation plan、已执行验证与必要背景，避免把实现阶段推理过程一并带入。
- 若主会话已经先做过一次 review，不能把该结论当作最终 gate evidence；需要补做 subagent review，并以新的 review archive 作为完成依据。

## Rollback / Recovery Notes

- 后端和前端改动都可独立回退，无 schema 变更。
- 若冲突策略需要放宽，可仅回退导入服务和 CLI 的冲突判定，不影响题目编辑链路。

## Residual Risks

- 当前 `internal/app` 基线存在与本任务无关的编译断点，无法把整条 router integration 测试作为最终证据。
- 如果后续要把“重复 slug”前移到预览阶段，还需要扩展预览 JSON 或错误展示位，这次不处理。
