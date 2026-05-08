# AWD 检查器试跑架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`contracts`
- 关联模块：
  - `internal/module/contest/application/commands`
  - `internal/module/contest/application/jobs`
  - `frontend/src/components/platform/contest`
  - `frontend/src/features/contest-awd-config`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-11-awd-engine-phase6-checker-preview`
- 最后更新：`2026-05-07`

## 1. 背景与问题

管理端现在已经不是“保存后再看 readiness”的被动流程。当前代码允许管理员在 AWD 题目配置对话框里直接试跑 checker，并把试跑结果通过短期 token 绑定到保存动作上。这里需要明确的不是交互草图，而是当前真实执行链：

- 试跑与赛中巡检共用同一套 checker 执行逻辑
- 试跑结果不会写入 `awd_team_services`，也不会产生轮次副作用
- 已保存校验状态不是前端本地状态，而是由后端消费 `preview_token` 后落库

## 2. 架构结论

- 试跑入口固定为 `POST /admin/contests/:id/awd/checker-preview`。
- `AWDService.PreviewChecker` 是试跑编排 owner，`AWDRoundUpdater.PreviewServiceCheck` 是实际执行 owner。
- 试跑默认执行 3 轮，请求结果按多数派聚合，不再以单次偶发成功/失败作为最终结论。
- 试跑结果通过 Redis 中的 `preview_token` 与保存动作绑定；token 未被消费前，只是临时结果。
- 当前配置草稿一旦变化，前端会主动废弃本地 `preview_token`，避免旧结果为新草稿背书。

## 3. 模块边界与职责

### 3.1 模块清单

- `AWDChallengeConfigDialog`
  - 负责：收集草稿、触发试跑、展示实时进度、在保存时附带 `awd_checker_preview_token`
  - 不负责：自行决定校验状态

- `runAwdCheckerPreview` / `runContestAWDCheckerPreview`
  - 负责：前端 API 调用与 contract 归一化
  - 不负责：解析 checker 业务结果

- `AWDService.PreviewChecker`
  - 负责：校验参数、准备 preview runtime、执行多轮试跑、生成 `preview_token`
  - 不负责：直接写入赛事 service 持久化记录

- `storeAWDCheckerPreviewToken` / `consumeCheckerPreviewValidationState`
  - 负责：以短 TTL 保存试跑结果，并在保存配置时一次性消费
  - 不负责：保留历史试跑记录

### 3.2 事实源与所有权

- 临时试跑结果事实源：Redis `AWDCheckerPreviewTokenKey`
- 已保存校验结果事实源：`contest_awd_services.awd_checker_last_preview_result`
- 试跑进度事实源：preview realtime 事件流与后端 `reportAWDPreviewProgress`

## 4. 关键模型与不变量

### 4.1 核心实体

- `PreviewAWDCheckerReq`
  - 关键字段：`service_id`、`awd_challenge_id`、`checker_type`、`checker_config`、`access_url`、`preview_flag`、`preview_request_id`

- `AWDCheckerPreviewResp`
  - 关键字段：`checker_type`、`service_status`、`check_result`、`preview_context`、`preview_token`

- `storedAWDCheckerPreviewToken`
  - 绑定字段：`contest_id`、`service_id`、`awd_challenge_id`、`checker_type`、`checker_config`、`checker_token_env`

### 4.2 不变量

- 试跑上下文固定使用 `round_number = 0`、`team_id = 0`，不会污染正式轮次数据。
- `preview_flag` 为空时统一回退到 `flag{preview}`。
- 试跑 token TTL 固定为 30 分钟，且成功消费后立即删除。
- 试跑 token 必须同时匹配 contest、service、challenge、checker type、checker config、checker token env。
- 前端当前草稿签名变化后，原 `preview_token` 必须失效，不能继续参与保存。

## 5. 关键链路

### 5.1 试跑执行链路

1. 管理员在 `AWDChallengeConfigDialog` 中填写 checker 草稿。
2. 前端调用 `runContestAWDCheckerPreview`，可传显式 `access_url`，也可让后端自动拉起 preview runtime。
3. `AWDService.PreviewChecker` 校验 AWD 赛事上下文、checker 类型与配置。
4. 如果未提供 `access_url`，且题目是单容器部署并存在可用镜像，`prepareCheckerPreviewAccessURL` 会通过 `runtimeProbe.CreateContainer` 拉起临时实例。
5. `runPreviewCheckerAttempts` 连续执行 3 次 `PreviewServiceCheck`，再由 `aggregateAWDCheckerPreviewResults` 产出最终结果。
6. 后端把结果写入 Redis，返回 `preview_token` 给前端。

### 5.2 保存绑定链路

1. 前端保存时仅在 `previewSignature === currentCheckerSignature` 时附带 `awd_checker_preview_token`。
2. `CreateContestAWDService` / `UpdateContestAWDService` 消费 token，并把结果转换为 `passed` 或 `failed`。
3. token 不匹配或已过期时，保存请求会返回 `ErrAWDCheckerPreviewExpired`，不会静默接受旧结果。

## 6. 接口与契约

### 6.1 Preview 契约

- 请求：
  - `service_id` 用于编辑既有 service
  - `awd_challenge_id` 用于新建或显式指定题目
  - `preview_request_id` 用于关联实时进度事件

- 响应：
  - `preview_context.access_url` 是最终实际试跑目标
  - `preview_token` 是后续保存配置的唯一绑定凭证

### 6.2 进度契约

- 进度阶段至少包含：
  - `prepare`
  - `attempt-1`
  - `attempt-2`
  - `attempt-3`
  - `summary`
- 前端优先消费 realtime 事件；事件不可用时回退到本地定时进度提示

## 7. 兼容与迁移

- 试跑仍然兼容 `legacy_probe`，不会强制所有题目立即切到标准 checker。
- 无 `runtimeProbe`、镜像不可用或非单容器部署时，后端会要求管理员手动填写 `access_url`。
- Redis 不可用时，preview 请求会直接失败，因为当前实现必须返回可消费的 `preview_token`。
- 当前实现不保留“最近 N 次临时试跑历史”，只保留最终一次被保存的校验结果。

## 8. 代码落点

- `code/backend/internal/module/contest/application/commands/awd_service_run_commands.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- `code/backend/internal/module/contest/application/commands/awd_checker_preview_token_support.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/api/http/awd_round_check_handler.go`
- `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- `code/frontend/src/features/contest-awd-config/model/useAwdCheckerPreview.ts`
- `code/frontend/src/api/admin/contests.ts`

## 9. 验证标准

- 同一份 checker 草稿可在对话框内直接试跑，并返回结构化结果与 `preview_token`。
- 未提供 `access_url` 时，单容器题目能自动拉起 preview runtime；失败时会给出明确错误。
- 草稿改动后，旧 token 会在前端被废弃，保存请求不会继续附带它。
- 保存成功后，`awd_checker_last_preview_result` 与 `awd_checker_validation_state` 会同步更新；试跑本身不会写正式轮次数据。
