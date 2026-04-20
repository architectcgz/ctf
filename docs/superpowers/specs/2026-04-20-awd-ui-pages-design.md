# CTF AWD 前端页面全量升级设计

## 背景

当前仓库里的 AWD 前端能力已经分散落在 3 个入口里：

- 学生侧通过 `ContestDetail` 内嵌 `ContestAWDWorkspacePanel`
- 管理员侧通过 `ContestEdit` 里的 AWD 配置、赛前检查和运行段
- 教师侧通过 `TeacherAWDReviewDetail` 单页承载整场复盘

这些实现已经覆盖了一部分真实数据流，但页面职责混杂、入口不统一、文档里的 19 个 AWD 页面也没有真正形成独立路由体系。

本次升级的目标不是补几个子面板，而是把 AWD 前端整体收敛成新的工作台体系，完整覆盖设计文档 `docs/design/awd-ui-pages-vue3-ts-tailwind.md` 里的 19 个页面，并移除旧 AWD 页面作为主入口的角色。

## 目标

- 按文档完整落地学生端 5 页、管理员端 9 页、教师端 5 页，共 19 个 AWD 页面
- 以 3 套角色工作台替代旧 AWD 入口，不保留旧页面作为主 UI
- 能接真实接口的页面优先接真实数据，缺口页通过 adapter 层补齐 mock
- 三族页面统一使用共享 shell 结构：顶部赛事上下文、左侧页内导航、主内容画布
- 所有 AWD 页面显式处理 loading、error、empty、success 等非 happy path

## 非目标

- 不扩展后端接口范围，只消费当前已有接口并做前端适配
- 不顺带重做非 AWD 的普通竞赛、排行榜、题目页信息架构
- 不保留旧 AWD 单页 UI 作为并行方案

## 页面范围

### 学生端

- 战场总览
- 我的服务
- 目标目录
- 攻击记录
- 队伍协作

### 管理员端

- AWD 比赛总览
- Readiness 面板
- 轮次控制台
- Service 健康矩阵
- 攻击日志
- 流量态势
- 告警中心
- 实例运维
- 赛后复盘与导出

### 教师端

- 教学总览
- 队伍复盘
- Service 复盘
- 轮次回放
- 报告导出

## 总体方案

推荐采用“3 个角色工作台 + 19 个子页面视图”的结构。

理由：

- 与现有路由、Vue 组件边界和 API 形态最契合
- 可以复用学生战场、管理员运行段、教师复盘已经接好的真实数据流
- 缺后端字段的页面可以在同一路由体系下先通过 adapter 层补 mock
- 后续新增接口时只替换 adapter，不用推翻页面结构

不采用以下方案：

- `19` 个页面各自独立实现：会重复消费同一批 AWD 数据，样式和交互边界容易漂移
- 单一超级 AWD 页面：会把学生、管理员、教师三种使用场景强行压平，侵入现有站内结构过大

## 信息架构

### 共享 shell

三族页面都共享同一种骨架：

1. 顶部赛事上下文区
2. 左侧页内导航
3. 主内容画布

共享的不是颜色，而是页面结构。角色差异只放在：

- 强调色
- 文案语气
- 默认信息密度
- 默认首屏关注点

### 角色风格

#### 学生端

- 强调当前轮、目标切换、flag 提交反馈
- 默认首屏聚焦战场态势和我方动作
- 更强调单次操作反馈与快速切页

#### 管理员端

- 强调赛事健康度、异常信号和运营控制
- 默认首屏聚焦 KPI、轮次状态和风险告警
- 信息密度最高

#### 教师端

- 强调证据阅读、轮次回放和报告导出
- 默认首屏聚焦复盘摘要和案例线索
- 交互频率低于学生端与管理员端

## 路由重构

### 学生端路由

- `/contests/:id/awd/overview`
- `/contests/:id/awd/services`
- `/contests/:id/awd/targets`
- `/contests/:id/awd/attacks`
- `/contests/:id/awd/collab`

### 管理员端路由

- `/platform/contests/:id/awd/overview`
- `/platform/contests/:id/awd/readiness`
- `/platform/contests/:id/awd/rounds`
- `/platform/contests/:id/awd/services`
- `/platform/contests/:id/awd/attacks`
- `/platform/contests/:id/awd/traffic`
- `/platform/contests/:id/awd/alerts`
- `/platform/contests/:id/awd/instances`
- `/platform/contests/:id/awd/replay`

### 教师端路由

- `/academy/awd-reviews/:contestId/overview`
- `/academy/awd-reviews/:contestId/teams`
- `/academy/awd-reviews/:contestId/services`
- `/academy/awd-reviews/:contestId/replay`
- `/academy/awd-reviews/:contestId/export`

## 旧页面处理策略

旧 AWD 页面可以移除掉，不再保留为并行入口。

### 学生端

- 移除 `ContestDetail.vue` 里的 AWD 内嵌战场 UI
- 对 AWD 比赛，详情入口直接进入新的学生 AWD workspace

### 管理员端

- `ContestEdit.vue` 不再承载 AWD 运行工作台
- AWD 配置、赛前检查、运行态势拆到新的管理员 AWD 路由下
- `ContestEdit` 只保留普通竞赛编辑职责，或作为跳转壳存在

### 教师端

- `TeacherAWDReviewDetail.vue` 不再继续承载所有复盘场景
- 教师 AWD 复盘改为多页路由结构
- 旧复盘详情页拆薄为 shell 或路由跳转壳

### 保留策略

- 可以保留极薄的过渡重定向路由，避免站内已有链接立刻失效
- 旧 AWD 页面不再承载实际主 UI，不再作为功能入口存在

## 目录结构

建议新增独立模块目录：

```text
src/modules/awd/
  adapters/
    studentAwdPageAdapter.ts
    adminAwdPageAdapter.ts
    teacherAwdPageAdapter.ts
  components/
    AwdContextHero.vue
    AwdPageNav.vue
    AwdMetricStrip.vue
    AwdEventTimeline.vue
    AwdServiceHealthMatrix.vue
    AwdRoundRail.vue
    AwdTargetDirectory.vue
    AwdReplayPanel.vue
    AwdExportPanel.vue
  layouts/
    StudentAwdWorkspaceLayout.vue
    AdminAwdWorkspaceLayout.vue
    TeacherAwdWorkspaceLayout.vue
  views/
    student/
    admin/
    teacher/
```

## 组件分层

### layout 层

负责：

- 顶部赛事上下文
- 左侧页内导航
- 主内容插槽
- 通用 loading / error / empty / permission 壳

不负责：

- 页面业务数据拼装
- 细分场景交互

### adapter 层

负责：

- 将现有接口结构整理成页面消费的 view model
- 用集中 mock 填补当前后端缺失字段
- 统一字段格式与状态语义

不负责：

- 直接渲染 UI
- 到处散落页面私有 mock

### page view 层

负责：

- 单页展示和交互
- 消费 layout 提供的上下文与 adapter 输出的数据

每个页面只关注自己，不再自己拼赛事标题、轮次、导航和主壳结构。

## 数据承接策略

### 学生端

优先复用：

- `getContestAWDWorkspace`
- `getScoreboard`
- `useContestAWDWorkspace`

可直接支撑：

- 战场总览
- 我的服务
- 目标目录
- 攻击记录

`队伍协作` 先通过同源战场数据 + adapter 层补协作 view model，后续如果有专门接口再替换。

### 管理员端

优先复用：

- `usePlatformContestAwd`
- `AWDOperationsPanel` 已消费的数据流
- `AWDReadinessData`
- `AWDRoundSummaryData`
- `AWDTrafficSummaryData`
- attack / service / team / challenge 相关接口

可直接支撑：

- Readiness 面板
- 轮次控制台
- Service 健康矩阵
- 攻击日志
- 流量态势

需要 adapter 或 mock 补齐：

- AWD 比赛总览
- 告警中心
- 实例运维
- 赛后复盘与导出

### 教师端

优先复用：

- `getTeacherAWDReview`
- `useTeacherAwdReviewDetail`
- 现有导出链路

可直接支撑：

- 教学总览
- 队伍复盘
- 轮次回放
- 报告导出

需要 adapter 或 mock 补齐：

- Service 复盘

## 真实数据与 mock 的边界

- 只要现有后端接口能支撑，就必须接真实数据
- 缺字段的页面通过 adapter 层补齐 mock
- mock 只能集中放在 adapter 层，不允许散落到页面组件里
- UI 不显示“mock 数据”“演示数据”等说明性文字

对用户来说，19 个页面都应表现为正式产品页面。

## 非 happy path 设计

所有 AWD 页面必须显式覆盖以下状态：

- loading
- error
- empty
- success
- no-permission
- unavailable-by-status

高风险交互必须处理重复点击与竞态：

- 学生端启动服务、提交 stolen flag
- 管理员端创建轮次、补录检查、补录攻击、处理 readiness
- 教师端导出归档、导出报告

轮次切换、筛选切换和自动刷新要以“最后一次请求结果”为准，避免旧响应覆盖新视图。

## 验证口径

本次完成不能只看“页面能打开”，至少要验证：

- 19 个 AWD 页面路由全部可达
- 三族 shared shell 切页时赛事上下文稳定
- 真实接口页能正常加载、刷新、切筛选
- adapter 补齐页在缺真实字段时也能完整渲染
- 关键动作具备加载态、禁用态和错误回退

建议补充的测试：

- 路由与重定向测试
- adapter 映射测试
- workspace layout 渲染测试
- 学生提交动作 focused test
- 管理员操作 focused test
- 教师导出 focused test

## 完成定义

满足以下条件，才算本次 AWD UI 升级完成：

- 文档里的 19 个页面全部在新 AWD 路由体系下落地
- 三族共享 shell 完成并成为唯一主入口结构
- 能接真实接口的页面全部接入真实数据
- 不能接真实接口的页面通过 adapter 层补齐 mock，页面完整可用
- 旧 AWD 页面不再作为主入口存在

## 风险与权衡

### 风险

- 旧 AWD UI 与新 AWD 路由并存时，容易出现入口分裂和维护双轨
- 如果直接在旧页面里堆 tab，会继续扩大职责污染
- adapter 设计不清晰时，mock 容易渗入页面组件

### 应对

- 本次明确移除旧 AWD 页面主入口角色
- 通过模块化目录隔离 AWD UI 新结构
- 将真实字段映射与 mock 补齐统一收口到 adapter 层

## 实施顺序建议

1. 建立 AWD 路由族和三族 shared shell
2. 接入学生端 5 页并替换旧战场入口
3. 接入管理员端真实数据页并拆出剩余 4 页
4. 接入教师端真实数据页并补齐 Service 复盘
5. 清理旧 AWD 页面主入口和站内跳转
6. 完成 focused test 与最小充分验证
