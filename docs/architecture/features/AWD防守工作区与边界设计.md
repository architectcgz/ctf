# AWD 防守工作区与边界架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`runtime`、`contracts`
- 关联模块：
  - `code/backend/internal/module/challenge/domain`
  - `code/backend/internal/module/contest/application/queries`
  - `code/backend/internal/module/runtime/infrastructure`
  - `code/backend/internal/app/composition`
  - `code/frontend/src/components/contests`
  - `code/frontend/src/features/contest-awd-workspace`
- 过程追溯：旧稿 `awd-web-defense-workbench-design.md`、`awd-defense-content-page-design.md`
- 最后更新：`2026-05-07`

## 1. 背景与问题

旧稿把这个专题写成了浏览器文件工作台、受控修补片段和 SSH 入口的方案比较，但当前代码已经把学生防守入口固定下来。这里真正需要说明的，是当前产品表面和运行时边界：

- 学生页面只展示防守工作区状态摘要，不展示文件树和源码片段
- 深度防守入口是 SSH ticket，不是浏览器 IDE
- 运行时隔离依赖独立 `defense workspace`，不是直接进入比赛服务容器

## 2. 架构结论

- 学生侧当前正式入口只有 `POST /contests/:id/awd/services/:sid/defense/ssh`。
- 学生工作台通过 `GET /contests/:id/awd/workspace` 读取 `defense_connection.entry_mode / workspace_status / workspace_revision`，只获得连接状态摘要。
- 题包必须声明 `extensions.awd.runtime_config.defense_workspace`，且 `defense_workspace.entry_mode` 当前只支持 `ssh`。
- 旧的 `defense_scope.editable_paths` 已被后端明确拒绝，不再作为学生防守边界。
- SSH 登录目标是独立 `defense workspace` 容器，网关登录后工作目录固定为 `/workspace`，不是比赛服务容器根目录。
- 后端仍保留过往浏览器文件工作台的适配器和 DTO，但当前学生 HTTP 路由没有暴露 `defense/files`、`defense/directories`、`defense/commands`。

## 3. 模块边界与职责

### 3.1 模块清单

- `ContestAWDWorkspacePanel` / `useAwdWorkspaceAccessActions`
  - 负责：展示防守状态、触发 SSH 连接生成
  - 不负责：读取目录、文件内容或执行浏览器内命令

- `AWDWorkspaceQuery`
  - 负责：把 defense workspace 摘要写入 `services[].defense_connection`
  - 不负责：把工作区内部文件暴露给学生

- `awd_package_parser`
  - 负责：校验 `defense_workspace` 结构、挂载根目录和 entry mode
  - 不负责：为学生提供具体修补提示

- `awd_target_proxy_repository` / `AccessAWDDefenseSSH`
  - 负责：按赛事、轮次、实例和 workspace 状态签发 SSH 访问信息
  - 不负责：在 workspace 不可用时回退到服务容器

- `AWDDefenseSSHGateway`
  - 负责：校验 ticket、接管 SSH 会话并进入 `/workspace`
  - 不负责：提供浏览器文件编辑能力

### 3.2 事实源与所有权

- 工作区运行事实源：`awd_defense_workspaces`
- 学生工作台摘要事实源：`ContestAWDWorkspaceResp.services[].defense_connection`
- SSH 连接返回事实源：`AWDDefenseSSHAccessResp`
- 学生可见 HTTP 面事实源：`router_routes.go` 中当前已挂载的路由

## 4. 关键模型与不变量

### 4.1 核心实体

- `AWDDefenseWorkspace`
  - 关键字段：`contest_id`、`team_id`、`service_id`、`instance_id`、`workspace_revision`、`status`、`container_id`

- `AWDDefenseConnectionResp`
  - 关键字段：`entry_mode`、`workspace_status`、`workspace_revision`

- `AWDDefenseSSHAccessResp`
  - 关键字段：`host`、`port`、`username`、`password`、`command`、`expires_at`

### 4.2 不变量

- `defense_workspace.entry_mode` 当前只能是 `ssh`。
- `defense_scope.editable_paths` 已废弃，题包继续声明会被解析阶段拒绝。
- 只有在赛事 `running / frozen`、存在运行中的 round、目标实例 `running`、workspace `running` 且 `workspace_revision > 0` 时，后端才会签发 SSH 连接。
- SSH 登录后的默认工作目录固定为 `/workspace`。
- 学生当前正式接口面不暴露目录列表、文件内容和浏览器命令执行能力。

## 5. 关键链路

### 5.1 工作区摘要链路

1. 赛事工作台请求 `GET /contests/:id/awd/workspace`。
2. `AWDWorkspaceQuery` 读取服务定义和 `awd_defense_workspaces` 摘要。
3. `mergeAWDWorkspaceDefenseConnection` 把 `entry_mode`、`workspace_status`、`workspace_revision` 合并到对应服务。
4. 前端只渲染连接状态和入口按钮，不渲染文件系统内容。

### 5.2 SSH 连接链路

1. 学生在 `ContestAWDWorkspacePanel` 中点击“SSH 防守连接”。
2. 前端调用 `requestContestAWDDefenseSSH`。
3. 后端校验当前用户所属队伍、赛事状态、运行轮次、实例状态和 workspace 版本。
4. 成功后返回 `host / port / username / password / command / expires_at`。
5. 学生通过返回的 SSH 凭据连接网关。
6. `AWDDefenseSSHGateway` 鉴权通过后，在目标容器执行 `cd /workspace && exec /bin/sh`。

## 6. 接口与契约

### 6.1 学生工作台契约

- `GET /contests/:id/awd/workspace`
  - 当前服务摘要包含 `defense_connection`
  - `defense_connection` 只描述连接方式和运行状态，不包含目录或文件内容

### 6.2 SSH 访问契约

- `POST /contests/:id/awd/services/:sid/defense/ssh`
  - 返回短时连接凭据
  - 当前返回体固定包含：
    - `host`
    - `port`
    - `username`
    - `password`
    - `command`
    - `expires_at`

### 6.3 未暴露的旧能力

以下 DTO 或适配器仍在后端代码中保留，但当前学生产品表面未开放：

- `AWDDefenseFileResp`
- `AWDDefenseDirectoryResp`
- `AWDDefenseCommandResp`
- 任何 `defense/files`、`defense/directories`、`defense/commands` 路由

## 7. 兼容与迁移

- 旧的浏览器文件工作台和独立防守内容页已经不再是当前事实源。
- 当前学生深度防守入口已经收敛为 SSH，不再沿用 `editable_paths` 驱动的受控文件边界。
- 后端保留的旧 workbench 代码属于残留适配层，不应被当成学生当前已开放能力。
- 如果后续重新引入浏览器内防守编辑器，必须新开页面或契约专题，不能在本文里把残留代码写成已上线事实。

## 8. 代码落点

- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- `code/backend/internal/model/awd_defense_workspace.go`
- `code/backend/internal/dto/contest_awd_workspace.go`
- `code/backend/internal/dto/instance.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_query.go`
- `code/backend/internal/module/contest/application/queries/awd_workspace_result.go`
- `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
- `code/backend/internal/app/composition/runtime_adapter_compat.go`
- `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`

## 9. 验证标准

- 学生工作台能读取 `defense_connection.entry_mode / workspace_status / workspace_revision`。
- 学生页面只能生成 SSH 连接，不会请求目录、文件或浏览器命令接口。
- `defense_scope.editable_paths` 继续被后端解析器拒绝。
- SSH 鉴权成功后，工作目录落在 `/workspace`。
- 当前路由表中不存在学生可用的 `defense/files`、`defense/directories`、`defense/commands` 入口。
