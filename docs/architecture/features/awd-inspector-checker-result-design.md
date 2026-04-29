# AWD 巡检结果视图设计

## 目标

让管理端 AWD inspector 能正确展示标准 checker 结果，尤其是 `http_standard` 的 `put_flag / get_flag / havoc` 动作明细，以及 `SLA / 攻击 / 防守` 三段分。

这份文档承接 `docs/superpowers/plans/2026-04-11-awd-engine-phase3-inspector.md` 的已实现结果。

## 当前状态

- 前端 AWD API contracts 已保留 `checker_type`、`sla_score` 等字段。
- 管理端巡检结果视图能识别 `http_standard` 与 `legacy_probe`。
- 导出内容已包含 checker 类型和 SLA 分。
- 旧探活结果仍可读。

## 设计原则

### 1. 展示层消费结构化结果

巡检结果不再只展示“服务是否在线”。展示层通过 `check_result` 的结构化字段呈现：

- checker 类型
- 状态原因
- action 是否健康
- action 错误码
- target 级明细

这避免了前端从后端错误字符串中推断含义。

### 2. 三段分必须同时出现

AWD 管理视角统一展示：

- `sla_score`
- `attack_score`
- `defense_score`

轮次 summary、服务表、导出内容都按这个口径呈现。`total_score` 可以作为聚合结果出现，但不能替代三段分。

### 3. 兼容旧结果

`legacy_probe` 结果仍按探活语义展示。当前兼容要求：

- 没有 `checker_type` 时按旧结果处理。
- 旧 `probe` 字段不强行转换成 `put_flag / get_flag / havoc`。
- 导出中保留可读 checker 类型，避免旧数据看起来像标准 checker。

## 前端边界

当前主要落点：

- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/admin.ts`
- `code/frontend/src/composables/useAwdCheckResultPresentation.ts`
- `code/frontend/src/composables/useAwdInspectorExports.ts`
- `code/frontend/src/components/platform/contest/AWDRoundInspector.vue`

`useAwdCheckResultPresentation` 是巡检结果展示的归一化入口。页面组件不应散落解析 `check_result` 的细节。

## 展示内容

服务记录层至少展示：

- 服务状态：`up / down / compromised`
- checker 类型
- SLA 分
- 防守分
- 攻击分
- 状态原因

动作明细层至少展示：

- `put_flag`
- `get_flag`
- `havoc`
- 每个 action 的健康状态、状态码、错误码和耗时

导出层至少包含：

- `Checker类型`
- `SLA得分`
- `攻击得分`
- 防守得分和服务状态

## 风险与约束

- 不要在页面里重复实现 checker 结果解析。
- 新 checker 类型接入时，必须先扩展展示 helper 和测试。
- 旧探活数据只做兼容展示，不作为新 checker 契约继续扩展。

## 验收标准

- `http_standard` 能显示三段 action 摘要。
- `flag_mismatch / invalid_checker_config / flag_unavailable` 等状态能显示可读原因。
- 轮次 summary 和服务表都显示 SLA 分。
- CSV 导出包含 checker 类型和 SLA 分。
