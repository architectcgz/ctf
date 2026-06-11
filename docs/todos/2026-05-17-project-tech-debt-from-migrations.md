# 迁移过程中识别出的项目技术债

更新时间：2026-06-10

本文只记录当前仍成立、或仍需保留为 follow-up 的项目技术债。已经被代码和架构事实收口的条目不再继续保留在活动 backlog 中。

## P2：教学评估与 AWD 统一仍有残余边界

- [ ] 竞赛 / AWD 数据已回流到维度事实，但 recommendation / class review 与 practice 语义仍未完全统一。
  - 现状：个人 recommendation 与班级 class review 的 teaching fact snapshot 已吸收 AWD 成功覆盖、profile score 补充信号与 solved difficulty 覆盖，`difficulty_band` 也已经进入推荐候选查询 owner；但推荐与复盘主链路里的 `attempt / approved review evidence` 仍主要沿用 `contest_id IS NULL` 的练习侧语义。
  - 影响：这条债已经从“竞赛数据没有回流”收敛到“主画像里的练习 / 竞赛事实还没有完全统一”，当前仍不是一份完全统一的训练画像。
  - 依据：`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`、`docs/architecture/features/教学复盘建议生成架构.md`

## P2：模块迁移后的结构尾项

- [ ] `challenges.image_id` 仍未补数据库级 FK。
  - 现状：`image_id = 0` 的历史 no-image 哨兵值已经完成 schema / contract 清理，`000017_cleanup_challenge_image_id_zero` 负责把存量 `0` 清成 `NULL`；当前剩余问题只是不带物理 FK。
  - 影响：`challenges.image_id` 仍缺少数据库级完整性约束，非法引用目前仍主要依赖应用层保证。
  - 处理方向：单独补 `challenges.image_id -> images.id` 的 FK migration，并评估软删 / 导入链路上的约束策略。
  - 依据：`code/backend/migrations/000001_init_schema.up.sql`、`code/backend/migrations/000017_cleanup_challenge_image_id_zero.up.sql`

## 已核验并移出活动 backlog 的条目

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
- [x] `challenges.image_id = 0` 的历史 no-image 哨兵值已完成 schema / contract 清理，不再保留为当前活动技术债。
