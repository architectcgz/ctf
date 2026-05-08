# AWD 巡检结果视图架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`frontend`、`contracts`
- 关联模块：
  - `frontend/src/features/awd-inspector`
  - `frontend/src/components/platform/contest/AWDRoundInspector.vue`
  - `frontend/src/api/admin/contests.ts`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-11-awd-engine-phase3-inspector`
- 最后更新：`2026-05-07`

## 1. 背景与问题

管理端 AWD inspector 需要从“在线/离线”式探活展示升级到能稳定呈现标准 checker 结果、结构化动作明细和三段得分。原始问题在于：

- 页面不能再从后端错误字符串推断 checker 语义
- `SLA / 攻击 / 防守` 三段分必须成为统一管理视角
- 历史 `legacy_probe` 结果仍需兼容可读

## 2. 架构结论

- 巡检结果展示层统一消费结构化 `check_result`。
- `sla_score`、`attack_score`、`defense_score` 必须同时出现，`total_score` 只能作为聚合结果。
- `useAwdCheckResultPresentation` 是巡检结果展示归一化入口，页面组件不应散落解析 `check_result` 细节。
- `legacy_probe` 结果继续按兼容语义展示，不强行转换成标准三段动作。

## 3. 模块边界与职责

### 3.1 模块清单

- `useAwdCheckResultPresentation`
  - 负责：checker 类型标签、动作摘要、错误码与展示语义归一化
  - 不负责：承担导出格式或页面布局本身

- `useAwdInspectorExports` / `awdInspectorExportPayloads`
  - 负责：导出层的结构化字段映射
  - 不负责：重复实现 checker 结果语义解析

- `AWDRoundInspector.vue`
  - 负责：消费归一化后的展示结果
  - 不负责：页面内再次手写 `check_result` 解析规则

### 3.2 事实源与所有权

- 后端写入的结构化 `check_result` 和三段分是展示层事实源。
- 展示 helper 是前端语义映射唯一 owner。

## 4. 关键模型与不变量

### 4.1 核心实体

- `check_result`
  - 关键字段：`checker_type`、`status_reason`、action 级结果、`error_code`

- 三段分
  - `sla_score`
  - `attack_score`
  - `defense_score`

### 4.2 不变量

- 服务记录层必须能展示：服务状态、checker 类型、三段分、状态原因。
- 动作明细层至少能展示：`put_flag`、`get_flag`、`havoc` 的健康状态、状态码、错误码和耗时。
- 导出层必须包含可读 checker 类型和 SLA 分。

## 5. 关键链路

### 5.1 读路径

1. 后端返回带有 `checker_type`、`check_result` 和三段分的服务记录。
2. `api/admin/contests.ts` 将结构化结果映射到前端 contracts。
3. `useAwdCheckResultPresentation` 统一生成 checker 类型标签、动作摘要和状态原因。
4. 页面和导出层消费归一化结果。

## 6. 接口与契约

### 6.1 对外契约

巡检结果相关 API 至少需要暴露：

- `checker_type`
- `check_result`
- `sla_score`
- `attack_score`
- `defense_score`

### 6.2 导出契约

导出内容至少包含：

- `Checker类型`
- `SLA得分`
- `攻击得分`
- `防守得分`
- `服务状态`

## 7. 兼容与迁移

- 没有 `checker_type` 时按旧结果兼容处理。
- 旧 `probe` 字段不强行转换成 `put_flag / get_flag / havoc`。
- 新 checker 类型接入时，必须先补展示 helper 和相关测试。

## 8. 代码落点

- `code/frontend/src/features/awd-inspector/model/useAwdCheckResultPresentation.ts`
- `code/frontend/src/features/awd-inspector/model/awdInspectorExportPayloads.ts`
- `code/frontend/src/features/awd-inspector/model/useAwdCheckResultPresentation.test.ts`
- `code/frontend/src/components/platform/contest/AWDRoundInspector.vue`
- `code/frontend/src/api/admin/contests.ts`
- `code/backend/internal/model/awd.go`

## 9. 验证标准

- `http_standard` 能显示三段 action 摘要。
- `flag_mismatch`、`invalid_checker_config`、`flag_unavailable` 等状态能显示可读原因。
- 轮次 summary 和服务表都显示 SLA 分。
- CSV 导出包含 checker 类型和 SLA 分。
