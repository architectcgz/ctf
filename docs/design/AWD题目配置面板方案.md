# AWD 题目配置面板方案

> 状态：Draft
> 事实源：中间方案、交互取舍与待落地页面组织，不覆盖当前 `docs/architecture/` 最终事实
> 替代：无

## 定位

本文档保留后台 AWD service 配置入口的中间交互方案和页面组织取舍。

- 当前文档保留方案比较、交互取舍和待落地设计，不覆盖 `docs/architecture/` 中的最终事实。
- 如果后续能力落地并稳定，应把最终边界和页面事实回收到 `docs/architecture/features/` 或 `docs/architecture/frontend/pages/`。

## 目标

在现有后台 `ContestManage / ContestOperations -> AWD 运维视图` 内补上 AWD service 配置入口，让管理员可以在同一工作台里完成以下动作：

- 查看当前 AWD 赛事已关联 service 的 checker 配置
- 从 AWD service template 新增赛事 service
- 编辑 `checker_type / checker_config / awd_sla_score / awd_defense_score`
- 保持现有后台暗色工作台和 CTF 管理页的视觉语言一致

## 非目标

- 不新增独立路由或独立后台页面
- 不重做现有 AWD 巡检、攻击日志、流量态势页面结构
- 不把 AWD service 模板库做成独立赛场引擎
- 不扩展学员侧 AWD 比赛页

## 当前背景

后端这条链路当前已经收口到 `contest_awd_services`：

- `GET /api/v1/admin/contests/:id/awd/services` 返回赛事 service 列表
- `POST /api/v1/admin/contests/:id/awd/services` 创建赛事 service
- `PUT /api/v1/admin/contests/:id/awd/services/:sid` 更新赛事 service 配置
- `DELETE /api/v1/admin/contests/:id/awd/services/:sid` 删除赛事 service
- `contest_challenges` 只保留赛事题目关系字段，不再承载 AWD checker 运行配置

前端当前落点：

- `api/admin.ts` 提供 `listContestAWDServices / createContestAWDService / updateContestAWDService / deleteContestAWDService`
- `usePlatformContestAwd` 统一管理 AWD service 配置、readiness、轮次和 inspector 状态
- `AWDOperationsPanel` 与 `ContestChallengeOrchestrationPanel` 共同承接配置和运行态入口

因此当前文档以 `contest_awd_services` 为事实源，不再把 contest challenge 字段当作配置契约。

## 方案比较

### 方案 A：继续把题目配置堆进现有轮次态势页

做法：

- 在当前 AWD inspector 页面末尾继续叠加题目配置区块

优点：

- 代码改动最少

缺点：

- 当前 inspector 已经很长
- 轮次态势和题目配置是两类不同任务，继续堆叠会增加认知负担

### 方案 B：在 AWD 运维页内增加二级面板切换

做法：

- 继续保留“选择 AWD 赛事”的外层入口
- 在 AWD 运维页内部增加两个互斥面板：
  - `轮次态势`
  - `题目配置`

优点：

- 不增加新路由
- 把“看态势”和“配 service”拆开
- 最符合当前后台工作台的 top-tabs 交互习惯

缺点：

- 需要补一层局部 tab 状态与组件拆分

### 方案 C：单独新建 AWD 题目配置页面

做法：

- 为 AWD 题目配置新起一个后台页面或子路由

优点：

- 页面职责最清楚

缺点：

- 会把同一赛事的 AWD 运营工作割裂成两处入口
- 当前阶段改造面过大

## 决策

采用方案 B。

这轮在 `AWDOperationsPanel` 内增加局部二级 tab，把现有 inspector 留在 `轮次态势`，新增一个 `题目配置` 面板专门处理 `contest_awd_services` 配置。

## 交互设计

### 入口位置

入口保持在现有 `赛事管理 -> AWD 运维视图` 内，不新增页面跳转。

管理员先选择 AWD 赛事，然后在局部面板中切换：

- `轮次态势`
- `题目配置`

这样既保留当前检查面板的连续性，也避免把配置表单硬塞进巡检页面。

### 题目配置面板

面板由两部分构成：

1. 概览区
   - 已关联题目数
   - 已配置 Checker 数
   - `HTTP 标准 Checker` 数
   - 隐藏题目数
2. 目录区
   - 按行展示每个已关联题目
   - 每行包含：
     - 题目标题与分类
     - 可见性、顺序、分值
     - Checker 类型
     - `SLA / 防守` 分
     - Checker 配置摘要
     - 编辑操作

目录风格沿用当前后台 flat directory row，而不是再做一层卡片宫格。

### 新增与编辑

这轮统一使用一个对话框组件：

- `新增` 模式
  - 选择题库中的题目
  - 设置分值、顺序、可见性
  - 设置 AWD checker 配置
- `编辑` 模式
  - 锁定当前题目
  - 更新上述配置

对话框字段：

- `template_id`（新增时必选）
- `display_name`
- `points`
- `order`
- `is_visible`
- `checker_type`
- `awd_sla_score`
- `awd_defense_score`
- `checker_config` JSON

当前枚举先支持：

- `legacy_probe`
- `http_standard`

`checker_config` 继续以 JSON 文本框方式录入，避免这轮把 `put_flag/get_flag/havoc` 展开成更大的表单编辑器。

## 视觉与一致性约束

这轮 UI 必须遵守现有 CTF 后台工作台语言：

- 继续使用现有暗色 workspace surface
- 沿用 `top-tabs`、`workspace-directory-section`、`metric-panel-card` 等已验证模式
- 不做新的高饱和配色或独立设计语言
- 行操作保持“主动作强调、次动作中性”的后台风格
- 所有说明文案保持任务导向，不渲染实现说明

## 数据流设计

前端需要补齐四段数据流：

1. `AdminContestAWDServiceData` 接入 AWD 配置字段
2. `api/admin.ts` 新增：
   - `createContestAWDService`
   - `updateContestAWDService`
   - `deleteContestAWDService`
3. `usePlatformContestAwd` 增加：
   - AWD service template 列表加载
   - 新增/更新 `contest_awd_services`
   - 变更后刷新 service 列表、`challengeLinks` 映射和 readiness
4. `AWDOperationsPanel` 负责：
   - 局部 tab 状态
   - 打开新增/编辑对话框
   - 向题目配置面板透传数据和事件

## 校验与错误处理

对话框最小校验规则：

- 新增时必须选择 service template
- `points` 必须大于 0
- `order / awd_sla_score / awd_defense_score` 必须大于等于 0
- `checker_config` 必须是合法 JSON 对象

这轮不在前端深度校验 `http_standard` 内部字段完整性，保持与后端契约一致，由后端做最终校验和拒绝。

## 测试策略

至少覆盖：

1. API 契约
   - 管理端 AWD service 列表能归一化 AWD 配置字段
   - 新增和更新接口按后端当前契约发送字段
2. 运维页行为
   - AWD 运维页可切换到题目配置面板
   - 可打开新增与编辑对话框
   - 保存后会调用对应 API，并刷新展示
3. UI 一致性
   - 不破坏现有 AWD inspector
   - 空态、目录区、动作按钮符合现有后台模式

## 验收标准

完成后应满足：

1. 管理员在 AWD 运维页内即可查看每个 contest service 的 checker 配置
2. 管理员可以从 service template 新增一个赛事 service 并写入 AWD 配置
3. 管理员可以编辑已有题目的 checker 类型、配置 JSON、SLA 分和防守分
4. 原有 AWD 轮次态势、巡检导出和攻击日志面板继续正常工作
5. 新面板视觉上与当前 CTF 后台工作台保持一致
