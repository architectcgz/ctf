# 迁移过程中识别出的项目技术债

更新时间：2026-06-10

本文只记录当前仍成立、或仍需保留为 follow-up 的项目技术债。已经被代码和架构事实收口的条目不再继续保留在活动 backlog 中。

## P1：运行时状态边界迁移仍有残余边界

- [ ] `runtime` 仍保留 mixed persistence/state owner，后续需要继续拆到更明确的业务 owner。
  - 现状：`internal/module/container_runtime/{contracts,ports,application,domain,infrastructure,agentcontracts}` 已承接容器 capability contract、port、service、domain helper、本地 Docker host adapter 与 runtime-agent protocol / bridge；`internal/module/runtime/runtime` 兼容转发入口已删除。`container_runtime/runtime.Module` 的 runtime-owned persistence 依赖仍由 `internal/app/composition/runtime_module.go` 显式注入，不在 module 内部直接 new 宽 runtime repo；runtime node router 与 ACL migration 这条容器能力链也只依赖窄 runtime allocation / state 接口。port/subnet allocation 与 lifecycle release persistence 仍在 `runtime/infrastructure.AllocationRepository`，AWD defense workspace / AWD service operation persistence 仍在 `runtime/infrastructure.AWDRepository`，runtime managed instance lookup / active container inventory / container-to-node state lookup / ACL migration state update 仍在 `ManagedInstanceRepository`、`ActiveContainerInventoryRepository`、`ContainerNodeIndexRepository`、`ACLMigrationStateRepository`，proxy traffic recorder 仍收口在 `runtime/infrastructure.ProxyTrafficEventRecorder`。
  - 影响：后续继续拆 `runtime` 时，重点已经从容器 capability 迁移转为状态 owner 归属：`RuntimeManagedInstance`、AWD workspace / service operation / scope control、runtime node、allocation、ACL migration 与 proxy traffic 这些 mixed state 需要分别评估是否应迁往 `instance`、`contest`、`practice` 或保留为独立 state adapter。
  - 依据：`docs/design/backend-module-boundary-target.md`

## P2：教学评估与 AWD 统一仍有残余边界

- [ ] 竞赛 / AWD 数据已回流到维度事实，但 recommendation / class review 与 practice 语义仍未完全统一。
  - 现状：个人 recommendation 与班级 class review 的 teaching fact snapshot 已吸收 AWD 成功覆盖、profile score 补充信号与 solved difficulty 覆盖，`difficulty_band` 也已经进入推荐候选查询 owner；但推荐与复盘主链路里的 `attempt / approved review evidence` 仍主要沿用 `contest_id IS NULL` 的练习侧语义。
  - 影响：这条债已经从“竞赛数据没有回流”收敛到“主画像里的练习 / 竞赛事实还没有完全统一”，当前仍不是一份完全统一的训练画像。
  - 依据：`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`、`docs/architecture/features/教学复盘建议生成架构.md`

- [ ] `contest_awd_services.runtime_config.challenge_id` 只剩历史兼容清洗层，尚未彻底退场。
  - 现状：新写入已不再持久化该字段，查询返回也会过滤它；但历史数据兼容清洗逻辑仍保留在 query / response mapper 层。
  - 影响：兼容逻辑长期保留会继续增加 AWD 运行态契约的清理成本。
  - 依据：`docs/architecture/backend/design/awd-engine-migration.md`

## P2：模块迁移后的结构尾项

- [ ] `challenges.image_id = 0` 的历史 no-image 哨兵值还没有完成 schema / contract 清理，暂时不能继续补物理 FK。
  - 现状：当前 baseline `000001_init_schema.up.sql` 已吸收原 `000011_add_additional_foreign_keys` 里的 `audit_logs.user_id`、`awd_*` 的 creator / verifier / submitter、`instances` 主引用列、`submissions` 主引用列；`awd_* .service_id` 这批历史孤儿也已经在开发库回填完成，但 `challenges.image_id` 仍有 `76` 条历史 `image_id = 0` 记录。
  - 影响：`challenges.image_id` 继续缺少数据库级完整性约束，后续迁移和数据清理仍要保留逻辑层对 no-image 哨兵值的兼容。
  - 处理方向：先把 `image_id = 0` 的 schema / contract 语义清理完，再单独补 `image_id` 的 FK migration。
  - 依据：`code/backend/migrations/000001_init_schema.up.sql`

- [ ] `assessment / ops` 的副作用事件化边界仍有继续收口空间。
  - 现状：assessment 画像刷新、推荐缓存刷新、ops realtime relay 与通知等主链路已经挂到 event consumers；剩余问题主要是个别同步副作用、缓存失效 owner 和失败回退策略还没有彻底压平。
  - 影响：业务写路径虽然已显著变窄，后续迁移时仍可能在局部位置回流同步副作用细节。
  - 依据：`docs/design/backend-module-boundary-target.md`

## 已核验并移出活动 backlog 的条目

- [x] runtime HTTP adapter 的 compat 形态已经收口到 `internal/app/composition/runtime_http_service_adapter.go` 单一 facade，旧的 parallel compat wrapper 与 `runtime_adapter_compat.go` 均已移除，不再保留为当前技术债。
- [x] 推荐链路的 `difficulty_band` 已成为推荐候选查询的一部分，不再保留为当前技术债。
- [x] 班级时间段 owner 已收口到共享 `classwindow`，不再保留为当前技术债。
- [x] 班级报告导出结构已纳入 `summary / trend / review / category_distribution / difficulty_distribution / contest_migration`，原表述过期。
- [x] `contest_challenges.awd_*` 历史字段不再存在，原表述过期。
- [x] application 层 GORM / Redis concrete allowlist 未全局清空这条，按当前代码状态已不成立。
- [x] 历史 `awd_* .service_id` 孤儿引用已在开发库回填到真实 `contest_awd_services` 父记录，不再保留为当前活动技术债。
