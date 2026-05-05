# AWD 防守内容页设计

## 背景

当前学生侧 AWD 赛事已经有两块现成能力，但它们没有连成正确的用户路径：

- [ContestAWDWorkspacePanel.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue) 里已经有“我的防守 / Web 防守 / SSH / 重启”这些防守动作。
- [AWDDefenseFileWorkbench.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue) 已经具备左侧目录点击、右侧文件内容展示的基础结构。
- [ContestAWDDefenseWorkbench.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/views/contests/ContestAWDDefenseWorkbench.vue) 也已经有单独页面壳，但现在只是“入口已迁移”的占位。
- 后端 runtime adapter 已经实现 `ReadAWDDefenseFile` / `ListAWDDefenseDirectory`，但 HTTP handler 仍统一返回 `403 Forbidden`，所以前端页面即使存在也无法真正读内容。

结果就是：学生在战场页能看到“防守”语义，但点进去没有真正的内容页；文件工作台能力散落在仓库里，却没有形成可用入口。

## Gemini CLI 调用记录

本轮按要求在 worktree 内调用了 Gemini CLI 做页面设计探测，命令包括：

- `gemini -m gemini-2.5-flash -p "<防守内容页设计提示词>"`
- `timeout 20 gemini --skip-trust -y -m gemini-2.5-flash -o text -p "用一句话回复：测试成功。"`

结果：

- CLI 能启动，能输出本地启动提示。
- 模型正文在当前环境中长时间无响应。
- `timeout 20` 到期后仍未返回任何模型内容。
- 复测 TTY 模式同样没有正文返回。

因此本次没有取得可复用的 Gemini 设计文本。后续实现仍以本仓库现有 AWD 页面结构、已存在组件和本设计文档为准，不把外部模型无响应当成流程阻塞点。

## 目标

- 学生在 AWD 战场页点击某个服务的“防守”按钮后，进入该服务对应的独立防守内容页。
- 页面左侧显示当前服务允许浏览的防守文件列表与目录切换。
- 页面右侧显示被点击文件的内容，第一阶段只读。
- 页面保留返回战场入口，并保持当前 contest/service 上下文可直达、可刷新、可重载。
- 后端只开启受控只读文件浏览，不恢复保存文件和执行命令。

## 非目标

- 不做在线编辑器、命令执行器、浏览器 shell。
- 不开放任意文件写入、目录上传、新建文件或删除文件。
- 不在本轮恢复 `PUT /defense/files` 与 `POST /defense/commands`。
- 不把完整防守流程重新塞回 `ContestAWDWorkspacePanel.vue`。

## 为什么改成独立页

旧的 [awd-web-defense-workbench-design.md](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/docs/architecture/features/awd-web-defense-workbench-design.md) 选择把学生默认路径留在战场页里，主要是为了避免直接暴露完整源码浏览器。

这次需求已经明确变化：用户需要“点击防守按钮后，进入一个左侧选文件、右侧看内容”的独立内容页。因此本轮不再坚持“防守内容只能留在战场页”的约束，而是把风险收敛到更具体的两个边界上：

- 路由独立，但能力只读。
- 页面可浏览，但只走受控根目录和受控过滤规则。

也就是说，本轮改变的是入口组织，不是把浏览器防守能力无限放开。

## 用户路径

### 入口

学生仍从：

```text
/contests/:id?panel=challenges
```

进入 AWD 战场。

在“我的防守”服务卡上新增明确的“防守”按钮。点击后跳转到：

```text
/contests/:id/awd/defense/:serviceId
```

其中：

- `id` 表示 contest
- `serviceId` 表示当前被防守的服务

### 页面行为

进入页面后：

1. 先加载 contest 的 AWD workspace，用于确认当前用户确实属于该 contest 的当前队伍，并且 `serviceId` 属于自己的服务。
2. 自动请求根目录 `.` 的防守目录列表。
3. 左侧展示目录路径、目录项列表和服务摘要。
4. 点击目录项中的目录时，左侧切换到该目录。
5. 点击目录项中的文件时，右侧显示文件内容。
6. 顶部提供“返回战场”和“刷新当前目录”。

## 路由与页面边界

新增学生侧路由：

```text
name: ContestAWDDefenseWorkbench
path: /contests/:id/awd/defense/:serviceId
component: ContestAWDDefenseWorkbench.vue
```

页面 owner 保持在 `ContestAWDDefenseWorkbench.vue` 及其页面级 composable，不把目录加载、文件加载和路由同步下沉到展示组件里。

边界约定：

- `ContestAWDDefenseWorkbench.vue`
  - 持有 route params、页面级 loading/error、服务校验、目录状态、文件状态。
  - 负责 stale response 防护和重复点击防护。
  - 返回入口固定回到 `/contests/:id?panel=challenges`。
- `AWDDefenseFileWorkbench.vue`
  - 只负责展示左侧列表和右侧内容，不直接发请求。
  - 通过 props + emits 暴露目录点击、文件点击、刷新动作。
  - 不直接 import `@/api/contest` 或 `vue-router`。
- `ContestAWDWorkspacePanel.vue`
  - 只保留战场页内的防守入口，不再承担文件内容页本体。

## 信息架构

## 页面顶栏

- 返回战场
- 当前服务标题
- 当前目录路径
- 刷新按钮

## 左栏

- 服务状态摘要
  - 服务标题
  - `服务 #id`
  - 当前状态
- 目录区
  - 当前目录 breadcrumb
  - 目录项列表
  - 目录优先、文件次之

## 右栏

- 选中文件标题
- 文件大小
- 只读标记
- 文件内容区

## 状态设计

### 页面级

- `loading`
  - 首次进入时显示页面骨架或加载态
- `not_found`
  - `serviceId` 不属于当前用户的 AWD 服务
- `forbidden`
  - 后端未开启只读防守工作台
- `empty`
  - 目录为空
- `ready`
  - 左侧目录和右侧文件工作正常

### 交互级

- 切目录时立即清空右侧旧文件内容，避免旧内容残留在新目录语境里。
- 切文件时保留左侧目录，但右侧进入文件加载态。
- 连续快速点击目录或文件时，只允许最后一次响应生效。

## 后端策略

本轮只恢复两个只读接口：

```text
GET /api/v1/contests/:id/awd/services/:sid/defense/directories
GET /api/v1/contests/:id/awd/services/:sid/defense/files
```

保持继续禁用：

```text
PUT /api/v1/contests/:id/awd/services/:sid/defense/files
POST /api/v1/contests/:id/awd/services/:sid/defense/commands
```

处理原则：

- handler 不再无条件 `403`，而是调用现有 runtime service。
- runtime service 继续复用 `FindAWDDefenseSSHScope` 做 contest/team/service/container 边界校验。
- 目录和文件路径继续走 `normalizeAWDDefensePath` / `normalizeAWDDefenseDirectoryPath`。
- 只允许访问 `container.defense_workbench_root` 下的内容。

### 目录过滤

虽然本轮是独立页，但仍不能把敏感文件直接暴露给学生。目录返回前增加基础过滤规则，至少排除：

- `.env`
- `.env.*`
- `.ssh`
- `id_rsa`
- `id_ed25519`
- `authorized_keys`
- `known_hosts`

这层过滤应该落在后端目录结果整理处，而不是只靠前端隐藏。

## 前端实现方向

### 战场页入口调整

[AWDDefenseServiceList.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue) 当前只有选中、打开、SSH、重启按钮。本轮新增“防守”按钮，点击后直接 `router.push` 到独立页。

### 独立页复用现有组件

[AWDDefenseFileWorkbench.vue](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue) 不推倒重写，而是增强为真正的双栏内容页组件：

- 左侧固定为目录与文件列表
- 右侧固定为文件内容
- 保留刷新动作
- 增加空态、加载态、未选中文件态

### 新增页面级 composable

建议新增页面级 composable，例如：

```text
useContestAwdDefenseWorkbenchPage.ts
```

它负责：

- route param 解析
- backLink 生成
- workspace 拉取
- 目录请求与文件请求
- stale request 序号保护
- 页面错误态聚合

## 数据接口

前端补齐并使用以下 API client：

```text
requestContestAWDDefenseDirectory(contestId, serviceId, path)
requestContestAWDDefenseFile(contestId, serviceId, path)
```

其中 directory/file DTO 继续复用现有 `AWDDefenseDirectoryData` / `AWDDefenseFileData`。

## 响应式布局

- Desktop：左 22rem 文件栏，右侧内容自适应。
- Tablet：左栏收窄到 18rem，右栏维持主内容。
- Mobile：上下堆叠，文件列表在上，内容区在下。

移动端不隐藏功能，只改变布局。

## 验证要求

前端：

- 路由 source test 覆盖独立页路由和防守按钮文案
- 页面组件测试覆盖：
  - 首次加载根目录
  - 点击目录切换
  - 点击文件展示内容
  - 旧请求晚到不覆盖新状态
  - 403 时显示“未开启防守内容页”

后端：

- handler test 从“始终 forbidden”改为“read/list 成功，save/command 继续 forbidden”
- runtime adapter test 补目录过滤断言

## 文档关系

本设计不是否定 [awd-web-defense-workbench-design.md](/home/azhi/workspace/projects/ctf/.worktrees/feat/awd-defense-content-page/docs/architecture/features/awd-web-defense-workbench-design.md) 的全部内容，而是覆盖其中两项约束：

- 不再坚持“学生防守只留在战场页内，不新增独立路由”
- 不再坚持“学生页完全不挂浏览器文件工作台”

仍然保留它的核心安全原则：

- 不开放写文件
- 不开放执行命令
- 不暴露任意路径
- 不脱离队伍/服务边界
