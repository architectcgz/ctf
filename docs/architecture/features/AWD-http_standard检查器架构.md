# AWD `http_standard` Checker 架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`contracts`
- 关联模块：
  - `internal/module/contest/application/jobs`
  - `internal/module/contest/application/commands`
  - `internal/module/challenge/domain`
- 过程追溯：`practice/superpowers-plan-index.md` 中的 `2026-04-11-awd-engine-phase2-http-standard`
- 最后更新：`2026-05-07`

## 1. 背景与问题

原有 AWD 轮次巡检主要依赖基础探活语义，无法稳定表达 `put_flag / get_flag / havoc` 三段动作，也无法把服务可用性、失陷语义和结构化结果统一写回每轮 service 结果。

需要收口的核心问题是：

- `http_standard` 必须成为标准化 HTTP checker 执行链路
- readiness 与赛中轮次必须使用同一份 checker 配置口径
- checker 流量必须和选手攻击流量严格区分

## 2. 架构结论

- 轮次执行继续复用 `AWDRoundUpdater`，不引入第二套独立 checker 调度器。
- 平台 readiness 的正式依据是平台 runner 的执行结果，不是题目包内本地 `check.py`。
- `http_standard` 以 `put_flag -> get_flag -> havoc` 的顺序执行标准动作语义。
- checker 流量只代表平台裁判流量，不承担攻击流量归因。
- `legacy_probe` 仅保留兼容角色。

## 3. 模块边界与职责

### 3.1 模块清单

- `AWDRoundUpdater`
  - 负责：读取 service definition、目标实例和轮次 flag，并按 checker 类型分派执行
  - 不负责：引入新的独立调度事实源

- `checker-preview` / readiness
  - 负责：以赛事服务当前保存态配置执行正式验证，并写入 validation 状态
  - 不负责：直接以本地 `check.py` 作为 readiness 依据

- 攻击流量记录链路
  - 负责：记录选手通过代理、VPN 或跨队入口产生的攻击流量
  - 不负责：消费 checker runner 的裁判流量

### 3.2 事实源与所有权

- checker 类型与配置 owner：赛事 service 配置
- 轮次 flag owner：平台生成与缓存的官方轮次 flag
- 轮次结果 owner：`awd_team_services.check_result`

## 4. 关键模型与不变量

### 4.1 核心实体

- `http_standard`
  - 语义：标准 HTTP checker 类型
  - 动作：`put_flag`、`get_flag`、`havoc`

- `check_result`
  - 至少包含：`checker_type`、`status_reason`、`put_flag`、`get_flag`、`havoc`、`targets`、`latency_ms`、`error_code`、`error`

### 4.2 不变量

- `put_flag` 失败记为 `down`
- `get_flag` 校验失败记为 `compromised`
- `havoc` 启用且失败记为 `down`
- 三段动作均通过才记为 `up`
- checker 结果只能消费平台官方 flag
- checker 流量不得直接写成攻击流量事实

## 5. 关键链路

### 5.1 赛前校验链路

1. 出题人本地运行题目包内 `docker/check/check.py` 做自测。
2. 平台导入题包，读取 `extensions.awd.checker.type/config`。
3. 赛事挂题后，赛事 service 保存当前 checker 配置。
4. `checker-preview` 以赛事服务当前配置执行，并通过 preview token 写入 validation 状态。
5. readiness 只依据赛事服务保存态配置和校验状态判定是否可开赛。

### 5.2 赛中轮次链路

1. `AWDRoundUpdater` 读取当前轮上下文、service definition 和目标实例。
2. runner 读取平台官方轮次 flag。
3. 按 `put_flag -> get_flag -> havoc` 顺序执行动作。
4. 结果写入结构化 `check_result`，并同步更新服务状态与计分。

## 6. 接口与契约

### 6.1 配置契约

第一版只支持以下 HTTP 动作最小字段：

- `method`
- `path`
- `headers`
- `body_template`
- `expected_status`
- `expected_substring`

当前允许的模板变量：

- `{{FLAG}}`
- `{{ROUND}}`
- `{{TEAM_ID}}`
- `{{CHALLENGE_ID}}`

### 6.2 计分契约

- `up`：保留配置的 `sla_score` 与 `defense_score`
- `down`：SLA 和防守得分按失败语义清零
- `compromised`：防守得分清零，并记录失陷原因
- 攻击分仍由攻击提交流程写入，不由 checker 直接产生

## 7. 兼容与迁移

- `legacy_probe` 保留给旧配置和过渡数据。
- 当前实现不引入外置脚本沙箱，也不支持任意脚本执行。
- `http_standard` 只覆盖可表达成 HTTP 请求/响应断言的逻辑。
- TCP / Binary 服务后续应接入 `tcp_standard` 或受控 `script_checker`，而不是挤进 `http_standard`。

## 8. 代码落点

- `code/backend/internal/module/contest/application/jobs/awd_round_updater.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_runner.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_config.go`
- `code/backend/internal/module/contest/application/jobs/awd_http_checker_result.go`
- `code/backend/internal/module/contest/application/jobs/awd_checker_preview.go`
- `code/backend/internal/module/contest/application/jobs/status_awd_readiness.go`
- `code/backend/internal/module/contest/application/jobs/awd_round_updater_test.go`
- `code/backend/internal/module/challenge/domain/awd_package_parser.go`

## 9. 验证标准

- `http_standard` 成功路径能写入 `up` 和结构化 action 结果。
- 错误 flag 能写入 `compromised` 和 `flag_mismatch`。
- `havoc` 失败能写入 `down` 和对应错误码。
- `legacy_probe` 旧路径仍可运行。
