# AWD Defense Workbench Design And Implementation Plan

> Status: Superseded for student default UI by `docs/architecture/features/awd-web-defense-workbench-design.md` and `docs/plan/impl-plan/2026-05-04-awd-web-defense-workbench-implementation-plan.md`.
> The browser file workbench described here must not be mounted on the student AWD battlefield by default. Any future code-fragment or patch workflow must use service-side allowlisted fragments instead of generic directory/file browsing.

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将学生端 AWD 战场的“防守服务卡片”升级为面向真实攻防防守流程的防守工作台，让学生能快速连接服务、判断风险、定位攻击痕迹、完成修复并验证服务状态。

**Architecture:** 前端以 `contest-awd-workspace` feature 为页面状态 owner，拆出防守服务、连接、情报和攻击提交的展示组件，避免继续扩大 `ContestAWDWorkspacePanel.vue`。后端第一阶段复用已有 AWD workspace、scoreboard、SSH ticket 和 admin/teacher 侧已有流量读模型；第二阶段再按受控权限启用已存在但当前 forbidden 的 defense files/directories/commands 路由。

**Tech Stack:** Go 1.24, Gin, GORM, PostgreSQL, Docker runtime, Vue 3, TypeScript, Vue Router, Pinia, Vitest.

---

## Plan Summary

### Objective

- 让 `/contests/:id?panel=challenges` 的 AWD 面板更贴近真实防守：服务状态、连接入口、攻击痕迹、流量线索、修复动作和验证结果放在同一个决策路径里。
- 修正 VS Code Remote-SSH 的使用方式：页面主入口复制 `ssh user@host -p port` 命令；同时提供 OpenSSH `Host` alias 配置，避免把 `Host ...` 配置块当命令粘进 VS Code。
- 为学生防守方提供“当前应先处理哪个服务”的判断依据：SLA、checker、攻击次数、最近战报、异常流量摘要、服务重启状态。
- 为后续浏览器内轻量防守工作台预留边界：目录浏览、文件读取/保存、命令执行必须受限、审计、可回滚，不在第一阶段直接开放高风险能力。

### Non-goals

- 不做完整 Web IDE、在线 VS Code、终端模拟器或持久 shell 会话。
- 不做自动漏洞修复、自动上传补丁、自动生成最终防守结论。
- 不在第一阶段开放任意命令执行或任意文件写入。
- 不把管理员 AWD 交通分析页完整搬给学生；学生侧只展示与本队服务相关、可操作的精简线索。
- 不改变 AWD checker、轮次推进、计分和 SLA 规则。

### External Practice References

- AWD/Attack-Defense 常见流程：按回合维护服务可用性、攻击对手、修复自身服务并防止 flag 泄露。
- 防守动作通常包含：登录靶机、源码备份、弱口令修改、代码审计、漏洞修复、服务恢复。
- Web 防守常见线索：异常请求路径、静态资源 POST、可疑 webshell 文件、web server 进程拉起 shell/curl/wget、近期文件变更。
- 平台设计取舍：界面应把“连接、定位、修复、验证”排成连续工作流，而不是只给一个 SSH 字符串。

### Local Evidence

- 学生 AWD 面板：`code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- 学生 AWD 状态 owner：`code/frontend/src/features/contest-awd-workspace/model/useContestAWDWorkspace.ts`
- SSH ticket action：`code/frontend/src/features/contest-awd-workspace/model/useAwdWorkspaceAccessActions.ts`
- SSH presentation helper：`code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.ts`
- 后端 SSH ticket：`code/backend/internal/app/composition/runtime_module.go`
- 后端 SSH route：`code/backend/internal/module/runtime/api/http/handler.go`
- 已存在但禁用的防守文件/命令接口 DTO：`code/backend/internal/dto/instance.go`
- 已存在但当前返回 forbidden 的路由：
  - `GET /api/v1/contests/:id/awd/services/:sid/defense/files`
  - `GET /api/v1/contests/:id/awd/services/:sid/defense/directories`
  - `PUT /api/v1/contests/:id/awd/services/:sid/defense/files`
  - `POST /api/v1/contests/:id/awd/services/:sid/defense/commands`
- AWD traffic 管理侧能力：`code/frontend/src/api/admin/contests.ts`、`code/backend/internal/module/contest/application/queries/awd_traffic_*`
- 教师 AWD 复盘侧已有流量/攻击读模型：`code/frontend/src/api/teacher/awd-reviews.ts`、`code/backend/internal/module/assessment/application/queries/teacher_awd_review_service.go`

### Architecture Evaluation

- 当前问题不只是“按钮文案不清楚”，而是学生防守路径缺少一个稳定的信息架构：学生需要先判断服务风险，再连接，再修复，再验证，而不是在三个列里来回找信息。
- `ContestAWDWorkspacePanel.vue` 已同时承担状态派生、服务列表、攻击目标、战报、SSH 展示和样式，继续直接堆新 UI 会扩大回归风险。第一阶段必须先拆组件和 presentation helpers。
- 后端已有 defense file/command DTO 和路由，但 handler 当前返回 forbidden，说明项目曾经预留浏览器防守工作台能力。直接开启这些能力风险过高，必须先做 allowlist、审计、大小限制、备份和测试。
- VS Code Remote-SSH 对带 `+` 的用户名有解析坑：直接 `vscode-remote://ssh-remote+student+8+21@127.0.0.1:2222` 会被解析成 host `student`。页面应优先给 `ssh user@host -p port` 命令和 `Host alias` 配置，不建议生成这类 URI。

### Design Direction

#### Information Architecture

- 顶部 HUD 保留，但语义收敛为“本轮态势”：
  - 当前回合
  - 本队排名
  - 服务健康数
  - 最近同步时间
  - 手动刷新
- 左侧改为“防守服务”主工作区：
  - 服务列表按风险排序：`compromised`、`down`、`attack_received > 0`、`pending/creating`、`up`
  - 每个服务卡展示：题目名、服务编号、checker 状态、实例状态、攻击次数、最近事件时间
  - 主动作：打开服务、生成 SSH、重启
  - 连接生成后展示：VS Code 命令、密码、OpenSSH 配置折叠区、过期时间
- 中间改为“目标攻击”或“攻击向量”，保留当前攻击提交流程，但降低视觉权重，让防守页优先服务防守。
- 右侧改为“情报与痕迹”：
  - 最近战报
  - 本队相关异常流量摘要
  - 可疑路径 Top 5
  - 最近失败 checker / 5xx / 4xx 线索
- 后续二阶段可在服务卡展开“修复面板”：
  - 目录浏览
  - 文件查看/编辑
  - 保存时自动备份
  - 受控命令执行
  - 最近备份/恢复

#### Interaction Flow

1. 学生进入 AWD 面板。
2. 页面按风险排序服务，默认选中最高风险服务。
3. 学生点击 `SSH` 生成临时 ticket。
4. 页面主按钮复制 VS Code 可用命令，例如 `ssh student+8+21@127.0.0.1 -p 2222`。
5. 页面展示密码和过期时间；OpenSSH 配置放进折叠区，按钮文案为 `复制配置`。
6. 学生修复后点击 `重启` 或 `刷新`。
7. 页面显示 checker/SLA/最近流量是否恢复。

#### Copy Rules

- 使用终端用户能直接行动的短文案：
  - `复制 VS Code 命令`
  - `OpenSSH 配置`
  - `票据将在 14:05 过期`
  - `最近 5 分钟出现 3 次 5xx`
- 不在 UI 中出现设计解释：
  - 不写“这里用于帮助你快速判断防守态势”
  - 不写“本模块支持...”
  - 不写“建议先...再...”

#### Responsive Behavior

- Desktop：`防守服务` 宽列 + `目标攻击` 主列 + `情报` 右列。
- Medium：`防守服务` 和 `情报` 上下堆叠，攻击区独立一行。
- Mobile：先显示服务风险与连接动作，再显示攻击目标和战报；按钮保持不低于 36px 触控高度。

#### Accessibility

- 服务卡动作组使用 `role="group"` 和 `aria-label`。
- 复制按钮必须有文本，不只显示图标。
- 折叠区使用原生 `details/summary` 或明确 `aria-expanded`。
- 刷新、SSH、重启等异步按钮必须在 handler 内防重复。
- 错误状态靠近触发动作显示 toast 或 inline message，不能只 `console.error`。

## File Structure

### Frontend

- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - 保留 route-level panel composition，移出服务卡、情报列、攻击区等大块模板。
- Create: `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
  - 风险排序、状态 label、服务卡 VM、连接过期 label、风险摘要。
- Modify: `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.ts`
  - 保持 VS Code command 和 OpenSSH config 的分离；增加 `buildVSCodeRemoteSSHNotes` 之类的非 UI helper 只供测试/文档使用时需谨慎。
- Create: `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.ts`
  - 默认选中高风险服务、处理用户手动选择、服务列表刷新后保持选择。
- Create: `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
  - 防守服务列表和服务卡。
- Create: `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
  - SSH command、密码、OpenSSH config、复制状态和过期时间。
- Create: `code/frontend/src/components/contests/awd/AWDDefenseIntelPanel.vue`
  - 最近战报、本队相关流量摘要、可疑路径。
- Create: `code/frontend/src/components/contests/awd/AWDAttackTargetPanel.vue`
  - 从当前中间攻击区抽离。
- Modify: `code/frontend/src/features/contest-awd-workspace/index.ts`
  - 统一导出新增 presentation/composable，避免组件穿透内部路径。
- Test: `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts`
- Test: `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.test.ts`
- Test: `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.test.ts`
- Test: `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- Test: optional component tests under `code/frontend/src/components/contests/awd/__tests__/`

### Backend Phase 1

- Modify: `code/backend/internal/dto/instance.go`
  - 如有必要扩展 `AWDDefenseSSHAccessResp`：增加 `expires_in_seconds` 或 `usage_hint_key` 时必须保持向后兼容。
- Modify: `code/backend/internal/app/composition/runtime_module.go`
  - 保持 SSH command 和 profile alias 的后端生成逻辑稳定。
- Test: `code/backend/internal/module/runtime/api/http/handler_test.go`
  - 断言返回 `command`、`ssh_profile.alias`、`host_name`、`port`、`user`。
- Test: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - 断言 `parseAWDDefenseSSHUsername` 支持普通用户名 + `contestID` + `serviceID`，并记录 `+` 用户名约束。

### Backend Phase 2

- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
  - 将 defense files/directories/commands 从 forbidden 改为受控调用。
- Modify: `code/backend/internal/app/composition/runtime_module.go`
  - 复用已存在的 `ReadAWDDefenseFile`、`ListAWDDefenseDirectory`、`SaveAWDDefenseFile`、`RunAWDDefenseCommand`。
- Modify/Create: runtime authorization or policy helper near `code/backend/internal/module/runtime/...`
  - 限制路径、文件大小、命令 allowlist、输出大小、审计字段。
- Test: backend focused tests for:
  - 非本队成员 forbidden
  - 路径逃逸 forbidden
  - 保存文件前生成 backup
  - 命令超长 rejected
  - 命令输出截断

## Task 1: Stabilize SSH / VS Code Connection UX

**Files:**
- Modify: `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.ts`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Test: `code/frontend/src/features/contest-awd-workspace/model/sshAccessPresentation.test.ts`
- Test: `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

- [x] **Step 1: Add failing test for VS Code command separation**

Run: `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model/sshAccessPresentation.test.ts`

Expected: test asserts VS Code command starts with `ssh ` and not `Host `.

- [x] **Step 2: Ensure UI labels separate command and config**

Implementation:

```vue
<button @click="copySSHCommand(serviceId)">复制 VS Code 命令</button>
<details>
  <summary>OpenSSH 配置</summary>
  <button @click="copySSHConfig(serviceId)">复制配置</button>
</details>
```

- [x] **Step 3: Add copy failure feedback**

Expected: unsupported clipboard or rejection shows toast: `复制失败，请手动选择文本`.

- [x] **Step 4: Verify**

Run:

```bash
cd code/frontend
npm run test:run -- src/features/contest-awd-workspace/model/sshAccessPresentation.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
npm run typecheck
```

Expected: pass.

- [ ] **Step 5: Manual VS Code verification**

Run real API flow:

```bash
curl -c /tmp/ctf-cookie.txt -H 'Content-Type: application/json' \
  -d '{"username":"student","password":"Password123"}' \
  http://127.0.0.1:8080/api/v1/auth/login
curl -b /tmp/ctf-cookie.txt -X POST \
  http://127.0.0.1:8080/api/v1/contests/8/awd/services/21/defense/ssh
```

Expected:

- `command` is `ssh student+8+21@127.0.0.1 -p 2222`.
- OpenSSH profile alias is `ctf-awd-8-21`.
- VS Code Remote-SSH should use alias config, not direct `vscode-remote://ssh-remote+student+8+21@...`.

## Task 2: Extract Defense Presentation Model

**Files:**
- Create: `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- Test: `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts`
- Modify: `code/frontend/src/features/contest-awd-workspace/model/index.ts`
- Modify: `code/frontend/src/features/contest-awd-workspace/index.ts`

- [x] **Step 1: Write tests for risk ordering**

Test cases:

- `compromised` before `down`
- `down` before `attack_received > 0`
- active operation states before healthy `up`
- stable `up` last

- [x] **Step 2: Implement `toDefenseServiceCards`**

Shape:

```ts
interface AWDDefenseServiceCard {
  serviceId: string
  challengeId: string
  title: string
  riskLevel: 'critical' | 'warning' | 'watch' | 'stable'
  riskReasons: string[]
  serviceStatusLabel: string
  instanceStatusLabel: string
  canOpenService: boolean
  canRequestSSH: boolean
  canRestart: boolean
}
```

- [x] **Step 3: Export through feature public API**

Expected: components import from `@/features/contest-awd-workspace`, not internal model path.

- [x] **Step 4: Verify**

Run:

```bash
cd code/frontend
npm run test:run -- src/features/contest-awd-workspace/model/awdDefensePresentation.test.ts
npm run typecheck
```

Expected: pass.

## Task 3: Extract Defense Service Components

**Files:**
- Create: `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
- Create: `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Test: optional `code/frontend/src/components/contests/awd/__tests__/AWDDefenseServiceList.test.ts`
- Test: `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

- [x] **Step 1: Define component contracts**

`AWDDefenseServiceList.vue` props:

```ts
interface Props {
  services: AWDDefenseServiceCard[]
  selectedServiceId: string
  openingServiceKey: string
  openingSSHKey: string
  serviceActionPendingById: Record<string, boolean>
}
```

Emits:

- `select-service`
- `open-service`
- `request-ssh`
- `restart-service`

- [x] **Step 2: Move service list template**

Expected: parent still owns API calls and async handlers; child owns display and emits.

- [x] **Step 3: Move SSH block into `AWDDefenseConnectionPanel.vue`**

Props:

- `access`
- `serviceId`
- `copiedCommand`
- `copiedConfig`

Emits:

- `copy-command`
- `copy-config`

- [x] **Step 4: Keep async guards in parent/composable**

Expected: duplicate SSH/restart/open clicks still blocked by existing handler state.

- [x] **Step 5: Verify**

Run:

```bash
cd code/frontend
npm run test:run -- src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
npm run typecheck
```

Expected: pass.

## Task 4: Add Student-Safe Defense Intel Panel

**Files:**
- Create: `code/frontend/src/components/contests/awd/AWDDefenseIntelPanel.vue`
- Create/Modify: `code/frontend/src/features/contest-awd-workspace/model/awdDefenseIntelPresentation.ts`
- Modify: `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- Option A API: extend `GET /api/v1/contests/:id/awd/workspace`
- Option B API: add student-scoped `GET /api/v1/contests/:id/awd/defense/traffic`

- [ ] **Step 1: Decide API source**

Recommended first pass: extend workspace response only with small, scoped fields:

```ts
interface ContestAWDWorkspaceServiceData {
  recent_attack_count?: number
  recent_error_count?: number
  suspicious_paths?: Array<{ path: string; count: number; status_group: string }>
}
```

Expected: no admin-only raw traffic exposure.

- [ ] **Step 2: Add frontend presentation tests**

Expected:

- 5xx paths are `warning`.
- repeated POST to static-like path is `warning`.
- empty traffic shows compact empty state, not explanatory prose.

- [ ] **Step 3: Implement panel**

Sections:

- 最近战报
- 异常路径
- 最近 checker / 5xx

- [ ] **Step 4: Verify**

Run:

```bash
cd code/frontend
npm run test:run -- src/features/contest-awd-workspace/model src/views/contests/__tests__/ContestDetail.test.ts
npm run typecheck
```

Expected: pass.

## Task 5: Backend Student Defense Traffic Summary

**Files:**
- Modify: `code/backend/internal/dto/contest.go` or current AWD workspace DTO location
- Modify: `code/backend/internal/module/contest/application/queries/awd_service.go`
- Modify: `code/backend/internal/module/contest/infrastructure/...` exact repository after exploration
- Test: focused contest AWD workspace query tests

- [ ] **Step 1: Locate current AWD workspace query owner**

Run:

```bash
cd code/backend
rg -n "GetAWDWorkspace|recent_events|ContestAWDWorkspace" internal
```

Expected: identify service and repository owner.

- [ ] **Step 2: Write tests for student scoping**

Expected:

- student sees only own team service defense summary.
- student does not see unrelated teams' raw request bodies.
- admin/teacher traffic APIs remain unchanged.

- [ ] **Step 3: Implement minimal scoped summary**

Rules:

- aggregate by current round when available.
- limit paths to top 5.
- include only method/path/status/count/source/time, no sensitive request body by default.
- normalize times to UTC in API response.

- [ ] **Step 4: Verify**

Run:

```bash
cd code/backend
go test ./internal/module/contest/...
go test ./internal/app -run 'AWD|Router'
```

Expected: pass.

## Task 6: Controlled Browser Defense Workbench (Phase 2)

**Files:**
- Modify: `code/backend/internal/module/runtime/api/http/handler.go`
- Modify: `code/backend/internal/app/composition/runtime_module.go`
- Modify/Create: runtime defense policy helper
- Modify: `code/frontend/src/api/contest.ts`
- Create: `code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue`

- [ ] **Step 0: Ship read-only Phase 2A first**

Scope:

- Enable only directory listing and file reading.
- Keep file saving and command execution forbidden at the HTTP handler.
- Gate read-only workbench with `container.defense_workbench_readonly_enabled`, default `false`.
- Require `container.defense_workbench_root` to be an absolute non-root path when the gate is enabled.
- Development config may enable it for local testing and defaults to `/app`, matching the current AWD challenge image workdir.
- Frontend does not expose source browsing in the student battlefield by default; browser file access remains backend-gated for future controlled experiments.

Backend rules:

- `GET defense/directories` and `GET defense/files` call the runtime service.
- Student authorization remains owned by `FindAWDDefenseSSHScope`.
- Browser paths stay relative in the API, but backend resolves them under `defense_workbench_root` before reading from the container.
- Reject traversal, absolute paths, `.ssh`, env files, `/proc`, `/sys`, `/dev`, `/run`, `/var/run`.
- Max read size remains <= 256 KiB and directory list <= 300.

Frontend rules:

- Workbench belongs to the selected defense service.
- Route panel/composable owns async loading and errors.
- Component is display-only: directory entries, selected file, refresh/open events.
- No write editor and no command console in Phase 2A.

Validation:

```bash
cd code/backend
go test ./internal/module/runtime/api/http -run 'AWDDefenseWorkbench'
go test ./internal/app/composition -run 'AWDDefenseWorkbench'
go test ./internal/app -run 'Router|AWD|Runtime'
cd ../frontend
npm run test:run -- src/api/__tests__/contest.test.ts src/components/contests/awd src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts
npm run typecheck
```

- [ ] **Step 1: Keep feature flag disabled by default**

Expected: browser file workbench is hidden unless backend capability says enabled.

- [ ] **Step 2: Define backend allowlist**

Rules:

- readable roots: challenge workdir only.
- block `/proc`, `/sys`, `/dev`, `/run`, `/var/run`, `.ssh`, env files by default.
- max file read remains <= 256 KiB.
- command execution allowlist starts with read-only commands such as `ls`, `find`, `grep`, `sed -n`, `tail`, `cat` with path normalization.

- [ ] **Step 3: Add audit events**

Expected audit fields:

- user_id
- contest_id
- service_id
- action
- path or command hash
- result
- request_id

- [ ] **Step 4: Save file with backup**

Expected: `PUT defense/files` with `backup=true` returns `backup_path`.

- [ ] **Step 5: Verify**

Run:

```bash
cd code/backend
go test ./internal/module/runtime/...
go test ./internal/app -run 'AWDDefense|Runtime'
cd ../frontend
npm run test:run -- src/api/__tests__/contest.test.ts src/components/contests/awd
npm run typecheck
```

Expected: pass.

## Integration Checks

- [ ] `cd code/frontend && npm run test:run -- src/features/contest-awd-workspace/model src/views/contests/__tests__/ContestDetail.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- [ ] `cd code/frontend && npm run typecheck`
- [ ] `cd code/frontend && npm run check:theme-tail`
- [ ] `cd code/backend && go test ./internal/module/contest/...`
- [ ] `cd code/backend && go test ./internal/module/runtime/...`
- [ ] `cd code/backend && go test ./internal/app -run 'AWD|Runtime|Router'`

## Manual Verification

- [ ] 登录 `student / Password123`。
- [ ] 打开 `http://localhost:5173/contests/8?panel=challenges`。
- [ ] 生成服务 `21` 的 SSH 防守连接。
- [ ] 点击 `复制 VS Code 命令`，确认剪贴板内容是 `ssh student+8+21@127.0.0.1 -p 2222`。
- [ ] 展开 `OpenSSH 配置`，确认 `Host ctf-awd-8-21` 配置可复制。
- [ ] 在 VS Code Remote-SSH 中优先使用 `Host ctf-awd-8-21` alias；不要使用直接包含 `student+8+21@...` 的 remote URI。
- [ ] 输入页面返回的临时密码，确认不再出现 `Could not resolve hostname host`。
- [ ] 修复或重启服务后刷新，确认服务状态和最近同步时间更新。

## Rollback / Recovery Notes

- 前端组件拆分可单独 revert，不影响后端 SSH ticket。
- 后端 Phase 1 只扩展读模型/DTO 时应保持向后兼容。
- 后端 Phase 2 启用文件/命令接口前必须保留 feature flag；发现风险时可关闭浏览器工作台，只保留 SSH。
- 不执行破坏性数据库操作，不需要 migration rollback。

## Residual Risks

- VS Code Remote-SSH 的 GUI 密码输入结果无法完全通过 WSL 终端自动断言；可以通过 Remote-SSH 日志确认 host alias 解析正确，通过裸 SSH 验证密码票据和远端 exec。
- 学生侧流量摘要如果过细，可能泄露其他队伍攻击细节；第一版必须按本队服务聚合，并限制字段。
- 浏览器内文件保存/命令执行属于高风险能力，不应和第一阶段 UI 改造混在同一个提交里。
- 当前 `ContestAWDWorkspacePanel.vue` 已较大，拆分时要保持父组件拥有路由/刷新/API/错误策略，子组件只负责展示和事件上抛。

## Plan Review

- Self-review result: architecture boundary is explicit. Frontend page owner remains `ContestAWDWorkspacePanel.vue` / `useContestAWDWorkspace`; new child components own display only. Backend Phase 1 and Phase 2 are separated so high-risk file/command execution does not block immediate SSH/defense UX repair.
- Independent subagent review: not dispatched in this session because the active tool policy only allows spawning subagents when the user explicitly asks for subagents or parallel agent work.
