---
name: ctf-ui-theme-system
description: >
  Use when building or refactoring CTF frontend pages in this repo to keep
  layout, typography, color, copy tone, and interaction patterns aligned with
  the approved academy / challenges / platform-admin workspace style. Activate
  when the task touches page shells, workspace heroes, directory/list pages,
  tables, toolbars, metric panels, modals/drawers, theme tokens or pills, copy
  tone, or admin/teacher backend-style surfaces — even when the user only says
  "做个管理页 / 优化这个列表 / 统一样式 / 这个页面太乱". For dark-theme color
  leakage specifically, use `ctf-dark-surface-alignment` instead.
---

# CTF UI Theme System

本 skill 是 CTF 前端设计语言的**导航中心**。它只做路由：判断该读哪个 reference，
然后从磁盘按需读。详细规则全部在 `references/`，不在本文件内联。

核心结果：**专业、克制、技术感、高可读的 workspace UI**（light 优先、支持 dark）。

## Positioning
- 学生流程优先：发现题目 → 起环境 → 提交 flag → 复盘题解。
- 教师/管理员页要分析化、操作化，不是企业 OA 仪表盘。
- 认知负载低：结构先于装饰。
- 气质关键词：technical / professional / restrained / reliable。禁止 anime/game-shop/neon/cyberpunk/OA 风。

## Use When
- 新建或重构 academy / challenges / platform 前端页面。
- 触达 page shell、workspace hero、目录/列表/表格、toolbar、metric 面板、modal/drawer。
- 调整主题 token、category/difficulty pill、copy 语气、管理端/教师端密集后台面。

## Do Not Use
- 暗色主题颜色泄漏对齐 → `ctf-dark-surface-alignment`。
- 通用前端工程（composable/状态/生命周期/测试）→ `frontend-engineer`。
- 后端 / API / 服务 → `ctf-backend-patterns`、`go-backend`。
- 前端项目本地硬规则（路由命名空间、SFC 块顺序、entities/features 归属、UI 测试粒度）
  → 以 `AGENTS.md` 的 CTF Frontend Local Rules 为准，本 skill 不复制。

## Reference Map
| 任务涉及 | 读这个 |
|---|---|
| shell 结构 / hero / 目录列表 / 间距 / filter bar / 行动 / 响应式降级 | `references/layout-rules.md` |
| 标题 / eyebrow / 字体栈 / 文字尺度 | `references/typography-rules.md` |
| 颜色 / surface / token / category·difficulty pill / warning·danger | `references/color-surface-tokens.md` |
| 复用列表 toolbar（`WorkspaceDirectoryToolbar`） | `references/directory-toolbar.md` |
| 是否该用 card | `references/card-usage-rules.md` |
| metric / `metric-panel-*` / 精确 selector·XPath | `references/metric-panels.md` |
| modal / side-over drawer / 滚动安全 | `references/modal-drawer.md` |
| 标识符 / 机器值 / 时间·数字·状态格式 / 去杂 | `references/data-display.md` |
| copy 语气 / 可访问性 / 收尾检查 | `references/copy-accessibility-checklist.md` |
| 后台"总览+列表"工作台结构 | `references/saas-workbench-pattern.md` |
| 管理端密集后台设计语言 | `references/admin-design-system.md` |
| 已建立的页面族（challenge/admin/env/student/teacher） | `references/page-presets.md` |
| 收尾自查清单 | `references/implementation-checklist.md` |

## Common Tasks
- 新建管理/教师列表页 → `saas-workbench-pattern.md` + `layout-rules.md` + `admin-design-system.md`。
- 优化已有列表/表格 → `layout-rules.md`（目录/长文本/操作列/响应式）+ `data-display.md`。
- 套用页面族风格 → `page-presets.md` 找对应族，再读其指向的 reference。
- 改 metric 卡 / 精确节点 → `metric-panels.md`。
- 改主题色 / pill / 去硬编码 → `color-surface-tokens.md`。
- 收尾前 → 逐项过 `implementation-checklist.md`；若改了共享 selector，补一条断言精确边界的回归测试。
- 不在列表内 → 先读本文件 Reference Map，再按主题匹配 reference。

## Known Gotchas（命中即停）
- 用 card 包 filter bar / form / 目录容器 → 信息密度下降。见 `card-usage-rules.md`。
- 用 tab 把"总览"和它描述的"列表"拆开 → 反模式。见 `saas-workbench-pattern.md`。
- 页面采用共享样式但保留同类局部覆盖 → 漂移。见 `feedback/2026-05-09-shared-ui-style-adoption-must-remove-local-overrides.md`。
- 共享 selector 改动无回归测试 → 静默回归。见 `copy-accessibility-checklist.md` Ship checklist。
- 历史事故复盘见 `incidents.md`。

## 添加新规则
当 `feedback/` 出现通过 2/3 录入标准的前端 pattern 时：归到对应 reference，加 5–10 行；
若是高代价陷阱，同时在 Known Gotchas 加一行锚点；本 SKILL.md 保持导航中心，不内联规则正文。
