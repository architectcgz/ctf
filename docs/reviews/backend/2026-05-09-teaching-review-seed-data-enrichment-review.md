# 2026-05-09 教学复盘样本数据补量 Review

## Scope

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `docs/plan/impl-plan/2026-05-08-teaching-review-seed-data-implementation-plan.md`

## Findings

无 blocker。

## Checked Points

- 样本班级从 6 名学生扩到 7 名学生，并补了更明确的画像标签，命令输出可以直接对照“稳定闭环 / 高试错 / 低活跃 / AWD 迁移”等场景。
- 练习样本补到了 12 个 `practice session`、33 条 `audit_logs`、18 条 `submissions`、3 条 `writeups`、42 条 `skill_profiles`，足以稳定驱动班级复盘与个人归档分化。
- 新增 1 组独立的 AWD 样本赛事，包含 `1 contest / 1 round / 2 teams / 2 attack logs`，成功覆盖 `awd_participation` 观察分支。
- 班级复盘现在能稳定打出 5 类核心结论：`activity_risk`、`weak_dimension_cluster`、`training_closure_gap`、`retry_cost_high`、`trend_watch`。
- AWD 样本清理走专用 title 前缀，并对 `teams / contests` 使用 `Unscoped` 硬删除，避免软删除残留把 6 位 `invite_code` 唯一键卡死，seed 复跑边界保持明确。

## Residual Risks

- 当前开发库公开题目仍然偏少，推荐题会重复命中同一批 published challenge；它足够做论文演示，但不代表更大题库下的推荐分布。
- AWD 迁移样本目前只覆盖 1 名学生；如果后续需要展示班级级别的 AWD 聚类结论，还需要再补 1 到 2 名 AWD 参与学生。
- 本轮没有补 HTTP/UI 端到端验证，证据仍以 seed 命令内部读取结果和数据库核对为主。

## Validation Evidence

- `go test ./cmd/seed-teaching-review-data ./internal/module/assessment/... ./internal/module/teaching_readmodel/...`
- `go run ./cmd/seed-teaching-review-data`
- `docker exec ctf-postgres psql -U postgres -d ctf ...` 核对 `users / instances / audit_logs / submissions / submission_writeups / skill_profiles / contests / awd_rounds / teams / awd_attack_logs`
