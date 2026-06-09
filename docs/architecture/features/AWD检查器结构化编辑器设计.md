# AWD 检查器结构化编辑器架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`frontend`、`backend`、`contracts`
- 关联模块：
  - `code/frontend/src/pages/platform/contests/ContestAwdConfigRoutePage.vue`
  - `code/frontend/src/features/platform/contest-awd-config/model/useContestAwdConfigPage.ts`
  - `code/frontend/src/features/platform/contest-awd-config/model/useAwdCheckerConfigDraft.ts`
  - `code/frontend/src/features/platform/contest-awd-config/model/useAwdCheckerDraftHydration.ts`
  - `code/frontend/src/features/platform/contest-awd-config/model/awdCheckerConfigSupport.ts`
  - `code/frontend/src/features/platform/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue`
  - `code/frontend/src/features/contest-awd-admin`
  - `code/backend/internal/module/contest/application/commands`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-11-awd-engine-phase5-structured-config-editor`
- 最后更新：`2026-06-01`

## 1. 背景与问题

当前 AWD checker 配置已经不是 textarea 里塞一段 JSON 的模式。管理端真实架构是“按 checker 类型维护结构化草稿，再统一编译成后端保存的 `checker_config`”。这篇文档要明确：

- 结构化编辑器的 owner 已经收口到 `ContestAwdConfigRoutePage.vue`、`ContestAwdConfigWorkspaceShell.vue` 与 `useContestAwdConfigPage.ts`
- 题目包默认 checker 配置与赛事级覆盖已经在 UI 中分层
- 前后端都对 checker 类型和配置做归一化，避免自由 JSON 漂移成不可验证状态
- 同一份输入语义的 normalize/default/validate 必须有唯一 owner，不能散在 route/page/support/transport 四层重复兜底

## 2. 架构结论

- `ContestAwdConfigRoutePage.vue` 只作为 route entry；`useContestAwdConfigPage.ts` + `ContestAwdConfigWorkspaceShell.vue` 组成当前正式的赛事级 checker 编辑入口。
- `awdCheckerConfigSupport.ts` 负责四种 checker 的草稿初始化、结构化校验和配置构建，是前端 checker payload normalize 的唯一 owner。
- 新建题目时默认继承 AWD 题库题目的 checker 配置；只有开启 `checkerOverrideEnabled` 后，赛事级草稿才脱离题目包默认值。
- 保存接口始终提交结构化字段 `checker_type`、`checker_config`、`awd_sla_score`、`awd_defense_score`，后端再统一写入 `runtime_config`。
- 试跑 token 与当前 checker 草稿签名绑定，编辑器本身承担 token 失效判断。
- `useContestAwdConfigDataLoader.ts`、`useAwdCheckerSaveFlow.ts`、`useAwdCheckerPreview.ts` 只承接 transport / workflow，不再各自做一份业务 normalize。

## 3. 模块边界与职责

### 3.1 模块清单

- `ContestAwdConfigRoutePage.vue`
  - 负责：作为 `/platform/contests/:id/awd/config` 的 route entry 挂载 feature page shell
  - 不负责：读取 route query、持有草稿、保存或试跑 workflow

- `useContestAwdConfigPage.ts`
  - 负责：页面级 route owner、breadcrumb、服务选择协调、loader / preview / save flow 组合
  - 不负责：字段级 checker payload normalize

- `useAwdCheckerDraftHydration.ts`
  - 负责：把持久化 `checker_config` 与 service 分值回填成结构化草稿
  - 不负责：决定 route query、请求发送或保存 payload

- `useAwdCheckerConfigDraft.ts`
  - 负责：页面草稿状态、字段错误、draft signature 与 validate orchestration
  - 不负责：重复实现 transport 或路由选择逻辑

- `awdCheckerConfigSupport.ts`
  - 负责：`create*Draft`、`build*CheckerConfig`、字段级校验、preset 与 checker payload normalize/default
  - 不负责：请求发送与服务列表刷新

- `useContestAwdConfigDataLoader.ts` / `useAwdCheckerSaveFlow.ts` / `useAwdCheckerPreview.ts`
  - 负责：数据读取、保存提交、试跑请求与错误反馈
  - 不负责：route query normalize、service fallback 或 checker payload builder

- `ContestAWDServiceService`
  - 负责：后端归一化 checker type/config，并将 challenge runtime 合成到 `runtime_config`
  - 不负责：维护前端结构化草稿形态

### 3.2 事实源与所有权

- 前端草稿事实源：`useContestAwdConfigPage.ts` 持有的 `legacyProbeDraft`、`httpStandardDraft`、`tcpStandardDraft`、`scriptCheckerDraft`
- 持久化事实源：`contest_awd_services.runtime_config`
- 题目包默认配置事实源：`AdminAwdChallengeData.checker_type/checker_config`

### 3.3 normalize / default / validate 唯一 owner

当前 `contest-awd-config` 明确按下面规则收口：

- route 输入 normalize
  - owner：`useContestAwdConfigPage.ts` 与 `useAwdChallengeSelection.ts`
  - 范围：`params.id -> contestId`、`query.service -> selectedServiceId`、service 不存在时回退到目录首项并回写 query
  - 禁止：把这层逻辑放进 `shared/model/navigation/*` 或 UI 组件

- 持久化配置 -> 草稿回填 normalize
  - owner：`useAwdCheckerDraftHydration.ts`
  - 范围：`service.checker_config`、`sla_score`、`defense_score` 回填到四类结构化 draft
  - 禁止：在 `useContestAwdConfigDataLoader.ts`、`useAwdCheckerPreview.ts`、`useAwdCheckerSaveFlow.ts` 再做第二份回填

- checker 草稿 default / payload normalize / 字段 validate
  - owner：`awdCheckerConfigSupport.ts`
  - 范围：`create*Draft()`、`build*CheckerConfig()`、HTTP method normalize、headers JSON parse、TCP/script 字段校验、preset default
  - 禁止：在 page model、UI field 组件或 transport flow 里重复写同语义 normalize/default/validate

- DOM 输入 -> 原始 primitive 转换
  - owner：最靠近字段的 UI 或 draft owner
  - 范围：例如 input 字符串转 `number`、checkbox 转 `boolean`
  - 约束：这里只允许做浏览器事件层的机械转换，不承接业务 fallback、checker 语义 normalize 或 payload 组装

- transport / 持久化 normalize
  - owner：后端 `ContestAWDServiceService`
  - 范围：checker type/config 最终合法化、`runtime_config` 合成、preview token 消费
  - 禁止：前端假定后端 fallback 规则后再自行复制一份“安全兜底”

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
- 同一份字段语义只允许一个 normalize/default/validate owner；其它层最多消费结果，不再复制同语义 fallback。

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

- `code/frontend/src/pages/platform/contests/ContestAwdConfigRoutePage.vue`
- `code/frontend/src/features/platform/contest-awd-config/model/useContestAwdConfigPage.ts`
- `code/frontend/src/features/platform/contest-awd-config/model/useAwdCheckerConfigDraft.ts`
- `code/frontend/src/features/platform/contest-awd-config/model/useAwdCheckerDraftHydration.ts`
- `code/frontend/src/features/platform/contest-awd-config/model/useAwdCheckerSaveFlow.ts`
- `code/frontend/src/features/platform/contest-awd-config/model/useAwdCheckerPreview.ts`
- `code/frontend/src/features/platform/contest-awd-config/model/useContestAwdConfigDataLoader.ts`
- `code/frontend/src/features/platform/contest-awd-config/model/awdCheckerConfigSupport.ts`
- `code/frontend/src/features/platform/contest-awd-config/ui/ContestAwdConfigWorkspaceShell.vue`
- `code/frontend/src/features/platform/contest-awd-config/ui/ContestAwdDebugStation.vue`
- `code/frontend/src/features/contest-awd-admin/model/useAwdChallengeLinkOperations.ts`
- `code/backend/internal/module/contest/application/commands/challenge_awd_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_support.go`
- `code/backend/internal/module/contest/application/commands/contest_awd_service_service.go`
- `code/backend/internal/module/contest/application/queries/contest_awd_service_query.go`

## 9. 验证标准

- 新建和编辑题目都能正确回填并保存结构化 checker 草稿。
- 题目包默认配置与赛事级覆盖切换后，草稿来源和保存结果保持一致。
- HTTP/TCP/脚本三类字段错误能在前端被结构化拦截，而不是等到后端再报泛化错误。
- 保存后的 `runtime_config` 同时保留归一化 `checker_config` 与 `checker_config_raw`。
