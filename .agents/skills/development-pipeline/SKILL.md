---
name: development-pipeline
description: Use when engineering work is multi-step, cross-module, high-risk, or needs formal review gates, validation evidence, and disciplined branch finishing before handoff.
---

# Development Pipeline

本仓库里的这份 `development-pipeline` 是项目补充入口，不再作为通用工程流水线规则的主体。

## 主体来源

- 通用主体：`~/.agents/skills/development-pipeline`
- 项目补充：
  - `requirements-analyst`
  - `code-reviewer`
  - `runtime-ops-safety`

## 在 CTF 仓库中的使用方式

做跨模块、结构性或需要 review gate 的任务时：

1. 先使用全局 `development-pipeline`
2. 把计划、评审和验证证据落到仓库事实源：
   - 实施计划：`docs/plan/impl-plan/`
   - 评审记录：`docs/reviews/`
3. 收尾时按仓库守卫脚本补跑一致性与完成度检查

## 本地补充关注点

- touched surface 上已经存在的结构债，在本仓库默认按 blocker 处理，不能写成 follow-up 混过去
- 方案文档是编码前入口，不是事后归档材料；计划没收口前不要直接进实现
- 如果当前任务命中 reuse-first、文档规范或 review 归档规则，要让对应脚本和目录一起收口

## 项目内附带参考

- `references/`
- `docs/plan/impl-plan/`
- `docs/reviews/`
- `scripts/check_impl_plan_done.sh`
- `scripts/check-task-intake.sh`
- `scripts/check-consistency.sh`
