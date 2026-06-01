# CTF 架构收口待修复清单

更新日期：2026-03-14

## 背景

本清单只记录“架构边界收口”相关剩余工作，用于约束后续修复顺序，避免在持续修复过程中遗漏已发现问题。

当前判断口径：

1. 优先收口后端 `Service -> Repository` 边界，减少 service 直连 `gorm.DB`
2. 优先收口后端 composition root，避免同类 service 被多处重复实例化
3. 前端按“`views` 负责装配，业务编排下沉到 composable / page model”继续统一

## 已完成（2026-03-14 第一轮）

1. [x] `internal/app` 装配层复用同一套 `containerService` / `assessmentService`
   - 已避免 HTTP 路由和后台 cleaner / updater 各自重复创建一套实例
   - 相关代码：
     - `code/backend/internal/app/router.go`
     - `code/backend/internal/app/http_server.go`

2. [x] `contest.ParticipationService` 去除直接 `gorm.DB` 访问
   - 报名、审核、公告、个人竞赛进度查询已下沉到 `ParticipationRepository`
   - 相关代码：
     - `code/backend/internal/module/contest/participation_service.go`
     - `code/backend/internal/module/contest/participation_repository.go`

3. [x] `contest.SubmissionService` 去除直接 `gorm.DB` 访问
   - 竞赛提交、事务内加锁、首血更新、动态计分回写已下沉到 `SubmissionRepository`
   - 相关代码：
     - `code/backend/internal/module/contest/submission_service.go`
     - `code/backend/internal/module/contest/submission_repository.go`

4. [x] `assessment.RecommendationService` 去除 `repo.db` 回流访问
   - 已解题目查询改为走 repository 方法
   - 相关代码：
     - `code/backend/internal/module/assessment/recommendation_service.go`
     - `code/backend/internal/module/assessment/repository.go`

5. [x] `practice.Service` 去除直接 `gorm.DB` 持有
   - 实例创建事务、竞赛作用域查询、实例锁、端口预留/绑定已下沉到 `practice.Repository`
   - service 仅保留业务编排，并改为依赖更小的跨模块接口
   - 相关代码：
     - `code/backend/internal/module/practice/service.go`
     - `code/backend/internal/module/practice/repository.go`
     - `code/backend/internal/module/practice/contest_instance_service_test.go`

6. [x] `contest.AWDService` 去除直接 `gorm.DB` 持有
   - 轮次、服务检查、攻击日志、队伍/报名/题目查询已下沉到 `AWDRepository`
   - service 保留规则编排，`AWDRoundUpdater` 和分数缓存同步改由 repository 包装调用
   - 相关代码：
     - `code/backend/internal/module/contest/awd_service.go`
     - `code/backend/internal/module/contest/awd_repository.go`
     - `code/backend/internal/module/contest/awd_service_test.go`

7. [x] `practice.ScoreService` 去除直接 `gorm.DB` 持有
   - 题目分值读取、已解题目查询、用户得分 upsert、排行榜回源、用户名批量查询已下沉到 `practice.Repository`
   - `UpdateUserScore` 同步改为批量读取题目信息，避免循环内逐题查库
   - 相关代码：
     - `code/backend/internal/module/practice/score_service.go`
     - `code/backend/internal/module/practice/score_repository.go`
     - `code/backend/internal/module/practice/score_service_test.go`

8. [x] `challenge.FlagService` 去除直接 `gorm.DB` 持有
   - 题目查询与 Flag 配置更新改为走 `challenge.Repository`
   - service 保留静态 / 动态 Flag 规则、格式校验与动态 Flag 生成逻辑
   - 相关代码：
     - `code/backend/internal/module/challenge/flag_service.go`
     - `code/backend/internal/module/challenge/repository.go`
     - `code/backend/internal/module/challenge/flag_service_test.go`

## 待修复项

### P0：收口阶段结果

1. [x] 当前 `code/backend/internal/module/*service.go` 已无直接持有 `*gorm.DB` 的实现
   - 本轮复查确认：
     - `code/backend/internal/module/assessment/report_service.go`
     - `code/backend/internal/module/system/dashboard_service.go`
   - 说明：
     - 上述模块当前已经分别通过 repository / runtime 接口访问底层资源
     - 后续若新增 `service -> gorm.DB` 回流，应继续补记到本清单

### P1：统一后端装配与生命周期边界

1. [x] `internal/app` 运行时组件复用关系已完成复查
   - 当前确认：
     - `containerService` 已由统一 composition root 创建，并复用于 handler / cleaner / AWD flag injector
     - `assessmentService` 已由统一 composition root 创建，并复用于 handler / assessment cleaner
   - 结论：
     - 当前未发现同类核心运行时组件被多处重复创建但未共享的情况
     - 后续新增后台任务仍需继续遵守统一 composition root 约束

2. [x] 后端架构文档已同步到当前实现边界
   - 已补充内容：
     - `internal/app` 的 composition root
     - HTTP 链路与后台 cleaner / updater 的共享依赖关系
     - 统一 lifecycle 管理与关闭顺序
   - 文档位置：
     - `docs/architecture/backend/01-system-architecture.md`

### P1：统一前端页面层模式

1. [ ] 把仍然偏“胖页面”的核心页面下沉到 composable / page model
   - 优先级建议：
     1. [x] `code/frontend/src/views/admin/ChallengeManage.vue`
        - 本轮已下沉到 `code/frontend/src/composables/useAdminChallenges.ts`
        - 页面保留装配和展示映射，列表加载、镜像预取、表单草稿、Flag 配置、发布/删除动作已下沉
     2. [x] `code/frontend/src/views/contests/ContestDetail.vue`
        - 本轮已下沉到 `code/frontend/src/composables/useContestDetailPage.ts`
        - 页面保留模板装配和展示函数，初始并发加载、倒计时、队伍操作、Flag 提交与弹窗状态已下沉
     3. [x] `code/frontend/src/views/instances/InstanceList.vue`
        - 本轮已下沉到 `code/frontend/src/composables/useInstanceListPage.ts`
        - 页面保留列表模板和展示映射，实例加载、倒计时、即将过期提醒、复制/打开/延时/销毁动作已下沉
     4. [x] `code/frontend/src/views/profile/SkillProfile.vue`
        - 本轮已下沉到 `code/frontend/src/composables/useSkillProfilePage.ts`
        - 页面保留展示模板，教师学员切换、画像/推荐加载、薄弱维度派生与跳转动作已下沉
   - 现状影响：
     - 页面直接承担 API 调用、表单编排、定时器、副作用、状态恢复
     - 同一仓库内已同时存在“view + composable”与“胖页面直调 API”两种实现方式，后续会继续分叉
   - 目标：
     - `views` 只负责页面装配
     - 业务编排统一收进 composable

2. [ ] 继续约束组件层不要直接承担页面级数据拉取
   - 当前需重点关注：
     - [x] `code/frontend/src/components/common/InstancePanel.vue`
       - 已改为 props / emits 驱动的展示组件，不再直接拉取实例数据或发起实例动作
     - [x] `code/frontend/src/components/layout/NotificationDropdown.vue`
       - 已下沉到 `code/frontend/src/composables/useNotificationDropdown.ts`
       - 组件保留弹层展示与交互绑定，通知 store / 已读 API / 浮层副作用编排已下沉
     - [x] `code/frontend/src/components/admin/topology/ChallengeTopologyStudioPage.vue`
       - 已下沉到 `code/frontend/src/composables/useChallengeTopologyStudioPage.ts`
       - 组件保留拓扑编排模板，拓扑/模板 API、画布状态、模板动作与全局快捷键编排已下沉
     - [x] `code/frontend/src/components/admin/writeup/ChallengeWriteupEditorPage.vue`
       - 已下沉到 `code/frontend/src/composables/useChallengeWriteupEditorPage.ts`
       - 组件保留题解编辑模板，题解加载、保存、删除、表单恢复与校验编排已下沉
   - 说明：
     - 组件层允许保留必要的局部交互
     - 但不应继续扩散为页面级 API 编排中心

## 推荐修复顺序

1. 当前清单项已全部完成，后续若新增 “view / component 直接编排 API” 回流，应继续补记到本清单

## 关闭条件

满足以下条件后，本清单可转为“已完成”：

1. 后端高复杂度 service 不再直接持有 `*gorm.DB`
2. 后端后台任务与 HTTP 链路共享统一 composition root
3. 前端核心页面已统一为“view 装配 + composable 编排”模式
4. 架构文档已同步到当前实现边界

## 当前结论

1. [x] 本清单记录的后端与前端架构边界收口项已全部完成
