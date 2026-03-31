# CTF 前端任务拆分与状态

## 文档信息

- 范围：`/home/azhi/workspace/projects/ctf/code/frontend`
- 目标：记录前端任务拆分、当前完成状态、剩余待办与后续执行顺序
- 流水线：`leader -> requirements-analyst（按需） -> frontend-engineer -> code-reviewer -> test-engineer`
- 最后更新：2026-03-31

## 当前判断

- 前端主线能力已完成当前主线后端的联调准备：鉴权守卫、学员端、管理端、教师端、通知链路和竞赛详情关键交互均已接上真实接口。
- 教师端相关后端接口已补齐并合入主线，不再阻塞 `TeacherDashboard` / `ClassManagement`。
- 通知能力已完成前后端闭环，并补齐管理员发布能力：
  - 后端：`POST /auth/ws-ticket`、`GET /notifications`、`PUT /notifications/:id/read`、`GET /ws/notifications`、`POST /admin/notifications`
  - 前端：通知下拉、通知中心、WebSocket 实时同步、管理员在 `/notifications` 页内发布通知
- 管理端 `UserManage`、`CheatDetection` 已从降级态切回真实接口页；报告导出也已补状态查询与轮询。
- 当前如果继续联调，重点已经从“补接口缺口”转向“按业务流实际验证”。

## 已完成任务

| 任务                    | 状态    | 结果                                  | 对应提交  |
| ----------------------- | ------- | ------------------------------------- | --------- |
| F0 基础对齐             | ✅ 完成 | 恢复路由守卫、校正前后端契约          | `fe30bcc` |
| F1 学员侧页面完善       | ✅ 完成 | 完成 Dashboard、Profile、个人报告导出 | `c1e0006` |
| F2 管理侧页面完善       | ✅ 完成 | 完成系统概览、审计日志、作弊研判页    | `d5d09fa` |
| F3 教师侧报告导出       | ✅ 完成 | 完成教师端报告导出                    | `f666799` |
| F4 教师侧班级与教学概览 | ✅ 完成 | 完成教学概览、班级管理                | `766f29e` |
| F6 通知实时链路         | ✅ 完成 | 完成 ws-ticket + WebSocket 通知同步   | `6fccfb8` |
| F8 管理员通知发布       | ✅ 完成 | 在通知中心完成发布抽屉与目标范围投递 | 当前工作树 |
| F7 联调缺口补齐         | ✅ 完成 | 补齐用户管理、作弊检测、报告状态查询  | `df2a7ab` |

## 已完成质量闭环

- 前端 review 文档：
  - [ctf-frontend-code-review-contract-guard-round1-11e5349.md](/home/azhi/workspace/projects/ctf/code/frontend/docs/reviews/frontend/ctf-frontend-code-review-contract-guard-round1-11e5349.md)
  - [ctf-frontend-code-review-student-pages-round1-821a2be.md](/home/azhi/workspace/projects/ctf/code/frontend/docs/reviews/frontend/ctf-frontend-code-review-student-pages-round1-821a2be.md)
  - [ctf-frontend-code-review-admin-pages-round1-da5c98a.md](/home/azhi/workspace/projects/ctf/code/frontend/docs/reviews/frontend/ctf-frontend-code-review-admin-pages-round1-da5c98a.md)
  - [ctf-frontend-code-review-teacher-report-round1-8ae3904.md](/home/azhi/workspace/projects/ctf/code/frontend/docs/reviews/frontend/ctf-frontend-code-review-teacher-report-round1-8ae3904.md)
  - [ctf-frontend-code-review-teacher-pages-round1-a2c09af.md](/home/azhi/workspace/projects/ctf/code/frontend/docs/reviews/frontend/ctf-frontend-code-review-teacher-pages-round1-a2c09af.md)
- 当前主线验证通过：
  - `npm run test:run`
  - `npm run typecheck`

## 已收口任务

### R1 管理端用户管理页去 mock并接真实接口

- 状态：`completed`
- 目标：将 `UserManage` 从静态 mock 页面改成真实可接入页面，或在后端接口缺失时显式降级。
- 范围：
  - `src/views/admin/UserManage.vue`
  - 可能涉及 `src/api/admin.ts`、类型定义与测试
- 当前现状：
  - 后端已补齐 `/admin/users`、`/admin/users/import`
  - `UserManage` 已恢复真实列表、筛选、分页、创建、编辑、删除、CSV 导入
- 验收标准：
  - 不保留假数据或降级态
  - 页面操作全部接真实接口
  - 已补测试

### R2 管理端非核心页真实性核对

- 状态：`completed`
- 目标：继续清理管理端中可能残留的“半成品页面”与假交互。
- 范围：
  - `src/views/admin/ContestManage.vue`
  - `src/views/admin/ImageManage.vue`
  - `src/views/admin/ChallengeManage.vue`
  - 相关 API 和测试
- 当前现状：
  - `ContestManage` 已移除 `mockContests`，改为真实接入 `/admin/contests`
  - `admin.ts` 已补竞赛字段归一化，修正 `start_time/end_time/status=registration` 与前端契约差异
  - `ContestManage` 只保留后端已提供的列表、创建、编辑能力；删除能力不再伪造
  - `CheatDetection` 已切换为真实聚合结果页，接入 `/admin/cheat-detection`
- 验收结果：
  - 管理端视图中残留的 mock 数据已清理完毕
  - 不再保留“有按钮但无真实行为”的竞赛管理入口
  - 已补页面测试与 API 归一化测试

### R3 任务文档与 review 留档同步

- 状态：`completed`
- 目标：把任务文档、最新实现状态和后续待办保持一致，避免文档滞后于代码。
- 范围：
  - `code/docs/tasks/*.md`
  - 必要时补充 `code/frontend/docs/reviews/frontend/`
- 当前现状：
  - 本文档已更新为最新状态
  - 本轮“竞赛管理真实性回归”已补 review 文档：
    - [ctf-frontend-code-review-admin-contests-round1-74e5b72.md](/home/azhi/workspace/projects/ctf/code/frontend/docs/reviews/frontend/ctf-frontend-code-review-admin-contests-round1-74e5b72.md)
- 验收标准：
  - 文档状态与主线代码一致
  - 后续新增批次继续按 review 文档归档

## 已解除的阻塞项

1. 教师端接口阻塞已解除：
   - `GET /teacher/classes`
   - `GET /teacher/classes/:name/students`
   - `GET /teacher/students/:id/progress`
   - `GET /teacher/students/:id/recommendations`
2. 通知能力阻塞已解除：
   - `POST /auth/ws-ticket`
   - `GET /notifications`
   - `PUT /notifications/:id/read`
   - `GET /ws/notifications`
   - `POST /admin/notifications`

## 当前风险与注意事项

1. `ContestManage` 仍有一个显式边界：后端未提供删除竞赛接口，所以前端继续隐藏删除能力。
2. 公开竞赛 API 的历史枚举差异已经在客户端做了归一化；如果后续继续扩展竞赛域，建议抽成共享 mapper，避免多处维护。
3. 管理员通知发布当前前端只开放单一目标范围选择；后端契约已支持 union 规则模型，后续若要做混合投递，需要再补交互设计。
4. 本轮“联调缺口补齐”已单独补 review 文档，后续继续扩展时沿用相同留档方式。

## 建议执行顺序

1. 当前前端任务拆分中的联调缺口已经收口，可以直接按业务流做后续联调。
2. 若通知中心后续继续扩展，优先补“发布历史 / 批次详情 / 混合目标范围”而不是重写收件箱页。
3. 若后续继续做公开竞赛流程扩展，优先统一 `contest` 共享契约。
4. 后续新增批次继续保持 `code-reviewer -> test-engineer` 闭环。

## 交付物

- 任务状态文档：本文档
- 已归档前端 review 文档：`code/frontend/docs/reviews/frontend/`
- 前端验证命令：
  - `npm run test:run`
  - `npm run typecheck`
  - `npm run build`
