# 攻击过程还原与全流量回放演进设计

## 目标

在现有攻防证据链基础上，补出面向教师复盘和后续流量取证扩展的“攻击过程自动还原”能力。该能力第一阶段不追求完整抓包和命令级录制，而是把平台已经记录的访问、代理请求、提交、AWD 攻击和流量摘要整理成可阅读的攻击会话。

设计需要同时满足两个目标：

1. 当前版本可以低风险落地，用于展示学生从访问靶机、尝试利用到提交结果的关键过程。
2. 后续可以自然接入完整 HTTP 请求响应、HAR、pcap 或对象存储证据包，而不推翻现有读模型和前端展示。

## 当前基础

平台已经具备以下事实数据：

- `audit_logs`：记录实例访问 `instance_access` 和平台代理请求 `instance_proxy_request`。
- `submissions`：记录普通训练和 Jeopardy 题目的 Flag 提交结果。
- `awd_attack_logs`：记录 AWD 攻击提交、受害队伍、提交人、成功状态和得分。
- `awd_traffic_events`：记录 AWD 代理访问摘要，包括攻击队伍、受害队伍、服务、路径、方法和状态码。
- `submission_writeups`：记录学生对解题过程的人工复盘材料。
- 教学读模型已经提供 `GET /api/v1/teacher/students/:id/evidence`，用于教师查看事件级证据链。

因此本设计不从底层重新采集所有行为，而是在现有事实数据上增加一层会话化读模型。

## 设计原则

- 事实数据和派生视图分离：原始业务事实仍保存在各自模块，攻击会话只是读模型聚合结果。
- 先做事件级还原：第一版只还原关键行为顺序，不承诺完整流量回放、命令录制或浏览器录屏。
- 预留证据引用：事件结构中预留 `capture_ref`，以后可以挂载完整请求响应、HAR、pcap 或脱敏证据包。
- 不把大对象放进 PostgreSQL：数据库只保存索引、归属、摘要、hash 和存储引用，完整流量进入对象存储或文件存储。
- 前端依赖稳定 DTO：前端只理解统一事件结构，不直接拼接 `audit_logs`、`submissions`、`awd_attack_logs` 等多张表。

## 分层模型

攻击过程还原分为三层：

```text
AttackSession  攻击会话，表示一次围绕某题、某服务或某目标的连续尝试
AttackEvent    关键事件，表示访问、请求、提交、AWD 攻击、Writeup 等行为
TrafficCapture 全流量或请求响应证据包，作为 AttackEvent 的外部附件
```

第一阶段只实现 `AttackSession + AttackEvent`。`TrafficCapture` 仅在 DTO 和数据结构中预留引用，不立即落地全量存储。

## 统一事件结构

后端对外返回统一的攻击事件结构，屏蔽底层事实来源差异：

```json
{
  "id": "evt_20260501_0001",
  "session_id": "sess_12_20260501_0001",
  "type": "proxy_request",
  "stage": "exploit",
  "source": "audit_logs",
  "occurred_at": "2026-05-01T10:30:00+08:00",
  "actor": {
    "user_id": 1001,
    "team_id": 2001
  },
  "target": {
    "challenge_id": 12,
    "contest_id": null,
    "round_id": null,
    "service_id": null,
    "victim_team_id": null
  },
  "summary": "POST /login 返回 200",
  "meta": {
    "method": "POST",
    "path": "/login",
    "query": "",
    "status_code": 200,
    "payload_preview": "username=admin&password=***"
  },
  "capture_available": false,
  "capture_ref": null
}
```

字段说明：

- `type` 表示事件类别，例如 `instance_access`、`proxy_request`、`challenge_submission`、`awd_attack_submission`、`awd_traffic`、`writeup`。
- `stage` 表示过程阶段，例如 `access`、`exploit`、`submit`、`review`。
- `source` 表示事实来源，例如 `audit_logs`、`submissions`、`awd_attack_logs`、`awd_traffic_events`。
- `actor` 保存行为主体，普通训练以 `user_id` 为主，竞赛和 AWD 场景可补充 `team_id`。
- `target` 保存目标上下文，普通题使用 `challenge_id`，AWD 场景补充 `contest_id`、`round_id`、`service_id` 和 `victim_team_id`。
- `meta` 只保存安全截断后的结构化元数据，不保存完整敏感内容。
- `capture_ref` 是后续全流量回放的外部证据引用。

## 攻击会话结构

攻击会话是从事件聚合出的派生结果：

```json
{
  "id": "sess_12_20260501_0001",
  "mode": "jeopardy",
  "student_id": 1001,
  "team_id": null,
  "challenge_id": 12,
  "contest_id": null,
  "round_id": null,
  "service_id": null,
  "victim_team_id": null,
  "title": "SQL 注入",
  "started_at": "2026-05-01T10:20:00+08:00",
  "ended_at": "2026-05-01T10:45:00+08:00",
  "result": "success",
  "event_count": 8,
  "capture_count": 0,
  "events": []
}
```

`result` 由事件推导：

- 存在正确提交或成功 AWD 攻击时为 `success`。
- 存在错误提交但没有成功事件时为 `failed`。
- 只有访问和请求、没有提交结果时为 `in_progress`。
- 只有零散事件且无法判断状态时为 `unknown`。

## 会话切分规则

第一版采用简单规则，避免过早引入复杂状态机：

- 普通训练：按 `student_id + challenge_id` 聚合，再按时间排序。
- Jeopardy 竞赛：按 `student_id/team_id + contest_id + challenge_id` 聚合。
- AWD 攻击：按 `student_id/team_id + contest_id + service_id + victim_team_id` 聚合。
- 如果同一目标两次事件间隔超过配置阈值，例如 30 或 60 分钟，可以拆成两个 session。

会话切分只影响展示，不改变底层事实数据。

## 接口设计

建议新增教师侧读接口：

```text
GET /api/v1/teacher/students/:id/attack-sessions
```

查询参数：

- `challenge_id`：按普通题过滤。
- `contest_id`：按竞赛过滤。
- `round_id`：按 AWD 轮次过滤。
- `mode`：`practice`、`jeopardy`、`awd`。
- `with_events`：是否返回事件明细，默认返回。
- `limit`、`offset`：分页。

响应结构：

```json
{
  "summary": {
    "total_sessions": 3,
    "success_count": 1,
    "failed_count": 1,
    "in_progress_count": 1,
    "capture_available_count": 0
  },
  "sessions": []
}
```

现有 `GET /api/v1/teacher/students/:id/evidence` 可以继续保留，作为事件级证据接口。`attack-sessions` 负责会话化组织，两者共用底层事件构建逻辑。

## 后端落点

推荐在教学读模型中实现，不侵入运行时和竞赛写入链路：

- `internal/dto/teacher_attack_session.go`：新增响应 DTO。
- `internal/module/teaching_readmodel/ports/query.go`：增加攻击会话查询接口。
- `internal/module/teaching_readmodel/infrastructure/repository.go`：从现有表读取并转换为统一事件。
- `internal/module/teaching_readmodel/application/queries/service.go`：完成权限检查、事件聚合、会话切分和摘要计算。
- `internal/module/teaching_readmodel/api/http/handler.go`：新增 HTTP handler。
- `internal/app/router_routes.go`：挂载教师侧路由。

这样可以保持运行时模块只负责记录事实，教学读模型负责面向教师的复盘组织。

## 前端展示

前端在教师学生分析页新增“攻击过程还原”面板，展示粒度以会话为主：

- 会话列表：展示题目、模式、开始时间、结束时间、结果和事件数量。
- 会话详情：按时间线展示访问、请求、提交、AWD 攻击和 Writeup。
- 事件详情：展示方法、路径、状态码、payload 预览、攻击目标和得分等安全摘要。
- 全流量入口：当 `capture_available=true` 时显示“查看流量证据”，否则不展示入口。

前端必须对未知事件类型提供通用展示兜底，避免后续新增事件类型时页面不可用。

## 全流量回放扩展

后续如果要支持完整流量回放，不应把完整流量塞进 `AttackEvent.meta`。推荐新增独立证据包模型：

```text
traffic_captures
```

建议字段：

- `id`
- `event_id`
- `session_id`
- `kind`：`http_exchange`、`har`、`pcap`、`command_log`
- `storage`：`local`、`minio`、`s3`
- `object_key`
- `size`
- `sha256`
- `content_type`
- `redaction_status`
- `created_at`
- `expires_at`

`AttackEvent` 只保存引用：

```json
{
  "capture_available": true,
  "capture_ref": {
    "capture_id": 1001,
    "kind": "http_exchange",
    "redaction_status": "redacted"
  }
}
```

采集层优先从 runtime proxy 扩展，因为平台代理天然拥有用户、实例、题目、队伍和服务上下文。旁路 pcap 可以作为更后期能力，但需要额外处理容器网络定位、数据量、权限和脱敏问题。

## 安全与合规边界

- 不记录完整 `Authorization`、`Cookie`、会话令牌和平台内部凭据。
- `payload_preview` 必须受长度限制，并对 `password`、`token`、`secret`、`flag` 等字段脱敏。
- 全流量证据默认只允许教师、管理员或授权评审访问。
- 证据包应具备过期清理策略，避免长期保存敏感流量。
- 下载或查看完整证据应进入审计日志。
- 对象存储中的证据包应保存 hash，便于校验完整性和排查篡改。

## 演进路线

## 详细实施计划

### Phase 1：事件级会话还原落地

后端先在 `teaching_readmodel` 内完成读模型聚合，不新增写入链路和数据库表：

- 在 `TeacherEvidenceEvent` 的内部记录结构中补充 `source`、`stage`、`student_id`、`team_id`、`contest_id`、`round_id`、`service_id`、`victim_team_id` 等归属字段。
- 扩展教师证据仓库读取范围，纳入 `submissions`、`submission_writeups`、人工评审提交、`awd_attack_logs` 和 `awd_traffic_events`。
- 新增 `GET /api/v1/teacher/students/:id/attack-sessions`，复用证据事件构建结果，按目标和时间间隔聚合为会话。
- 会话结果只从明确事件推导：正确提交或成功 AWD 攻击为 `success`，失败提交或失败 AWD 攻击为 `failed`，仅访问/请求为 `in_progress`，其余为 `unknown`。
- DTO 中预留 `capture_available` 和 `capture_ref`，第一阶段固定无完整流量证据。

### Phase 2：教师前端会话视图

前端在学生分析页证据区域接入会话接口：

- 页面加载学生详情时并行拉取 `evidence` 和 `attack-sessions`。
- 复盘区默认展示会话摘要、结果、模式、目标和关键事件路径。
- 会话内按时间线展示事件明细，并针对代理请求、AWD 攻击、AWD 流量、Writeup、人工评审提供字段化摘要。
- 对未知事件类型保留通用标签和 summary 展示，不让后续事件扩展破坏页面。

### Phase 3：筛选与归档对齐

在会话读模型稳定后继续收敛复盘体验：

- 为证据接口和会话接口补齐 `contest_id`、`round_id`、`event_type`、`result`、时间范围和分页筛选。
- 将复盘归档的证据构建逻辑迁移到同一事件构建器，确保实时页面和导出归档口径一致。
- 增加教师观察点，例如只访问未提交、连续错误提交、成功后缺少 Writeup、AWD 成功但缺少复盘材料。

### Phase 4：全流量证据包

当 runtime proxy 或对象存储证据能力准备好后，再引入 `traffic_captures`：

- 数据库只保存证据包索引、hash、大小、脱敏状态和对象存储引用。
- `AttackEvent` 只暴露 `capture_ref`，不把完整请求响应写入事件 `meta`。
- 查看和下载完整证据需要教师/管理员权限，并写入审计日志。

### v1：事件级攻击过程还原

- 聚合 `audit_logs`、`submissions`、`awd_attack_logs`、`awd_traffic_events`。
- 返回统一 `AttackEvent` 和 `AttackSession`。
- 教师页面按时间线展示主要攻击过程。
- 不保存完整请求响应和 pcap。

### v2：统一事件构建器

- 把证据接口和攻击会话接口共用的事件转换逻辑收敛到教学读模型内部。
- 对 `meta` 字段做稳定 schema 和脱敏处理。
- 增加未知事件类型兜底展示。

### v3：会话切分增强

- 支持按时间间隔拆分同一题的多次尝试。
- 支持按 AWD 轮次、目标队伍、服务实例聚合。
- 支持按结果、题目、竞赛和时间范围筛选。

### v4：流量证据包

- 新增 `traffic_captures` 索引表。
- runtime proxy 可选保存脱敏 HTTP exchange 或 HAR。
- 对象存储保存大对象，数据库只保存引用和校验信息。

### v5：回放与行为识别

- 前端提供请求响应查看、HAR 下载或回放入口。
- 后端增加行为标签识别，例如 SQL 注入尝试、文件上传、疑似命令执行。
- 报告和复盘归档可引用会话和证据包。

## 不做项

第一阶段明确不做：

- 全流量 pcap 抓取。
- 命令级终端录制。
- 浏览器录屏。
- 完整 HTTP body 长期持久化。
- 自动判断漏洞类型和攻击成功原因。

这些能力可以在 `TrafficCapture` 和行为标签体系成熟后继续扩展。

## 最小验收标准

- 教师可以打开某个学生的攻击会话列表。
- 每个会话能展示访问目标、平台代理请求、Flag 提交、AWD 攻击结果等关键事件。
- 普通训练、Jeopardy 和 AWD 场景至少都有清晰的数据归属字段。
- 事件中不暴露完整敏感凭据和完整 Flag。
- `capture_ref` 字段已预留，但没有全流量证据时不会影响页面展示。
