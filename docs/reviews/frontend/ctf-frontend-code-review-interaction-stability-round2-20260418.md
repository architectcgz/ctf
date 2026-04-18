# CTF 前端代码 Review（交互稳定性与架构质量专项 第 2 轮）

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | 前端交互稳定性、异步边界、性能与组件耦合 |
| 轮次 | 第 2 轮 |
| 审查范围 | `code/frontend/src` 主要视图、composables、stores、通用样式原语 |
| 审查日期 | 2026-04-18 |
| 审查方式 | 静态代码审查，聚焦竞态、边界状态、泄漏风险、响应式误区、组件职责 |
| 审查状态 | 持续修复中 |

## 当前状态

- 已完成：
  - `H1` 高频查询取消链路已落地到管理员用户目录、审计日志、教师实例、通知发布用户搜索。
  - `H2` 题目管理发布状态二次拼装的旧结果回写问题已通过请求代次保护收口，并补了回归测试。
  - `H3` 后台学生目录已切到服务端分页查询，不再在页面层聚合全量班级学生。
  - `M1` 题目管理页已补错误态和重试入口。
  - `M2` 教师实例筛选延迟请求已在卸载时清理。
  - `M3` 后台题目详情的延迟跳转已补 timer 清理。
  - `M6` 通知详情已移除静态禁用占位按钮，改为说明块。
  - `L2` 后台页面级按钮原语已完成收敛，目标范围内不再残留 `admin-btn` / `publish-btn` / `template-action-btn` / `topology-toolbar-btn` 等私有按钮族。
  - `L1` 共享布局层与通用弹层已继续收敛一批裸魔法值，`Sidebar`、`TopNav`、`AppToast`、`CLightActionPopover`、`CFocusedInputDialog`、`CImmersiveConfirmDialog`、`MinimalFloatingModal`、`SlideOverDrawer`、`ClassicCenteredModal`、`CContextTooltip`、`ModalTemplateShell` 不再保留 `w-[260px]`、`max-w-[1600px]`、`rounded-[22px]`、局部绿色按钮色或大段亮色硬编码。
  - `L1` 个人中心页已继续清理 `UserProfile`、`SkillProfile` 的骨架圆角、图表高度和内文字色魔法值，改为语义类承接，不再把 `rounded-[24px]`、`h-[30rem]`、`text-[11px]` 和模板内 `text-[var(...)]` 直接写在结构节点上。
  - `L1` 学生端 journal soft 工作区已补共享空态面板类，`StudentCategoryProgressPage`、`StudentDifficultyPage`、`StudentRecommendationPage`、`StudentTimelinePage` 不再重复内联 `rounded-[22px]`、`border-[var(--journal-shell-border)]` 这类空态块样式；教师 `TeacherAWDReviewIndex`、`TeacherAWDReviewDetail` 的加载骨架也已切到语义类承接。
  - 教师班级工作区已补 `TeacherClassWorkspaceSection` 桥接视图，旧明细路由会回落到统一 `TeacherClassStudents` 工作区，并通过 `panel` 查询参数恢复目标标签页。
- 未完成：
  - `M4` 教师分析面板和 AWD 轮次面板的 props 透传与职责过载。
  - `M5` 超大组件继续拆分。
  - `L1` Tailwind 裸魔法数清理。

## 问题清单

### 🔴 高优先级

- [H1] 高频查询链路没有真正接入请求取消，`useAbortController` 处于“只写未用”状态
  - 状态：已完成
  - 处理结果：
    - `useAdminUsers`、`AuditLog`、`useTeacherInstances`、`useAdminNotificationPublisher` 均已接入取消或请求代次保护。
    - 通知发布抽屉的延迟搜索已改为显式 timer，并在关闭/卸载时清理，避免迟到搜索。

- [H2] 题目管理发布状态的二次拼装存在真实竞态，列表切换后可能回写错误状态
  - 状态：已完成
  - 处理结果：
    - 已引入独立的 `latestPublishRequestsToken`，旧批次结果不会覆盖当前列表。
    - 已补“翻页后旧发布状态结果被忽略”的回归测试。

- [H3] 后台学生目录默认拉取全班级学生后再本地排序/分页，数据量一大就会拖垮页面
  - 状态：已完成
  - 处理结果：
    - 后台学生目录已改为 `getStudentsDirectory` 服务端分页查询，并通过 `useStudentDirectoryQuery` 统一调度。

### 🟡 中优先级

- [M1] 题目管理页缺失错误态，接口失败时会退化成“空目录”
  - 状态：已完成
  - 处理结果：
    - `ChallengeManage` 已补 `error` 分支，接口失败时显示错误态和重试按钮，不再伪装成空目录。

- [M2] 教师实例筛选的 debounce 没有在卸载时取消，存在迟到请求风险
  - 状态：已完成
  - 处理结果：
    - 教师实例页已改成显式 timer 管理，并在 `onUnmounted` 中清理定时器与中止请求。

- [M3] 后台题目详情的延迟跳转未清理，存在卸载后迟到导航
  - 状态：已完成
  - 处理结果：
    - 题目详情已收口 `redirectTimer` 生命周期，卸载时不会再触发迟到跳转。

- [M4] 教师分析面板与 AWD 轮次面板已经出现明显的 props 透传与职责过载
  - 状态：未完成
  - 说明：
    - 这是结构性问题，不适合和本轮稳定性修复混在一个提交里，需要单独拆分实施。

- [M5] 管理端存在超大页面文件，虽然已抽出 composable，但单文件职责仍然过重
  - 状态：未完成
  - 说明：
    - 需要按工作区块继续拆组件，属于较大规模重构，不是这一轮的最小修复项。

- [M6] 通知详情仍保留静态禁用按钮，属于低信息度占位
  - 状态：已完成
  - 处理结果：
    - 相关区域已改为说明态文案，不再渲染无来源的禁用按钮。

### 🟢 低优先级

- [L1] Tailwind 任意值与魔法数仍是系统性问题，但需区分 token bridge 与裸像素
  - 状态：未完成
  - 说明：
    - 已完成 `NotificationDropdown`、`PageHeader`、`AppCard`、`Sidebar`、`TopNav`、`AppToast`、`CLightActionPopover`、`CFocusedInputDialog`、`CImmersiveConfirmDialog`、`MinimalFloatingModal`、`SlideOverDrawer`、`ClassicCenteredModal`、`CContextTooltip`、`ModalTemplateShell`、`UserProfile`、`SkillProfile`、`StudentCategoryProgressPage`、`StudentDifficultyPage`、`StudentRecommendationPage`、`StudentTimelinePage`、`TeacherAWDReviewIndex`、`TeacherAWDReviewDetail` 这批共享组件/壳层/通用弹层/个人中心页面/学生工作区/教师 AWD 页面 的低信息度裸字号、裸尺寸、亮色硬编码和 `1px` 任意值清理，相关回归测试已补齐。
    - 剩余项主要集中在表格列宽声明、实验/参考页面，以及少量需要转成 CSS token 或共享类的布局尺寸，仍需分批替换，避免把 token bridge 与真正的裸像素混改。

- [L2] 按钮原语已经存在，但页面级按钮体系仍然碎片化
  - 状态：已完成
  - 处理结果：
    - 已完成 `AdminNotificationPublishDrawer`、`ChallengeWriteupManagePanel`、`ChallengeTopologyStudioPage`、`ChallengeDetail`、`ImageManage`、`AdminDashboardPage`、`ChallengeWriteupViewPage`、`ChallengeWriteupEditorPage`、`UserGovernancePage` 的页面私有按钮族清理。
    - 后台目标范围已统一切到 `ui-btn` 原语，并通过页面级 `--ui-btn-*` token 覆盖保留各自的深色工作台视觉。
    - 经扫描，`src/views/admin`、`src/components/admin`、`src/components/notifications` 中不再残留 `admin-btn` / `publish-btn` / `publish-inline-btn` / `template-action-btn` / `topology-toolbar-btn`。

## 已完成提交

- `126caba4 fix(frontend): 收口学生目录与交互稳定性`
- `c1f95ea3 fix(frontend): 收口通知发布搜索竞态`
- `50ab319a fix(frontend): 收敛后台页面按钮原语`
- `526ed7cb fix(frontend): 收敛个人中心页面魔法值`

## 下一批建议

1. 继续做 `L1`，只清理纯魔法数，不碰 token bridge。
2. `M4`、`M5` 单独立项拆分，不与稳定性修复混提。
