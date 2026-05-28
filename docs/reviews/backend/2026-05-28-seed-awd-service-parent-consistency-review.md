# 2026-05-28 seed AWD service parent consistency review

## 范围

- 任务：修复 `cmd/seed-teaching-review-data` 生成 AWD 复盘样本时 `service_id` 指向不存在父记录的问题。
- 主要改动：
  - `code/backend/cmd/seed-teaching-review-data/main.go`
  - `code/backend/cmd/seed-teaching-review-data/main_test.go`

## Review 结论

- 当前变更已经把 `AWDTeamService`、`AWDAttackLog`、`AWDTrafficEvent` 的 `service_id` 收口到真实创建的 `contest_awd_services.id`。
- `resetSeededAWDData` 也补上了 `contest_awd_services` 清理，避免重复 seed 后继续残留父表数据。
- 自查过程中发现一次测试回归：`main_test.go` 曾被错误覆盖成只剩新增测试。该问题已在同一回合修正，现已恢复原有测试并追加新的父子关系回归测试。

## Findings

- 无未修复 findings。

## 验证证据

- `go test ./cmd/seed-teaching-review-data -count=1`
  - 结果：通过

## 门禁说明

- 本次 review 为同上下文自查记录，不计作 `development-pipeline` 所要求的独立 reviewer gate。
- 当前会话工具策略不允许在未经用户明确授权的情况下拉起独立 reviewer agent，因此独立 review 门禁未满足；如需严格补齐，可在后续用户明确授权后再补一轮独立 review。
