# 2026-05-28 AWD service orphan backfill

## 背景

历史 `seed-teaching-review-data` 在生成 AWD 教学复盘样本时，曾直接把 `awd_team_services`、`awd_attack_logs`、`awd_traffic_events` 的 `service_id` 写成伪造值 `7016`，但没有创建对应的 `contest_awd_services` 父记录。

P1 已经修复 seed 脚本，不再继续制造新孤儿；本次操作负责修复当前开发库中已存在的历史残留。

## 操作范围

- 数据库：开发库 `ctf`
- 赛事范围：`contest_id IN (42,43,44)`
- 题目范围：`awd_challenge_id = 16`

## 备份与快照

- 整库备份：
  - `backups/db/ctf-before-awd-service-backfill-20260528-171051.sql`
- before snapshot：
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/awd_attack_logs.csv`
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/awd_team_services.csv`
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/awd_traffic_events.csv`
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/contest_awd_services.csv`
- after snapshot：
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/awd_attack_logs_after.csv`
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/awd_team_services_after.csv`
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/awd_traffic_events_after.csv`
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/contest_awd_services_after.csv`

## 执行动作

1. 为 `contest_id = 42, 43, 44` 分别补写 `contest_awd_services` 父记录，均指向 `awd_challenge_id = 16`。
2. 生成的新父记录 ID：
   - `42 -> 33`
   - `43 -> 34`
   - `44 -> 35`
3. 回填子表：
   - `awd_team_services`：`17` 行
   - `awd_attack_logs`：`12` 行
   - `awd_traffic_events`：`17` 行

## 验证结果

- `awd_team_services.service_id` orphan：`0`
- `awd_attack_logs.service_id` orphan：`0`
- `awd_traffic_events.service_id` orphan：`0`

## 剩余问题

- `challenges.image_id` 仍有 `76` 条历史 `image_id = 0` 的 no-image 哨兵值，不适合直接补物理 FK。
- 这条需要后续按 schema / contract 清理线单独处理。

## 回退方式

如需整库回退，可使用：

```bash
docker exec -i ctf-postgres psql -U postgres -d postgres -c "DROP DATABASE ctf;"
docker exec -i ctf-postgres psql -U postgres -d postgres -c "CREATE DATABASE ctf;"
docker exec -i ctf-postgres psql -U postgres -d ctf < backups/db/ctf-before-awd-service-backfill-20260528-171051.sql
```
