# 攻防证据链与教学复盘闭环设计

## 目标

在现有学生时间线和审计日志基础上，补出教师视角的“攻击会话复盘”能力，让平台能够回答三个问题：

1. 学生何时开始进入某题的实战利用阶段
2. 利用过程中发起了哪些关键请求与动作
3. 这次尝试最终是否形成有效提交、解题成功或中断

## 当前基础

- 运行时代理已经记录 `instance_access` 与 `instance_proxy_request`
- 学生时间线已经展示读题、访问目标、提示解锁、提交记录、利用轨迹
- 教师侧已存在学员分析页、学员时间线接口

## 核心问题

- 教师视角缺少“攻击会话”聚合与筛选
- 事件明细分散在时间线和审计日志中，不便教学复盘
- `meta/detail` 字段不够标准化，难以做稳定前端展示

## 设计范围

### 包含

- 基于现有数据源聚合“攻击会话”
- 新增教师侧复盘查询 API
- 在教师学生分析页新增“攻击复盘”面板
- 补齐必要 DTO、查询服务、前端 contracts 和单测/集成测试

### 不包含

- 原始流量包抓取
- 完整 HTTP body 全量持久化
- 真实浏览器录屏
- 外部 SIEM/日志平台集成

## 核心设计

### 1. 先复用现有审计日志，不新增底层事件表

第一版不新建底层事件表，继续以 `audit_logs`、`submissions`、`challenge_hint_unlocks`、`instances` 为证据源。

### 2. 引入“攻击会话”读模型

- 会话起点：`instance_access`
- 会话内事件：`instance_proxy_request`、`hint_unlock`、`flag_submit`
- 会话终点：
  - 正向结束：成功提交
  - 中断结束：实例销毁/过期
  - 未知结束：长时间无活动

第一版按“实例 ID + 时间窗口”聚合，确保与现有代理审计天然对齐。

### 3. 教师侧复盘接口

- `GET /api/v1/teacher/students/:id/attack-sessions`
  - 支持 `challenge_id`、`status`、`limit`
- `GET /api/v1/teacher/students/:id/attack-sessions/:session_id`
  - 返回单次会话详情，包括关键步骤、请求摘要、提交结果、会话总结

### 4. 教师侧复盘页面

- 在学员分析页增加“攻击复盘”分区
- 支持：
  - 会话列表
  - 会话步骤
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

### AttackSessionListItem

- `id`
- `student_id`
- `challenge_id`
- `challenge_title`
- `instance_id`
- `started_at`
- `ended_at`
- `status`
- `request_count`
- `submit_count`
- `last_event_summary`

### AttackSessionDetail

- `session`
- `events[]`
- `summary`

## 数据与接口原则

- 优先复用现有 `audit_logs`
- 不单独引入重型事件存储
- 查询侧在 `teaching_readmodel` 增加聚合 DTO
- 第一阶段不做完整 HTTP body 全量留存，只保留安全截断预览

## 风险

- 审计字段不统一会导致前端大量兜底判断
- 教师页面如果直接读取原始审计日志，交互会很差
- 证据过多时需要分页或按会话折叠

## 最小交付

- 教师可在学员分析页查看某学员某题的攻击证据链
- 可看到访问、代理请求、提示、提交四类关键证据
- 可展开查看每次请求的核心元数据
- 可复制为答辩中演示的“漏洞利用过程复盘”
