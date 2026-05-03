# 教学复盘优化设计

## 目标

当前教师复盘已经具备学员进度、能力画像、时间线、攻防证据链、Writeup、人工评审和复盘归档等基础能力，但整体仍偏“数据罗列”。本设计目标是把教师复盘从“查看若干列表”优化为“围绕学生、题目、攻击过程和教学观察的复盘工作台”。

优化后，教师应能回答以下问题：

1. 这个学生完成了哪些题、在哪些方向薄弱。
2. 某次解题或攻击过程是怎样推进的。
3. 哪些请求、提交、AWD 攻击和 Writeup 可以作为复盘证据。
4. 这次训练是否形成了“实操、命中、复盘输出”的闭环。
5. 哪些学生需要课堂讲解、补题或重点关注。

## 当前代码状态

现有教师学生分析页主要由以下链路组成：

- 前端页面：`TeacherStudentAnalysis.vue`
- 页面状态：`useTeacherStudentAnalysisPage`
- 展示组件：`StudentAnalysisPage`、`StudentInsightPanel`
- 教师证据接口：`GET /api/v1/teacher/students/:id/evidence`
- 后端落点：`teaching_readmodel.QueryService.GetStudentEvidence`
- 复盘归档：`assessment.ReportService.BuildStudentReviewArchive`
- AWD 复盘页：`TeacherAWDReviewDetail.vue`

代码中已经具备以下数据来源：

- `audit_logs`：实例访问和平台代理请求。
- `submissions`：Flag 提交结果。
- `submission_writeups`：学生 Writeup。
- `manual review submissions`：人工评审型提交。
- `awd_attack_logs`：AWD 攻击提交。
- `awd_traffic_events`：AWD 流量摘要。
- `skill profile`：能力画像。

## 主要问题

### 1. 实时复盘与归档复盘能力不一致

教师实时证据接口当前主要聚合：

- `instance_access`
- `instance_proxy_request`
- `challenge_submission`

复盘归档仓库已经额外纳入 `awd_attack_logs`。这会造成同一名学生在“教师学员分析页”和“复盘归档页”看到的证据范围不一致。教师日常查看时可能看不到 AWD 攻击提交，但导出归档时又能看到相关记录。

### 2. 证据链没有会话化

当前证据接口返回平铺事件列表。教师需要自行从多个事件中判断一次训练或攻击过程的开始、尝试、提交和结果。这个结构适合证明“系统记录了事件”，但不适合课堂复盘。

### 3. 筛选维度不足

后端证据接口支持 `challenge_id`，但前端页面默认一次性拉取全部证据，没有按题目、竞赛、时间范围、事件类型或结果筛选。事件数量增加后，页面会退化成流水账。

### 4. 前端展示信息不足

现有证据卡片主要展示标题、描述、阶段和时间。对教师复盘更有价值的字段没有充分呈现，例如：

- 请求方法
- 请求路径
- 状态码
- payload 预览
- 是否命中
- 得分
- AWD 受害队伍
- AWD 服务标识
- 轮次

此外，后端代理请求 meta 中使用的是 `request_method`、`target_path`、`status_code` 等字段，而前端部分展示逻辑读取的是 `method`，存在字段语义未对齐的问题。

### 5. AWD 复盘和学生复盘割裂

AWD 复盘页偏比赛态势和队伍视角，适合观察某一轮服务、攻击和流量情况；学生分析页偏个人训练视角。两者没有统一到“某个学生在某场 AWD 中攻击了哪些目标、产生了哪些流量、结果如何”的过程复盘。

## 总体设计

教学复盘优化分为三层：

```text
EvidenceEvent       统一证据事件，屏蔽底层表差异
AttackSession       攻击/解题会话，将事件组织成过程
ReviewWorkspace     教师复盘工作台，围绕学生、题目、会话和观察点展示
```

### EvidenceEvent

统一证据事件是教学复盘的基础读模型。它不改变原始业务写入链路，只把不同事实源转换为统一结构：

```json
{
  "id": "evt_1",
  "type": "proxy_request",
  "stage": "exploit",
  "source": "audit_logs",
  "occurred_at": "2026-05-01T10:30:00+08:00",
  "student_id": 1001,
  "challenge_id": 12,
  "contest_id": null,
  "round_id": null,
  "service_id": null,
  "victim_team_id": null,
  "title": "SQL 注入",
  "summary": "POST /login 返回 200",
  "meta": {
    "request_method": "POST",
    "target_path": "/login",
    "status_code": 200,
    "payload_preview": "username=admin&password=***"
  }
}
```

事件类型第一阶段至少覆盖：

- `instance_access`
- `proxy_request`
- `challenge_submission`
- `awd_attack_submission`
- `awd_traffic`
- `writeup`
- `manual_review`

### AttackSession

攻击会话用于把零散事件整理成教师可阅读的过程。会话不作为第一阶段的事实表保存，而是查询时聚合：

```json
{
  "id": "sess_12_1",
  "mode": "practice",
  "student_id": 1001,
  "challenge_id": 12,
  "title": "SQL 注入",
  "started_at": "2026-05-01T10:20:00+08:00",
  "ended_at": "2026-05-01T10:45:00+08:00",
  "result": "success",
  "event_count": 6,
  "events": []
}
```

第一版会话切分规则：

- 普通训练按 `student_id + challenge_id` 聚合。
- Jeopardy 竞赛按 `student_id/team_id + contest_id + challenge_id` 聚合。
- AWD 按 `student_id/team_id + contest_id + service_id + victim_team_id` 聚合。
- 同一目标两次事件间隔超过配置阈值时可拆分为多个会话。

### ReviewWorkspace

教师复盘工作台以“学生”为入口，但展示不再只是列表堆叠，而是按复盘任务组织：

- 总览：完成度、正确提交、实操事件、Writeup、人工评审和最近活动。
- 薄弱方向：能力画像与推荐练习。
- 攻击过程：按会话展示访问、请求、提交、AWD 攻击和结果。
- 证据详情：展示事件级元数据和安全截断内容。
- 复盘输出：Writeup、人工评审、教师观察点和归档导出。

## 接口设计

### 保留现有证据接口

```text
GET /api/v1/teacher/students/:id/evidence
```

增强方向：

- 支持 `challenge_id`
- 支持 `contest_id`
- 支持 `round_id`
- 支持 `event_type`
- 支持 `from`、`to`
- 支持分页
- 纳入 `awd_attack_logs` 和必要的 `awd_traffic_events`

该接口继续用于事件级证据列表。

### 新增攻击会话接口

```text
GET /api/v1/teacher/students/:id/attack-sessions
```

查询参数：

- `mode`：`practice`、`jeopardy`、`awd`
- `challenge_id`
- `contest_id`
- `round_id`
- `result`
- `with_events`
- `limit`
- `offset`

响应结构：

```json
{
  "summary": {
    "total_sessions": 4,
    "success_count": 2,
    "failed_count": 1,
    "in_progress_count": 1,
    "event_count": 21
  },
  "sessions": []
}
```

### 归档接口对齐

复盘归档应复用与实时页面一致的事件构建逻辑，避免“页面看到一套、导出看到另一套”。如果短期无法完全复用，也要保证事件类型、字段名和 summary 统计口径一致。

## 后端改造方案

### 1. 收敛事件构建逻辑

在 `teaching_readmodel` 内部增加统一事件构建器：

```text
buildEvidenceEvents(userID, query) -> []EvidenceEvent
```

它负责从以下来源读取并转换事件：

- `audit_logs`
- `submissions`
- `awd_attack_logs`
- `awd_traffic_events`
- `submission_writeups`
- 人工评审提交

`GetStudentEvidence`、`GetAttackSessions` 和复盘归档都应尽量复用这套事件转换逻辑。

### 2. 补齐 AWD 个人证据

教师学生分析页需要看到学生个人维度的 AWD 证据：

- 通过 `awd_attack_logs.submitted_by_user_id` 找到学生提交的攻击。
- 通过 `awd_traffic_events.attacker_team_id` 关联学生所在队伍，展示相关攻击流量摘要。
- 对于没有具体 `submitted_by_user_id` 的历史记录，可以降级展示队伍级证据，并在 meta 中标记 `scope=team`。

### 3. 增加会话聚合服务

在应用层增加会话聚合逻辑：

```text
events -> group by target -> split by time gap -> derive result -> build sessions
```

结果判断规则：

- 存在正确普通提交或成功 AWD 攻击：`success`
- 存在错误提交或失败 AWD 攻击，但没有成功事件：`failed`
- 只有访问或请求，没有提交：`in_progress`
- 无法判断：`unknown`

### 4. 统一 meta 字段

建议规范代理请求字段：

- `request_method`
- `target_path`
- `target_query`
- `status_code`
- `payload_preview`
- `payload_truncated`

前端不再读取模糊的 `method` 字段。对于兼容旧数据，可以在前端或后端做一次 fallback。

## 前端改造方案

### 1. 重组学生分析页复盘区

将当前“攻防证据链”区升级为“复盘工作台”：

- 顶部展示复盘摘要：会话数、实操请求数、正确提交数、AWD 攻击数、Writeup 数。
- 左侧或顶部提供筛选：题目、模式、结果、事件类型、时间范围。
- 主体默认展示攻击会话，而不是平铺全部事件。
- 每个会话可展开查看事件明细。

### 2. 会话卡片信息结构

每个会话卡片展示：

- 题目或 AWD 服务名称。
- 模式：训练、Jeopardy、AWD。
- 结果：成功、失败、进行中、未知。
- 开始与结束时间。
- 事件数量。
- 关键路径摘要，例如 `访问目标 -> POST /login 200 -> 提交成功`。

### 3. 事件明细展示

事件明细按类型展示重点字段：

- 访问事件：实例、题目、访问时间。
- 代理请求：方法、路径、状态码、payload 预览。
- 普通提交：正确性、得分。
- AWD 攻击：攻击队伍、受害队伍、服务、轮次、是否成功、得分。
- AWD 流量：方法、路径、状态码、攻击方、受害方。
- Writeup：标题、状态、更新时间。

### 4. 复盘观察点

前端展示由后端计算或前端轻量推导的观察点：

- 是否有实操请求。
- 是否存在连续错误提交。
- 是否只访问未提交。
- 是否成功后提交了 Writeup。
- AWD 是否存在有效攻击但缺少复盘材料。

观察点只做辅助提示，不直接替代教师判断。

## 与攻击过程还原设计的关系

`attack-session-replay-evolution-design.md` 关注攻击过程还原与后续全流量回放演进；本文关注教师复盘工作台整体体验。二者关系如下：

- 本文使用 `AttackSession` 作为教师复盘的核心展示单元。
- 全流量回放不是本次优化的一期目标。
- 未来当 `TrafficCapture` 落地后，教师复盘工作台只需要在事件详情中显示“查看流量证据”入口。

## 分阶段交付

## 详细实施计划

### Phase 1：修齐实时证据源

目标是让教师学生分析页和复盘归档看到同一类事实：

- 后端 `teaching_readmodel.Repository.GetStudentEvidence` 纳入 `awd_attack_logs`、`awd_traffic_events`、`submission_writeups` 和人工评审提交。
- AWD 攻击按 `submitted_by_user_id` 优先归属到学生；历史或队伍级事件通过 `team_members` 降级归属到学生所在队伍，并在 `meta.scope` 中标记。
- 代理请求 meta 统一使用 `request_method`、`target_path`、`target_query`、`status_code`、`payload_preview`。
- 教师证据摘要把普通提交和 AWD 攻击提交统一计入提交/成功统计。

### Phase 2：新增攻击会话接口

目标是把平铺证据组织成课堂可读的过程：

- 新增 `GET /api/v1/teacher/students/:id/attack-sessions`。
- 应用层复用证据事件，按 `practice / jeopardy / awd` 三类目标聚合。
- 同一目标两次事件间隔超过 1 小时拆分为不同会话。
- 支持 `mode`、`challenge_id`、`contest_id`、`round_id`、`result`、`with_events`、`limit`、`offset` 查询参数。
- 响应返回 summary 与 sessions，便于前端先展示会话列表，再展开事件明细。

### Phase 3：前端复盘工作台

目标是把学生分析页证据区从“事件列表”改为“复盘工作台”：

- `useTeacherStudentAnalysisPage` 并行加载攻击会话。
- `StudentInsightPanel` 默认展示会话数、事件数、成功会话和实操请求数。
- 会话卡片展示题目/服务、模式、结果、开始时间、事件数量和关键路径摘要。
- 事件明细展示请求方法、路径、状态码、payload 预览、得分、受害队伍、服务、轮次、Writeup 标题和人工评审状态。
- 没有完整流量证据时不展示流量入口；后续 `capture_available=true` 后再补入口。

### Phase 4：归档和观察点

目标是让实时复盘、归档导出和教学判断口径一致：

- 将复盘归档切换为同一事件构建器或至少同步事件类型、字段名和 summary 统计口径。
- 基于会话和证据事件生成观察点：实操请求、连续错误提交、访问未提交、成功后 Writeup、AWD 命中但缺少复盘材料。
- 增加筛选控件，并将有分享价值的筛选条件同步到路由 query。

### Phase 1：修齐证据源和展示字段

- 教师实时证据接口纳入 `awd_attack_logs`。
- 统一代理请求 meta 字段。
- 前端展示请求方法、路径、状态码、payload 预览、成功状态和得分。
- 修复前端读取 `method` 与后端 `request_method` 不一致的问题。

### Phase 2：攻击会话读模型

- 新增 `attack-sessions` 接口。
- 按题目、竞赛、AWD 目标聚合事件。
- 前端将证据区默认展示为会话列表。
- 保留事件列表作为展开详情。

### Phase 3：筛选与教学观察点

- 增加题目、模式、结果、时间和事件类型筛选。
- 增加复盘摘要和教师观察点。
- 将 Writeup、人工评审和攻击会话关联展示。

### Phase 4：归档复用同一读模型

- 复盘归档和实时复盘页面共用事件构建逻辑。
- 归档导出包含会话摘要、关键事件和教师观察点。
- 确保页面展示、JSON 归档和论文描述口径一致。

### Phase 5：流量证据扩展

- 接入 `TrafficCapture` 索引。
- 在事件详情中展示 `capture_ref`。
- 支持脱敏请求响应摘要或 HAR 证据下载。

## 验收标准

- 教师能在学生分析页看到普通训练和 AWD 两类证据。
- 教师能按会话理解一次攻击或解题过程，而不是只能看平铺事件。
- 代理请求能显示方法、路径、状态码和安全截断 payload。
- AWD 攻击事件能显示轮次、受害队伍、服务和成功状态。
- 复盘归档与实时页面的事件类型和统计口径一致。
- 没有完整流量数据时页面仍然可用；后续有 `capture_ref` 时可以自然扩展。

## 不做项

本阶段不做：

- 完整流量回放。
- pcap 抓包。
- 命令级录制。
- 自动判定漏洞利用类型。
- 自动生成最终教师评价结论。

这些能力应在攻击会话读模型稳定后继续演进。
