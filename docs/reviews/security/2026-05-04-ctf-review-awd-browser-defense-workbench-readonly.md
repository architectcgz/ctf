# AWD Browser Defense Workbench Readonly Review

- Review target:
  - Repository: `ctf`
  - Worktree: `/home/azhi/workspace/projects/.worktrees/ctf-awd-browser-defense-workbench`
  - Diff source: current uncommitted local diff reviewed on 2026-05-04
  - Files reviewed:
    - `code/backend/configs/config.yaml`
    - `code/backend/configs/config.dev.yaml`
    - `code/backend/configs/config.prod.yaml`
    - `code/backend/internal/config/config.go`
    - `code/backend/internal/app/composition/runtime_module.go`
    - `code/backend/internal/app/composition/runtime_module_test.go`
    - `code/backend/internal/module/runtime/api/http/handler.go`
    - `code/backend/internal/module/runtime/api/http/handler_test.go`
    - `code/backend/internal/module/runtime/infrastructure/awd_target_proxy_repository.go`
    - `code/backend/internal/module/runtime/infrastructure/engine.go`
    - `code/backend/internal/module/runtime/infrastructure/engine_test.go`
    - `code/frontend/src/api/contest.ts`
    - `code/frontend/src/api/contracts.ts`
    - `code/frontend/src/api/__tests__/contest.test.ts`
    - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
    - `code/frontend/src/components/contests/awd/AWDDefenseFileWorkbench.vue`
    - `code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts`
    - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
    - `docs/plan/impl-plan/2026-05-04-awd-defense-workbench-design-implementation-plan.md`
  - Validation executed:
    - `cd code/backend && go test ./internal/app/composition ./internal/module/runtime/api/http ./internal/module/runtime/infrastructure`
    - `cd code/backend && go test ./internal/app/composition -run 'AWDDefenseWorkbench'`
    - `cd code/frontend && npm run test:run -- src/api/__tests__/contest.test.ts src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`

- Classification: agree with non-trivial security review gate. This slice re-opens previously forbidden browser read routes, adds a new runtime feature flag, and introduces a new student-facing defense file workflow across backend and frontend.
- Verdict: pass with minor issues

## Findings

### Minor

1. `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue:397-438`, `code/frontend/src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts:1-54`, `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts:33-40`
   - Risk: 当前实现已经在发起新目录/文件请求前清空 `defenseFile`，旧文件内容不会继续残留；但相关测试仍然停留在 source 字符串断言和静态 emit 断言，没有真正覆盖 deferred promise、切服务、切目录、旧响应晚到这些异步状态路径。
   - Impact: 这是测试债，不是当前实现缺陷。代码层面这轮修复已经把之前的 stale-content/blocking loading 问题收掉，但未来重构时这条行为仍然容易回归。
   - Fix direction: 后续补一个组件级异步测试，显式断言“旧请求晚于新请求返回时，不会覆盖当前服务/当前文件状态，且 loading 只由最新 seq 负责清理”。

## Material Findings

无。

## Senior Implementation Assessment

- 这轮修复把之前的两个 blocker 都处理到了正确方向上。后端现在由 `FindAWDDefenseSSHScope` 守住租户/队伍边界，再叠加 `container.defense_workbench_root` 做正向 rooted 约束；adapter 也已经把所有 read/list 请求拼到受控根目录下，不再依赖容器 `WorkingDir`。
- 配置层和组合层的证据也足够直接：`container.defense_workbench_readonly_enabled` 为 true 时必须提供绝对且非根路径；composition 测试覆盖了空 root、`.`、`/`、相对 root 拒绝，以及 `src/app.py -> /home/student/src/app.py`、根目录 listing -> `/home/student` 这两条关键路径。
- 前端 stale-content 问题也已经按最小变更修掉：同服务内切目录/文件时会先清空 `defenseFile`，旧文件内容不会继续挂在当前界面。剩下的问题主要是测试还不够行为化，不再构成 gate blocker。

## Required Re-validation

- `cd code/backend && go test ./internal/app/composition ./internal/module/runtime/api/http ./internal/module/runtime/infrastructure -count=1`
- `cd code/backend && go test ./internal/app/composition -run 'AWDDefenseWorkbench' -count=1`
- `cd code/frontend && npm run test:run -- src/api/__tests__/contest.test.ts src/components/contests/awd/__tests__/AWDDefenseFileWorkbench.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `cd code/frontend && npm run typecheck`
- 手工或自动化补充以下行为验证：
  - 根目录下的只读文件/目录请求在真实容器里与测试一致，实际访问路径稳定落在 `container.defense_workbench_root` 下。
  - 目录 listing 在存在 `.env`、`prod.env`、`.ssh` 等 entry 时会稳定过滤，且不会影响非敏感 sibling entry。
  - 补一个真正覆盖 stale response 的前端异步行为测试。

## Residual Risk

- 本次 review 没有跑真实容器复现；后端 rooted 策略主要依赖配置校验和 composition 测试证据，尚未补到端到端验证。
- 当前 handler 与前端接线都显示保存文件和命令执行仍未开放，这一项在最新 diff 中保持正确。
- `FindAWDDefenseSSHScope` 的现有仓储查询仍然把 contest/team/service/running instance 边界绑定在一起；本 review 没有发现它在本切片中被削弱。
