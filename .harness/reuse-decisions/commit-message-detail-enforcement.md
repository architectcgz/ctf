# Reuse Decision

## Change type
repo workflow / commit message detail enforcement

## Existing code searched
- `code/frontend/src/features`
- `code/frontend/src/api`
- `scripts/...`
- `.githooks/...`
- `docs/...`
- `scripts/check-commit-message.sh`
- `.githooks/commit-msg`
- `.githooks/README.md`
- `AGENTS.md`

## Similar implementations found
- 现有 `scripts/check-commit-message.sh` 只校验提交标题格式和中文描述，没有要求正文。
- `.githooks/README.md` 和 `AGENTS.md` 也只说明了单行标题约束，没有要求提交详情。

## Decision
refactor_existing

## Reason
当前最小正确切片是直接把详细提交说明收口到 `commit-msg` 校验脚本，并同步更新仓库规则说明：

- 标题继续使用 `英文类型(可选 scope): 中文描述`
- 普通提交额外要求正文
- 正文至少两行有效内容，避免只补一句空泛说明

这样可以：

- 从 механized hook 层面阻止只有简要标题的提交进入历史
- 让项目规则和 hook 文档保持一致

本轮不做：

- 不改 `pre-commit` 流程
- 不改 merge / revert 的豁免策略
- 不引入新的外部提交模板工具

## Files to modify
- `.harness/reuse-decisions/commit-message-detail-enforcement.md`
- `docs/plan/impl-plan/2026-05-31-commit-message-detail-enforcement-plan.md`
- `scripts/check-commit-message.sh`
- `AGENTS.md`
- `.githooks/README.md`

## After implementation
- commit-msg hook 会要求普通提交同时带规范标题和足够详细的正文说明。
- 项目文档会同步说明推荐使用多个 `-m` 组织提交信息。
