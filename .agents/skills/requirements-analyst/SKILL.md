---
name: requirements-analyst
description: Use before implementation when requirements are vague, risky, cross-module, or likely to hide edge cases and need explicit scope, assumptions, acceptance criteria, dependencies, and non-functional constraints.
---

# Requirements Analyst

本仓库里的这份 `requirements-analyst` 是项目补充入口，不再作为通用需求分析 workflow 的主体。

## 主体来源

- 通用主体：`~/.agents/skills/requirements-analyst`
- 项目补充：
  - `development-pipeline`
  - `code-reviewer`

## 在 CTF 仓库中的使用方式

需求不清、跨模块或容易藏 owner 漂移时：

1. 先使用全局 `requirements-analyst`
2. 再把产出落到本仓库事实源：
   - 需求基线：`docs/requirements/`
   - 实施计划：`docs/plan/impl-plan/`
   - 契约事实：`docs/contracts/`
   - 架构事实：`docs/architecture/`

## 本地补充关注点

- 对 `API / filter / sort / pagination`，要明确 `normalize / default / validate` 的唯一 owner
- 区分需求事实、实现计划和历史评审，不要把其中一个目录当成全部事实源
- 学生、教师、管理员三类角色的行为变化，要分别核对 `/academy/*` 与 `/platform/*` 的边界

## 项目内附带参考

- `docs/requirements/`
- `docs/plan/impl-plan/`
- `docs/contracts/`
- `docs/architecture/`
