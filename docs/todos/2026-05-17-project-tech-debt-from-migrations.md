# 迁移过程中识别出的项目技术债

更新时间：2026-06-11

本文只记录当前仍成立、或仍需保留为 follow-up 的项目技术债。已经被代码和架构事实收口的条目不再继续保留在活动 backlog 中。

## P2：模块迁移后的结构尾项

- [ ] `challenges.image_id = 0` 的历史 no-image 哨兵值还没有完成 schema / contract 清理，暂时不能继续补物理 FK。
  - 现状：当前 baseline `000001_init_schema.up.sql` 已吸收原 `000011_add_additional_foreign_keys` 里的 `audit_logs.user_id`、`awd_*` 的 creator / verifier / submitter、`instances` 主引用列、`submissions` 主引用列；`awd_* .service_id` 这批历史孤儿也已经在开发库回填完成，但 `challenges.image_id` 仍有 `76` 条历史 `image_id = 0` 记录。
  - 影响：`challenges.image_id` 继续缺少数据库级完整性约束，后续迁移和数据清理仍要保留逻辑层对 no-image 哨兵值的兼容。
  - 处理方向：先把 `image_id = 0` 的 schema / contract 语义清理完，再单独补 `image_id` 的 FK migration。
  - 依据：`code/backend/migrations/000001_init_schema.up.sql`

## 已核验并移出活动 backlog 的条目

- [x] recommendation / class review 的 live teaching fact 语义已统一到 teaching snapshot + `internal/teaching/advice` owner：challenge submission 不再限于 practice-only，学生 scoped `awd_attack_logs(source=submission)` 会进入 attempt / success 事实，`contest.flag_accepted` 也会触发 recommendation cache 失效和 profile 增量更新。
- [x] legacy `runtime` 模块已退役，runtime node / allocation 归 `container_runtime`，实例清理 / startup recovery / ACL 读写归 `instance` 与 app composition，AWD workspace / operation / scope / proxy traffic 归 `contest`；`moduleDependencyBaseline` 中已无 runtime 相关边。
- [x] runtime HTTP adapter 的 compat 形态已经收口到 `internal/app/composition/runtime_http_service_adapter.go` 单一 facade，旧的 parallel compat wrapper 与 `runtime_adapter_compat.go` 均已移除，不再保留为当前技术债。
- [x] `contest_awd_services.runtime_config.challenge_id` 历史兼容层已退场：`000015_remove_legacy_awd_runtime_config_challenge_id` 清理历史 key，query / response mapper 不再过滤该字段，新写路径测试继续保证不会重新持久化。
- [x] `assessment / ops` 事件消费边界按当前模块 baseline 已收口；ops 仅消费 challenge / contest / practice 事件，assessment 的剩余问题归入“练习 / 竞赛训练事实语义统一”而不是模块事件化边界。
- [x] 推荐链路的 `difficulty_band` 已成为推荐候选查询的一部分，不再保留为当前技术债。
- [x] 班级时间段 owner 已收口到共享 `classwindow`，不再保留为当前技术债。
- [x] 班级报告导出结构已纳入 `summary / trend / review / category_distribution / difficulty_distribution / contest_migration`，原表述过期。
- [x] `contest_challenges.awd_*` 历史字段不再存在，原表述过期。
- [x] application 层 GORM / Redis concrete allowlist 未全局清空这条，按当前代码状态已不成立。
- [x] 历史 `awd_* .service_id` 孤儿引用已在开发库回填到真实 `contest_awd_services` 父记录，不再保留为当前活动技术债。
