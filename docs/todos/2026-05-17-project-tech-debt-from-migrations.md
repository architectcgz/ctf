# 迁移过程中识别出的项目技术债

更新时间：2026-05-18

本文只记录当前仍成立、或仍需保留为 follow-up 的项目技术债。已经被代码和架构事实收口的条目不再继续保留在活动 backlog 中。

## P1：运行时边界迁移仍有 compat 遗留

- [ ] `container_runtime` capability port 的最终物理落点还没有完全定型。
  - 现状：`runtime/ports/container_runtime.go` 仍集中承载 provisioning / cleanup / file / image / stats / interactive 等 capability interface，`internal/app/composition/runtime_module.go` 继续以 `ContainerRuntimeModule` 做物理聚合。
  - 影响：后续继续拆 `runtime` 时，仍可能再次引入过渡 owner 或把 capability 继续堆在组合层。
  - 依据：`docs/design/backend-module-boundary-target.md`

- [ ] runtime HTTP adapter 的 compat 形态还没有彻底收口。
  - 现状：`internal/app/composition/runtime_http_service_adapter.go` 仍在桥接 `instance` contracts 与 runtime HTTP surface，`instance_module.go` 仍通过该 adapter 装配实例 HTTP 服务。
  - 影响：边界虽然已经明显比迁移前清晰，但最终唯一 owner 仍没有完全压实。
  - 依据：`docs/reviews/backend/2026-05-11-runtime-instance-boundary-slice9-review.md`、`docs/design/backend-module-boundary-target.md`

## P2：教学评估与 AWD 统一仍有残余边界

- [ ] 竞赛 / AWD 数据已回流到维度事实，但 recommendation / class review 与 practice 语义仍未完全统一。
  - 现状：个人与班级 teaching fact snapshot 已吸收 AWD 成功覆盖、profile score 补充信号与 solved difficulty 覆盖；但推荐与复盘主链路里，attempt / review evidence 仍主要沿用练习侧语义。
  - 影响：当前已经不是“完全没回流”，但仍不是一份完全统一的训练画像。
  - 依据：`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`、`docs/architecture/features/教学复盘建议生成架构.md`

- [ ] `contest_awd_services.runtime_config.challenge_id` 只剩历史兼容清洗层，尚未彻底退场。
  - 现状：新写入已不再持久化该字段，查询返回也会过滤它；但历史数据兼容清洗逻辑仍保留在 query / response mapper 层。
  - 影响：兼容逻辑长期保留会继续增加 AWD 运行态契约的清理成本。
  - 依据：`docs/architecture/backend/design/awd-engine-migration.md`

## P2：模块迁移后的结构尾项

- [ ] `assessment / ops` 的副作用事件化边界仍有继续收口空间。
  - 现状：assessment 画像刷新、推荐缓存刷新、ops realtime relay 与通知等主链路已经挂到 event consumers，但仍可继续压缩个别同步副作用与缓存失效 owner。
  - 影响：业务写路径虽然已显著变窄，后续迁移时仍可能在局部位置回流同步副作用细节。
  - 依据：`docs/design/backend-module-boundary-target.md`

## 已核验并移出活动 backlog 的条目

- [x] 推荐链路的 `difficulty_band` 已成为推荐候选查询的一部分，不再保留为当前技术债。
- [x] 班级时间段 owner 已收口到共享 `classwindow`，不再保留为当前技术债。
- [x] 班级报告导出结构已纳入 `summary / trend / review / category_distribution / difficulty_distribution / contest_migration`，原表述过期。
- [x] `contest_challenges.awd_*` 历史字段不再存在，原表述过期。
- [x] application 层 GORM / Redis concrete allowlist 未全局清空这条，按当前代码状态已不成立。
