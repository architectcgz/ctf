# 攻防证据链与教学复盘闭环设计

## 目标

在现有学生时间线和审计日志基础上，补出教师视角的“攻防证据链”能力，让平台能够回答三个问题：

1. 学生何时开始进入某题的实战利用阶段
2. 利用过程中发起了哪些关键请求与动作
3. 这次尝试最终是否形成有效提交、解题成功或中断

## 当前基础

- 运行时代理已经记录 `instance_access` 与 `instance_proxy_request`
- 学生时间线已经展示读题、访问目标、提示解锁、提交记录、利用轨迹
- 教师侧已存在学员分析页、学员时间线接口

## 当前实现状态

当前代码没有落成独立 `attack-sessions` 接口，而是通过教学读模型提供统一证据链接口：

- `GET /api/v1/teacher/students/:id/evidence`
- 可选查询参数：`challenge_id`
- 前端调用：`getStudentEvidence`
- 后端落点：`teaching_readmodel.QueryService.GetStudentEvidence`

该接口返回 `summary + events[]`，按事件类型聚合访问、代理请求、提交和 AWD 攻击等证据。当前最终事实源以这个统一证据接口为准。

## 核心问题

- 教师视角缺少“攻击会话”聚合与筛选
- 事件明细分散在时间线和审计日志中，不便教学复盘
- `meta/detail` 字段不够标准化，难以做稳定前端展示

## 设计范围

### 包含

- 基于现有数据源聚合“攻防证据事件”
- 新增教师侧证据链查询 API
- 在教师学生分析页新增“攻防证据”面板
- 补齐必要 DTO、查询服务、前端 contracts 和单测/集成测试

### 不包含

- 原始流量包抓取
- 完整 HTTP body 全量持久化
- 真实浏览器录屏
- 外部 SIEM/日志平台集成

## 核心设计

### 1. 先复用现有审计日志，不新增底层事件表

第一版不新建底层事件表，继续以 `audit_logs`、`submissions`、`challenge_hint_unlocks`、`instances` 为证据源。

### 2. 引入证据链读模型

- 证据事件：`instance_access`、`instance_proxy_request`、`challenge_submission`、AWD 攻击事件等
- 支持按 `challenge_id` 过滤
- 返回总事件数、代理请求数、提交数、成功数等摘要

第一版不再对外暴露单独会话对象，而是让前端在学生分析页和复盘归档中按事件类型整理阅读结构。

### 3. 教师侧复盘接口

- `GET /api/v1/teacher/students/:id/evidence`
  - 支持 `challenge_id`
  - 返回 `summary` 与 `events[]`
  - `events[].meta` 保存可展示的安全截断元数据

### 4. 教师侧复盘页面

- 在学员分析页增加“攻防证据”分区
- 支持：
  - 证据事件列表
  - 请求摘要展开
  - 按题目筛选

### 5. 教学评价闭环

- 复盘页展示“教师观察点”：
  - 是否频繁盲打
  - 是否先读题再访问目标
  - 是否通过提示推进
  - 是否存在连续错误提交
- 第一阶段不做自动评分，只做可解释摘要

## 数据结构建议

### TeacherEvidenceResp

- `summary.total_events`
- `summary.proxy_request_count`
- `summary.submit_count`
- `summary.success_count`
- `summary.challenge_id`
- `events[].type`
- `events[].challenge_id`
- `events[].title`
- `events[].detail`
- `events[].timestamp`
- `events[].meta`

## 数据与接口原则

- 优先复用现有 `audit_logs`
- 不单独引入重型事件存储
- 查询侧在 `teaching_readmodel` 增加证据链 DTO
- 第一阶段不做完整 HTTP body 全量留存，只保留安全截断预览

## 风险

- 审计字段不统一会导致前端大量兜底判断
- 教师页面如果直接读取原始审计日志，交互会很差
- 证据过多时需要分页或按会话折叠

## 最小交付

- 教师可在学员分析页查看某学员某题的攻防证据链
- 可看到访问、代理请求、提示、提交四类关键证据
- 可展开查看每次请求的核心元数据
- 可复制为答辩中演示的“漏洞利用过程复盘”
