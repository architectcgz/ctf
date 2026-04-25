# CTF 前端专项审查（2026-04-22）

## 审查信息

| 字段 | 说明 |
|------|------|
| 审查主题 | 前端架构质量、异步边界、状态同步、主题一致性与门禁健康度 |
| 审查范围 | `code/frontend/src` 路由、关键视图、composables、stores、共享样式与验证门禁 |
| 审查日期 | 2026-04-22 |
| 审查方式 | 静态代码审查 + 最小验证基线检查 |
| 审查状态 | 已记录，已推进至第六十五轮最小高收益修复 |

## 当前结论

- 本轮专项审查起手时暴露出的高收益问题，已经有大半完成收口：
  - `typecheck` 基线已恢复，当前专项修复工作树可稳定作为静态门禁。
  - 题目详情竞态、实例页状态同步、`/ui-lab` 正式路由暴露、认证表单基础可访问性、`window.confirm` 混用、登录态 token 本地持久化等问题已完成修复，其中登录态已切到服务端 session 与 `HttpOnly` cookie。
  - 本轮继续收口了平台 AWD 编排组件簇、教师端 `workspace / tabs / surface` 漂移，以及 shared pagination 的 review 基线漂移。
- 当前仍未整体关闭专项审查，剩余问题主要集中在：
  - `P2-1` 主题硬编码清理仍未全量收口
  - `P2-5` 仍有部分大页面和管理端 surface 漂移未完成系统性减重

## 优先级结论

### P1

- [P1-1] 静态门禁失效，`code/frontend` 当前 `typecheck` 失败
  - 直接影响：
    - 无法把“本轮修改是否引入新的类型错误”作为可靠门禁。
    - 后续修复缺少稳定基线，风险会持续叠加。
  - 已确认位置：
    - `code/frontend/src/views/platform/ContestEdit.vue`
    - `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
    - `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
    - `code/frontend/src/composables/__tests__/useContestAWDWorkspace.test.ts`
  - 修复顺序：
    - 先恢复 `typecheck` 通过，再继续更高层的页面修复。

- [P1-2] 题目详情页请求缺少“只接受最后一次结果”的保护，切换题目时可能串页
  - 直接影响：
    - 题面、题解、个人题解、提交记录可能被旧请求回写。
    - 用户快速切换题目时会看到不属于当前题目的内容。
  - 已确认位置：
    - `code/frontend/src/views/challenges/ChallengeDetail.vue`
    - `code/frontend/src/composables/useChallengeDetailInteractions.ts`
  - 修复方向：
    - 为详情加载链路补 request token 或取消机制，并让多路并发加载共享同一代次。

- [P1-3] 实例列表页初始化后几乎不再和服务端同步，真实状态可能长期陈旧
  - 直接影响：
    - `pending / creating / failed / crashed` 等状态会停留在旧值。
    - 页面只更新倒计时，不更新实例真实生命周期，容易误导操作。
  - 已确认位置：
    - `code/frontend/src/composables/useInstanceListPage.ts`
    - `code/frontend/src/views/instances/InstanceList.vue`
  - 修复方向：
    - 为实例页补显式刷新与等待态轮询策略，至少保证等待中实例能自动重新拉取。

- [P1-4] 登录态令牌保存在 `localStorage`，与富文本渲染并存，爆炸半径偏大
  - 直接影响：
    - 当前未证实已有 XSS，但一旦未来任意富文本链路或第三方输入点失守，账户接管成本会很低。
  - 已确认位置：
    - `code/frontend/src/stores/auth.ts`
    - `code/frontend/src/api/request.ts`
    - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - 处理状态：
    - 已在第二十一轮切到服务端 session 与 `HttpOnly` cookie，前端不再持有认证 token；旧 `ctf_access_token / ctf_refresh_token` 仅保留清理逻辑。

### P2

- [P2-1] 主题系统没有真正收口，生产代码里仍大量使用硬编码色值
  - 已确认位置：
    - `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
    - `code/frontend/src/components/layout/BackofficeSubNav.vue`
    - `code/frontend/src/composables/useNotificationDropdown.ts`
    - `code/frontend/src/components/common/WorkspaceDataTable.vue`
  - 影响：
    - 暗色模式、品牌切换和样式一致性都会持续出现漂移。

- [P2-2] `/ui-lab` 被挂进正式公开路由，设计实验页与正式产品边界混杂
  - 已确认位置：
    - `code/frontend/src/router/index.ts`
    - `code/frontend/src/router/guards.ts`
  - 影响：
    - 设计试验内容进入正式构建和公开访问面，后续容易演变成误暴露入口。

- [P2-3] 认证表单基础可访问性不足，`label` 与 `input` 没有显式关联
  - 已确认位置：
    - `code/frontend/src/views/auth/LoginView.vue`
    - `code/frontend/src/views/auth/RegisterView.vue`
  - 影响：
    - 屏幕阅读器体验和表单可用性都偏弱。

- [P2-4] 关键交互仍混用 `window.confirm`，没有统一到平台确认对话框原语
  - 已确认位置：
    - `code/frontend/src/views/platform/UserManage.vue`
    - `code/frontend/src/composables/useContestDetailPage.ts`
    - `code/frontend/src/composables/useChallengeTopologyStudioPage.ts`
  - 影响：
    - 交互体验、主题、键盘焦点与测试方式都不一致。

- [P2-5] 路由文件和多个核心页面已经明显过重，继续叠加需求会放大维护成本
  - 已确认位置：
    - `code/frontend/src/router/index.ts`
    - `code/frontend/src/views/challenges/ChallengeDetail.vue`
    - `code/frontend/src/views/challenges/ChallengeList.vue`
    - `code/frontend/src/views/contests/ContestDetail.vue`
    - `code/frontend/src/views/platform/ChallengeDetail.vue`
    - `code/frontend/src/views/platform/ImageManage.vue`
    - `code/frontend/src/views/platform/AuditLog.vue`
    - `code/frontend/src/views/platform/ChallengeManage.vue`
    - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - 影响：
    - 阅读和回归成本高，局部修复也更容易带出连锁问题。

## 本轮修复范围

- 第一优先级：
  - 恢复 `typecheck` 通过
  - 修复 `ChallengeDetail` 请求竞态
  - 修复实例页状态同步策略
- 暂不在本轮直接处理：
  - 认证令牌存储方案重构
  - 全量主题硬编码清理
  - 超大页面系统性拆分

## 第一轮修复进展

- 已完成：
  - `P1-1` 静态门禁恢复，`code/frontend` 的 `typecheck` 已可通过。
  - `P1-2` `ChallengeDetail` 已补“旧请求作废”保护，详情与提交记录快速切换不再被旧路由回写。
  - `P1-3` 实例页已补等待态远端状态轮询，等待中的实例会重新同步服务端状态。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/ContestEdit.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - `code/frontend/src/composables/useChallengeDetailInteractions.ts`
  - `code/frontend/src/composables/useInstanceListPage.ts`
  - `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`
  - `code/frontend/src/views/instances/__tests__/InstanceList.test.ts`
  - `code/frontend/src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts`
  - `code/frontend/src/composables/__tests__/useContestAWDWorkspace.test.ts`

## 验证基线

- 主验证命令：
  - `pnpm -C code/frontend typecheck`
- 当前基线状态：
  - 主工作区审查时已确认 `typecheck` 失败。
  - 修复 worktree 中已完成依赖安装并重新建立基线。
- 第一轮修复后的验证：
  - `npm run typecheck`
  - `npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts src/views/instances/__tests__/InstanceList.test.ts`

## 第二轮修复进展

- 已完成：
  - `P2-2` `/ui-lab` 不再作为匿名公开入口，已收拢为需要登录且仅管理员可访问的内部页面。
  - `P2-3` 登录页与注册页的 `label` 已显式关联对应 `input`，基础表单可访问性补齐。
  - `P2-4` 用户删除、竞赛队员踢出、拓扑与模板删除/覆盖确认已统一接入平台确认对话框原语，不再依赖浏览器原生 `window.confirm`。
- 本轮涉及文件：
  - `code/frontend/src/router/index.ts`
  - `code/frontend/src/router/guards.ts`
  - `code/frontend/src/views/auth/LoginView.vue`
  - `code/frontend/src/views/auth/RegisterView.vue`
  - `code/frontend/src/views/platform/UserManage.vue`
  - `code/frontend/src/composables/useContestDetailPage.ts`
  - `code/frontend/src/composables/useChallengeTopologyStudioPage.ts`
  - `code/frontend/src/router/__tests__/guards.test.ts`
  - `code/frontend/src/router/__tests__/sharedRoutes.test.ts`
  - `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
  - `code/frontend/src/views/auth/__tests__/RegisterView.test.ts`
  - `code/frontend/src/views/platform/__tests__/UserManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

## 第二轮验证

- 已执行：
  - `npm run typecheck`
  - `npm run test:run -- src/router/__tests__/guards.test.ts src/router/__tests__/sharedRoutes.test.ts src/views/auth/__tests__/LoginView.test.ts src/views/auth/__tests__/RegisterView.test.ts src/views/platform/__tests__/ChallengeTopologyStudio.test.ts src/views/contests/__tests__/ContestDetail.test.ts`
  - `npm run test:run -- src/views/platform/__tests__/UserManage.test.ts -t "删除用户失败时不应抛到全局错误页"`


## 第三轮修复进展

- 已完成：
  - `P2-1` 主题 token 收口继续推进，后台二级导航、工作区数据表和通知下拉的通知类型配色已去掉局部硬编码颜色，统一回到共享主题变量与 `color-mix`。
  - 通知下拉现有结构测试里一处对模板单行排版的脆弱断言已改成检查关键结构，避免后续仅因格式变化造成假失败。
- 本轮涉及文件：
  - `code/frontend/src/components/layout/BackofficeSubNav.vue`
  - `code/frontend/src/components/common/WorkspaceDataTable.vue`
  - `code/frontend/src/composables/useNotificationDropdown.ts`
  - `code/frontend/src/components/layout/__tests__/BackofficeSubNav.test.ts`
  - `code/frontend/src/components/common/__tests__/WorkspaceDataTable.test.ts`
  - `code/frontend/src/composables/__tests__/useNotificationDropdown.test.ts`
  - `code/frontend/src/components/layout/__tests__/NotificationDropdown.test.ts`

## 第三轮验证

- 已执行：
  - `npm run test:run -- src/components/layout/__tests__/BackofficeSubNav.test.ts src/components/common/__tests__/WorkspaceDataTable.test.ts src/composables/__tests__/useNotificationDropdown.test.ts`
  - `npm run test:run -- src/components/layout/__tests__/NotificationDropdown.test.ts`
  - `npm run typecheck`

## 第四轮修复进展

- 已完成：
  - `P2-1` 教师端 AWD 复盘详情页已切回共享深色 surface 体系，移除亮色硬编码、`text-slate-*` 工具类和局部状态色，页面标题与赛事标题语义已拆开。
  - 教师端共享 surface 回归里暴露的旧漂移一并收口：班级管理页按钮体系改回 `teacher-btn`，列表状态列改回共享目录标签结构；`ClassStudentsPage` 补齐教师端 surface token 形态；`TeacherAWDReviewIndex` 根壳补齐 `workspace-shell`；共享 `teacher-surface-hero` 改为直角版本。
- 本轮涉及文件：
  - `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
  - `code/frontend/src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
  - `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
  - `code/frontend/src/assets/styles/teacher-surface.css`

## 第四轮验证

- 已执行：
  - `npm run test:run -- src/views/teacher/__tests__/TeacherAWDReviewDetail.test.ts`
  - `npm run test:run -- src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherSurface.test.ts`
  - `npm run typecheck`

## 第五轮修复进展

- 已完成：
  - `P2-5` 路由装配层开始减重，`router/index.ts` 已从单文件明细声明改为只负责装配认证路由、应用壳路由、错误路由和杂项路由。
  - 学生、教师、平台三组子路由已拆到独立模块，原有 canonical path、redirect 和 meta 字段保持不变，后续继续做大页面收口时可以按分组单独演进。
  - 新增源码级回归测试，防止后续再把三套路由明细重新堆回 `router/index.ts`。
- 本轮涉及文件：
  - `code/frontend/src/router/index.ts`
  - `code/frontend/src/router/routes/authRoutes.ts`
  - `code/frontend/src/router/routes/appShellRoute.ts`
  - `code/frontend/src/router/routes/studentRoutes.ts`
  - `code/frontend/src/router/routes/teacherRoutes.ts`
  - `code/frontend/src/router/routes/platformRoutes.ts`
  - `code/frontend/src/router/routes/errorRoutes.ts`
  - `code/frontend/src/router/routes/utilityRoutes.ts`
  - `code/frontend/src/router/routes/route-helpers.ts`
  - `code/frontend/src/router/__tests__/routeModuleExtraction.test.ts`

## 第五轮验证

- 已执行：
  - `npm run test:run -- src/router/__tests__/routeModuleExtraction.test.ts src/router/__tests__/sharedRoutes.test.ts src/router/__tests__/guards.test.ts src/router/__tests__/errorRoutes.test.ts`
  - `npm run typecheck`

## 第六轮修复进展

- 已完成：
  - `P2-5` `ContestDetail.vue` 继续减重，公告内容区和队伍内容区已抽到独立 `components/contests` 组件，路由页保留状态装配、tab 切换和核心交互。
  - 队伍按钮原语和公告列表结构的 source 护栏已跟随抽取边界更新，避免继续把实现细节绑死在单一路由页源码上。
- 本轮涉及文件：
  - `code/frontend/src/views/contests/ContestDetail.vue`
  - `code/frontend/src/components/contests/ContestAnnouncementsPanel.vue`
  - `code/frontend/src/components/contests/ContestTeamPanel.vue`
  - `code/frontend/src/views/contests/__tests__/contestDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/contests/__tests__/contestStudentActionPrimitives.test.ts`

## 第六轮验证

- 已执行：
  - `npm run test:run -- src/views/contests/__tests__/contestDetailPanelExtraction.test.ts src/views/contests/__tests__/contestStudentActionPrimitives.test.ts src/views/contests/__tests__/ContestDetail.test.ts`
  - `npm run typecheck`

## 第七轮修复进展

- 已完成：
  - `P2-5` `ImageManage.vue` 继续减重，镜像详情弹窗和创建弹窗已抽到独立 `components/platform/images` 组件，页面本体只保留筛选、排序、轮询与增删流程。
  - `ImageManage` 的源码护栏已改到新的真实承载组件上，同时补回目录头和工具栏的共享样式约束，避免测试继续被模板排版和旧边界绑死。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/ImageManage.vue`
  - `code/frontend/src/components/platform/images/ImageDetailModal.vue`
  - `code/frontend/src/components/platform/images/ImageCreateModal.vue`
  - `code/frontend/src/views/platform/__tests__/imageManageModalExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ImageManage.test.ts`

## 第七轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/imageManageModalExtraction.test.ts src/views/platform/__tests__/ImageManage.test.ts`
  - `npm run typecheck`

## 第八轮修复进展

- 已完成：
  - `P2-5` `ChallengeDetail.vue` 继续减重，提交记录面板和个人题解面板已抽到独立 `components/challenge` 组件，路由页继续只保留数据装配、分页状态、tab 切换和核心交互。
  - 与源码结构相关的护栏已改成面向组合源码，避免既有测试继续把 `workspace overline` 语义和面板实现细节硬绑定在单一路由文件上。
  - `ChallengeDetail.vue` 本体行数已从 `1686` 行降到 `1441` 行，后续再拆 `solution` 面板时切面会更稳定。
- 本轮涉及文件：
  - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - `code/frontend/src/components/challenge/ChallengeSubmissionRecordsPanel.vue`
  - `code/frontend/src/components/challenge/ChallengeWriteupPanel.vue`
  - `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

## 第八轮验证

- 已执行：
  - `npm run test:run -- src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts`
  - `npm run typecheck`

## 第九轮修复进展

- 已完成：
  - `P2-5` `ChallengeDetail.vue` 继续减重，题解区已抽到独立 `components/challenge/ChallengeSolutionsPanel.vue`，父页面继续保留题解加载、tab 状态、选中项和键盘导航 ownership。
  - 题解区原有的次级 tabs、空态、锁定态、富文本渲染和焦点可访问性语义都保持在新组件内，原有行为测试无需按交互重写。
  - `ChallengeDetail.vue` 本体行数已从 `1441` 行继续降到 `1213` 行，`question` 面板和右侧工具区之外的主要内容区已基本拆完。
- 本轮涉及文件：
  - `code/frontend/src/composables/useChallengeDetailPresentation.ts`
  - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - `code/frontend/src/components/challenge/ChallengeSolutionsPanel.vue`
  - `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

## 第九轮验证

- 已执行：
  - `npm run test:run -- src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts`
  - `npm run typecheck`

## 第十轮修复进展

- 已完成：
  - `P2-5` `ChallengeDetail.vue` 继续减重，题目面板已抽到独立 `components/challenge/ChallengeQuestionPanel.vue`，题面、附件入口、提示展开和分值侧栏都已从路由页移出。
  - `ChallengeDetail` 的源码护栏已同步更新为四块内容面板都通过子组件装配，`Question / Statement / Hints` 的共享 `workspace overline` 语义也已切到组合源码检查。
  - `ChallengeDetail.vue` 本体行数已从 `1213` 行继续降到 `913` 行，目前路由页主要只剩页级装配和右侧工具区。
- 本轮涉及文件：
  - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - `code/frontend/src/components/challenge/ChallengeQuestionPanel.vue`
  - `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

## 第十轮验证

- 已执行：
  - `npm run test:run -- src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts`
  - `npm run typecheck`

## 第十一轮修复进展

- 已完成：
  - `P2-5` `ChallengeDetail.vue` 继续减重，右侧工具区已抽到独立 `components/challenge/ChallengeActionAside.vue`，Flag 提交区和实例操作区都已从路由页移出。
  - `ChallengeDetail` 的源码护栏已同步更新为五块主要装配区都通过子组件接线，`Primary Action` 的 `workspace overline` 语义检查也已切到新侧栏组件。
  - `ChallengeDetail.vue` 本体行数已从 `913` 行继续降到 `732` 行，页面本体现在基本只剩页级状态、布局装配和样式变量层。
- 本轮涉及文件：
  - `code/frontend/src/views/challenges/ChallengeDetail.vue`
  - `code/frontend/src/components/challenge/ChallengeActionAside.vue`
  - `code/frontend/src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

## 第十一轮验证

- 已执行：
  - `npm run test:run -- src/views/challenges/__tests__/challengeDetailPanelExtraction.test.ts src/views/challenges/__tests__/ChallengeDetail.test.ts`
  - `npm run typecheck`

## 第十二轮修复进展

- 已完成：
  - `P2-5` `ContestDetail.vue` 继续减重，概览面板已抽到独立 `components/contests/ContestOverviewPanel.vue`，路由页不再直接承载 hero、统计卡、赛程信息和公告预览模板。
  - `ContestDetail` 的源码护栏已同步更新为概览、公告、队伍三区都通过子组件装配，`Contest / Rules / Schedule / Announcements` 的 `workspace overline` 语义改为面向组合源码检查。
  - `ContestDetail.vue` 本体行数已从 `1231` 行降到 `948` 行，后续继续拆普通题目工作区时边界会更清晰。
- 本轮涉及文件：
  - `code/frontend/src/views/contests/ContestDetail.vue`
  - `code/frontend/src/components/contests/ContestOverviewPanel.vue`
  - `code/frontend/src/views/contests/__tests__/contestDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

## 第十二轮验证

- 已执行：
  - `npm run test:run -- src/views/contests/__tests__/contestDetailPanelExtraction.test.ts src/views/contests/__tests__/ContestDetail.test.ts`
  - `npm run typecheck`

## 第十三轮修复进展

- 已完成：
  - `P2-5` `ContestDetail.vue` 继续减重，普通竞赛题目工作区已抽到独立 `components/contests/ContestChallengeWorkspacePanel.vue`，路由页不再直接承载题目列表、题目聚焦区和 Flag 提交表单模板。
  - 这轮拆分保持了状态 ownership 不变：选题、Flag 输入、提交动作和提交结果仍由 `useContestDetailPage` 与路由页持有，新组件只通过 props 和 emits 装配普通竞赛交互。
  - `ContestDetail` 的源码护栏已同步更新为概览、题目工作区、公告和队伍四块区域都通过子组件接线，`Selected / Primary Action` 的 `workspace overline` 语义也已切到组合源码检查。
  - `ContestDetail.vue` 本体行数已从 `948` 行继续降到 `671` 行，普通竞赛内容区的主要模板负担已经移出，后续若继续收口可优先看对话框和少量页级样式。
- 本轮涉及文件：
  - `code/frontend/src/views/contests/ContestDetail.vue`
  - `code/frontend/src/components/contests/ContestChallengeWorkspacePanel.vue`
  - `code/frontend/src/views/contests/__tests__/contestDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

## 第十三轮验证

- 已执行：
  - `npm run test:run -- src/views/contests/__tests__/contestDetailPanelExtraction.test.ts src/views/contests/__tests__/ContestDetail.test.ts`
  - `npm run typecheck`

## 第十四轮修复进展

- 已完成：
  - `P2-5` `ChallengeList.vue` 继续减重，题目目录工作区已抽到独立 `components/challenge/ChallengeDirectoryPanel.vue`，路由页不再直接承载筛选区、空错态和题目目录表格模板。
  - 这轮拆分保持了状态 ownership 不变：路由 query 同步、分页、搜索刷新和跳转动作仍由 `ChallengeList.vue` 持有，新组件只通过 props 和 emits 装配目录交互。
  - `ChallengeList` 的源码护栏已同步更新为题目目录工作区通过独立组件接线，原有目录式结构、共享 UI 原语和标题溢出约束都改为面向组合源码检查。
  - `ChallengeList.vue` 本体行数已从 `946` 行降到 `318` 行，页面本体现在基本只剩路由同步、分页装配、概况区和导航动作。
- 本轮涉及文件：
  - `code/frontend/src/views/challenges/ChallengeList.vue`
  - `code/frontend/src/components/challenge/ChallengeDirectoryPanel.vue`
  - `code/frontend/src/views/challenges/__tests__/ChallengeList.test.ts`
  - `code/frontend/src/views/challenges/__tests__/challengeListPanelExtraction.test.ts`

## 第十四轮验证

- 已执行：
  - `npm run test:run -- src/views/challenges/__tests__/challengeListPanelExtraction.test.ts src/views/challenges/__tests__/ChallengeList.test.ts`
  - `npm run typecheck`

## 第十五轮修复进展

- 已完成：
  - `P2-5` `platform/ChallengeDetail.vue` 继续减重，题目详情 tab 已抽到独立 `components/platform/challenge/AdminChallengeProfilePanel.vue`，路由页不再直接承载详情头、基础信息、提示区和判题模式配置模板。
  - 这轮拆分保持了状态 ownership 不变：路由页继续持有加载、下载、Flag 草稿、保存动作与 query tabs，新组件只通过 props 和 emits 承载详情 tab 展示与输入。
  - `Admin ChallengeDetail` 的源码护栏已补到 detail tab 抽离层，现有行为测试仍覆盖拓扑入口、query tab 切换、共享实例提示、附件下载和 Flag 保存边界。
  - `ChallengeDetail.vue` 本体行数已从 `969` 行降到 `402` 行，路由页现在主要只剩顶层导航、tab 装配和远端交互 owner。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/ChallengeDetail.vue`
  - `code/frontend/src/components/platform/challenge/AdminChallengeProfilePanel.vue`
  - `code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`

## 第十五轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts`
  - `npm run typecheck`

## 第十六轮修复进展

- 已完成：
  - `P2-5` `ImageManage.vue` 继续减重，镜像目录工作区已抽到独立 `components/platform/images/ImageDirectoryPanel.vue`，路由页不再直接承载目录头、筛选排序、空错态、表格与分页模板。
  - 这轮拆分保持了状态 ownership 不变：轮询、创建、删除、详情弹窗、状态整理和排序结果仍由 `ImageManage.vue` 持有，新组件只通过 props 和 emits 承载目录展示与交互。
  - `ImageManage` 的源码护栏已同步更新为目录工作区通过独立平台组件接线，目录标题、共享工具栏、共享表格和长文本省略样式都改为面向组合源码检查。
  - `ImageManage.vue` 本体行数已从 `873` 行降到 `545` 行，页面本体现在主要只剩头部状态条、轮询 owner、创建/删除流程和两个弹窗装配。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/ImageManage.vue`
  - `code/frontend/src/components/platform/images/ImageDirectoryPanel.vue`
  - `code/frontend/src/views/platform/__tests__/ImageManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/imageManageWorkspaceExtraction.test.ts`

## 第十六轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/imageManageWorkspaceExtraction.test.ts src/views/platform/__tests__/imageManageModalExtraction.test.ts src/views/platform/__tests__/ImageManage.test.ts`
  - `npm run typecheck`

## 第十七轮修复进展

- 已完成：
  - `P2-5` `AuditLog.vue` 继续减重，操作流水工作区已抽到独立 `components/platform/audit/AuditLogDirectoryPanel.vue`，路由页不再直接承载目录头、筛选排序、空错态、表格与分页模板。
  - 这轮拆分保持了状态 ownership 不变：路由 query 同步、请求取消与 stale request 防护、筛选自动应用、执行人详情弹窗和摘要卡片仍由 `AuditLog.vue` 持有，新组件只通过 props 和 emits 承载目录展示与交互。
  - `AuditLog` 的源码护栏已同步更新为父子组合源码检查，目录工作区、共享工具栏、共享表格和深色 surface 边框变量都改为面向组合源码断言。
  - `AuditLog.vue` 本体行数已从 `831` 行降到 `542` 行，页面本体现在主要只剩摘要区、路由同步、远端加载 owner 和执行人详情弹窗。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/AuditLog.vue`
  - `code/frontend/src/components/platform/audit/AuditLogDirectoryPanel.vue`
  - `code/frontend/src/views/platform/__tests__/AuditLog.test.ts`
  - `code/frontend/src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第十七轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts src/views/platform/__tests__/AuditLog.test.ts`
  - `npm run typecheck`

## 第十八轮修复进展

- 已完成：
  - `P2-5` `ChallengeManage.vue` 继续减重，题目目录工作区已抽到独立 `components/platform/challenge/ChallengeManageDirectoryPanel.vue`，路由页不再直接承载筛选区、空错态、表格动作菜单与分页模板。
  - 这轮拆分保持了状态 ownership 不变：题目列表加载、排序状态、发布检查、删除确认、路由跳转和 action menu open state 仍由 `ChallengeManage.vue` 持有，新组件只通过 props 和 emits 承载目录展示与交互。
  - 同页顺手把管理端 shell 和摘要区收回统一工作区原语，`ChallengeManage` 的源码护栏也同步更新为父子组合源码检查。
  - `ChallengeManage.vue` 本体行数已从 `709` 行降到 `339` 行，页面本体现在主要只剩 workspace shell、摘要区、目录状态装配和动作 owner。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/ChallengeManage.vue`
  - `code/frontend/src/components/platform/challenge/ChallengeManageDirectoryPanel.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第十八轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts src/views/platform/__tests__/ChallengeManage.test.ts`
  - `npm run typecheck`

## 第十九轮修复进展

- 已完成：
  - `P2-5` `AWDReviewIndex.vue` 继续减重，复盘赛事目录工作区已抽到独立 `components/platform/awd-review/AwdReviewDirectoryPanel.vue`，路由页不再直接承载目录头、筛选区、空错态与表格模板。
  - 这轮拆分保持了状态 ownership 不变：赛事列表加载、自动筛选、路由跳转、刷新和摘要统计仍由 `AWDReviewIndex.vue` 持有，新组件只通过 props 和 emits 承载目录展示与交互。
  - 同页摘要区已切回统一 `admin-summary-grid + progress-strip + metric-panel-*` 原语，`AWDReviewIndex` 的源码护栏也同步更新为父子组合源码检查。
  - `AWDReviewIndex.vue` 本体行数已从 `535` 行降到 `217` 行，页面本体现在主要只剩 workspace shell、摘要区、目录状态装配和动作 owner。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/components/platform/awd-review/AwdReviewDirectoryPanel.vue`
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/platform/__tests__/awdReviewDirectoryExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第十九轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/awdReviewDirectoryExtraction.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `npm run typecheck`
- 额外检查：
  - `npm run test:run -- src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts` 仍存在仓内既有失败，当前失败面主要落在 `UserGovernancePage`、`ContestOrchestrationPage`、共享 `journal-notes.css` 等非本轮 AWD 变更文件，本次提交未扩大处理范围。

## 第二十轮修复进展

- 已完成：
  - 平台 AWD 组件簇的 review 漂移已收口：
    - `AWDReadinessSummary` 补回 readiness 决策文案、零题目空态和可开赛 / 强制开赛 / 不可开赛语义
    - `ContestChallengeOrchestrationPanel` 修正 AWD create/edit 的 payload 边界，本地刷新改为合并题目关系与 service 数据，恢复筛选 / 菜单 / 配置入口契约
    - `ContestChallengeEditorDialog` 对齐 `AdminSurfaceModal`，避免继续挂在旧 drawer 壳层
    - `AWDOperationsPanel`、`AWDRoundHeaderPanel`、`AWDRoundSelectionPanel`、`AWDScoreboardSummaryPanel`、`AWDTrafficPanel`、`AWDServiceStatusPanel`、`AWDAttackLogPanel` 的测试依赖语义与页面文案已回到一致状态
  - 教师端 `workspace / tabs / surface` 漂移已收口：
    - `TeacherAWDReviewIndex.vue`、`TeacherAWDReviewDetail.vue` 去掉手写 `rounded-[30px]` 根壳，回到共享 `workspace-shell + teacher-management-shell`
    - `ClassManagementPage.vue`、`ClassStudentsPage.vue`、`StudentManagementPage.vue`、`TeacherDashboardPage.vue` 补齐 soft-border token / 共享壳层契约
    - 班级 workspace tab 的 review 基线改到真实 owner `ClassStudentsPage.vue`，不再把目录页误判为 tab 同步实现页
  - shared pagination review 基线已对齐到抽层后的真实 owner：
    - `ChallengeList.vue` 的分页检查转移到 `ChallengeDirectoryPanel.vue`
    - 避免继续盯住已完成抽层的 view 层 raw 源码，减少假失败
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDReadinessSummary.vue`
  - `code/frontend/src/components/platform/contest/AWDOperationsPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDRoundHeaderPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDRoundSelectionPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDScoreboardSummaryPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDTrafficPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDServiceStatusPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDAttackLogPanel.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - `code/frontend/src/views/teacher/TeacherAWDReviewIndex.vue`
  - `code/frontend/src/views/teacher/TeacherAWDReviewDetail.vue`
  - `code/frontend/src/components/teacher/class-management/ClassManagementPage.vue`
  - `code/frontend/src/components/teacher/class-management/ClassStudentsPage.vue`
  - `code/frontend/src/components/teacher/dashboard/TeacherDashboardPage.vue`
  - `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`
  - `code/frontend/src/views/__tests__/sharedPaginationControls.test.ts`
  - `code/frontend/src/views/teacher/__tests__/classManagementTabsAdoption.test.ts`

## 第二十轮验证

- 已执行：
  - `npm run test:run -- src/components/platform/__tests__/AWDChallengeConfigDialog.test.ts src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/components/platform/__tests__/AWDOperationsPanel.test.ts src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/components/platform/__tests__/AWDReadinessSummary.test.ts src/views/teacher/__tests__/teacherBaseSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherDarkSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherDetailSurfaceAlignment.test.ts src/views/teacher/__tests__/teacherEyebrowSharedStyles.test.ts src/views/teacher/__tests__/teacherPanelShellAdoption.test.ts src/views/teacher/__tests__/teacherRootShellCleanup.test.ts src/views/teacher/__tests__/teacherSharedDirectoryStyles.test.ts src/views/teacher/__tests__/teacherSurface.test.ts src/views/teacher/__tests__/teacherWorkspaceSubpanelAdoption.test.ts src/views/teacher/__tests__/classManagementTabsAdoption.test.ts src/router/__tests__/sharedRoutes.test.ts src/views/__tests__/sharedPaginationControls.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/surfaceBackground.test.ts src/components/layout/__tests__/NotificationDropdown.test.ts`
  - 结果：`21` 个测试文件、`104` 条测试通过
  - `npm run typecheck`

## 当前剩余未收口项

- `P2-1` 仍未全量处理：
  - 这轮只进一步收口了 AWD / 教师端 / 通知 / shared pagination 周边的主题与基线漂移
  - 生产代码里仍可能存在其他绕过共享 token 的硬编码样式
- `P2-5` 仍未全量处理：
  - 本轮修的是 AWD 和教师端 review 漂移，不是对剩余超重页面做完系统性减重
  - 仍需结合管理端剩余页面和共享 surface 基线继续推进

## 第二十一轮修复进展

- 已完成：
  - `P1-4` 登录态已经切到服务端 session，前端不再持有认证 token。
  - 登录恢复改为依赖服务端 `HttpOnly session cookie`：
    - 启动阶段直接请求 `/auth/profile`
    - 受保护路由在内存里没有用户信息时，会先等待同一条 restore promise，再决定是否跳去登录页
    - `/login`、`/register` 在已有有效 session cookie 时，也会先恢复会话再跳转回角色工作台
  - 旧的 `ctf_access_token / ctf_refresh_token` 只保留清理逻辑，不再作为运行时持久化来源。
- 本轮涉及文件：
  - `code/frontend/src/api/auth.ts`
  - `code/frontend/src/stores/auth.ts`
  - `code/frontend/src/router/guards.ts`
  - `code/frontend/src/main.ts`
  - `code/frontend/src/stores/__tests__/auth.test.ts`
  - `code/frontend/src/router/__tests__/guards.test.ts`

## 第二十一轮验证

- 已执行：
  - `npm run test:run -- src/stores/__tests__/auth.test.ts src/router/__tests__/guards.test.ts`
  - 结果：`2` 个测试文件、`12` 条测试通过
  - `npm run typecheck`

## 第二十二轮修复进展

- 已完成：
  - `P2-1` 平台 AWD 运行态剩余的亮色硬编码继续收口：
    - `AWDAttackLogPanel` 的工具栏、筛选器、时间线、结果徽标和空态已切回共享主题 token，不再直接写死浅色背景、边框和成功态色值。
    - `AWDRuntimePendingState` 的根壳、卡片、图标容器和说明文案已切回共享 surface / text / border token，等待态不再依赖白底卡片和浅色边框。
  - `P2-1` 导入页的局部主题回退值继续收口：
    - `ChallengeImportManage` 的品牌色已改为从 `journal-accent` 派生，不再在页面里手写蓝色品牌常量。
    - 页面操作按钮已切回共享 `ui-btn` 原语，不再维护一套本地亮色按钮壳层。
    - 页尾 `light / dark` 专用硬编码覆盖已删除，改由现有 `journal-* / color-*` token 直接驱动 surface、边框和文字层级。
    - `ChallengePackageImportReview` 的确认 / 重置操作也已切回共享 `ui-btn` 原语，移除局部 `#fff` 按钮文字色和旧按钮壳层。
  - 管理端题目目录里一组已经失效的 presentation helper 已清理：
    - `useChallengeManagePresentation` 里未再被页面消费的分类 / 难度 / 状态 / 发布检查颜色辅助函数与旧导入页残留 helper 已移除，避免继续保留无主的硬编码主题色。
  - 补上源码级主题护栏：
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `AWDAttackLogPanel`、`AWDRuntimePendingState`、`ChallengeImportManage`、`useChallengeManagePresentation` 的硬编码色值回归检查，避免后续再把这批亮色值带回 AWD 面板、导入页和管理端题目目录。
    - `ChallengePackageImportReview.test.ts` 已补按钮原语护栏，防止导入预览页重新回退到局部主按钮实现。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDAttackLogPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDRuntimePendingState.vue`
  - `code/frontend/src/views/platform/ChallengeImportManage.vue`
  - `code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`
  - `code/frontend/src/composables/useChallengeManagePresentation.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `code/frontend/src/components/platform/__tests__/ChallengePackageImportReview.test.ts`

## 第二十二轮验证

- 已执行：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/components/platform/__tests__/AWDRoundInspector.test.ts`
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`
  - `npm run test:run -- src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `npm run test:run -- src/components/platform/__tests__/ChallengePackageImportReview.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 结果：`11` 个测试文件、`58` 条测试通过
  - `npm run build`
  - `npm run typecheck`

## 第二十三轮修复进展

- 已完成：
  - `P2-1` 平台 AWD 题目配置面板继续收口共享主题原语：
    - `AWDChallengeConfigPanel` 的就绪验证状态徽标已切回共享 `ui-badge`，不再局部写死 `success / danger / warning + white text`。
    - 面板内遗留的本地 `ops-btn` 主按钮样式已移除，避免继续保留一套未被消费的亮色按钮壳层。
  - `P2-1` 赛事工具残留硬编码继续清理：
    - `utils/contest.ts` 中无消费者的 `getStatusBadgeClass` 已删除，避免继续保留 `#06b6d4 / #f59e0b / #30363d` 这组失效色值。
  - 补上源码级护栏：
    - `contestUiPrimitiveAdoptionPhase4.test.ts` 已新增 `AWDChallengeConfigPanel` 对共享 badge 原语的断言。
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `AWDChallengeConfigPanel` 与 `utils/contest.ts` 的硬编码色值回归检查。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDChallengeConfigPanel.vue`
  - `code/frontend/src/utils/contest.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十三轮验证

- 已执行：
  - `npm run test:run -- src/components/platform/__tests__/AWDOperationsPanel.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 结果：`3` 个测试文件、`20` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第二十四轮修复进展

- 已完成：
  - `P2-1` AWD readiness 与赛事目录残留的亮色局部实现继续收口：
    - `AWDReadinessSummary` 的阻塞状态徽标已切回共享 `ui-badge`，不再保留 `success / danger / warning + white text` 的本地胶囊实现。
    - `AWDReadinessSummary` 的最近校验列已移除 `text-slate-500 / text-[11px]` 这类局部 Tailwind 任意值写法，改回 token 驱动的文本层级。
    - `PlatformContestTable` 的 AWD 运维台入口按钮已删除本地 `white` 文本覆盖，继续沿用共享 `ui-btn--primary` 默认前景色。
  - `P2-1` 赛事编排页的失效局部样式继续清理：
    - `ContestChallengeOrchestrationPanel` 中未被消费的 `ops-btn` 亮色按钮壳层已删除。
    - 快捷筛选计数胶囊已从 `rgba(0,0,0,0.05)` 改回共享 token 派生的 `color-mix`。
  - 主题护栏继续加固：
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `AWDReadinessSummary`、`PlatformContestTable`、`ContestChallengeOrchestrationPanel` 的硬编码回归检查。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDReadinessSummary.vue`
  - `code/frontend/src/components/platform/contest/PlatformContestTable.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十四轮验证

- 已执行：
  - `npm run test:run -- src/components/platform/__tests__/AWDReadinessSummary.test.ts src/components/platform/__tests__/PlatformContestTable.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 结果：`5` 个测试文件、`25` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第二十五轮修复进展

- 已完成：
  - `P2-1` 导航与通知共享壳层的 light theme 字面量继续收口：
    - `NotificationDropdown` 的 light theme surface / line / text 变量已改成共享 token 派生，不再直接写 `white / #f8fafc / #e2e8f0 / #0f172a`。
    - `TopNav` 的 backoffice shell light theme 变量已改成 `journal-* / color-*` token 派生，焦点描边不再依赖 `white` 混色。
    - `Sidebar` 的 backoffice shell light theme 变量已改成共享 token 派生，按钮和激活态的局部阴影也已从硬编码 `rgba(...)` 收口到 token 化表达。
  - 共享主题护栏继续加固：
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `NotificationDropdown`、`TopNav`、`Sidebar` 这轮字面量 light token 回归的检查。
- 本轮涉及文件：
  - `code/frontend/src/components/layout/NotificationDropdown.vue`
  - `code/frontend/src/components/layout/TopNav.vue`
  - `code/frontend/src/components/layout/Sidebar.vue`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十五轮验证

- 已执行：
  - `npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/components/layout/__tests__/TopNav.test.ts src/components/layout/__tests__/NotificationDropdown.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 结果：`4` 个测试文件、`35` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第二十六轮修复进展

- 已完成：
  - `P2-1` 共享导航壳层残留的低层级样式类继续收口：
    - `TopNav` 的 backoffice breadcrumb 已不再复用 `text-slate-*` 工具类，改为语义化 `topnav-breadcrumb*` token 类。
    - `NotificationDropdown` 的未读角标已移除 `text-white`，改回组件内 token 驱动的前景色。
    - `Sidebar` 的 academy/student 导航激活态残留 `rgba(...)` 阴影已改成 token 化表达。
  - `P2-1` 赛事编排页剩余的 Tailwind 任意值与旧色类继续清理：
    - `ContestChallengeOrchestrationPanel` 的分值列、AWD 分值摘要和预览摘要已改为语义类，不再保留 `text-slate-* / text-[10px]`。
  - 主题护栏继续加固：
    - `sharedThemeTokenAdoption.test.ts` 已新增对上述 breadcrumb / badge / 表格文案残留类名和阴影字面量的回归检查。
- 本轮涉及文件：
  - `code/frontend/src/components/layout/TopNav.vue`
  - `code/frontend/src/components/layout/NotificationDropdown.vue`
  - `code/frontend/src/components/layout/Sidebar.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十六轮验证

- 已执行：
  - `npm run test:run -- src/components/layout/__tests__/Sidebar.test.ts src/components/layout/__tests__/TopNav.test.ts src/components/layout/__tests__/NotificationDropdown.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 结果：`6` 个测试文件、`45` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第二十七轮修复进展

- 已完成：
  - `P2-1` AWD 运行面板簇剩余的表格 utility class 继续收口：
    - `AWDTrafficPanel` 的路径行、趋势空态、时间列、向量分隔符和空表提示已改为语义类，不再保留 `text-slate-* / text-[10px] / text-[11px]`。
    - `AWDServiceStatusPanel` 的本轮表现表已改为语义类，不再在模板里写 `text-slate-* / text-emerald-* / text-red-* / text-orange-*`。
    - `AWDScoreboardSummaryPanel` 的上下文 HUD、总分列、解题进度列和最后命中时间已改为语义类，不再保留 `text-slate-* / text-blue-* / text-orange-* / text-[11px]`。
  - 共享主题护栏继续加固：
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `AWDTrafficPanel`、`AWDServiceStatusPanel`、`AWDScoreboardSummaryPanel` 的回归检查。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDTrafficPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDServiceStatusPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDScoreboardSummaryPanel.vue`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十七轮验证

- 已执行：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/components/platform/__tests__/AWDRoundInspector.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
  - 结果：`4` 个测试文件、`19` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第二十八轮修复进展

- 已完成：
  - `P2-1` AWD 赛前/轮次页残留的本地按钮死样式继续收口：
    - `ContestAwdPreflightPanel` 已删除未再使用的 `.ops-btn / .ops-btn--primary` 本地实现，强制放行入口完全落回共享 `ui-btn` 原语。
    - `AWDRoundHeaderPanel` 已删除未再使用的 `.ops-btn / .ops-btn--neutral / .ops-btn--primary` 本地实现，轮次工具栏完全落回共享按钮体系。
  - `P2-1` 竞赛编辑页继续收口共享按钮与 surface token：
    - `ContestEdit` 的保存按钮已改为只覆盖 `ui-btn` 变量，不再本地硬写 `color: white;`。
    - `ContestEdit` 的工作台画布已移除 `background: var(--color-bg-surface, #ffffff)` fallback，统一回到共享 surface token。
  - `P2-1` 题目编排菜单交互的测试契约已重新对齐当前实现：
    - `ContestChallengeOrchestrationPanel` 的动作菜单 DOM id 已重新稳定到 `challenge_id` 维度，避免 service/link id 混入测试与交互定位。
    - `ContestEdit.test.ts` 中 AWD 题目池创建桩已补齐 `challenge_id` 契约，并同步更新运行段降级壳文案、AWD 分值摘要格式等现行断言。
  - 共享主题护栏继续加固：
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `ContestAwdPreflightPanel`、`AWDRoundHeaderPanel`、`ContestEdit` 的回归检查。
    - `contestUiPrimitiveAdoptionPhase4.test.ts` 已新增对 `ContestAwdPreflightPanel`、`AWDRoundHeaderPanel` 不再残留 `.ops-btn--primary` 的断言。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestAwdPreflightPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDRoundHeaderPanel.vue`
  - `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - `code/frontend/src/views/platform/ContestEdit.vue`
  - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十八轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ContestEdit.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase18.test.ts src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 结果：`8` 个测试文件、`72` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第二十九轮修复进展

- 已完成：
  - `P2-1` AWD 运行簇剩余的局部按钮原语继续收口：
    - `AWDRoundInspector` 的“导出复盘包”按钮已切回共享 `ui-btn ui-btn--secondary`，不再维护本地 `ops-btn`。
    - `AWDServiceStatusPanel` 的“导出报告”按钮已切回共享 `ui-btn ui-btn--secondary`，本地按钮壳层已删除。
    - `AWDAttackLogPanel` 的“导出审计日志”按钮已切回共享 `ui-btn ui-btn--secondary`，同时保留面板自身的 secondary token 覆写。
    - `AWDTrafficPanel` 中已无模板消费的 `.ops-btn` 死样式已删除。
    - `AWDRoundInspector` 的告警 pill 激活态前景色已从直接 `white` 改为共享 `var(--color-text-inverse)`。
  - `P2-1` 平台后台零散 utility / token 回退继续收口：
    - `InstanceManage` 的状态列已从 `bg-green-100 / bg-slate-100` utility class 改为语义化 `instance-status-pill`。
    - `PlatformContestFormPanel` 的时间轴分隔符已移除 `text-slate-300`，改回 token 驱动的 `timeline-divider`。
    - `AWDServiceTemplateLibraryPage` 的页签计数 badge 前景色已从 `white` 改为共享 `var(--color-text-inverse)`。
  - 测试护栏继续加固：
    - `contestUiPrimitiveAdoptionPhase4.test.ts` 已新增对 `AWDRoundInspector`、`AWDServiceStatusPanel`、`AWDAttackLogPanel` 不再回退到 `ops-btn` 的断言。
    - `contestUiPrimitiveAdoption.test.ts` 已新增 `PlatformContestFormPanel` 不再保留 `text-slate-300` 的断言。
    - `InstanceManage.test.ts` 已新增实例状态 pill 不再回退到 `bg-green-100 / bg-slate-100` 的断言。
    - `sharedThemeTokenAdoption.test.ts` 已新增对 `AWDRoundInspector`、`AWDServiceStatusPanel`、`AWDAttackLogPanel`、`AWDTrafficPanel`、`AWDServiceTemplateLibraryPage` 的回归检查。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDRoundInspector.vue`
  - `code/frontend/src/components/platform/contest/AWDServiceStatusPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDAttackLogPanel.vue`
  - `code/frontend/src/components/platform/contest/AWDTrafficPanel.vue`
  - `code/frontend/src/views/platform/InstanceManage.vue`
  - `code/frontend/src/components/platform/contest/PlatformContestFormPanel.vue`
  - `code/frontend/src/components/platform/awd-service/AWDServiceTemplateLibraryPage.vue`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第二十九轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase4.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/components/platform/__tests__/AWDRoundInspector.test.ts src/components/platform/__tests__/AWDRoundInspectorExtraction.test.ts`
  - 结果：`4` 个测试文件、`19` 条测试通过
  - `npm run test:run -- src/views/platform/__tests__/InstanceManage.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - 结果：`4` 个测试文件、`30` 条测试通过
  - `npm run typecheck`
  - `npm run build`
- 备注：
  - `build` 仍会打印仓内既有 `[esbuild css minify] Expected identifier but found "." [css-syntax-error]` 警告，本轮未扩大到定位该历史问题。

## 第三十轮修复进展

- 已完成：
  - `P2-1` AWD 就绪摘要中系统级阻塞图标残留的 `text-red-500` 已收口为组件内语义样式，不再依赖 Tailwind 状态色硬编码。
  - `sharedThemeTokenAdoption.test.ts` 已补对 `AWDReadinessSummary` 的 `text-red-500` 回退断言，继续把主题清理结果固化为源码门禁。
  - `P2-1` `ChallengeTopologyStudioPage` 中主按钮变量残留的 `#fff` 回退值已改回共享 `var(--color-text-inverse)`。
  - `sharedThemeTokenAdoption.test.ts` 已补对 `ChallengeTopologyStudioPage` 这两处 `#fff` 主按钮变量回退的源码断言。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/AWDReadinessSummary.vue`
  - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第三十轮验证

- 已执行：
  - `npm run test:run -- src/components/platform/__tests__/AWDReadinessSummary.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第三十一轮修复进展

- 已完成：
  - `P2-1` `ThemePreview.vue` 已从亮色静态样张收口为受控的暗色 preview surface，根壳、侧栏、头部、统计卡、工具栏、表格和滚动条都改回语义类与共享 token 驱动。
  - 页面内残留的 `bg-white`、`text-slate-*`、`border-slate-*`、`bg-indigo-*`、`#e2e8f0 / #cbd5e0` 等硬编码颜色，以及 `text-[10px]`、`rounded-[20px]`、`rounded-[24px]`、`rounded-[18px]`、`max-w-[1400px]` 等任意值工具类已清理。
  - `sharedThemeTokenAdoption.test.ts` 已新增 `ThemePreview` 的源码护栏，防止实验页再次回退到亮色工具类和局部魔法值。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/ThemePreview.vue`
  - `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`

## 第三十一轮验证

- 已执行：
  - `npm run test:run -- src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - `npm run typecheck`

## 第三十二轮修复进展

- 已完成：
  - `P2-5` `AuditLog.vue` 中剩余内联的“执行人详情”弹窗已抽到独立组件 `AuditActorDetailModal.vue`，页面本体继续只保留路由 query、筛选、分页和数据装配。
  - `auditLogWorkspaceExtraction.test.ts` 已补对新弹窗组件边界的源码护栏，防止后续再把 `AdminSurfaceModal` 和详情结构塞回路由页。
  - `AuditLog.test.ts` 已改为同时校验 `AuditLogDirectoryPanel` 与 `AuditActorDetailModal` 的边界职责，运行行为保持不变。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/audit/AuditActorDetailModal.vue`
  - `code/frontend/src/views/platform/AuditLog.vue`
  - `code/frontend/src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/AuditLog.test.ts`

## 第三十二轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/AuditLog.test.ts src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`
  - `npm run typecheck`

## 第三十三轮修复进展

- 已完成：
  - `P2-5` `AuditLog.vue` 的 hero 区和审计摘要卡已继续抽到独立组件 `AuditLogHeroPanel.vue`，路由页进一步收敛到查询同步、加载、排序、分页和目录/弹窗装配。
  - `auditLogWorkspaceExtraction.test.ts` 已补对 `AuditLogHeroPanel` 的源码边界断言，避免后续再把头部说明和进度卡堆回路由页。
  - `AuditLog.test.ts` 已切换为检查 `AuditLogHeroPanel` 的共享壳层和摘要卡契约，现有运行行为保持不变。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/audit/AuditLogHeroPanel.vue`
  - `code/frontend/src/views/platform/AuditLog.vue`
  - `code/frontend/src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/AuditLog.test.ts`

## 第三十三轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/AuditLog.test.ts src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`
  - `npm run typecheck`

## 第三十四轮修复进展

- 已完成：
  - `P2-5` `ImageManage.vue` 的头部说明、操作按钮和状态条已抽到独立组件 `ImageManageHeroPanel.vue`，路由页继续收敛到轮询、筛选、分页和弹窗状态装配。
  - `imageManageWorkspaceExtraction.test.ts` 已补对 `ImageManageHeroPanel` 的源码边界断言，避免后续再把头部摘要塞回路由页。
  - `ImageManage.test.ts` 已改为检查 `ImageManageHeroPanel` 的头部状态条、按钮原语和分隔线契约，现有运行行为保持不变。
  - 本轮中途暴露的 `statusSummary` `tone` 类型放宽问题已改成显式联合类型 `ImageStatusSummaryItem`，`typecheck` 基线已恢复。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/images/ImageManageHeroPanel.vue`
  - `code/frontend/src/views/platform/ImageManage.vue`
  - `code/frontend/src/views/platform/__tests__/imageManageWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ImageManage.test.ts`

## 第三十四轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ImageManage.test.ts src/views/platform/__tests__/imageManageWorkspaceExtraction.test.ts`
  - `npm run typecheck`

## 第三十五轮修复进展

- 已完成：
  - `P2-5` `ChallengeManage.vue` 的头部说明、导入入口和四张摘要卡已抽到独立组件 `ChallengeManageHeroPanel.vue`，路由页继续收敛到筛选、排序、分页和目录动作装配。
  - `challengeManageDirectoryExtraction.test.ts` 已补对 `ChallengeManageHeroPanel` 的源码边界断言，避免后续再把头部摘要堆回路由页。
  - `ChallengeManage.test.ts` 已改为检查 `ChallengeManageHeroPanel` 的 workspace 头部和摘要卡契约，现有交互保持不变。
  - 本轮中途暴露的 `Calendar` 导入缺失已补回，`ChallengeManage` 挂载和 `typecheck` 基线已恢复。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/ChallengeManageHeroPanel.vue`
  - `code/frontend/src/views/platform/ChallengeManage.vue`
  - `code/frontend/src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`

## 第三十五轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`
  - `npm run typecheck`

## 第三十六轮修复进展

- 已完成：
  - `P2-5` `platform/ChallengeDetail.vue` 的顶栏壳层已抽到独立组件 `AdminChallengeTopbarPanel.vue`，路由页继续收敛到 query tabs、详情/题解装配与远端交互 owner。
  - `challengeDetailPanelExtraction.test.ts` 已补对 `AdminChallengeTopbarPanel` 的源码边界断言，避免后续再把 workspace 顶栏和导航按钮塞回路由页。
  - `ChallengeDetail.test.ts` 已改为检查新顶栏组件的按钮原语契约，`platformManagementSurfaceAlignment.test.ts` 也已同步切到组合源码检查，现有行为保持不变。
  - `ChallengeDetail.vue` 本体行数已从 `402` 行进一步降到 `372` 行，页面本体现在主要只剩顶层导航装配、tab 切换与详情远端动作。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/AdminChallengeTopbarPanel.vue`
  - `code/frontend/src/views/platform/ChallengeDetail.vue`
  - `code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第三十六轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeDetail.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第三十七轮修复进展

- 已完成：
  - `P2-5` `CheatDetection.vue` 的 hero 顶栏和摘要卡已分别抽到独立组件 `CheatDetectionHeroPanel.vue`、`CheatDetectionSummaryPanel.vue`，路由页继续只持有风控数据加载、审计跳转、空错态与目录工作区装配。
  - 新增并补强 `cheatDetectionPanelExtraction.test.ts`，对 `CheatDetectionHeroPanel` 和 `CheatDetectionSummaryPanel` 都加上源码边界断言，避免后续再把工作区头部说明、刷新按钮和 KPI 模板塞回路由页。
  - `cheatDetectionSurfaceAlignment.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到组合源码检查，抽取后深色 surface 与目录护栏保持不变。
  - `CheatDetection.vue` 本体行数已从 `572` 行降到 `409` 行，页面本体现在主要只剩风险目录、错误态与跳转 owner。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/cheat/CheatDetectionHeroPanel.vue`
  - `code/frontend/src/components/platform/cheat/CheatDetectionSummaryPanel.vue`
  - `code/frontend/src/views/platform/CheatDetection.vue`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`

## 第三十七轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/CheatDetection.test.ts src/views/platform/__tests__/cheatDetectionPanelExtraction.test.ts src/views/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第三十八轮修复进展

- 已完成：
  - `P2-5` `InstanceManage.vue` 的头部说明、返回/刷新动作和实例摘要卡已抽到独立组件 `InstanceManageHeroPanel.vue`，路由页继续只持有 teacher 实例列表加载、销毁流程、分页与状态列装配。
  - 新增 `instanceManagePanelExtraction.test.ts`，补对 `InstanceManageHeroPanel` 的源码边界断言，避免后续再把 hero 与摘要卡堆回路由页。
  - `InstanceManage.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到组合源码检查，现有行为与深色管理页护栏保持不变。
  - `InstanceManage.vue` 本体行数已从 `393` 行降到 `288` 行，页面本体现在主要只剩实例目录、销毁动作和分页 owner。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/instance/InstanceManageHeroPanel.vue`
  - `code/frontend/src/views/platform/InstanceManage.vue`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/instanceManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第三十八轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/InstanceManage.test.ts src/views/platform/__tests__/instanceManagePanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第三十九轮修复进展

- 已完成：
  - `P2-5` `StudentManage.vue` 的头部说明、刷新动作和学生摘要卡已抽到独立组件 `StudentManageHeroPanel.vue`，路由页继续只持有筛选、分页、teacher 学生目录请求与学员跳转 owner。
  - 新增 `studentManagePanelExtraction.test.ts`，补对 `StudentManageHeroPanel` 的源码边界断言，避免后续再把 hero 与摘要卡回塞父页。
  - `StudentManage.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到组合源码检查，现有目录交互保持不变。
  - `StudentManage.vue` 本体行数已从 `373` 行降到 `273` 行，页面本体现在主要只剩学生目录筛选、分页和跳转 owner。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/student/StudentManageHeroPanel.vue`
  - `code/frontend/src/views/platform/StudentManage.vue`
  - `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/studentManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第三十九轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/StudentManage.test.ts src/views/platform/__tests__/studentManagePanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第四十轮修复进展

- 已完成：
  - `P2-5` `ClassManage.vue` 的头部说明、刷新动作和班级摘要卡已抽到独立组件 `ClassManageHeroPanel.vue`，路由页继续只持有班级目录加载、分页和班级详情跳转 owner。
  - 新增 `classManagePanelExtraction.test.ts`，补对 `ClassManageHeroPanel` 的源码边界断言，避免后续再把 hero 与摘要卡塞回父页。
  - `ClassManage.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到组合源码检查，管理页摘要卡护栏保持不变。
  - `ClassManage.vue` 本体行数已从 `281` 行降到 `180` 行，页面本体现在主要只剩班级目录与分页装配。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/class/ClassManageHeroPanel.vue`
  - `code/frontend/src/views/platform/ClassManage.vue`
  - `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/classManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第四十轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ClassManage.test.ts src/views/platform/__tests__/classManagePanelExtraction.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第四十一轮修复进展

- 已完成：
  - `P2-5` `ContestOperationsHub.vue` 的运维目录 hero 与摘要卡已抽到独立组件 `ContestOperationsHubHeroPanel.vue`，路由页继续只持有赛事列表加载、状态筛选、空态与跳转 owner。
  - 新增 `contestOperationsHubPanelExtraction.test.ts`，补对 `ContestOperationsHubHeroPanel` 的源码边界断言，避免后续再把运维头部与摘要卡塞回父页。
  - `ContestOperationsHub.test.ts` 与 `contestUiPrimitiveAdoption.test.ts` 已同步切到组合源码检查，现有运维目录交互与 primitive 护栏保持不变。
  - `ContestOperationsHub.vue` 本体行数已从 `403` 行降到 `339` 行，页面本体现在主要只剩目录列表、空态和跳转 owner。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestOperationsHubHeroPanel.vue`
  - `code/frontend/src/views/platform/ContestOperationsHub.vue`
  - `code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts`

## 第四十一轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ContestOperationsHub.test.ts src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts`
  - `npm run typecheck`

## 第四十二轮修复进展

- 已完成：
  - `P2-5` `ContestAnnouncements.vue` 的竞赛公告顶栏已抽到独立组件 `ContestAnnouncementsTopbarPanel.vue`，路由页继续只持有公告加载、发布、删除、只读分支和列表 owner。
  - 新增 `contestAnnouncementsPanelExtraction.test.ts`，补对 `ContestAnnouncementsTopbarPanel` 的源码边界断言，避免后续再把返回入口和状态条塞回父页。
  - `ContestAnnouncements.test.ts` 已同步校验新顶栏组件的结构契约，页面行为保持不变。
  - `ContestAnnouncements.vue` 本体行数已从 `403` 行降到 `350` 行，页面本体现在主要只剩表单、列表和权限分支 owner。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestAnnouncementsTopbarPanel.vue`
  - `code/frontend/src/views/platform/ContestAnnouncements.vue`
  - `code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts`

## 第四十二轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ContestAnnouncements.test.ts src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts`
  - `npm run typecheck`

## 第四十三轮修复进展

- 已完成：
  - `P2-5` `ChallengeImportManage.vue` 的导入页头部和三颗动作按钮已抽到独立组件 `ChallengeImportHeroPanel.vue`，路由页继续只持有上传、回执、导入队列与预览跳转 owner。
  - 新增 `challengeImportManagePanelExtraction.test.ts`，补对 `ChallengeImportHeroPanel` 的源码边界断言，避免后续再把导入页 hero 与动作按钮塞回父页。
  - `ChallengeImportManage.test.ts` 已同步校验新 hero 组件的结构契约，`journalPlatformShellStyles.test.ts` 基线保持不变。
  - `ChallengeImportManage.vue` 本体行数已从 `545` 行降到 `478` 行，页面本体现在主要只剩上传工作区、最近上传结果与待确认导入目录。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/ChallengeImportHeroPanel.vue`
  - `code/frontend/src/views/platform/ChallengeImportManage.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/challengeImportManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts`

## 第四十三轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/challengeImportManagePanelExtraction.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts`
  - `npm run typecheck`

## 第四十四轮修复进展

- 已完成：
  - `P2-5` `ChallengeImportManage.vue` 的最近上传结果工作区已抽到独立组件 `ChallengeImportUploadResultsPanel.vue`，路由页继续只持有上传回执数据和时间格式化 owner。
  - 新增 `challengeImportUploadResultsExtraction.test.ts`，补对 `ChallengeImportUploadResultsPanel` 的源码边界断言，避免后续再把上传回执卡片回塞父页。
  - `ChallengeImportManage.test.ts` 已继续覆盖上传后回执显示与预览跳转，`journalPlatformShellStyles.test.ts` 基线保持不变。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/ChallengeImportUploadResultsPanel.vue`
  - `code/frontend/src/views/platform/ChallengeImportManage.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/challengeImportUploadResultsExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/journalPlatformShellStyles.test.ts`

## 第四十四轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/challengeImportManagePanelExtraction.test.ts src/views/platform/__tests__/challengeImportUploadResultsExtraction.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts`
  - `npm run typecheck`

## 第四十五轮修复进展

- 已完成：
  - `P2-5` `ChallengeImportManage.vue` 的待确认导入队列工作区已抽到独立组件 `ChallengeImportQueuePanel.vue`，路由页继续只持有队列加载、标签映射和预览跳转 owner。
  - 新增 `challengeImportQueueExtraction.test.ts`，补对 `ChallengeImportQueuePanel` 的源码边界断言，并把“继续查看预览”按钮样式 ownership 一并移入子组件，修复 hero 抽离后按钮类名失去样式的问题。
  - `ChallengeImportManage.vue` 本体行数已从第四十三轮的 `478` 行降到 `210` 行，页面本体现在主要只剩上传入口、上传结果与队列组件装配。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/ChallengeImportQueuePanel.vue`
  - `code/frontend/src/views/platform/ChallengeImportManage.vue`
  - `code/frontend/src/views/platform/__tests__/challengeImportQueueExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts`
  - `code/frontend/src/components.d.ts`

## 第四十五轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeImportManage.test.ts src/views/platform/__tests__/challengeImportManagePanelExtraction.test.ts src/views/platform/__tests__/challengeImportUploadResultsExtraction.test.ts src/views/platform/__tests__/challengeImportQueueExtraction.test.ts src/views/platform/__tests__/journalPlatformShellStyles.test.ts`
  - `npm run typecheck`

## 第四十六轮修复进展

- 已完成：
  - `P2-5` `CheatDetection.vue` 的三段风险目录工作区已继续抽到独立组件 `CheatDetectionReviewPanels.vue`，路由页继续只持有风控数据加载、审计跳转、空错态与兜底空态 owner。
  - 新增 `cheatDetectionReviewPanelsExtraction.test.ts`，补对 `CheatDetectionReviewPanels` 的源码边界断言，避免后续再把高频提交、共享 IP 和审计联动模板塞回父页。
  - `cheatDetectionSurfaceAlignment.test.ts` 已同步切到组合源码检查，新组件接手目录行、badge 和空态边框样式后，现有深色 surface 护栏保持不变。
  - `CheatDetection.vue` 本体行数已从第三十七轮的 `409` 行降到 `172` 行，页面本体现在主要只剩加载、错误态和跳转装配。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/cheat/CheatDetectionReviewPanels.vue`
  - `code/frontend/src/views/platform/CheatDetection.vue`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionReviewPanelsExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`

## 第四十六轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/cheatDetectionReviewPanelsExtraction.test.ts src/views/platform/__tests__/CheatDetection.test.ts src/views/platform/__tests__/cheatDetectionPanelExtraction.test.ts src/views/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第四十七轮修复进展

- 已完成：
  - `P2-5` `ContestAnnouncements.vue` 的公告发布区、只读提示和历史公告列表已抽到独立组件 `ContestAnnouncementsWorkspacePanel.vue`，路由页继续只持有竞赛加载、公告 management owner 和返回工作台跳转。
  - 新增 `contestAnnouncementsWorkspaceExtraction.test.ts`，补对 `ContestAnnouncementsWorkspacePanel` 的源码边界断言，避免后续再把发布表单和历史公告模板塞回父页。
  - `ContestAnnouncements.test.ts` 保持行为验证不变，发布入口、只读态和历史列表都继续走现有断言。
  - `ContestAnnouncements.vue` 本体行数已从第四十二轮的 `350` 行降到 `162` 行，页面本体现在主要只剩加载失败态、顶栏和工作区装配。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestAnnouncementsWorkspacePanel.vue`
  - `code/frontend/src/views/platform/ContestAnnouncements.vue`
  - `code/frontend/src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ContestAnnouncements.test.ts`
  - `code/frontend/src/components.d.ts`

## 第四十七轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestAnnouncementsWorkspaceExtraction.test.ts src/views/platform/__tests__/contestAnnouncementsPanelExtraction.test.ts src/views/platform/__tests__/ContestAnnouncements.test.ts`
  - `npm run typecheck`

## 第四十八轮修复进展

- 已完成：
  - `P2-5` `ContestEdit.vue` 的顶部工作台 header 已抽到独立组件 `ContestEditTopbarPanel.vue`，路由页继续只持有竞赛详情加载、阶段切换、AWD 工作台 owner 和保存/跳转动作。
  - 新增 `contestEditTopbarExtraction.test.ts`，补对 `ContestEditTopbarPanel` 的源码边界断言，避免后续再把返回入口、公告入口和保存按钮模板塞回父页。
  - `ContestEdit.test.ts` 保持现有行为验证不变，抽取后公告入口、顶部保存按钮和其余 AWD 工作台交互基线都保持不变。
  - `ContestEdit.vue` 本体行数已从 `903` 行降到 `732` 行，页面本体进一步收敛到 stage 内容装配与 AWD 相关 owner。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestEditTopbarPanel.vue`
  - `code/frontend/src/views/platform/ContestEdit.vue`
  - `code/frontend/src/views/platform/__tests__/contestEditTopbarExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
  - `code/frontend/src/components.d.ts`

## 第四十八轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestEditTopbarExtraction.test.ts src/views/platform/__tests__/ContestEdit.test.ts`
  - `npm run typecheck`

## 第四十九轮修复进展

- 已完成：
  - `P2-5` `platform/ChallengeDetail.vue` 的 tab rail、loading 态和题目管理/题解两个 workspace 壳层已抽到独立组件 `AdminChallengeWorkspaceTabs.vue`，路由页继续只持有 query tab owner、详情加载、附件下载和 Flag 配置保存。
  - 新增 `challengeDetailWorkspaceExtraction.test.ts`，补对 `AdminChallengeWorkspaceTabs` 的源码边界断言，避免后续再把 tab rail 与题解工作区模板塞回父页。
  - `challengeDetailPanelExtraction.test.ts` 已同步切到新的 workspace 组件源码边界，`ChallengeDetail.test.ts` 的 tab 切换、题解页 query、附件下载与共享实例 Flag 护栏行为保持不变。
  - `ChallengeDetail.vue` 本体行数已从 `372` 行降到 `292` 行，页面本体进一步收敛到顶栏、数据 owner 与路由同步。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/AdminChallengeWorkspaceTabs.vue`
  - `code/frontend/src/views/platform/ChallengeDetail.vue`
  - `code/frontend/src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeDetail.test.ts`
  - `code/frontend/src/components.d.ts`

## 第四十九轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/challengeDetailWorkspaceExtraction.test.ts src/views/platform/__tests__/challengeDetailPanelExtraction.test.ts src/views/platform/__tests__/ChallengeDetail.test.ts`
  - `npm run typecheck`

## 第五十轮修复进展

- 已完成：
  - `P2-5` `ChallengePackageFormat.vue` 的 hero 文案、目录结构示例和 `challenge.yml` 示例已抽到独立组件 `ChallengePackageFormatGuidePanel.vue`，路由页继续只持有返回导入页导航和页面壳。
  - 新增 `challengePackageFormatGuideExtraction.test.ts`，补对 `ChallengePackageFormatGuidePanel` 的源码边界断言，避免后续再把整块示例文档模板塞回父页。
  - `ChallengePackageFormat.test.ts` 已同步切到新的 guide 组件源码边界，现有文档内容展示与 overline 护栏保持不变。
  - `ChallengePackageFormat.vue` 本体行数已从 `220` 行降到 `61` 行，页面本体现在主要只剩导航与展示组件装配。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/ChallengePackageFormatGuidePanel.vue`
  - `code/frontend/src/views/platform/ChallengePackageFormat.vue`
  - `code/frontend/src/views/platform/__tests__/challengePackageFormatGuideExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengePackageFormat.test.ts`
  - `code/frontend/src/components.d.ts`

## 第五十轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/challengePackageFormatGuideExtraction.test.ts src/views/platform/__tests__/ChallengePackageFormat.test.ts`
  - `npm run typecheck`

## 第五十一轮修复进展

- 已完成：
  - `P2-5` `ContestManage.vue` 的赛事结果导出、报表轮询和下载链路已抽到独立 composable `useContestExportFlow.ts`，页面继续只持有赛事目录加载、创建/编辑弹窗、公告抽屉和工作区装配。
  - 新增 `contestManageExportFlowExtraction.test.ts`，补对 `useContestExportFlow` 抽取边界的源码断言，避免后续再把导出状态机和报表下载逻辑塞回父页。
  - `ContestManage.test.ts` 保持现有行为验证不变，导出失败不抛全局错误页、竞赛目录筛选和 AWD 启动 gate 相关交互基线都保持不变。
  - `ContestManage.vue` 本体行数已从 `218` 行降到 `137` 行，页面本体进一步收敛到目录 owner、弹窗开关和子组件事件编排。
- 本轮涉及文件：
  - `code/frontend/src/composables/useContestExportFlow.ts`
  - `code/frontend/src/views/platform/ContestManage.vue`
  - `code/frontend/src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ContestManage.test.ts`

## 第五十一轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestManageExportFlowExtraction.test.ts src/views/platform/__tests__/ContestManage.test.ts`
  - `npm run typecheck`

## 第五十二轮修复进展

- 已完成：
  - `P2-5` `ImageManage.vue` 的页面 owner 状态已抽到独立 composable `useImageManagePage.ts`，把镜像筛选、排序、轮询刷新、创建/删除和格式化辅助从路由页移出，页面继续只持有工作区、详情弹窗和创建弹窗装配。
  - 新增 `imageManagePageStateExtraction.test.ts`，补对 `useImageManagePage` 抽取边界的源码断言，避免后续再把页面级状态机、轮询和创建删除逻辑塞回父页。
  - `ImageManage.test.ts`、`imageManageWorkspaceExtraction.test.ts` 和 `imageManageModalExtraction.test.ts` 保持行为验证不变，镜像列表筛选排序、自动轮询、创建删除与 modal 装配基线都保持不变。
  - `ImageManage.vue` 本体行数已从 `396` 行降到 `145` 行，页面本体进一步收敛到布局、组件绑定和样式壳层。
- 本轮涉及文件：
  - `code/frontend/src/composables/useImageManagePage.ts`
  - `code/frontend/src/views/platform/ImageManage.vue`
  - `code/frontend/src/views/platform/__tests__/imageManagePageStateExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ImageManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/imageManageWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/imageManageModalExtraction.test.ts`

## 第五十二轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/imageManagePageStateExtraction.test.ts src/views/platform/__tests__/ImageManage.test.ts src/views/platform/__tests__/imageManageWorkspaceExtraction.test.ts src/views/platform/__tests__/imageManageModalExtraction.test.ts`
  - `npm run typecheck`

## 第五十三轮修复进展

- 已完成：
  - `P2-5` `AuditLog.vue` 的页面 owner 状态已抽到独立 composable `useAuditLogPage.ts`，把路由 query 同步、筛选节流、请求代次控制、排序与执行人详情弹窗状态从路由页移出，页面继续只持有工作区与 modal 装配。
  - 新增 `auditLogPageStateExtraction.test.ts`，补对 `useAuditLogPage` 抽取边界的源码断言，避免后续再把页面级状态机和 query 同步逻辑塞回父页。
  - `AuditLog.test.ts` 与 `auditLogWorkspaceExtraction.test.ts` 保持行为验证不变，预置 query 加载、筛选同步、执行人详情弹窗与目录工作区基线都保持不变。
  - `AuditLog.vue` 本体行数已从 `323` 行降到 `89` 行，页面本体进一步收敛到布局、组件绑定和样式壳层。
- 本轮涉及文件：
  - `code/frontend/src/composables/useAuditLogPage.ts`
  - `code/frontend/src/views/platform/AuditLog.vue`
  - `code/frontend/src/views/platform/__tests__/auditLogPageStateExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/AuditLog.test.ts`
  - `code/frontend/src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`

## 第五十三轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/auditLogPageStateExtraction.test.ts src/views/platform/__tests__/AuditLog.test.ts src/views/platform/__tests__/auditLogWorkspaceExtraction.test.ts`
  - `npm run typecheck`

## 第五十四轮修复进展

- 已完成：
  - `P2-5` `ChallengeManage.vue` 的页面 owner 状态已抽到独立 composable `useChallengeManagePage.ts`，把题目排序、空态/错态文案、筛选状态和路由跳转交互从路由页移出，页面继续只持有 hero 与目录组件装配。
  - 新增 `challengeManagePageStateExtraction.test.ts`，补对 `useChallengeManagePage` 抽取边界的源码断言，避免后续再把页面级状态与交互 owner 塞回父页。
  - `ChallengeManage.test.ts` 与 `challengeManageDirectoryExtraction.test.ts` 保持行为验证不变，排序切换、错误态、action menu、目录工具栏与 hero 工作区基线都保持不变。
  - `ChallengeManage.vue` 本体行数已从 `220` 行降到 `115` 行，页面本体进一步收敛到布局、组件绑定和样式壳层。
- 本轮涉及文件：
  - `code/frontend/src/composables/useChallengeManagePage.ts`
  - `code/frontend/src/views/platform/ChallengeManage.vue`
  - `code/frontend/src/views/platform/__tests__/challengeManagePageStateExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`

## 第五十四轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/challengeManagePageStateExtraction.test.ts src/views/platform/__tests__/ChallengeManage.test.ts src/views/platform/__tests__/challengeManageDirectoryExtraction.test.ts`
  - `npm run typecheck`

## 第五十五轮修复进展

- 已完成：
  - `P2-5` 相关的管理端 surface 回归基线已补齐，`platformManagementSurfaceAlignment.test.ts` 不再继续盯着旧页面文件，而是跟随当前抽取边界校验 `ContestEditTopbarPanel`、`AdminChallengeProfilePanel`、`AdminChallengeWorkspaceTabs` 与 `CheatDetectionReviewPanels`。
  - `ChallengeDetail` 相关断言已从脆弱的单行源码匹配改成对当前组件源码的稳定匹配，避免仅因模板换行或抽取组件导致假失败。
  - 本轮只修验证护栏，不改业务实现；目标是让现有的页面减重结果能继续被这条回归测试正确保护。
- 本轮涉及文件：
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第五十五轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第五十六轮修复进展

- 已完成：
  - `P2-5` `ContestEdit.vue` 的 stage 工作区模板已抽到独立组件 `ContestEditWorkspacePanel.vue`，把 `basics / pool / awd-config / preflight / operations` 五段工作区渲染和对应滚动/过渡样式从路由页移出，页面继续只持有竞赛详情加载、AWD 数据 owner、保存流程与对话框状态。
  - 新增 `contestEditWorkspaceExtraction.test.ts`，补对 `ContestEditWorkspacePanel` 抽取边界的源码断言，避免后续再把 stage 工作区模板回塞父页。
  - `contestUiPrimitiveAdoptionPhase2.test.ts`、`platformManagementSurfaceAlignment.test.ts` 已同步切到新的工作区组件源码边界，抽取后共享按钮、滚动容器和管理端 surface 护栏都保持不变。
  - `ContestEdit.vue` 本体行数已从 `732` 行降到 `599` 行，路由页进一步收敛到顶部导航、工作区状态 owner 与 AWD 配置保存链路。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestEditWorkspacePanel.vue`
  - `code/frontend/src/views/platform/ContestEdit.vue`
  - `code/frontend/src/views/platform/__tests__/contestEditWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts`

## 第五十六轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestEditWorkspaceExtraction.test.ts src/views/platform/__tests__/contestEditTopbarExtraction.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/platformRootShellCleanup.test.ts`
  - `npm run typecheck`

## 第五十七轮修复进展

- 已完成：
  - `P2-5` `ContestEdit.vue` 的 AWD 工作区 owner 状态已抽到独立 composable `useContestEditAwdWorkspace.ts`，把 AWD 配置列表同步、赛前检查同步、题目聚焦导航、服务模板目录加载和配置保存链路从路由页移出，页面继续只持有竞赛详情加载、stage 选择和开赛 gate。
  - 新增 `contestEditAwdWorkspaceExtraction.test.ts`，补对 `useContestEditAwdWorkspace` 抽取边界的源码断言，避免后续再把 AWD 工作区状态机和动作回塞父页。
  - `ContestEdit.test.ts` 与管理端 surface 护栏保持行为验证不变，题目池跳转 AWD 配置、赛前检查回跳、模板目录加载、AWD 配置保存和阶段级 loading/error 基线都保持不变。
  - `ContestEdit.vue` 本体行数已从 `599` 行降到 `437` 行，路由页进一步收敛到赛事详情 owner、保存流程与 AWD 启动 override 对话框。
- 本轮涉及文件：
  - `code/frontend/src/composables/useContestEditAwdWorkspace.ts`
  - `code/frontend/src/views/platform/ContestEdit.vue`
  - `code/frontend/src/views/platform/__tests__/contestEditAwdWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestEditWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestEditTopbarExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts`

## 第五十七轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestEditAwdWorkspaceExtraction.test.ts src/views/platform/__tests__/contestEditWorkspaceExtraction.test.ts src/views/platform/__tests__/contestEditTopbarExtraction.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase2.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/platformRootShellCleanup.test.ts`
  - `npm run typecheck`

## 第五十八轮修复进展

- 已完成：
  - `P2-5` `InstanceManage.vue` 的实例目录工作区已抽到独立组件 `InstanceManageWorkspacePanel.vue`，把实例表格、空态、分页、错误态和状态 pill 样式从路由页移出，页面继续只持有 teacher 实例列表加载、销毁动作与返回概览跳转。
  - 新增 `instanceManageWorkspaceExtraction.test.ts`，补对 `InstanceManageWorkspacePanel` 抽取边界的源码断言，避免后续再把目录模板和行内状态样式回塞父页。
  - `InstanceManage.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到新的工作区组件源码边界，实例页的 teacher API 接口复用、销毁流程、目录 spacing 和 dark surface 护栏都保持不变。
  - `InstanceManage.vue` 本体行数已从 `288` 行降到 `176` 行，路由页进一步收敛到实例数据 owner、分页状态和销毁交互。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/instance/InstanceManageWorkspacePanel.vue`
  - `code/frontend/src/views/platform/InstanceManage.vue`
  - `code/frontend/src/views/platform/__tests__/InstanceManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/instanceManageWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/instanceManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第五十八轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/instanceManageWorkspaceExtraction.test.ts src/views/platform/__tests__/instanceManagePanelExtraction.test.ts src/views/platform/__tests__/InstanceManage.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第五十九轮修复进展

- 已完成：
  - `P2-5` `ContestOperationsHub.vue` 的赛事运维目录工作区已抽到独立组件 `ContestOperationsHubWorkspacePanel.vue`，把 loading/error/empty 三段状态、运维赛事目录卡片、进入运维台与返回编辑按钮模板从路由页移出，页面继续只持有赛事列表加载和路由跳转 owner。
  - 新增 `contestOperationsHubWorkspaceExtraction.test.ts`，补对 `ContestOperationsHubWorkspacePanel` 抽取边界的源码断言，避免后续再把运维目录工作区模板回塞父页。
  - `ContestOperationsHub.test.ts`、`contestUiPrimitiveAdoption.test.ts` 和 `contestUiPrimitiveAdoptionPhase27.test.ts` 已同步切到新的工作区组件源码边界，现有共享按钮原语、目录文案和运维入口跳转基线都保持不变。
  - `ContestOperationsHub.vue` 本体行数已从 `339` 行降到 `129` 行，路由页进一步收敛到 AWD 赛事筛选、推荐赛事计算和进入具体工作区的导航动作。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestOperationsHubWorkspacePanel.vue`
  - `code/frontend/src/views/platform/ContestOperationsHub.vue`
  - `code/frontend/src/views/platform/__tests__/ContestOperationsHub.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts`

## 第五十九轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/contestOperationsHubWorkspaceExtraction.test.ts src/views/platform/__tests__/contestOperationsHubPanelExtraction.test.ts src/views/platform/__tests__/ContestOperationsHub.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoption.test.ts src/views/platform/__tests__/contestUiPrimitiveAdoptionPhase27.test.ts`
  - `npm run typecheck`

## 第六十轮修复进展

- 已完成：
  - `P2-5` `StudentManage.vue` 的学生目录工作区已抽到独立组件 `StudentManageWorkspacePanel.vue`，把搜索栏、班级筛选、表格、空态、分页和错误态从路由页移出，页面继续只持有 teacher 学员目录 query owner、班级列表加载和学员分析跳转。
  - 新增 `studentManageWorkspaceExtraction.test.ts`，补对 `StudentManageWorkspacePanel` 抽取边界的源码断言，避免后续再把学生目录工作区模板回塞父页。
  - `StudentManage.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到新的工作区组件源码边界，学员目录检索、班级筛选、学员分析跳转和 dark surface spacing 护栏都保持不变。
  - `StudentManage.vue` 本体行数已从 `273` 行降到 `180` 行，路由页进一步收敛到筛选参数 owner、目录请求调度和路由跳转。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/student/StudentManageWorkspacePanel.vue`
  - `code/frontend/src/views/platform/StudentManage.vue`
  - `code/frontend/src/views/platform/__tests__/StudentManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/studentManageWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/studentManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第六十轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/studentManageWorkspaceExtraction.test.ts src/views/platform/__tests__/studentManagePanelExtraction.test.ts src/views/platform/__tests__/StudentManage.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第六十一轮修复进展

- 已完成：
  - `P2-5` `ClassManage.vue` 的班级目录工作区已抽到独立组件 `ClassManageWorkspacePanel.vue`，把目录标题、表格、空态、分页和错误态从路由页移出，页面继续只持有班级列表加载、分页状态、班级详情跳转和总学生数汇总。
  - 新增 `classManageWorkspaceExtraction.test.ts`，补对 `ClassManageWorkspacePanel` 抽取边界的源码断言，避免后续再把班级目录工作区模板回塞父页。
  - `ClassManage.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到新的工作区组件源码边界，班级目录 teacher API 复用、班级详情跳转和 dark surface spacing 护栏都保持不变。
  - `ClassManage.vue` 本体行数已从 `180` 行降到 `107` 行，路由页进一步收敛到列表 owner、分页调度和跳转动作。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/class/ClassManageWorkspacePanel.vue`
  - `code/frontend/src/views/platform/ClassManage.vue`
  - `code/frontend/src/views/platform/__tests__/ClassManage.test.ts`
  - `code/frontend/src/views/platform/__tests__/classManageWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/classManagePanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第六十一轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/classManageWorkspaceExtraction.test.ts src/views/platform/__tests__/classManagePanelExtraction.test.ts src/views/platform/__tests__/ClassManage.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第六十二轮修复进展

- 已完成：
  - `P2-5` `AWDReviewIndex.vue` 的复盘页头部与摘要卡已抽到独立组件 `AwdReviewHeroPanel.vue`，把返回概览、刷新目录和三张 summary 卡的展示模板从路由页移出，页面继续只持有 `useTeacherAwdReviewIndex` 返回的目录 owner、筛选重置和复盘详情跳转。
  - 新增 `awdReviewHeroExtraction.test.ts`，补对 `AwdReviewHeroPanel` 抽取边界的源码断言，避免后续再把 hero 和 summary 模板回塞父页。
  - `AWDReviewIndex.test.ts` 与 `platformManagementSurfaceAlignment.test.ts` 已同步切到新的组合源码边界，返回概览、刷新目录、目录工作区渲染和管理端 summary panel 护栏都保持不变。
  - `AWDReviewIndex.vue` 本体行数已从 `217` 行降到 `97` 行，路由页进一步收敛到筛选状态、目录数据映射和目录组件装配。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/awd-review/AwdReviewHeroPanel.vue`
  - `code/frontend/src/views/platform/AWDReviewIndex.vue`
  - `code/frontend/src/views/platform/__tests__/AWDReviewIndex.test.ts`
  - `code/frontend/src/views/platform/__tests__/awdReviewHeroExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/awdReviewDirectoryExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`

## 第六十二轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/awdReviewHeroExtraction.test.ts src/views/platform/__tests__/awdReviewDirectoryExtraction.test.ts src/views/platform/__tests__/AWDReviewIndex.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `npm run typecheck`

## 第六十三轮修复进展

- 已完成：
  - `P2-5` `ContestOperations.vue` 的赛事运维工作区顶部壳层已抽到独立组件 `ContestOperationsTopbarPanel.vue`，把返回目录、赛事标题和进入竞赛工作室按钮的展示模板从路由页移出，页面继续只持有 `contestId`、`activeTab`、`route.query.activeTab` 同步、赛事详情加载和主路由动作。
  - 父页新增 `activeTab` 归一化逻辑，非法 `query.activeTab` 会统一回退到 `matrix`，避免路由层把任意字符串继续透传给 `AWDOperationsPanel`。
  - 新增 `ContestOperations.test.ts`，补对父页 owner 边界的行为断言，覆盖合法 tab 透传、非法 tab 回退，以及顶栏子组件事件触发后的返回目录和进入编辑页跳转。
  - `contestOperationsTopbarExtraction.test.ts` 已作为源码护栏固定 `ContestOperationsTopbarPanel` 抽取边界，避免后续再把顶部壳层塞回父页。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/contest/ContestOperationsTopbarPanel.vue`
  - `code/frontend/src/views/platform/ContestOperations.vue`
  - `code/frontend/src/views/platform/__tests__/ContestOperations.test.ts`
  - `code/frontend/src/views/platform/__tests__/contestOperationsTopbarExtraction.test.ts`

## 第六十三轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ContestOperations.test.ts src/views/platform/__tests__/contestOperationsTopbarExtraction.test.ts`
  - `npm run typecheck`

## 第六十四轮修复进展

- 已完成：
  - `P2-5` `ChallengeImportPreview.vue` 的导入预览工作区壳层已抽到独立组件 `ChallengeImportPreviewWorkspacePanel.vue`，把顶部返回区、页面说明、loading/empty 状态和预览展示壳从路由页移出，页面继续只持有 `importId` 路由同步、预览加载/重置、确认导入和返回导航动作。
  - `ChallengeImportPreview.test.ts` 已补对父页 owner 边界的行为断言，覆盖工作区子组件触发返回导入页、返回待确认队列和确认导入三条主动作时，路由页仍负责导航与提交链路。
  - 新增 `challengeImportPreviewWorkspaceExtraction.test.ts`，补对 `ChallengeImportPreviewWorkspacePanel` 抽取边界的源码断言，避免后续再把预览壳层和空态模板回塞父页。
  - `platformRootShellCleanup.test.ts` 已同步切到新的工作区组件源码边界，继续约束导入预览页沿用共享管理员根壳，不因抽取而丢失 surface 护栏。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/challenge/ChallengeImportPreviewWorkspacePanel.vue`
  - `code/frontend/src/views/platform/ChallengeImportPreview.vue`
  - `code/frontend/src/views/platform/__tests__/ChallengeImportPreview.test.ts`
  - `code/frontend/src/views/platform/__tests__/challengeImportPreviewWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts`

## 第六十四轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/ChallengeImportPreview.test.ts src/views/platform/__tests__/challengeImportPreviewWorkspaceExtraction.test.ts src/views/platform/__tests__/platformRootShellCleanup.test.ts`
  - `npm run typecheck`

## 第六十五轮修复进展

- 已完成：
  - `P2-5` `CheatDetection.vue` 的作弊检测工作区壳层已抽到独立组件 `CheatDetectionWorkspacePanel.vue`，把 hero、loading、summary、风险目录、错误态和空态展示从路由页移出，页面继续只持有风险数据加载、刷新、错误状态、时间格式化和审计日志跳转。
  - 新增 `cheatDetectionWorkspaceExtraction.test.ts`，补对 `CheatDetectionWorkspacePanel` 抽取边界的源码断言，避免后续再把工作区壳层模板回塞父页。
  - `CheatDetection.test.ts` 已补对父页 owner 边界的行为断言，覆盖工作区子组件触发刷新和审计跳转时，路由页仍负责请求与导航链路。
  - `cheatDetectionPanelExtraction.test.ts`、`cheatDetectionReviewPanelsExtraction.test.ts`、`cheatDetectionSurfaceAlignment.test.ts`、`platformManagementSurfaceAlignment.test.ts` 与 `platformRootShellCleanup.test.ts` 已同步切到新的工作区组件源码边界，surface、summary、目录和根壳护栏保持不变。
- 本轮涉及文件：
  - `code/frontend/src/components/platform/cheat/CheatDetectionWorkspacePanel.vue`
  - `code/frontend/src/views/platform/CheatDetection.vue`
  - `code/frontend/src/views/platform/__tests__/CheatDetection.test.ts`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionWorkspaceExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionPanelExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionReviewPanelsExtraction.test.ts`
  - `code/frontend/src/views/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts`
  - `code/frontend/src/views/platform/__tests__/platformRootShellCleanup.test.ts`

## 第六十五轮验证

- 已执行：
  - `npm run test:run -- src/views/platform/__tests__/CheatDetection.test.ts src/views/platform/__tests__/cheatDetectionWorkspaceExtraction.test.ts src/views/platform/__tests__/cheatDetectionPanelExtraction.test.ts src/views/platform/__tests__/cheatDetectionReviewPanelsExtraction.test.ts src/views/platform/__tests__/cheatDetectionSurfaceAlignment.test.ts src/views/platform/__tests__/platformManagementSurfaceAlignment.test.ts src/views/platform/__tests__/platformRootShellCleanup.test.ts`
  - `npm run typecheck`

## 备注

- 本文件用于记录本轮前端专项审查结论与修复顺序。
- 后续每修完一个高优先级问题，都应回写验证结果，避免文档和实际状态脱节。
