# commit message detail enforcement 计划

## Objective

- 修改 `scripts/check-commit-message.sh`，阻止只有简短标题、没有详细正文的提交信息进入历史。
- 同步更新仓库 `AGENTS.md` 和 `.githooks/README.md`，明确提交正文要求。

## Non-goals

- 不调整 `pre-commit` 检查项。
- 不变更 merge / revert commit 的放行规则。
- 不引入额外 commit template 文件或第三方工具。

## Source Inputs

- `scripts/check-commit-message.sh`
- `.githooks/commit-msg`
- `.githooks/README.md`
- `AGENTS.md`

## Plan Review Result

- 直接在 `commit-msg` hook 校验正文是最小改动、最强约束的位置。
- 规则采用“标题 + 正文”结构，其中正文至少两行有效内容，避免只追加一句无信息量文本。

## Task Slices

### Slice 1: 强化 commit message 校验脚本

- 目标：新增正文存在性和最小详细度校验。
- 风险：
  - 规则过严会误伤正常提交；需要保留 merge / revert 豁免，并用多个示例覆盖。

### Slice 2: 同步项目规则与 hook 文档

- 目标：让 `AGENTS.md` 与 `.githooks/README.md` 说明新的提交流程。
- 风险：
  - 如果文档不更新，开发者仍会按旧的一行标题习惯提交。

### Slice 3: 手工验证脚本

- 目标：用临时 commit message 文件验证通过/失败场景。
- 风险：
  - 只跑 happy path 不足以证明校验收紧真的生效。

## Validation Plan

- `bash scripts/check-task-intake.sh --reuse-decision commit-message-detail-enforcement`
- `bash scripts/check-commit-message.sh <tmp-valid-message>`
- `bash scripts/check-commit-message.sh <tmp-invalid-message>`
- `bash scripts/check-consistency.sh`
- `git diff --check`

## Review Focus

- 标题规则是否保持兼容。
- 正文规则是否足够明确，且能实际拦住“只有简要信息”的提交。
- 文档说明是否与脚本一致。

## Rollback / Recovery

- 如果正文约束过严，可以放宽正文行数或字数门槛，但必须继续保留“不能只有简短标题”的硬约束。
