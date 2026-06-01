# 前端按钮原语迁移审计

更新日期：2026-05-10

## 背景

本清单记录 `code/frontend/src` 内仍需要迁移或收口的按钮样式。触发原因是学生 dashboard 与错误页先后出现按钮夜间模式边框、hover 强度和局部按钮体系漂移问题。

本轮只做全仓静态扫描和迁移分组，不做批量替换。后续迁移应按页面和共享原语分批完成，每批都补对应测试。

## 扫描口径

扫描范围：

- `code/frontend/src/**/*.vue`
- `code/frontend/src/**/*.ts`
- `code/frontend/src/**/*.css`

扫描重点：

- `<button>`、`<RouterLink>`、`<a>` 上的按钮类
- 类名包含 `btn`、`button`、`action` 的模板使用和样式定义
- 已有共享按钮原语：`ui-btn`、`journal-btn-*`、`header-btn`、`workspace-directory-row-btn`、`c-action-menu__*`

本轮静态结果：

- 前端 `.vue / .ts` 文件：1035 个
- 按钮相关类名：291 个
- `ui-btn` 相关使用：619 次
- 非共享按钮形态元素候选：153 处，分布在 72 个文件

这些数字是迁移审计入口，不等价于最终缺陷数量。Tabs、菜单、关闭按钮、分页、画布工具等特殊控件需要按交互语义单独判断。

## 已确认的共享原语

### `ui-btn`

通用工作区按钮原语。普通 primary、secondary、ghost、danger、link 动作默认使用它。

迁移规则：

- 页面只组合 `ui-btn ui-btn--*`。
- 页面可通过 `--ui-btn-*` 变量调整尺寸、边框、hover 强度。
- 不在页面 scoped style 中重写完整 `border / background / color / hover / focus` 状态。

### `journal-btn-*`

学生 soft journal 页面按钮原语。当前允许：

- `journal-btn-primary`
- `journal-btn-secondary`
- `journal-btn-outline`

迁移规则：

- 学生 dashboard / journal surface 内优先使用这些语义类。
- light / dark 下都必须保留可见边框。
- dark hover 只做柔和增强，不直接打满 accent 边框。

### `header-btn`

工作区 hero/header 区域按钮原语，当前在平台端和教师端复用较多。短期保留，不强行折叠到 `ui-btn`。

后续要求：

- `header-btn` 的状态实现继续集中在全局样式。
- 页面只使用 `header-btn--primary / header-btn--ghost / header-btn--compact`。
- 若后续发现它与 `ui-btn` 的变量模型可以统一，应先设计兼容方案，再迁移调用点。

### `workspace-directory-row-btn`

目录行 CTA 原语。短期保留，用于列表行内 pill 形入口。

后续要求：

- 行内普通动作优先使用 `ui-btn ui-btn--xs`。
- 需要整行 pill 视觉或目录行 CTA 时才使用 `workspace-directory-row-btn`。
- danger 状态必须保留明确边框和柔和 hover，不在页面内重写。

### `c-action-menu__*`

菜单按钮和菜单项原语。属于菜单交互，不纳入普通按钮迁移。

## 需要迁移或收口的候选组

### P0：普通动作按钮应直接迁移到 `ui-btn`

这些按钮承担普通确认、取消、恢复、查看、跳转、重试等动作，不应在局部组件里自建按钮体系。

- `code/frontend/src/components/common/DeleteConfirmModal.vue`
  - 当前：`delete-confirm-modal__action`
  - 建议：底部动作改为 `ui-btn ui-btn--secondary` / `ui-btn ui-btn--danger`

- `code/frontend/src/components/common/modal-templates/CImmersiveConfirmDialog.vue`
  - 当前：`c-immersive-confirm__button`
  - 建议：改为 `ui-btn`，沉浸式弹窗只通过变量调整尺寸和 danger 语义

- `code/frontend/src/components/common/modal-templates/CLightActionPopover.vue`
  - 当前：`c-light-action-popover__action`
  - 建议：确认普通 action 是否属于菜单项；若是按钮动作，迁移到 `ui-btn`；若是菜单项，改归 `c-action-menu` 体系

- `code/frontend/src/views/dashboard/DashboardView.vue`
  - 当前：`workspace-alert-action`
  - 建议：迁移到 `ui-btn ui-btn--secondary` 或 `ui-btn ui-btn--ghost`

- `code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`
  - 当前：`import-review__ghost`、`import-review__primary`
  - 建议：迁移到 `ui-btn ui-btn--ghost` / `ui-btn ui-btn--primary`

- `code/frontend/src/components/teacher/student-insight/StudentInsightWriteupsSection.vue`
  - 当前：`writeup-action-button`
  - 建议：迁移到 `ui-btn ui-btn--secondary`；warning 状态用 `ui-btn--danger` 或新增共享 warning 变体，不在组件内复制按钮状态

- `code/frontend/src/entities/challenge/ui/ChallengeProfileMetaGrid.vue`
  - 当前：`challenge-link challenge-link-button`
  - 建议：若视觉是按钮，改用 `ui-btn--link` 或 `ui-btn--secondary`；若是文本链接，去掉 button 形态类

### P1：重复出现的局部按钮体系应抽成共享原语或归并到 `ui-btn`

这些不是一两个按钮，而是多个组件重复定义一套状态。迁移前应先确定它们是否真的需要专用原语。

- `ops-btn`
  - 使用文件包括：
    - `code/frontend/src/components/platform/contest/AWDAttackLogPanel.vue`
    - `code/frontend/src/components/platform/contest/AWDInstanceOrchestrationPanel.vue`
    - `code/frontend/src/components/platform/contest/AWDOperationsPanel.vue`
    - `code/frontend/src/components/platform/contest/AWDRoundHeaderPanel.vue`
    - `code/frontend/src/components/platform/contest/AWDRoundInspector.vue`
    - `code/frontend/src/components/platform/contest/AWDServiceStatusPanel.vue`
    - `code/frontend/src/components/platform/contest/projector/ContestProjectorToolbar.vue`
  - 当前问题：多个文件各自定义 `.ops-btn` 的尺寸、边框、hover、disabled。
  - 建议：优先迁移到 `ui-btn ui-btn--sm`。若 AWD 操作区确实需要专用密度，新增全局 `ops-btn` 原语并让它复用 `--ui-btn-*` 变量模型。

- `topology-*btn` / Tailwind inline button classes
  - 主要文件：
    - `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
    - `code/frontend/src/components/platform/topology/TopologyConnectivitySections.vue`
    - `code/frontend/src/components/platform/topology/TopologyNetworkSection.vue`
    - `code/frontend/src/components/platform/topology/TopologyTemplateSidePanel.vue`
  - 当前问题：拓扑编辑器内存在 `topology-toolbar-btn`、`topology-action-btn`、`template-action-btn` 以及大量 `rounded-xl border px-3 py-2 ...` 内联类。
  - 建议：拓扑编辑器先设计一个局部但共享的 topology action primitive，使用主题变量；不要继续在各 section 中复制 Tailwind 按钮组合。

- `asset-btn` / `defense-ops__button` / `asset-ssh__copy`
  - 主要文件：
    - `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
    - `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
    - `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
  - 当前问题：学生 AWD 侧资产动作按钮自成一套。
  - 建议：优先复用 `ui-btn` 或学生侧比赛专用变量；如需要紧凑资产按钮，抽成同一组共享类。

### P2：页面适配类保留，但内部按钮状态应回到共享原语

这些类可能只是布局或容器命名，不一定要删除；重点是它们不应再负责完整按钮状态。

- `overview-action-main`、`overview-anchor-btn`
  - 文件：`code/frontend/src/components/platform/dashboard/PlatformOverviewPage.vue`
  - 当前已组合 `header-btn`，后续仅保留布局适配。

- `user-action-btn`
  - 文件：`code/frontend/src/components/platform/user/UserGovernancePage.vue`
  - 需要确认是否已组合 `ui-btn`；若只是尺寸适配可保留，若重写状态需迁移。

- `contest-form-button`
  - 文件：`code/frontend/src/components/platform/contest/PlatformContestFormPanel.vue`
  - 建议确认是否只是测试辅助类；真实按钮应使用 `ui-btn`。

- `instance-link-btn`
  - 文件：`code/frontend/src/views/instances/InstanceList.vue`
  - 当前与 `ui-btn--link` 相关，保留前需确认不要覆盖主题状态。

- `auth-submit-btn`
  - 文件：`code/frontend/src/views/auth/LoginView.vue`
  - 当前已有 `ui-btn--block` 使用迹象；只允许作为登录页尺寸适配类。

### P3：暂不迁移的特殊控件

这些控件不是普通按钮，迁移时应按具体交互设计，不直接套 `ui-btn`。

- 顶栏图标按钮：`topnav-icon-button`、`topnav-notification-button`
- 关闭按钮：`close-btn`、`modal-template-*__close`、`app-toast-close`
- tab / chip / segmented controls：`tab-btn`、`top-tab`、`sub-tab`、`teacher-directory-chip`
- 分页按钮：`page-pagination-controls__button`
- 菜单触发器和菜单项：`c-action-menu__*`
- 画布、投影、拓扑编辑器中的非普通动作控件
- `UILab.vue`、`ThemePreview.vue` 这类预览/实验页面，除非它们会进入正式路由

## 推荐迁移顺序

1. P0 普通动作按钮先迁移到 `ui-btn`，每个页面或组件一批。
2. P1 里先处理 `ops-btn`，因为它重复定义最多，且 AWD 操作页容易继续复制。
3. 拓扑编辑器单独做一批，先设计 topology action primitive，再迁移散落的 inline button classes。
4. 学生 AWD 资产按钮单独做一批，避免和 backoffice `ui-btn` 密度混在一起。
5. P2 逐项确认适配类只负责布局变量，不负责状态。

## 每批迁移的完成标准

1. 普通按钮必须组合已有共享原语，或新增一个明确归属的共享原语。
2. light / dark 下边框都可见，hover 边框柔和，focus ring 清晰。
3. 页面 scoped style 不再复制完整按钮状态机。
4. 对应测试覆盖类名采用、dark mode token 或视觉契约。
5. 若新增共享原语，更新 `docs/architecture/frontend/06-components.md`。

## 后续验证建议

- 静态：新增按钮类时禁止 `xxx-action-primary / xxx-action-secondary` 这类页面私有按钮体系继续扩散。
- 单测：为迁移批次补充类名采用和 dark token 断言。
- 视觉：重点检查 dark mode hover、focus、disabled 和 danger 状态。
