# 2026-05-28 seed awd service parent consistency implementation plan

> 状态：Current
> 类型：实施过程

## Plan Summary

- Objective
  - 修复 `seed-teaching-review-data` 生成 AWD 样本时遗漏 `contest_awd_services` 父记录的问题，避免继续制造 `service_id` 孤儿数据。
- Non-goals
  - 不处理 `image_id = 0` 的历史 nullable 语义迁移。
  - 不在本次把 seed 命令重构成独立应用 service。
- Source architecture or design docs
  - `code/backend/cmd/seed-teaching-review-data/main.go`
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
  - `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- Dependency order
  - 先补 seed 内部 parent builder，再回填子表写入，再补 reset 清理。
- Expected specialist skills
  - `go-backend`

## Task 1

- Goal
  - 扩展 AWD challenge catalog 与 seed helper，使脚本能构建真实 `contest_awd_services` 父记录。
- Touched modules or boundaries
  - `cmd/seed-teaching-review-data`
  - AWD challenge metadata -> seeded contest service snapshot/runtime config
- Dependencies
  - 正常 AWD service 创建路径的字段语义
- Validation
  - 编译 seed 命令
- Review focus
  - 父记录字段是否足够支撑现有 AWD 查询读取
- Risk notes
  - 不能再把 `service_id` 作为跨 contest 复用的常量

## Task 2

- Goal
  - 让 `AWDTeamService / AWDAttackLog / AWDTrafficEvent` 使用真实父记录 ID，并补强 reset 清理旧 service 行。
- Touched modules or boundaries
  - `seedStudentAWDScenario`
  - `resetSeededAWDData`
- Dependencies
  - Task 1
- Validation
  - 运行 seed 命令后检查 `contest_awd_services` 与三张子表是否一致
- Review focus
  - reset 后重跑是否仍保持幂等
- Risk notes
  - 需要确认显式删除 `contest_awd_services` 不与现有 FK/软删策略冲突

## Integration Checks

- seed 命令重跑后，教学 AWD 样本 contest 必须存在 `contest_awd_services` 父记录
- `awd_team_services / awd_attack_logs / awd_traffic_events` 不再出现缺失父记录的 `service_id`

## Rollback / Recovery Notes

- 仅修改 seed 命令和本地样本数据生成逻辑，回滚可直接恢复代码并重新 seed
- 若验证过程中重跑 seed 命令，会改写教学样本数据，不影响正式业务数据

## Residual Risks

- 现有开发库已经存在的 orphan 需要通过本次重跑 seed 或独立清理动作消除；代码修复本身不会自动修补线上/历史库
