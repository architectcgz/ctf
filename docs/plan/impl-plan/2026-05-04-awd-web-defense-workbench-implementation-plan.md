# AWD Web 防守工作台实现计划

## 目标

在学生 AWD 攻防战场内实现默认 Web 防守工作台，替代学生侧完整文件浏览入口。第一阶段不暴露完整 `app.py`、目录树或任意文件 API，只提供服务风险、修复动作、验证动作和高级 SSH 入口。

## 输入文档

- `docs/architecture/features/awd-student-battle-workspace-design.md`
- `docs/architecture/features/awd-web-defense-workbench-design.md`
- `docs/plan/impl-plan/2026-05-04-awd-defense-workbench-design-implementation-plan.md`

## 非目标

- 不启用通用 `files/directories/commands` 学生 UI。
- 不新增完整源码查看、完整目录树、任意文件保存或浏览器 shell。
- 不做后端片段策略 API；该能力进入二期。
- 不改变 AWD 计分、checker、实例重启和 SSH ticket 语义。

## 当前风险

- 当前分支仍包含 `AWDDefenseFileWorkbench` 和 `requestContestAWDDefenseDirectory/File` 的学生页挂载路径，会暴露完整文件工作台。
- `ContestAWDWorkspacePanel.vue` 已经承载大量模板和状态，新增 UI 必须控制在小组件内。
- SSH 仍可查看容器内文件；本阶段只降低 Web UI/API 默认暴露面，不声明完全阻止源码访问。

## Task 1: 移除学生侧完整文件工作台挂载

**目标：** 学生 AWD 战场不再调用目录/文件读取接口，不展示完整目录树或完整源码。

**文件：**

- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

**步骤：**

- 删除 `AWDDefenseFileWorkbench` import 和模板挂载。
- 删除 `requestContestAWDDefenseDirectory`、`requestContestAWDDefenseFile` import 和相关 state/watch/handler。
- 更新 source test，断言学生战场不包含文件工作台和文件 API。

**验证：**

```bash
cd code/frontend
npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
```

## Task 2: 增加 Web 防守操作面板

**目标：** 选中防守服务后展示非源码型防守工作台。

**文件：**

- 新增 `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- 修改 `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- 修改 `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`

**组件契约：**

`AWDDefenseOperationsPanel` 接收：

- `service`
- `loading`
- `access`
- `copiedCommand`
- `copiedConfig`

发出：

- `open-service`
- `request-ssh`
- `restart-service`
- `refresh`
- `copy-command`
- `copy-config`

**UI 内容：**

- 风险状态：服务状态、实例状态、风险标签。
- 攻击线索：`attack_received` 派生风险原因。
- 代码片段区：显示受控片段空态，不显示文件名、路径或源码。
- 验证区：打开服务、刷新、重启。
- 高级连接：复用 `AWDDefenseConnectionPanel`。

**验证：**

```bash
cd code/frontend
npm run test:run -- src/components/contests/awd/__tests__/AWDDefenseConnectionPanel.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
```

## Task 3: 页面布局和入口文案调整

**目标：** 学生能明确从攻防战场进入防守位置，SSH 不再显得是唯一防守页。

**文件：**

- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- `code/frontend/src/features/contest-detail/model/useContestDetailRoutePage.ts`
- `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`

**步骤：**

- AWD 页签文案使用“攻防战场”。
- 防守区标题使用“我的防守”。
- 服务列表标题使用“防守服务”。
- 维持攻击区和情报区同页展示。
- 修复 SSH 展开内容撑高服务卡的问题，面板内部滚动。

**验证：**

```bash
cd code/frontend
npm run test:run -- src/views/contests/__tests__/ContestDetail.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/components/contests/awd/__tests__/AWDDefenseConnectionPanel.test.ts
```

## Task 4: 全量前端验证和配置检查

**验证：**

```bash
cd code/frontend
npm run typecheck
npm run check:theme-tail
```

如涉及后端配置默认值，再运行：

```bash
cd code/backend
go test ./internal/config -run 'Validate|Config'
```

## Review Gate

本任务属于非平凡前端功能和安全边界调整。实现后必须进行独立 review，重点检查：

- 学生页是否仍有完整文件工作台挂载。
- 是否仍调用目录/文件读取 API。
- UI 是否把 SSH 作为高级入口而非唯一防守路径。
- 防守操作面板是否没有显示 `app.py`、真实路径、内部实现说明。
- 异步动作是否由当前页面 owner 捕获，按钮是否防重复。

Review 记录归档到：

```text
docs/reviews/frontend/2026-05-04-awd-web-defense-workbench-review.md
```
