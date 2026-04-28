# CTF 前端代码 Review（交互稳定性与架构质量专项 第 2 轮）

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | 前端交互稳定性、异步边界、性能与组件耦合 |
| 轮次 | 第 2 轮 |
| 审查范围 | `code/frontend/src` 主要视图、composables、stores、通用样式原语 |
| 审查日期 | 2026-04-18 |
| 审查方式 | 静态代码审查，聚焦竞态、边界状态、泄漏风险、响应式误区、组件职责 |
| 审查状态 | 持续补充中，本文件先记录当前已确认问题 |

## 当前结论

- 已确认问题主要集中在 4 类：
  - 高频筛选与轮询链路只做了“旧响应丢弃”，但真正的请求取消没有落地。
  - 个别后台目录页仍存在“拉全量后本地筛选/分页”的重型实现。
  - 少数页面仍缺失错误态承接，接口失败时会误呈现为空列表。
  - 管理端和教师端存在超大组件与高密度 props 透传，维护成本持续升高。
- 本轮没有发现明显的 `const { foo } = props` 响应式误用。
- 本轮没有发现明显把纯局部 UI 状态滥塞进 Pinia 的情况，现有 `auth` / `contest` / `notification` store 仍以业务共享状态为主。

## 问题清单

### 🔴 高优先级

- [H1] 高频查询链路没有真正接入请求取消，`useAbortController` 处于“只写未用”状态
  - 文件：
    - `code/frontend/src/composables/useAbortController.ts:1-24`
    - `code/frontend/src/composables/useAdminUsers.ts:60-87`
    - `code/frontend/src/views/admin/AuditLog.vue:164-213`
  - 问题描述：
    - 仓库里已经有 `useAbortController`，但当前实际引用只出现在测试中，生产代码没有接入。
    - `useAdminUsers`、`AuditLog` 这类高频筛选页面仍依赖 debounce 和重新发请求，缺少 `AbortController` 或显式的“只接受最后一次请求结果”封装。
  - 影响范围/风险：
    - 用户快速切换筛选或离开页面时，旧请求仍继续占用网络与服务端资源。
    - 复杂链路里很容易出现 toast、loading、分页状态与当前筛选条件不同步的问题。
  - 修正建议：
    - 对所有高频筛选列表建立统一的“可取消请求”查询封装，优先把 `useAbortController` 真正接到 API 层。
    - 至少先覆盖管理员用户目录、审计日志、教师实例、通知发布用户搜索这几类高频场景。

- [H2] 题目管理发布状态的二次拼装存在真实竞态，列表切换后可能回写错误状态
  - 文件：`code/frontend/src/composables/useAdminChallenges.ts:88-124`
  - 问题描述：
    - 题目列表本身通过 `usePagination` 带了 `latestRequestId` 防护，但发布状态是后续用 `Promise.all` 二次拼出来的，没有请求代次校验，也没有取消。
    - 快速切页、切筛选或轮询过程中，上一批 `getLatestChallengePublishRequest` 返回后仍可能覆写当前页的 `latestPublishRequests`。
  - 影响范围/风险：
    - 题目行上的发布检查状态、轮询开关、完成提示可能对应错题目集合。
    - 这是后台运维视图，状态错位会直接误导管理操作。
  - 修正建议：
    - 给 `loadLatestPublishRequests` 增加独立请求 token 或 abort 信号。
    - 让发布状态与当前列表快照绑定，切页/切筛选时旧批次结果必须作废。

- [H3] 后台学生目录默认拉取全班级学生后再本地排序/分页，数据量一大就会拖垮页面
  - 文件：`code/frontend/src/views/admin/StudentManage.vue:168-190`
  - 问题描述：
    - 选中“全部班级”时，页面会对每个班级并发调用 `getClassStudents`，再把结果 `flat()` 到本地。
    - 后续排序、分页都在浏览器内存里做，不是基于服务端分页。
  - 影响范围/风险：
    - 班级和学生规模变大后，初始化耗时、切筛选延迟和内存占用都会线性放大。
    - 这类目录页属于后台高频页，卡顿会直接影响运维和教学管理效率。
  - 修正建议：
    - 改成服务端统一分页查询学生目录，不要在页面层聚合全量班级数据。
    - 如果短期无法改 API，至少给“全部班级”模式加结果上限、分页提示和显式加载策略。

### 🟡 中优先级

- [M1] 题目管理页缺失错误态，接口失败时会退化成“空目录”
  - 文件：
    - `code/frontend/src/views/admin/ChallengeManage.vue:423-428`
    - `code/frontend/src/composables/usePagination.ts:18-50`
  - 问题描述：
    - `usePagination` 已经暴露 `error`，但 `ChallengeManage` 只判断 `loading` 和 `list.length === 0`，没有单独渲染失败态。
    - 当接口报错且列表为空时，用户看到的是“当前还没有题目”，而不是加载失败。
  - 影响范围/风险：
    - 管理员会误判为题库为空，掩盖真实接口故障。
  - 修正建议：
    - 接入 `error` 分支，使用统一 `AppEmpty` 错误态并提供重试入口。

- [M2] 教师实例筛选的 debounce 没有在卸载时取消，存在迟到请求风险
  - 文件：`code/frontend/src/composables/useTeacherInstances.ts:91-127`
  - 问题描述：
    - `scheduleInstanceSearch` 被创建后只在 watch 中触发，没有 `onUnmounted` 清理。
    - 页面切走时仍可能触发一次延迟的 `loadInstances()`。
  - 影响范围/风险：
    - 会产生额外请求，极端情况下还可能在离开页面后继续弹出错误 toast。
  - 修正建议：
    - 统一要求所有 `useDebounceFn` 在 composable 卸载时执行 `cancel()`。

- [M3] 后台题目详情的延迟跳转未清理，存在卸载后迟到导航
  - 文件：`code/frontend/src/views/admin/ChallengeDetail.vue:530-538`
  - 问题描述：
    - 加载失败时通过裸 `setTimeout` 延迟 1.5 秒跳回列表，但没有保存 timer 引用，也没有在卸载时 `clearTimeout`。
  - 影响范围/风险：
    - 用户在这 1.5 秒内切换到别的页面，仍可能被旧定时器强行跳回。
  - 修正建议：
    - 把 timer 引用挂到组件作用域，并在 `onBeforeUnmount` 中清理。

- [M4] 教师分析面板与 AWD 轮次面板已经出现明显的 props 透传与职责过载
  - 文件：
    - `code/frontend/src/components/teacher/StudentInsightPanel.vue:34-69`
    - `code/frontend/src/components/teacher/class-management/StudentAnalysisPage.vue:242-266`
    - `code/frontend/src/components/admin/contest/awdInspector.types.ts:14-50`
  - 问题描述：
    - `StudentInsightPanel` 需要接收一整组学生画像、题解、人工评审、证据流 props。
    - `AWDRoundInspector` 也已经发展到高密度输入/输出接口，父组件与子组件之间靠大包 props/emits 维持。
  - 影响范围/风险：
    - 任何一个字段变化都要同步修改父子组件与测试，回归面扩大。
    - 组件复用和拆分难度持续升高。
  - 修正建议：
    - 以领域拆分子面板，并把跨层共享但不适合全局 store 的只读上下文改成 `provide/inject` 或专用 composable。

- [M5] 管理端存在超大页面文件，虽然已抽出 composable，但单文件职责仍然过重
  - 文件：
    - `code/frontend/src/components/admin/topology/ChallengeTopologyStudioPage.vue`（2885 行）
    - `code/frontend/src/components/admin/contest/AWDRoundInspector.vue`（1475 行）
    - `code/frontend/src/components/teacher/StudentInsightPanel.vue`（1082 行）
  - 问题描述：
    - `ChallengeTopologyStudioPage` 虽然把核心逻辑挪进 `useChallengeTopologyStudioPage`，但页面层仍一次性承接大量状态、操作和样式分支。
  - 影响范围/风险：
    - 阅读、测试和局部演进成本都偏高，后续很容易继续长成“上帝组件”。
  - 修正建议：
    - 继续按工作区块拆成独立子组件，例如工具条、模板目录、画布侧栏、摘要条等。

- [M6] 通知详情仍保留静态禁用按钮，属于低信息度占位
  - 文件：`code/frontend/src/views/notifications/NotificationDetail.vue:210-221`
  - 问题描述：
    - 页面直接渲染一个固定 `disabled` 的“暂无关联对象”按钮，没有状态来源，也不提供解释。
  - 影响范围/风险：
    - 这是典型的静态 UI 占位，既不表达真实业务状态，也会误导后续实现。
  - 修正建议：
    - 改成条件渲染的说明块，或等关联对象能力接入后再恢复可操作入口。

### 🟢 低优先级

- [L1] Tailwind 任意值与魔法数仍是系统性问题，但需区分 token bridge 与裸像素
  - 文件：
    - `code/frontend/src/components/layout/NotificationDropdown.vue:63-85`
    - `code/frontend/src/components/admin/contest/AWDRoundInspector.vue:197-210`
    - `code/frontend/src/components/common/SkillRadar.vue:53-80`
  - 问题描述：
    - 仓库里仍存在 `w-[1px]`、`rounded-[28px]`、`text-[11px]`、`h-[280px]` 这类任意值。
    - 其中一部分是与设计 token 对接的 bridge，另一部分则是直接把像素写死。
  - 影响范围/风险：
    - 会削弱设计系统的可维护性，后续统一调整尺寸和节奏时成本偏高。
  - 修正建议：
    - 先清理纯魔法数，再决定哪些 token bridge 需要保留。

- [L2] 按钮原语已经存在，但页面级按钮体系仍然碎片化
  - 文件：
    - `code/frontend/src/assets/styles/workspace-shell.css:169-240`
    - `code/frontend/src/components/admin/writeup/ChallengeWriteupManagePanel.vue:263-265`
    - `code/frontend/src/components/notifications/AdminNotificationPublishDrawer.vue:286-293`
    - `code/frontend/src/components/admin/topology/ChallengeTopologyStudioPage.vue:50-125`
  - 问题描述：
    - 仓库已有 `ui-btn` / `teacher-btn` / `journal-btn` 这些基础原语，但页面里仍并存 `admin-btn`、`publish-btn`、`template-action-btn`、`topology-toolbar-btn`。
  - 影响范围/风险：
    - 暗色模式、尺寸体系、交互反馈和禁用态会持续出现不一致。
  - 修正建议：
    - 先按后台、教师端、学生端三类视图收敛按钮族，再逐步删除页面私有按钮实现。

## 已验证的正向模式

- `usePagination`、`useContestDetailPage`、`useContestAWDWorkspace`、`useTeacherInstances` 已经开始使用请求序号保护，说明团队已经有“只接受最后一次响应”的意识。
- `TopNav`、`useInstanceListPage`、`useChallengeTopologyStudioPage`、`useWebSocket` 等事件与定时器场景大多有成对清理，本轮未发现系统性的 `addEventListener` / `setInterval` 泄漏。
- `auth` / `contest` / `notification` store 目前仍以共享业务状态为主，没有明显把页面级弹窗、局部筛选一股脑塞进全局 store。
- 针对 `props` 解构丢失响应式的专项扫描里，本轮未发现明确误用。

## 后续建议

1. 先统一高频查询的取消策略，把 `AbortController` 从“工具存在”变成“生产可用”。
2. 把后台学生目录、题目目录、审计日志这三类后台列表页统一成同一套边界状态规范：`loading / error / empty / retry`。
3. 按领域把 `StudentInsightPanel`、`AWDRoundInspector`、`ChallengeTopologyStudioPage` 继续拆小，避免后续修复都落在超大组件上。
4. 在按钮、间距和 Tailwind 任意值层面继续做设计系统收敛，减少页面私有实现。

## 变更摘要

- 新增：`docs/reviews/ctf-frontend-code-review-interaction-stability-round2-20260418.md`
