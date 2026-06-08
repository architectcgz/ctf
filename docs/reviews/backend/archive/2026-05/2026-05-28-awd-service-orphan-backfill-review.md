# 2026-05-28 AWD service orphan backfill review

## 范围

- 任务：修复开发库里历史 `awd_* .service_id` 孤儿引用。
- 受影响范围：
  - `contest_id IN (42,43,44)`
  - `awd_challenge_id = 16`

## Review 结论

- 本次数据修复已经把历史样本中的 `service_id = 7016` 定向回填到真实父记录：
  - `42 -> 33`
  - `43 -> 34`
  - `44 -> 35`
- 三张子表回填后，`awd_team_services`、`awd_attack_logs`、`awd_traffic_events` 的 `service_id` orphan 均为 `0`。
- 本次没有触碰 `challenges.image_id = 0` 的历史哨兵值，因为它已经不是简单 orphan，而是 schema / contract 兼容问题。

## Findings

- 无未修复 findings。

## 验证证据

- 实际整库备份已生成：
  - `backups/db/ctf-before-awd-service-backfill-20260528-171051.sql`
- before / after snapshot 已生成：
  - `backups/db/snapshots/awd-service-backfill-20260528-171051/`
- orphan 复查结果：
  - `awd_team_services.service_id = 0`
  - `awd_attack_logs.service_id = 0`
  - `awd_traffic_events.service_id = 0`
  - `challenges.image_id = 76` 仍保留，未纳入本次修复

## 门禁说明

- 本次 review 为同上下文自查记录，不计作 `development-pipeline` 所要求的独立 reviewer gate。
