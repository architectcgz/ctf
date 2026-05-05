# Touched structural debt must be closed in pipeline

## Status

agent-recorded

## Context

Recorded on 2026-05-05 after tightening pipeline and review prompts. When a task touches a previously tracked structural-debt surface such as an oversized owner-mixed frontend component, the debt is now mandatory scope for the same pipeline. Intake must classify it, the implementation plan must define debt-closure criteria, and final review must block if the touched debt remains. This rule was recorded into /home/azhi/workspace/projects/AGENTS.md and the global skills development-pipeline and code-reviewer.

## Problem

如果 pipeline 允许“我已经知道这里有结构债，但这次先把功能塞进去，后面再拆”，那 review 只会越来越像补记账，而不是风险闸门。对于超大组件、owner 混杂页面、已知待拆模块，这种处理会让每次变更都继续增厚同一块高风险表面，最后既难验证也难拆分。

## Suggested Direction

把规则前移到 intake、plan review 和 final review 三个阶段：

1. intake 必须识别本次是否触达已知结构债表面。
2. implementation plan 必须写明该债务在本次流水线中的收口标准，而不是只写 follow-up。
3. final review 若发现 touched debt 仍然存在，直接判 blocker，不允许降级成 residual risk。

## Target Owner

- skill: `development-pipeline`, `code-reviewer`
- agent: main coding agent policy
- docs: `/home/azhi/workspace/projects/AGENTS.md`
- code area: touched oversized or owner-mixed surfaces such as `ContestAWDWorkspacePanel.vue`

## Evidence

- file: `/home/azhi/workspace/projects/AGENTS.md`
- file: `/home/azhi/.codex/skills/development-pipeline/SKILL.md`
- file: `/home/azhi/.codex/skills/development-pipeline/references/review-gates.md`
- file: `/home/azhi/.codex/skills/development-pipeline/references/stage-definitions.md`
- file: `/home/azhi/.codex/skills/development-pipeline/references/task-slicing-rules.md`
- file: `/home/azhi/.codex/skills/code-reviewer/SKILL.md`
- behavior: the pipeline and review prompts now require touched known structural debt to be treated as in-scope work and block completion if it remains unresolved.

## Decision Log

- 2026-05-05: Created.
- 2026-05-05: Rule recorded into project AGENTS and global pipeline/review skills.
