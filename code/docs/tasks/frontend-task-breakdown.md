# CTF 前端任务拆分与状态

## 文档信息

- 范围：`/home/azhi/workspace/projects/ctf/code/frontend`
- 目标：记录前端任务拆分、当前完成状态、剩余待办与后续执行顺序
- 流水线：`leader -> requirements-analyst（按需） -> frontend-engineer -> code-reviewer -> test-engineer`
- 最后更新：2026-03-07

## 当前判断

- 前端主线能力已基本闭合：鉴权守卫、学员端、管理端核心页、教师端核心页、通知链路均已落地。
- 教师端相关后端接口已补齐并合入主线，不再阻塞 `TeacherDashboard` / `ClassManagement`。
- 通知能力已完成前后端闭环：
  - 后端：`POST /auth/ws-ticket`、`GET /notifications`、`PUT /notifications/:id/read`、`GET /ws/notifications`
  - 前端：通知下拉、通知中心、WebSocket 实时同步
- 当前剩余工作已不再是主流程打通，而是收尾型任务：清理仍依赖 mock 数据的页面、继续补 review 与文档闭环。

## 已完成任务

| 任务                    | 状态    | 结果                                  | 对应提交  |
| ----------------------- | ------- | ------------------------------------- | --------- |
| F0 基础对齐             | ✅ 完成 | 恢复路由守卫、校正前后端契约          | `fe30bcc` |
| F1 学员侧页面完善       | ✅ 完成 | 完成 Dashboard、Profile、个人报告导出 | `c1e0006` |
| F2 管理侧页面完善       | ✅ 完成 | 完成系统概览、审计日志、作弊研判页    | `d5d09fa` |
| F3 教师侧报告导出       | ✅ 完成 | 完成教师端报告导出                    | `f666799` |
| F4 教师侧班级与教学概览 | ✅ 完成 | 完成教学概览、班级管理                | `766f29e` |
| F6 通知实时链路         | ✅ 完成 | 完成 ws-ticket + WebSocket 通知同步   | `6fccfb8` |

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

## 剩余任务拆分

### R1 管理端用户管理页去 mock

- 状态：`completed`
- 目标：将 `UserManage` 从静态 mock 页面改成真实可接入页面，或在后端接口缺失时显式降级。
- 范围：
  - `src/views/admin/UserManage.vue`
  - 可能涉及 `src/api/admin.ts`、类型定义与测试
- 当前现状：
  - 已移除本地 `mockUsers`
  - 当前页面改为显式降级态，明确说明 `/admin/users` 接口缺失
  - 待后端补齐接口后再恢复真实列表、搜索、分页、编辑能力
- 验收标准：
  - 不保留假数据
  - 明确展示“接口未提供”的降级态
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
  - `CheatDetection` 与 `UserManage` 继续保持显式降级态，不再伪装成已接后端
- 验收结果：
  - 管理端视图中残留的 mock 数据已清理完毕
  - 不再保留“有按钮但无真实行为”的竞赛管理入口
  - 已补页面测试与 API 归一化测试

### R3 任务文档与 review 留档同步

- 状态：`in_progress`
- 目标：把任务文档、最新实现状态和后续待办保持一致，避免文档滞后于代码。
- 范围：
  - `code/docs/tasks/*.md`
  - 必要时补充 `code/frontend/docs/reviews/frontend/`
- 当前现状：
  - 本文档已更新为最新状态
  - 新增的“实时通知链路”尚无单独 review 文档
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

## 当前风险与注意事项

1. `UserManage` 的假数据已清除，但对应真实后端接口仍未提供。
2. 竞赛公共页使用的 contest 契约仍保留历史枚举（如 `registering`），后续如果继续扩展公开竞赛流程，建议统一收敛到共享映射层。
3. 前端 review 文档目前覆盖到并行交付阶段，最新“实时通知链路”“用户管理降级页”“竞赛管理真实性回归”如果继续扩展，建议补单独 review 记录。

## 建议执行顺序

1. 下一步聚焦 `R3`，把本轮新增实现与 review 留档同步。
2. 若后续继续做公开竞赛流程联调，优先统一 `contest` 共享契约。
3. 每完成一个批次，都进入 `code-reviewer -> test-engineer` 闭环。

## 交付物

- 任务状态文档：本文档
- 已归档前端 review 文档：`code/frontend/docs/reviews/frontend/`
- 前端验证命令：
  - `npm run test:run`
  - `npm run typecheck`
