# AWD 检查器校验状态架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`contracts`
- 关联模块：
  - `internal/module/contest/application/commands`
  - `internal/module/contest/application/queries`
  - `internal/module/contest/domain`
  - `frontend/src/components/platform/contest`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-11-awd-engine-phase7-checker-validation-state`
- 最后更新：`2026-05-07`

## 1. 背景与问题

AWD checker 的“最近通过 / 最近失败 / 未验证 / 待重新验证”已经不是界面提示词，而是 `contest_awd_services` 上的正式状态字段。当前文档需要纠正旧设计稿里“token 失配也允许保存”的过时说法，并明确真实状态机：

- 试跑结果只有在保存配置时消费 `preview_token` 后才会成为正式校验状态
- 状态流转围绕当前保存态配置展开，不围绕前端草稿展开
- `stale` 表示最近一次保存的试跑结果还在，但已经不再对应当前 checker 配置

## 2. 架构结论

- 校验状态 owner 是 `contest_awd_services`，不是前端本地缓存。
- 正式状态字段有且仅有：`pending`、`passed`、`failed`、`stale`。
- `preview_token` 现在是强绑定校验凭证；调用保存接口时若传了 token 但无法消费，请求会失败，而不是静默降级。
- 配置未变化时，更新题目的排序、显示状态等非 checker 字段，不会改写现有校验状态。
- `last_preview_result` 保存的是完整 preview 结果快照，供 service 列表、配置对话框和 readiness 共同复用。

## 3. 模块边界与职责

### 3.1 模块清单

- `consumeCheckerPreviewValidationState`
  - 负责：消费 Redis token，生成 `passed/failed + last_preview_*`
  - 不负责：判断配置变化后的 `stale` 逻辑

- `buildContestAWDServiceValidationUpdate`
  - 负责：比较新旧 checker 配置，决定保留、转 stale 或刷新 preview 结果
  - 不负责：执行 preview

- `ContestAWDServiceQueryService`
  - 负责：把保存态校验状态和 `last_preview_result` 暴露给前端
  - 不负责：推导 readiness 阻塞原因

- 前端 `AWDChallengeConfigPanel` / `AWDChallengeConfigDialog`
  - 负责：展示状态标签、最近校验时间和目标地址摘要
  - 不负责：自行推导状态流转

### 3.2 事实源与所有权

- 状态事实源：`contest_awd_services.awd_checker_validation_state`
- 最近校验事实源：
  - `contest_awd_services.awd_checker_last_preview_at`
  - `contest_awd_services.awd_checker_last_preview_result`
- 状态标签归一化入口：`NormalizeAWDCheckerValidationState`

## 4. 关键模型与不变量

### 4.1 核心实体

- `AWDCheckerValidationState`
  - `pending`
  - `passed`
  - `failed`
  - `stale`

- `AWDCheckerPreviewResult`
  - `checker_type`
  - `service_status`
  - `check_result`
  - `preview_context`

### 4.2 不变量

- `service_status == up` 时才会把正式状态写成 `passed`；其余 preview 结果统一写成 `failed`。
- 若保存时附带 `preview_token`，但 token 已过期或与当前 checker 配置不匹配，接口返回 `ErrAWDCheckerPreviewExpired`。
- 若 checker 配置发生变化且已有历史试跑结果，状态转为 `stale`，同时保留上一份 `last_preview_result` 供排障查看。
- 若 checker 配置发生变化但此前没有任何校验记录，状态保持 `pending`。
- 空值或未知状态在读路径上统一归一化为 `pending`。

## 5. 关键链路

### 5.1 写路径

1. 试跑成功后，结果先以 `preview_token` 形式存在 Redis。
2. `CreateContestAWDService` / `UpdateContestAWDService` 调用 `consumeCheckerPreviewValidationState` 尝试消费 token。
3. 消费成功则写入：
   - `awd_checker_validation_state`
   - `awd_checker_last_preview_at`
   - `awd_checker_last_preview_result`
4. 没有 token 时，`buildContestAWDServiceValidationUpdate` 对比新旧 checker 配置，必要时把状态转为 `stale` 或 `pending`。

### 5.2 读路径

1. `ListContestAWDServices` 返回 service 当前保存态。
2. HTTP handler 把 `last_preview_result` 解析成结构化 DTO。
3. `AWDChallengeConfigPanel`、`AWDChallengeConfigDialog`、`AWDReadinessChecklist` 复用同一份结果快照展示状态与最近访问目标。

## 6. 接口与契约

### 6.1 持久化契约

当前 `ContestAWDServiceResp` 至少公开：

- `validation_state`
- `last_preview_at`
- `last_preview_result`

### 6.2 展示契约

前端对外只消费四种正式状态文案：

- `pending` -> `未验证`
- `passed` -> `最近通过`
- `failed` -> `最近失败`
- `stale` -> `待重新验证`

## 7. 兼容与迁移

- 历史脏值或空值会在读路径上自动归一化为 `pending`。
- 当前没有“自动重新试跑后回写状态”的后台任务；状态更新只由保存配置时消费 token 或显式配置变更触发。
- `last_preview_result` 当前只保存最近一次正式绑定结果，不保存完整历史。

## 8. 代码落点

- `code/backend/internal/model/contest_challenge.go`
- `code/backend/internal/module/contest/domain/awd_checker_validation_support.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`
- `code/backend/internal/module/contest/api/http/awd_service_manage_handler.go`
- `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
- `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`

## 9. 验证标准

- 试跑通过并保存后，状态写成 `passed`，同时落库时间与结果快照。
- 试跑失败并保存后，状态写成 `failed`，且前端可读出最近失败目标。
- 修改 checker 草稿但未重新试跑时，已有结果会转成 `stale`；从未试跑过的配置保持 `pending`。
- 携带失效或失配 token 的保存请求会失败，不会错误更新状态。
