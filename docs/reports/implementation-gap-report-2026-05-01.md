# 当前实现与课题/架构文档差距报告

日期：2026-05-01

## 结论

当前平台已经覆盖课题主线能力：题目管理、题包导入、实例生命周期、动态 Flag、Jeopardy 竞赛、AWD 对抗、能力画像、报告导出和基础教学复盘均有代码落点。主要风险不在“系统不可用”，而在部分能力与课题表述、架构设计之间存在粒度差异。

优先需要补齐的是两类问题：

- 教师复盘偏弱：实时证据链、复盘归档、AWD 复盘和攻击过程还原尚未统一。
- 实验证据不足：并发、长稳、故障注入、容器隔离和资源限制仍缺少可复现实验记录。

## 对比来源

- 课题文档：`docs/毕业设计课题.md`
- 开题报告：`docs/开题报告/开题报告.md`
- 架构文档：`docs/architecture/features/攻击证据链与教学复盘架构.md`
- 架构文档：`docs/architecture/features/攻击会话读模型与复盘工作台架构.md`
- 架构文档：`docs/architecture/features/教学复盘优化设计.md`
- 架构文档：`docs/architecture/features/题包拓扑同步与导出架构.md`
- 后端路由：`code/backend/internal/app/router_routes.go`
- 教学读模型：`code/backend/internal/module/teaching_readmodel`
- 评估与报告：`code/backend/internal/module/assessment`
- 竞赛 AWD：`code/backend/internal/module/contest`
- 运行时与代理：`code/backend/internal/module/runtime`
- 前端教师页：`code/frontend/src/views/teacher`、`code/frontend/src/components/teacher`

## 差距总览

| 编号 | 方向 | 课题/架构要求 | 当前状态 | 判断 | 优先级 |
| --- | --- | --- | --- | --- | --- |
| 1 | 教师复盘 | 围绕学生、题目、攻击过程和教学观察形成复盘工作台 | 有证据链、时间线、归档、AWD 复盘，但未统一 | 需要优化 | P0 |
| 2 | 攻击步骤自动记录 | 自动记录攻击步骤与漏洞利用过程 | 已记录关键事件，未形成攻击会话 | 实现偏弱 | P0 |
| 3 | 全流量回放演进 | 为未来完整 HTTP/HAR/pcap 证据包预留 | 架构已设计，代码未落地 `TrafficCapture` | 缺失但可后置 | P2 |
| 4 | AWD 流量与防守统计 | 自动监控攻击流量、统计防守成功次数 | 已有 `awd_traffic_events`、服务状态、SLA/defense 计分 | 基础具备，语义需收敛 | P1 |
| 5 | 自定义拓扑 | 支持自定义拓扑与题包拓扑 | 已有拓扑导入、保存、导出和运行时创建基础 | 基础具备，需验证和说明 | P1 |
| 6 | 能力画像与推荐 | 生成能力画像、推荐练习、导出报告 | 已有 profile、recommendation、report 接口 | 基础具备，解释性可优化 | P2 |
| 7 | 容器隔离与资源限制 | 验证容器网络隔离、资源限制、安全边界 | 代码有隔离/调度/清理设计，实验记录不足 | 证据缺口 | P0 |
| 8 | 并发与稳定性 | 通过压测、长稳、故障注入验证 | 论文中仍标为待补 | 证据缺口 | P0 |
| 9 | 题目镜像供应链 | 降低题目包和镜像风险 | 题包校验较完整，镜像扫描/构建隔离仍弱 | 需要优化 | P2 |

## 1. 教师复盘：实时页、归档页和 AWD 复盘未统一

### 设计要求

`教学复盘优化设计.md` 要求把教师复盘从“数据罗列”升级为复盘工作台，核心结构是：

```text
EvidenceEvent
AttackSession
ReviewWorkspace
```

### 当前实现

后端已经提供：

- `GET /api/v1/teacher/students/:id/timeline`
- `GET /api/v1/teacher/students/:id/evidence`
- `GET /api/v1/teacher/students/:id/review-archive`
- `POST /api/v1/teacher/students/:id/review-archive/export`
- `GET /api/v1/teacher/awd/reviews`
- `GET /api/v1/teacher/awd/reviews/:id`

教师学生分析页通过 `getStudentEvidence` 拉取证据链，前端有 `StudentInsightPanel` 展示“攻防证据链”。复盘归档则通过 `ReportRepository.GetStudentEvidence` 重新组装证据。

### 差距

- 教师实时证据接口主要聚合 `instance_access`、`instance_proxy_request`、`challenge_submission`。
- 复盘归档已经纳入 `awd_attack_logs`，但教师实时学生分析页尚未对齐这部分 AWD 个人攻击证据。
- AWD 复盘页偏比赛/队伍/轮次视角，学生分析页偏个人视角，两者没有统一到“某个学生的一次攻击过程”。
- 前端证据卡片展示字段较少，对 `request_method`、`target_path`、`status_code`、`payload_preview` 等复盘关键字段展示不足。

### 建议

按 `教学复盘优化设计.md` 的 Phase 1 和 Phase 2 处理：

- 先让教师实时证据接口纳入 `awd_attack_logs`。
- 再新增或改造 `attack-sessions` 读模型，把证据从平铺事件升级为会话化复盘。
- 复盘归档和实时页面尽量复用同一套事件构建逻辑，避免两套口径继续漂移。

## 2. 攻击步骤自动记录：已有关键事件，但未形成自动还原

### 课题要求

`docs/毕业设计课题.md` 明确写到“系统自动记录攻击步骤与漏洞利用过程”。

### 当前实现

平台已经具备关键事件记录：

- 运行时代理记录 `instance_access` 与 `instance_proxy_request`。
- `submissions` 记录普通题提交结果。
- `awd_attack_logs` 记录 AWD 攻击提交。
- `awd_traffic_events` 记录 AWD 代理访问摘要。
- Writeup 和复盘归档能补充人工复盘材料。

### 差距

当前能力更准确地说是“关键事件级证据链”，还不是完整的“攻击步骤自动还原”：

- 没有 `GET /api/v1/teacher/students/:id/attack-sessions`。
- 没有 `AttackSession` 聚合结果。
- 没有按题目、竞赛、AWD 目标把访问、请求、提交、攻击结果串成一次过程。
- 没有请求之间的依赖关系、时序间隔、成功路径摘要。

### 建议

把论文和答辩口径定为“关键行为证据链 + 攻击过程事件级还原”。实现上优先落地：

- `EvidenceEvent` 统一事件结构。
- `AttackSession` 查询时聚合。
- 前端按 session 展示访问、利用请求、提交和结果。

完整命令录制、完整请求响应留存、全流量回放作为后续增强。

## 3. 全流量回放：架构已预留，代码未实现

### 架构要求

`攻击会话读模型与复盘工作台架构.md` 将能力拆为：

```text
AttackSession
AttackEvent
TrafficCapture
```

其中 `TrafficCapture` 用于未来保存完整 HTTP exchange、HAR、pcap 或命令记录引用。

### 当前实现

代码只保存摘要：

- 代理审计 detail 包含 `target_path`、`target_query`、`status`、`payload_preview` 等安全截断内容。
- AWD 流量表保存 method、path、status_code、attacker/victim/service 等摘要。

### 差距

- 没有 `traffic_captures` 表。
- 没有 `capture_ref` 字段。
- 没有对象存储或文件存储的证据包管理。
- 没有回放沙箱，也没有 HAR/pcap 查看或下载接口。

### 建议

此项不建议现在作为答辩前必做。当前应先完成事件级会话还原。后续做全流量时，应只在事件上挂 `capture_ref`，完整流量放对象存储，避免把大对象塞进 PostgreSQL。

## 4. AWD 流量监控与防守统计：基础具备，但需收敛复盘口径

### 课题要求

课题文档写到“支持团队对抗模式，自动监控攻击流量、统计防守成功次数，生成实时排行榜”。

### 当前实现

当前代码已有：

- `awd_attack_logs`
- `awd_traffic_events`
- `GetTrafficSummary`
- `ListTrafficEvents`
- `DefenseScore`
- `ServiceStatus`
- `DefenseSuccessCount`
- AWD 复盘详情页中的服务、攻击、流量三类证据

### 差距

- “自动监控攻击流量”当前是平台代理链路下的流量摘要，不是全流量抓包。
- 防守成功统计更接近 checker/service status 和 defense score 的统计，不是学生修补动作级别的防守过程记录。
- AWD 流量与个人学生复盘尚未完全联通。

### 建议

论文和答辩中建议使用准确表述：

- “记录平台代理访问摘要和 AWD 攻击日志”
- “基于 checker 结果统计服务状态、SLA 与防守得分”

后续如果要增强“防守过程”，需要补队伍防守动作记录，例如防守 SSH 登录、重启服务、补丁提交或文件变更摘要。

## 5. 自定义拓扑：基础具备，仍需验证和边界说明

### 课题要求

课题文档要求靶场管理支持“自定义拓扑”。

### 当前实现

代码和架构文档显示已经具备较多拓扑能力：

- 管理端拓扑 API：`GET/PUT/DELETE /api/v1/admin/challenges/:id/topology`
- 环境模板 API：`/api/v1/admin/environment-templates`
- 题包拓扑导入：`extensions.topology.source`
- 拓扑存储：`challenge_topologies`
- 拓扑导出：`docker/topology.yml`
- 运行时拓扑创建：`RuntimeTopologyCreateRequest`
- ACL 清理和拓扑运行资源回收

### 差距

这部分不属于缺失，但需要补清楚边界：

- 需要确认前端拓扑工作台是否覆盖了创建、编辑、预览、导出完整链路。
- 需要补拓扑运行时隔离验证，证明节点间策略、平台管理面隔离、资源回收有效。
- 需要明确一期支持的拓扑复杂度，避免被理解成任意复杂云网络编排。

### 建议

把该项定位为“已实现题包拓扑和平台拓扑基础能力，支持受限拓扑编排”。后续重点是验证和演示材料，而不是重做模块。

## 6. 能力画像、推荐与报告：已实现，解释性可增强

### 课题要求

课题文档要求“基于实训数据生成能力画像，推荐针对性靶场练习，导出实训报告供教学复盘”。

### 当前实现

后端已有：

- `/api/v1/users/me/skill-profile`
- `/api/v1/users/me/recommendations`
- `/api/v1/teacher/students/:id/skill-profile`
- `/api/v1/teacher/students/:id/recommendations`
- `/api/v1/reports/personal`
- `/api/v1/reports/class`
- 学生复盘归档导出
- AWD 复盘归档和教师报告导出

### 差距

- 能力画像和推荐的解释性还可以增强，例如推荐原因与具体错误/薄弱证据关联不足。
- 复盘归档和实时复盘口径存在差异，影响教师对学生过程的解释。
- 报告更多是数据汇总，自动教学观察点仍偏基础。

### 建议

短期不需要重做画像算法。优先把证据链、攻击会话、Writeup 和能力画像关联起来，让教师能看到“为什么推荐这些题”“为什么判定这个维度薄弱”。

## 7. 容器隔离与资源限制：实现有基础，实验记录不足

### 开题要求

开题报告把容器隔离、资源限制和安全边界列为关键问题，并要求通过安全隔离测试验证边界有效性。

### 当前实现

代码和文档中已有：

- 独立网络、动态端口、资源限制、非特权容器等设计。
- 两段式实例调度。
- 运行时清理、孤儿资源回收、ACL 清理。
- proxy ticket 限制访问入口。

### 差距

当前缺口主要是证据：

- 缺少容器访问平台 API、PostgreSQL、Redis、Docker socket 的隔离实验记录。
- 缺少跨实例访问、跨队访问、AWD 网络策略的验证记录。
- 缺少 CPU、内存、PIDs 限制的实测记录。
- 缺少异常容器、异常网络、端口泄漏的回收实验记录。

### 建议

答辩前建议补一组最小实验：

- 靶机访问平台内网服务失败。
- 两个普通训练实例互访失败。
- 容器资源限制触发后平台仍可用。
- 实例过期或异常后容器、网络、端口被回收。

这些实验比继续堆功能更能支撑论文可信度。

## 8. 并发、长稳与故障注入：论文已承认待补，仍是关键缺口

### 开题要求

开题报告技术路线包含并发压测、长稳运行和故障注入。

### 当前实现

论文第 5、6 章已经把这部分标为 TODO 或后续补充方向。代码中有调度、清理和维护逻辑，也有不少模块测试与集成测试。

### 差距

- 没有形成可复现实验记录。
- 没有明确实例启动并发、最大活跃实例、排行榜刷新延迟等量化结果。
- 没有故障注入案例，例如 Docker 启动失败、容器中途退出、Redis 短暂不可用、清理任务重试。

### 建议

优先补实验记录，不一定要做大规模压测：

- 10/20/50 个实例启动压测。
- 1 小时长稳运行。
- 手动停止容器后维护任务恢复或清理。
- Redis 排行榜缓存失败后的数据库回退。

## 9. 题目镜像供应链：题包校验较强，镜像安全仍可优化

### 架构要求

题包设计要求 manifest、附件、拓扑、checker 等进入导入预检；后续工作也提到镜像构建和漏洞扫描。

### 当前实现

题包导入、拓扑解析、checker 配置、预览和导出链路较完整。普通题和 AWD 题包已有较多校验逻辑。

### 差距

- 镜像构建隔离和镜像扫描不是当前主链路。
- 题包导入能检查结构和路径，但不能保证镜像内容安全。
- 没有形成“题目发布前安全检查”报告。

### 建议

该项可放到 P2。毕业设计当前只需说明题包结构预检和运行时隔离已经覆盖主要风险，镜像扫描作为后续增强。

## 建议优先级

### P0：答辩前建议优先补

1. 教师实时证据链纳入 AWD 攻击日志，修复实时页与归档页口径不一致。
2. 攻击会话读模型最小版，支撑“攻击过程自动还原”的答辩展示。
3. 容器隔离、资源限制、并发和长稳实验记录。

### P1：论文定稿前建议优化

1. 教师复盘前端展示关键 meta 字段。
2. AWD 流量、防守成功统计和学生个人复盘打通。
3. 自定义拓扑的运行时验证和演示材料。

### P2：后续增强

1. `TrafficCapture`、HAR/pcap/HTTP exchange 证据包。
2. 镜像扫描和隔离构建。
3. 更细的推荐解释和自动教学观察点。

## 对论文表述的建议

建议保持以下边界：

- 可以写：平台已实现关键行为证据链，记录实例访问、平台代理请求、Flag 提交、AWD 攻击日志、流量摘要、Writeup 和复盘归档。
- 可以写：平台支持事件级攻击过程复盘，并具备向攻击会话和全流量证据扩展的架构基础。
- 不建议写：已经完成完整攻击步骤自动还原、命令级录制、全流量回放或完整请求响应留存。
- 可以写：支持受限自定义拓扑和题包拓扑导入，不建议写成任意复杂网络仿真编排。
- 可以写：AWD 支持服务检查、攻击日志、流量摘要和防守得分统计，不建议写成完整防守操作取证。

