# AWD Phase 9 教师复盘归档与报告导出设计

## 目标

在现有 AWD 引擎 `phase2-8` 的基础上，补齐一条面向教师的正式复盘链路，让平台从“管理员可看运维态势”升级成“教师可按整场赛事做教学复盘，并导出正式归档材料”。

本轮要解决的问题：

- 教师端缺少一条面向 AWD 赛事的正式复盘入口
- 现有 AWD “导出当前轮复盘包”仍是前端本地拼 JSON，不是平台正式归档能力
- 现有 `/academy/reports` 是通用班级报告页，不适合承载 AWD 赛事复盘
- 证据包导出、教师阅读型报告、教师端复盘页目前不是同一份事实源
- 毕设要求中的“导出实训报告供教学复盘”还没有落成 AWD 对抗场景下的正式交付

## 非目标

- 不重做学生侧 AWD 比赛页面
- 不改 AWD 计分、checker、流量采集和 readiness gate 的既有语义
- 不引入独立下载系统或独立导出任务队列
- 不做班级维度的 AWD 权限隔离
- 不把教师复盘页做成管理员运维页的视觉复制
- 不在本轮引入复杂的赛后 AI 结论生成

## 当前背景

现有仓库已经具备：

- AWD 轮次、攻击日志、服务检查、流量事件、排行榜和 readiness gate
- 管理端 `ContestManage -> AWDOperationsPanel / AWDRoundInspector` 的运维视图
- 通用报告任务、状态轮询、文件下载和过期控制
- 教师端已有班级报告、学生复盘归档等教学型页面

但当前还有三个明显断点：

- AWD 证据还停留在管理员运维视角，没有教师端赛事复盘入口
- 管理端“导出当前轮复盘包”是浏览器侧临时 JSON，不是后端正式归档
- 教师端现有 `/academy/reports` 是通用班级报告页，和 AWD 赛事复盘的对象模型不一致

同时，课题文档要求平台同时满足：

- 团队对抗
- 自动监控攻击流量
- 实时排行榜
- 实训数据沉淀
- 导出实训报告供教学复盘

这决定了 phase9 不能只补一个下载按钮，而要形成“页面可读 + 归档可导 + 事实源统一”的完整链路。

## 已确认范围

本轮设计按下面这些用户确认的边界执行：

- 主交付：后端正式导出链路和教师端可用复盘页一起做
- 主使用方：教师端优先
- 主视角：按单场 AWD 竞赛组织，而不是按班级或单个学生组织
- 页面粒度：整场和单轮并重，整场为主，单轮作为一等切片能力
- 教师可见范围：允许查看整场比赛所有队伍与完整对抗数据
- 时间窗口：赛中可看，赛后是正式导出窗口
- 导出成品：同时提供证据归档包和教师阅读型报告
- 底层顺序：先做结构化证据归档包，再由同一份 archive 渲染教师报告
- 下载链路：复用现有 report task / status / download 基础设施
- 业务生成链路：从现有 `report_service` 中分层，避免把 AWD 复盘逻辑继续塞成大杂糅
- 导航调整：移除 `/academy/reports` 独立教师报告页

## 方案比较

### 方案 A：教师端直接复用 admin AWD inspector

做法：

- 直接把现有 AWD 运维查询接口暴露给教师端
- 教师页主要复用 `AWDRoundInspector` 的结构
- 后端只补一个正式导出接口

优点：

- 开发最快
- 前端复用率高

缺点：

- 教师端会绑定管理员运维视角
- 整场摘要、单轮切片、教学报告容易各自拼装
- 后续很难维持“页面与导出同事实源”

不采用。

### 方案 B：新增 AWD review archive 事实层，页面和导出都基于同一份 archive

做法：

- 新增教师端 AWD 赛事目录页和详情页
- 后端先统一构建 `AWDReviewArchive`
- 页面查询、证据包导出、教师报告导出都从同一份 archive 出发
- 继续复用现有 report task / status / download

优点：

- 最符合毕设里“教学复盘 + 导出报告”的要求
- 整场、单轮、导出三条链路事实一致
- 后续继续加教学摘要、队伍下钻也不会再拆数据口径

缺点：

- 需要补一层新的 archive builder / renderer
- 前后端都要新增教师端专用契约

推荐采用。

### 方案 C：先做导出工厂，页面只做轻量壳

做法：

- 先把后端证据包和教师报告导出做完
- 教师端页面只做任务入口和状态查看

优点：

- 后端归档链路清晰

缺点：

- 本轮教师端可用性不足
- 页面价值偏弱，不符合“直接交付可用后台复盘页”的目标

不采用。

## 决策

采用方案 B。

phase9 的核心不是“新增一个 AWD 下载接口”，而是新增一份正式的 `AWDReviewArchive` 事实模型，并由它同时支撑：

1. 教师端 AWD 赛事目录页
2. 教师端整场/单轮复盘详情页
3. 结构化证据归档包导出
4. 教师阅读型复盘报告导出

## 信息架构与路由

### 教师端新入口

新增教师工作台一级入口：

- `/academy/awd-reviews`

用途：

- 展示可复盘的 AWD / AWD+ 赛事目录
- 作为教师进入 AWD 复盘的统一入口

设计约定：

- 延续现有 teacher workspace 的 dark-surface 体系
- 不复用管理员页面的运维布局
- 侧边栏新增 `AWD复盘` 菜单项

### 详情页

新增详情页：

- `/academy/awd-reviews/:contestId`

路由查询参数：

- `round=<roundNumber>`：切换到单轮复盘
- `team=<teamId>`：打开队伍下钻视图

页面主骨架保持一套，不为整场和单轮再拆两张页面：

- 默认进入整场复盘总览
- 切换轮次时仍停留在同一路由，通过 query 切片
- 队伍视图只作为下钻层，不作为一级页面

### `/academy/reports` 迁移

本轮移除：

- `/academy/reports` 路由
- `ReportExport` 视图作为教师工作台独立页面
- 侧边栏中的通用“报告导出”入口

迁移原则：

- AWD 赛事导出统一收进 `/academy/awd-reviews/:contestId`
- 现有班级训练报告导出不删除后端能力，但不再保留独立总页面
- 原来从班级管理、班级详情、学生分析跳去 `/academy/reports` 的按钮，需要改成所在上下文内的局部导出动作或直接移除旧跳转

## 页面内容设计

### 赛事目录页

目录页负责“先选比赛，再进入复盘”，不直接堆日志。

每张赛事卡展示：

- 比赛标题
- 模式、状态、时间窗口
- 当前轮次或已结束标记
- 队伍数、总轮次数
- 最近证据更新时间
- 关键态势摘要：
  - 异常服务数
  - 成功攻击数
  - 当前是否允许正式导出

目录页支持：

- 状态筛选：进行中 / 已结束
- 关键词检索：按比赛标题
- 快捷进入：点击卡片进入详情页

### 详情页整场总览

详情页默认先看整场总览，服务教师的第一阅读目标是“看清一场比赛”。

整场总览分 5 块：

- 顶部赛事头部
  - 比赛状态、时间、轮次、队伍数
  - 当前是赛中观察态还是赛后复盘态
  - 导出入口：证据归档包 / 教师报告
- 总览指标
  - SLA / attack / defense / total 聚合
  - 成功攻击数、异常服务数、活跃攻击队伍数
- 趋势与排行榜
  - 当前整场排行榜
  - 分数按轮次变化
- 服务与对抗态势
  - 服务状态分布
  - 高频告警
  - 受攻击最重的 service / team
- 教学复盘摘要
  - 供教师阅读的观察结论
  - 不是原始日志平铺

### 单轮复盘

单轮是详情页的一等切片视图，不另起独立页面。

单轮内容至少包含：

- 本轮 KPI
- 本轮服务矩阵
- 本轮攻击日志列表
- 本轮流量热点与事件
- 本轮队伍对比

页面骨架和导出入口保持不变，只切换到当前轮次的数据切片。

### 队伍下钻

队伍视图只作为详情页内的下钻层。

推荐形态：

- 页内抽屉
- 或右侧扩展面板

内容包含：

- 该队整场表现摘要
- 当前选中轮次下的服务状态
- 攻击/被攻击记录
- 关键 checker / traffic 样本

## 数据模型

### 核心事实模型

新增统一事实结构：

- `AWDReviewArchive`

它既可作为页面查询响应的基础结构，也可作为导出任务的内部中间模型。

顶层结构至少包含：

- `generated_at`
- `scope`
  - `contest_id`
  - `round_id`（可选）
  - `snapshot_type`：`live` 或 `final`
- `contest`
- `overview`
- `readiness`
- `scoreboard`
- `rounds`
- `selected_round`
- `teams`
- `evidence`
- `teaching_observations`
- `warnings`

### `contest`

至少包含：

- `id`
- `title`
- `mode`
- `status`
- `started_at`
- `ended_at`
- `freeze_time`
- `team_count`
- `round_count`

### `overview`

至少包含：

- `total_sla_score`
- `total_attack_score`
- `total_defense_score`
- `total_score`
- `successful_attack_count`
- `failed_attack_count`
- `service_alert_count`
- `active_attacker_team_count`
- `victim_team_count`
- `latest_evidence_at`
- `export_ready`

### `rounds`

每轮摘要至少包含：

- `round_id`
- `round_number`
- `status`
- `started_at`
- `ended_at`
- `service_alert_count`
- `successful_attack_count`
- `failed_attack_count`
- `top_attack_team`
- `top_victim_team`

### `selected_round`

当页面切到单轮或导出按单轮切片时返回。

至少包含：

- `round`
- `summary`
- `service_matrix`
- `attacks`
- `traffic_summary`
- `traffic_events`
- `team_breakdown`

### `teams`

整场队伍索引与摘要，至少包含：

- `team_id`
- `team_name`
- `rank`
- `total_score`
- `sla_score`
- `attack_score`
- `defense_score`
- `successful_attack_count`
- `service_alert_count`

### `evidence`

统一放复盘证据，不让页面再各自查各自拼。

首版至少包含三类事实：

- `services`
  - 来自 `awd_team_services`
- `attacks`
  - 来自 `awd_attack_logs`
- `traffic`
  - 来自 `awd_traffic_events`

### `teaching_observations`

这是教师页和教师报告都会直接消费的一层教学摘要。

首版不做复杂生成，只做规则化归纳，至少包含：

- `summary`
- `items`
  - `key`
  - `label`
  - `severity`
  - `summary`
  - `evidence_refs`

## 导出设计

### 导出类型

本轮新增两类 report type：

- `awd_review_archive`
- `awd_review_report`

用途区分：

- `awd_review_archive`：结构化证据归档包
- `awd_review_report`：教师阅读型报告

### 证据归档包

首版采用：

- `zip`

包内建议至少包含：

- `manifest.json`
- `overview.json`
- `rounds/<round-number>.json`
- `services.csv`
- `attacks.csv`
- `traffic.csv`
- `teams.csv`

导出范围：

- 默认支持整场导出
- 当用户当前选中某一轮时，支持按当前轮切片导出
- 归档包 manifest 里需要显式声明导出范围和时间

### 教师阅读型报告

首版只做：

- `pdf`

内容组织：

- 赛事摘要
- 关键指标
- 轮次走势
- 队伍对比
- 关键异常与证据样本
- 教学观察结论

注意：

- 教师报告不重新查库
- 直接消费同一份 `AWDReviewArchive`

## 接口设计

### 教师目录接口

新增：

- `GET /api/v1/teacher/awd/reviews`

职责：

- 返回教师可查看的 AWD / AWD+ 赛事目录
- 首版按完整赛事范围返回，不做班级隔离

### 教师详情接口

新增：

- `GET /api/v1/teacher/awd/reviews/:id`

查询参数：

- `round_id`（可选）
- `team_id`（可选）

职责：

- 返回 `AWDReviewArchive` 查询态
- 支撑整场总览、单轮切片和队伍下钻

### 导出接口

新增：

- `POST /api/v1/teacher/awd/reviews/:id/export/archive`
- `POST /api/v1/teacher/awd/reviews/:id/export/report`

请求体建议支持：

- `round_id`（可选）

返回：

- 继续复用现有 `ReportExportData`

下载与轮询：

- 继续复用：
  - `GET /api/v1/reports/:id`
  - `GET /api/v1/reports/:id/download`

## 后端设计

### 责任分层

本轮按两层分工：

- 复用层
  - `ReportService`
  - report repository
  - report status / download / expiry
- 新增层
  - `AWDReviewArchiveBuilder`
  - `AWDReviewArchiveQueryService`
  - `AWDReviewArchiveRenderer`

边界约定：

- `ReportService` 继续只做任务创建、异步执行、状态变更和下载授权
- `AWDReviewArchiveBuilder` 负责聚合 contest / round / service / attack / traffic / scoreboard / readiness 数据
- `AWDReviewArchiveRenderer` 负责把 archive 渲染成 zip 或 pdf

### 组装来源

builder 复用现有数据源：

- contest 元信息
- AWD 轮次
- `awd_team_services`
- `awd_attack_logs`
- `awd_traffic_events`
- live scoreboard / summary
- readiness 摘要

首版不新增新的事实表，优先由查询层聚合。

### 任务创建

导出任务流程：

1. 教师发起导出请求
2. `ReportService` 创建 `reports` 记录
3. 异步任务调用 `AWDReviewArchiveBuilder`
4. 根据导出类型调用 renderer
5. 文件落盘后标记 `ready`
6. 前端轮询状态并下载

### 权限

首版教师 AWD 复盘权限按整场开放：

- 教师可以查看所选比赛的全场完整数据
- 不按班级过滤对手队伍与证据

这符合本轮已确认范围，但权限校验仍要求：

- 仅 `teacher` / `admin` 可访问教师 AWD 复盘接口
- 学生不可访问

## 前端设计

### 新页面

新增：

- `TeacherAWDReviewIndex`
- `TeacherAWDReviewDetail`

首版建议拆出独立 composable：

- `useTeacherAwdReviewIndex`
- `useTeacherAwdReviewDetail`

### 设计系统要求

前端必须显式沿用现有教师工作台视觉系统：

- teacher workspace dark-surface
- 现有标题、eyebrow、metric panel、workspace shell
- 不把 admin 的高密度运维表格直接搬进教师页

### 导出交互

详情页顶部提供：

- `导出证据归档包`
- `导出教师报告`

状态约定：

- 赛中：
  - 页面可看
  - 证据快照导出可选
  - 正式教师报告默认置灰或提示赛后开放
- 赛后：
  - 两类导出都开放

### `/academy/reports` 去除后的页面迁移

本轮前端还要完成：

- 删除教师侧边栏中的 `/academy/reports`
- 删除或迁移所有直接 `router.push({ name: 'ReportExport' })` 的旧入口

迁移原则：

- AWD 导出入口迁移到新详情页
- 班级报告导出保留为班级或学生上下文内动作，不再通过独立总页承载

## 错误处理

### 查询态错误

明确区分：

- 比赛不存在
- 比赛不是 AWD / AWD+
- 当前教师无访问权限
- 指定轮次不存在

这些错误要优先落到页面状态，不用统一 toast 糊掉。

### 构建态错误

archive 构建时允许部分证据缺失，但要把问题显式写进 `warnings`。

只有下面这些情况才让任务整体失败：

- 核心比赛元信息无法加载
- 选中轮次不存在
- 文件渲染失败
- 文件落盘失败

### 导出态错误

当 report 任务失败时，错误要通过现有 `ReportExportData.error_message` 回传给前端。

## 验证策略

### 后端

至少覆盖：

- 赛事目录查询
- 整场 archive 组装
- 单轮切片组装
- 教师权限校验
- `awd_review_archive` 导出任务创建、轮询和下载
- `awd_review_report` 导出任务创建、轮询和下载

### 前端

至少覆盖：

- 赛事目录页渲染与筛选
- 详情页整场 / 单轮切换
- 赛中 / 赛后导出按钮状态
- `/academy/reports` 移除后的侧边栏与旧入口迁移

### 验收标准

本轮完成后，教师应能在一条连续工作流里完成：

1. 进入 `AWD复盘`
2. 选择一场 AWD 比赛
3. 查看整场复盘总览
4. 切换某一轮查看单轮证据
5. 导出结构化证据归档包
6. 在赛后导出教师阅读型报告

## 风险与取舍

- 移除 `/academy/reports` 会牵动现有多个教师页面入口，必须同步迁移，不能只删路由
- 教师端开放全场数据会弱化班级边界，但这是当前已确认的产品口径
- 如果把 archive builder 直接塞进现有 `report_service.go`，后续维护成本会明显升高，因此本轮必须先把业务生成层分离
- 首版教师报告只做 PDF，可先保证正式归档可交付；后续再考虑更多格式
