# AWD 检查器结构化编辑器架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`frontend`、`backend`、`contracts`
- 关联模块：
  - `frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
  - `frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
  - `frontend/src/features/contest-awd-admin`
  - `internal/module/contest/application/commands`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-11-awd-engine-phase5-structured-config-editor`
- 最后更新：`2026-05-07`

## 1. 背景与问题

当前 AWD checker 配置已经不是 textarea 里塞一段 JSON 的模式。管理端真实架构是“按 checker 类型维护结构化草稿，再统一编译成后端保存的 `checker_config`”。这篇文档要明确：

- 结构化编辑器的 owner 是 `AWDChallengeConfigDialog`，不是独立页面
- 题目包默认 checker 配置与赛事级覆盖已经在 UI 中分层
- 前后端都对 checker 类型和配置做归一化，避免自由 JSON 漂移成不可验证状态

## 2. 架构结论

- `AWDChallengeConfigDialog` 是当前唯一正式的赛事级 checker 编辑入口。
- `awdCheckerConfigSupport.ts` 负责四种 checker 的草稿初始化、结构化校验和配置构建，是前端配置语义唯一 owner。
- 新建题目时默认继承 AWD 题库题目的 checker 配置；只有开启 `checkerOverrideEnabled` 后，赛事级草稿才脱离题目包默认值。
- 保存接口始终提交结构化字段 `checker_type`、`checker_config`、`awd_sla_score`、`awd_defense_score`，后端再统一写入 `runtime_config`。
- 试跑 token 与当前 checker 草稿签名绑定，编辑器本身承担 token 失效判断。

## 3. 模块边界与职责

### 3.1 模块清单

- `AWDChallengeConfigDialog`
  - 负责：草稿状态、题目包默认值回填、赛事级覆盖切换、保存与试跑交互
  - 不负责：后端 checker 语义归一化

- `awdCheckerConfigSupport.ts`
  - 负责：`create*Draft`、`build*CheckerConfig`、字段级校验和预设模板
  - 不负责：请求发送与服务列表刷新

- `useAwdChallengeLinkOperations`
  - 负责：把编辑器 payload 落到 `createContestAWDService` / `updateContestAWDService`
  - 不负责：解析复杂 preview 结果

- `ContestAWDServiceService`
  - 负责：后端归一化 checker type/config，并将 challenge runtime 合成到 `runtime_config`
  - 不负责：维护前端结构化草稿形态

### 3.2 事实源与所有权

- 前端草稿事实源：`AWDChallengeConfigDialog` 的 `legacyProbeDraft`、`httpStandardDraft`、`tcpStandardDraft`、`scriptCheckerDraft`
- 持久化事实源：`contest_awd_services.runtime_config`
- 题目包默认配置事实源：`AdminAwdChallengeData.checker_type/checker_config`

## 4. 关键模型与不变量

### 4.1 核心实体

- 结构化草稿类型：
  - `AWDLegacyProbeDraft`
  - `AWDHTTPStandardDraft`
  - `AWDTCPStandardDraft`
  - `AWDScriptCheckerDraft`

- 持久化输入：
  - `checker_type`
  - `checker_config`
  - `awd_sla_score`
  - `awd_defense_score`

### 4.2 不变量

- HTTP action 的 `headers_text` 必须是 JSON 对象；无效时前端直接报字段错误。
- TCP 步骤至少保留一个有效步骤，且 `send_hex` 不能与 `send` / `send_template` 同时出现。
- `script_checker.entry` 必须是题目包内相对路径，不能以 `/` 开头，也不能包含 `..`。
- 关闭赛事级覆盖后，编辑器立即回退到题目包默认 checker 配置。
- 后端只接受合法 `checker_type`，并将 `checker_config` 序列化为标准 JSON 字符串。

## 5. 关键链路

### 5.1 新建题目链路

1. 管理员在对话框里先选 AWD 题库题目。
2. 编辑器读取 `selectedAwdChallenge.checker_type/checker_config` 作为默认草稿。
3. 若不开启赛事级覆盖，保存时直接提交题目包默认 checker 语义。
4. 若开启覆盖，则以当前结构化草稿构建 `checker_config` 后保存。

### 5.2 编辑既有题目链路

1. 对话框打开时，从 `props.draft` 读取赛事 service 当前保存态。
2. `create*Draft` 系列函数把已保存配置反向映射回结构化表单。
3. 编辑器可继续试跑、保存，并把 `awd_checker_preview_token` 一并附带给后端。

### 5.3 持久化链路

1. 前端提交 `checker_type` 和结构化 `checker_config`。
2. 后端 `validateAndNormalizeContestAWDFields` 校验赛事模式、分数范围和 checker 类型合法性。
3. `buildContestAWDServiceRuntimeConfig` 把 `checker_type`、`checker_config`、`checker_config_raw` 与 `challenge_runtime` 统一写入 `runtime_config`。

## 6. 接口与契约

### 6.1 前端构建契约

- `http_standard`
  - 支持 `put_flag`、`get_flag`、`havoc`
  - 提供 `REST /api/flag`、`Form /flag`、`File /flag.txt` 三组预设

- `tcp_standard`
  - 支持 `timeout_ms` 与多步 `steps`

- `script_checker`
  - 支持 `runtime`、`entry`、`timeout_sec`、`args`、`env`、`output`

### 6.2 后端持久化契约

- `CreateContestAWDServiceReq`
- `UpdateContestAWDServiceReq`

当前都接受：

- `checker_type`
- `checker_config`
- `awd_sla_score`
- `awd_defense_score`
- `awd_checker_preview_token`

## 7. 兼容与迁移

- 旧 `legacy_probe` 仍保留最小 `health_path` 草稿，不强制迁成标准 checker。
- 前端结构化编辑器不是通用 JSON editor；新增 checker 类型时必须补草稿结构和字段验证。
- 查询返回 `runtime_config` 时会去掉 `challenge_id` 这类不应由前端重新提交的内部字段。

## 8. 代码落点

- `code/frontend/src/components/platform/contest/AWDChallengeConfigDialog.vue`
- `code/frontend/src/components/platform/contest/awdCheckerConfigSupport.ts`
- `code/frontend/src/features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts`
- `code/frontend/src/utils/platformContestAwdChallengeLinks.ts`
- `code/backend/internal/module/contest/application/commands/challenge_awd_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`

## 9. 验证标准

- 新建和编辑题目都能正确回填并保存结构化 checker 草稿。
- 题目包默认配置与赛事级覆盖切换后，草稿来源和保存结果保持一致。
- HTTP/TCP/脚本三类字段错误能在前端被结构化拦截，而不是等到后端再报泛化错误。
- 保存后的 `runtime_config` 同时保留归一化 `checker_config` 与 `checker_config_raw`。
