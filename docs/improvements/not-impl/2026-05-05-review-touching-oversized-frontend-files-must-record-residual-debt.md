# Review touching oversized frontend files must record residual debt

## Status

not-impl

## Context

Observed on 2026-05-05 during the AWD defense content page review. The diff touched ContestAWDWorkspacePanel.vue, which is already listed in the frontend audit TD-1 oversized-component backlog, but the independent review archive initially concluded 'No material findings' without explicitly recording that known decomposition debt as residual risk. Durable improvement needed: when a review touches a known oversized frontend owner file, the archive should explicitly state whether the current slice reduced ownership risk, whether further decomposition remains, and whether the debt is blocking or intentionally deferred.

## Problem

当 review 命中已知的超大前端 owner 文件，却只检查当前功能 diff 的正确性而不显式记录结构债，后续读 review 结论的人会误以为该文件已经达到可长期承载新需求的状态。这样会弱化 review 的风险沟通能力，也会让“本次切片只是避免继续堆逻辑进去”和“该组件已经不需要继续拆分”这两件事混在一起。

## Suggested Direction

把这条规则沉淀到 `code-reviewer` / `development-pipeline` 相关流程里：当 diff 触达已在 audit、review backlog 或 architecture note 中列出的超大组件时，review 归档必须额外写清三件事：

1. 当前切片是否降低了 owner 混杂风险。
2. 原有拆分债是否仍然存在，以及是否阻塞本次合并。
3. 若不阻塞，本次为何只做最小切片而不继续展开结构重排。

## Target Owner

- skill: `code-reviewer`, `development-pipeline`
- agent: main coding agent policy
- docs: project `AGENTS.md` and review archive conventions when applicable
- code area: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`

## Evidence

- file: `docs/reviews/general/2026-05-05-awd-defense-content-page-review.md`
- file: `docs/reviews/frontend/ctf-frontend-audit-20260422.md`
- behavior: the review archive originally ended at `No material findings` and did not explicitly preserve the already-known `ContestAWDWorkspacePanel.vue` decomposition debt, even though that file was touched by the diff and remains listed in the frontend oversized-component backlog.

## Decision Log

- 2026-05-05: Created.
- 2026-05-05: Review archive was amended to restore the missing residual-risk note, but the durable review-policy rule is still not recorded in the underlying skill/agent guidance.
- 2026-05-05: Superseded in practice by `docs/improvements/agent-recorded/2026-05-05-touched-structural-debt-must-be-closed-in-pipeline.md` after the rule was recorded into project AGENTS and the pipeline/review skills.
