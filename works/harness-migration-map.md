# Harness Migration Map

## 迁移原则

严格参考 `deusyu/harness-engineering` 的顶层结构，但不破坏 CTF 原有业务事实源。

## 已迁移内容

- historical improvements -> `feedback/`
- CTF UI theme skill -> `prompts/ctf-ui-theme-system-skill.md`
- superpowers README and plans -> `practice/superpowers-plan-index.md`
- planning archive -> `practice/planning-archive-index.md`
- instance lifecycle research refs -> `references/ctf-instance-lifecycle-research.md`

## 保留在原路径的内容

- `docs/architecture/`：当前架构和最终设计事实源。
- `docs/contracts/`：API 和题包合同。
- `docs/plan/impl-plan/`：结构性实施计划。
- `docs/reviews/`：review 证据。
- 历史过程资料可通过 Git 历史追溯；当前 harness 入口以顶层目录为准。

## 未删除旧文件的原因

旧路径已按用户确认清理。当前迁移保留内容归档和索引，历史细节通过 Git 历史追溯。
