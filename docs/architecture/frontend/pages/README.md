# 页面设计索引

## 读取规则

- 本目录保存页面级最终设计稿，主要用于布局、信息架构、交互形态和视觉方向。
- 页面设计不是接口契约；字段、权限和运行语义仍以 `docs/contracts/`、`docs/architecture/` 和当前代码为准。
- 如果页面设计与当前代码不同，不要默认认定代码错或设计错；先检查 `docs/design/README.md` 和相关专题文档，再决定是更新代码还是更新设计稿。
- `docs/reviews/` 中的设计评审只记录当时发现的问题，不作为当前页面设计事实源。
- 全局设计系统位于 `../design-system/MASTER.md`、`../design-system/LIGHT-MODE.md` 和 `../design-system/TECH-STACK.md`。

## 当前页面稿

- `admin.md`：管理后台框架与题目管理基础工作区。
- `admin-images.md`：镜像管理。
- `topology-editor.md`：网络拓扑编排器。
- `env-templates.md`：环境模板管理。
- `contest-challenges.md`、`contests.md`、`contest-announcements.md`、`contest-export.md`：赛事相关页面。
- `awd-contest.md`、`awd-monitor.md`：AWD 赛事与监控页面。
- `challenges.md`、`challenge-detail-v2.md`、`instances.md`：学生挑战与实例页面。
- `teacher.md`、`teacher-dark.md`：教师端页面。
- `dashboard.md`、`dashboard-dark.md`、`skill-profile.md`、`skill-profile-dark.md`：仪表盘与能力画像。
- `auth.md`、`auth-dark.md`、`profile.md`、`profile-dark.md`、`notifications.md`、`notifications-dark.md`：通用入口与个人/通知页面。

## 拓扑相关说明

- `topology-editor.md` 是拓扑编排器当前页面设计入口。
- `env-templates.md` 只描述环境模板列表、保存为模板、导入导出和空状态。
- `docs/superpowers` 当前没有独立的 topology editor 设计稿；不要再从 AWD 总设计或历史任务拆解里反推拓扑编辑器 UI。
