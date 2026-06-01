# 2026-05-28 AWD service orphan backfill implementation plan

> 状态：Current
> 类型：实施过程

## 目标与非目标

- 目标：修复开发库中历史 `seed-teaching-review-data` 留下的 `awd_* .service_id` 孤儿引用。
- 非目标：不改动业务代码，不处理 `challenges.image_id = 0` 的历史哨兵值问题，不直接补 `awd_* .service_id` 物理 FK migration。

## 输入事实

- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/internal/module/contest/entity/contest_awd_service.go`
- 当前开发库 orphan 复查结果

## 任务切片

1. 备份当前开发库，并导出受影响行 before snapshot。
2. 为 `contest_id IN (42,43,44)` 补齐 `awd_challenge_id = 16` 的 `contest_awd_services` 父记录。
3. 把 `awd_team_services`、`awd_attack_logs`、`awd_traffic_events` 中指向伪造 `service_id = 7016` 的记录回填到对应赛事的新父记录 ID。
4. 复查 orphan 与行数，记录本次修复结果和剩余债务。

## 预期改动面

- 开发库 `ctf`
- `.harness/reuse-decisions/awd-service-orphan-backfill.md`
- `docs/reviews/backend/2026-05-28-awd-service-orphan-backfill-review.md`
- `docs/operations/2026-05-28-awd-service-orphan-backfill.md`

## 兼容性与风险

- 这是高风险数据库写操作，必须先执行实际 `pg_dump` 备份。
- 只允许更新 `contest_id IN (42,43,44)` 且 `service_id = 7016` 的历史样本记录，避免误伤其他赛事。
- `contest_awd_services` 受 `(contest_id, awd_challenge_id)` 唯一约束保护，因此父记录补写必须按赛事逐条映射，不得重复创建。

## 验证

- 备份文件实际生成成功
- before snapshot 实际生成成功
- 写入后 orphan 复查：
  - `awd_team_services.service_id = 0 orphan`
  - `awd_attack_logs.service_id = 0 orphan`
  - `awd_traffic_events.service_id = 0 orphan`
- `bash scripts/check-workflow-complete.sh`
