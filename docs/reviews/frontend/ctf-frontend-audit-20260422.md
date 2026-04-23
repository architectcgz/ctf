# CTF 前端专项审查（2026-04-22）

## 审查信息

| 字段 | 说明 |
|------|------|
| 审查主题 | 前端架构质量、异步边界、状态同步、主题一致性与门禁健康度 |
| 审查范围 | `code/frontend/src` 路由、关键视图、composables、stores、共享样式与验证门禁 |
| 审查日期 | 2026-04-22 |
| 审查方式 | 静态代码审查 + 最小验证基线检查 |
| 审查状态 | 已记录，已完成八轮最小高收益修复 |

## 当前结论

- 当前前端不是“局部页面粗糙”，而是已经出现工程漂移：
  - 静态门禁失效，当前分支 `typecheck` 不能通过。
  - 关键数据页的异步边界处理不一致，存在真实竞态风险。
  - 部分状态面板只更新本地派生值，不再与服务端状态保持同步。
  - 主题系统和视觉 token 已经建立，但生产代码仍频繁绕过共享变量直接写死颜色。
  - 正式路由中混入设计实验页，产品边界不够干净。

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
  - 处理策略：
    - 本轮先记录为高风险设计问题，不在本次最小修复里直接重构认证方案。

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
    - `code/frontend/src/views/contests/ContestDetail.vue`
    - `code/frontend/src/views/platform/ImageManage.vue`
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

## 备注

- 本文件用于记录本轮前端专项审查结论与修复顺序。
- 后续每修完一个高优先级问题，都应回写验证结果，避免文档和实际状态脱节。
