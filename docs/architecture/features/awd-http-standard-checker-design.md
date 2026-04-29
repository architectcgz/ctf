# AWD `http_standard` Checker 设计

## 目标

将 AWD 轮次巡检从单纯 HTTP 探活升级为标准化 checker 执行链路，支持 `put_flag / get_flag / havoc` 三段语义，并把服务状态判定、SLA 分和防守分写入每轮 service 结果。

这份文档承接 `docs/superpowers/plans/2026-04-11-awd-engine-phase2-http-standard.md` 的已实现结果。

## 当前状态

- `http_standard` 已作为 AWD checker 类型进入运行链路。
- `AWDRoundUpdater` 仍是唯一轮次调度入口，没有新增独立 checker 服务。
- checker 配置、轮次 flag、目标实例和结果写入都在 contest 模块内完成。
- `legacy_probe` 仍保留为兼容类型。

## 核心设计

### 1. 继续复用 `AWDRoundUpdater`

轮次执行不引入新 worker。`AWDRoundUpdater` 在当前轮上下文中读取 service definition、队伍实例和轮次 flag，然后按 checker 类型分派：

- `http_standard`：执行标准 HTTP checker。
- `legacy_probe` 或空值：走旧探活兼容路径。

这样可以继续复用现有轮次、Redis flag、实例和得分重算链路，避免 checker 与比赛状态出现第二套调度事实源。

### 1.1 赛前验证与本地 `check.py` 的边界

平台赛前 readiness 的正式依据是平台 runner 的执行结果，不是题目包里的本地脚本。

当前链路固定为：

1. 出题人本地自测：运行题目包内 `docker/check/check.py`，验证业务链路、漏洞演示路径和 flag 闭环。
2. 平台导入：读取 `extensions.awd.checker.type/config`，写入 AWD 题库。
3. 赛事挂题：默认继承 AWD 题库 checker，写入 `contest_awd_services.runtime_config`。
4. 赛前验证：平台 `checker-preview` 使用当前赛事服务的 checker 配置执行，并通过 preview token 写入 `validation_state`。
5. readiness：只根据赛事服务保存态的 checker 配置和校验状态判定是否可开赛。
6. 赛中轮次：`AWDRoundUpdater` 使用同一份赛事服务 checker 配置执行正式轮次检查。

因此，`check.py` 当前只属于出题人本地验证与审计材料；它可以比 `http_standard` 覆盖更多业务步骤，但不会直接决定平台 readiness。

### 1.2 Checker 流量不是攻击流量

`http_standard` runner 只代表平台裁判流量，用于写入 flag、回读 flag 和探测服务可用性。它不承担选手攻击入口，也不写入攻击流量事实。

流量职责划分如下：

- Checker 流量：平台发起，影响 SLA、防守状态、validation 和 readiness。
- 选手攻击流量：选手通过 AWD 代理、VPN 或跨队网络入口访问目标服务，应进入 `awd_traffic_events` 等攻击流量记录。
- Flag 提交流量：选手把偷到的 flag 提交给平台，由攻击提交流程判分。

不要把 checker endpoint 当成选手攻击入口，也不要用 checker 结果推断攻击流量归因。

### 2. `http_standard` 动作语义

每个目标实例按顺序执行：

- `put_flag`：把当前轮 flag 写入目标服务。
- `get_flag`：从目标服务取回 flag，并校验返回内容。
- `havoc`：执行轻量业务探测，避免只返回 200 的假阳性。

判定规则：

- `put_flag` 失败：服务记为 `down`。
- `get_flag` 取回内容不包含可接受 flag：服务记为 `compromised`。
- `havoc` 启用且失败：服务记为 `down`。
- 三段均通过：服务记为 `up`。

### 3. 配置结构

第一版只支持 HTTP 动作的最小字段：

- `method`
- `path`
- `headers`
- `body_template`
- `expected_status`
- `expected_substring`

模板变量只允许当前实现支持的安全集合：

- `{{FLAG}}`
- `{{ROUND}}`
- `{{TEAM_ID}}`
- `{{CHALLENGE_ID}}`

后续如果扩展变量，必须同步更新 checker parser、执行器、测试和配置文档，不能在前端临时拼接。

### 4. 轮次 flag 读取

checker 使用平台生成并缓存的官方轮次 flag。当前读取规则：

- 优先读取 Redis 当前轮 flag。
- 在允许窗口内接受上一轮 flag。
- Redis 缺失时按既有 `flagSecret + BuildAWDRoundFlag` 规则回退生成。

只有平台官方 flag 会被 checker 和攻击提交流程认可。

## 结果结构

`awd_team_services.check_result` 保持 JSON 存储，但 `http_standard` 结果至少包含：

- `checker_type`
- `status_reason`
- `put_flag`
- `get_flag`
- `havoc`
- `targets`
- `latency_ms`
- `error_code`
- `error`

前端和导出只应消费这些结构化字段，不再从自由文本错误里解析状态。

## 计分影响

- `up`：保留配置的 `sla_score` 与 `defense_score`。
- `down`：SLA 和防守得分按失败语义清零。
- `compromised`：防守得分清零，并记录失陷原因。
- 攻击分仍由攻击提交流程写入，不由 checker 直接产生。

## 兼容边界

- `legacy_probe` 保留给旧配置和过渡数据。
- 新 AWD service 配置默认应使用 `http_standard`。
- 当前已实现版本不新增外置 checker 沙箱，不支持任意脚本执行。
- `http_standard` 只能验证能表达成 HTTP 请求/响应断言的逻辑；复杂多步骤业务、非 HTTP 协议和漏洞可利用性不应期待由它完整覆盖。
- TCP / Binary 服务不能直接使用 `http_standard` 验证协议交互，后续应接入 `tcp_standard` 或受控 `script_checker`。
- checker 不直接处理封禁、申诉或处罚，这些仍属于赛事治理能力。

## 验收标准

- `http_standard` 成功路径能写入 `up` 和结构化 action 结果。
- 错误 flag 能写入 `compromised` 和 `flag_mismatch`。
- `havoc` 失败能写入 `down` 和对应错误码。
- `legacy_probe` 旧路径仍可运行。
