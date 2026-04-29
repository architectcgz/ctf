# AWD Checker Runner 扩展设计

## 目标

在现有 `legacy_probe` / `http_standard` 基础上扩展 AWD checker 能力，支持：

- `tcp_standard`：面向 TCP / Binary 服务的结构化协议检查。
- `script_checker`：执行题目包提供的私有 checker 脚本，用于复杂业务逻辑、多步骤协议和非 HTTP 服务。
- `checker-runner`：独立安全沙箱 runner，作为所有脚本型 checker 的唯一执行环境。

这份文档描述目标架构与安全边界。当前已实现能力包含 `legacy_probe`、`http_standard`、`tcp_standard` 基础结构化 TCP runner、基础 Docker sandbox runner port、`script_checker` 的 preview / 赛中 runner 调用链，以及题目包私有 checker entry artifact 的导入、存储和只读注入。

## 现有边界

### `http_standard`

`http_standard` 是平台当前正式赛前验证和赛中轮次检查的主能力。它适合验证能表达为 HTTP 请求/响应断言的逻辑：

- `put_flag` 写入当前轮 flag。
- `get_flag` 回读同一个 flag。
- `havoc` 执行轻量健康或业务探测。
- 校验状态码、响应内容、headers 和 body 模板。

它不能完整验证：

- 漏洞是否真的可利用。
- 登录、上传、预览、越权等多步骤业务链。
- TCP / Binary 协议状态机。
- 选手攻击流量代理和攻击归因。
- 复杂依赖环境下的出题人自定义检查逻辑。

### 题目包 `docker/check/check.py`

当前 `check.py` 是出题人本地自测脚本，不直接参与平台 readiness。它可以验证比 `http_standard` 更完整的业务链路，但只在本地执行。

未来若要让平台执行教师 checker，不能直接在 API 进程或轮次调度进程里运行 `check.py`。必须通过 `script_checker` 契约导入，并交给独立 sandbox runner 执行。

## Checker 类型契约

### `tcp_standard`

`tcp_standard` 用于结构化验证 TCP 服务，不执行任意脚本。

建议第一版契约：

```yaml
extensions:
  awd:
    checker:
      type: tcp_standard
      config:
        timeout_ms: 3000
        connect:
          host: "{{TARGET_HOST}}"
          port: "{{TARGET_PORT}}"
        steps:
          - send: "PING\n"
            expect_contains: "PONG"
          - send_template: "SET_FLAG {{FLAG}}\n"
            expect_contains: "OK"
          - send: "GET_FLAG\n"
            expect_contains: "{{FLAG}}"
        havoc:
          - send: "STATUS\n"
            expect_contains: "READY"
```

第一版只支持：

- TCP connect。
- 文本或 hex payload。
- `send` / `send_template`。
- `expect_contains` / `expect_regex`。
- 每步超时和总超时。
- `{{FLAG}}`、`{{ROUND}}`、`{{TEAM_ID}}`、`{{CHALLENGE_ID}}`、`{{TARGET_HOST}}`、`{{TARGET_PORT}}` 模板变量。

不支持：

- 任意代码执行。
- 动态分支。
- 无限循环或长连接 soak test。
- 复杂二进制协议解析。

### `script_checker`

`script_checker` 用于复杂逻辑，执行题目包内私有 checker 文件。它必须通过 sandbox runner 执行。

建议契约：

```yaml
extensions:
  awd:
    checker:
      type: script_checker
      config:
        runtime: python3
        entry: docker/check/check.py
        files:
          - docker/check/check.py
          - docker/check/protocol.py
        timeout_sec: 10
        args:
          - "{{TARGET_URL}}"
        env:
          CHECKER_TOKEN: "{{CHECKER_TOKEN}}"
          FLAG: "{{FLAG}}"
        output: json
```

`script_checker` 只能访问平台注入的参数、环境变量和声明过的私有 checker 文件。`files` 未声明时默认只包含 `entry`；声明后 `entry` 必须位于 `files` 内，且所有文件都必须是题目包内相对路径。题目包中的 checker 文件不作为选手附件公开，也不允许被平台直接挂载到目标服务容器。

## Sandbox Runner 安全边界

`script_checker` 必须运行在独立 runner 中，不允许在 API 进程、轮次调度进程或宿主机 shell 中直接执行。当前基础 Docker runner 已落地，`script_checker` 的 preview / 赛中链路已通过 runner port 调用沙箱；题目包导入时会校验 `config.entry` / `config.files` 位于包内，并把声明文件复制到服务端私有 artifact 目录，runner 执行时再把文件只读注入沙箱。

第一版必须具备以下安全边界。

### 进程与容器

- 每次检查创建一次性容器或等价隔离环境。
- 容器使用只读 root filesystem。
- 禁止 privileged。
- 禁止挂载 Docker socket。
- 禁止挂载平台源码、配置、数据库凭据、SSH key、宿主机敏感目录。
- 只读挂载本次 checker 所需文件。
- 临时工作目录使用独立 tmpfs 或一次性目录，执行结束后销毁。

### 网络

- 默认拒绝外网和平台内网访问。
- 只允许访问本次目标服务地址和必要 DNS。
- 禁止访问 PostgreSQL、Redis、平台 API、Docker bridge 管理地址、宿主机元数据地址。
- 如果需要访问多个目标地址，必须由平台显式生成 allowlist。

### 资源限制

- 必须设置总超时。
- 必须限制 CPU、内存、进程数、文件描述符数量。
- 必须限制 stdout / stderr 输出大小。
- 必须限制 checker 包大小和展开文件数量。
- 超时、OOM、输出超限、非 0 退出码都视为 checker 失败，并写入结构化错误码。

### 运行时与依赖

- 第一版只支持平台内置运行时镜像，例如 `python3`。
- 不允许赛时联网安装依赖。
- 若允许额外依赖，必须在导入阶段构建固定 checker image，并记录镜像摘要。
- checker 执行环境版本必须可在结果中追踪。

### 输出契约

第一版只接受两种输出：

1. 退出码约定：
   - `0`：通过。
   - 非 `0`：失败。
2. JSON 输出约定：

```json
{
  "status": "ok",
  "reason": "flag_roundtrip_passed",
  "details": {
    "latency_ms": 42
  }
}
```

如果 `output=json`，runner 必须校验 JSON schema。无效 JSON 按 `invalid_checker_output` 处理。

### 审计与可观测

每次 `script_checker` 执行必须记录；`tcp_standard` 虽不进入脚本沙箱，也应在 target 结果里记录同一套基础定位字段：

- contest id、service id、team id、round number。
- checker type、checker package digest；`tcp_standard` 无 artifact 时该字段为空。
- duration_ms。
- exit_code、status_reason / error_code、resource_limit_hit。
- stdout/stderr 截断摘要。
- target allowlist 摘要。

审计记录不能保存完整 flag。输出中出现 flag 时，应在持久化前脱敏。

当前实现把审计摘要写入 `check_result.targets[].audit`，用于定位单个 target 的执行结果。脚本 runner 的 stdout/stderr 会在写入前截断并替换当前轮 flag；TCP runner 的连接错误、步骤错误和 target audit 也会经过同一套脱敏逻辑。

## Artifact 清理

`script_checker` artifact 按 `AWD_CHECKER_ARTIFACT_DIR/<slug>/<digest>/...` 存储。题目包重新导入同一个 slug 且 digest 变化时，平台在数据库事务提交成功后清理旧 digest 目录；事务失败时不清理旧目录，避免保存态仍引用旧 artifact。

清理函数必须先校验目标目录位于 `AWD_CHECKER_ARTIFACT_DIR` 内，且不能等于 artifact root 本身。清理只针对被新导入替换的旧 digest 目录，不扫描或删除未被当前操作确认的其它路径。

## Runner 接入点

### 赛前 preview

`POST /api/v1/admin/contests/:id/awd/checker-preview` 根据 checker type 分派：

- `legacy_probe`：现有路径。
- `http_standard`：现有 HTTP runner。
- `tcp_standard`：新增 TCP runner。
- `script_checker`：提交 runner job，等待结果或在超时内返回失败。

preview 通过后继续生成 preview token。保存赛事服务时消费 token，写入 `validation_state=passed`。

### 赛中轮次检查

`AWDRoundUpdater` 不直接执行脚本。它只负责：

1. 解析当前轮 service definition。
2. 为每个目标生成 checker job。
3. 调用 runner。
4. 把 runner 返回的结构化结果转换成 `awd_team_services` 状态和分数。

### readiness

readiness 不直接执行 checker。它只读取保存态：

- checker 类型和配置是否完整。
- 最近一次 preview token 是否已消费。
- `validation_state` 是否为 `passed`。

对于 `script_checker`，如果缺少 runner 校验结果，readiness 必须保持 `pending_validation` 或 `missing_checker`。

## 攻击流量边界

Checker runner 仍然不是攻击流量入口。

- Checker runner 代表平台裁判检查。
- 选手攻击流量仍走 AWD 代理、VPN 或跨队网络入口。
- Flag 提交仍走平台攻击提交接口。
- `awd_traffic_events` 应记录选手攻击代理流量，不应把 checker 流量混入选手攻击样本。

如果需要观测 checker 流量，应另建 `checker_run_logs` 或在 `check_result` 中记录，不写入攻击流量事实表。

## 迁移策略

1. 保持现有 `http_standard` 为默认能力。
2. 增加 `tcp_standard` 结构化 runner，覆盖简单 TCP 服务。
3. 增加 sandbox runner 基础设施。
4. 开放 `script_checker` 导入、preview 和赛中执行。
5. 前端结构化编辑器新增 checker type，隐藏高风险字段，只展示必要配置。
6. 题目包文档更新 `script_checker` 私有文件规则，明确不进入选手附件。

## 验收标准

- `http_standard` 现有测试不回退。
- `tcp_standard` 可以对本地 TCP fixture 完成 connect、send、expect、flag roundtrip。
- `script_checker` 只能在 sandbox runner 中执行，不能在 API 进程直接执行。
- sandbox runner 超时、OOM、非 0 退出、无效 JSON 输出都有稳定错误码。
- runner 网络 allowlist 生效，脚本无法访问平台 API、数据库、Redis 和 Docker socket。
- preview 通过后能生成 token，保存服务后 readiness 变为 passed。
- 赛中轮次能消费 runner 结果并写入 `awd_team_services`。
