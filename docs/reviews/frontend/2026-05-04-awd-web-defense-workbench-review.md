# AWD Web 防守工作台 Review

## Review Scope

- 学生 AWD 战场 Web 防守工作台默认入口。
- 通用防守文件/目录读取能力是否仍可从学生默认 API 暴露。
- SSH 从服务卡迁移到选中服务防守面板后的交互一致性。
- “攻防战场”页签和文档边界。

## Findings

### 1. 阻塞：通用文件/目录读取仍可从后端访问

结论：成立，已修复。

原问题：

- 前端移除了 `AWDDefenseFileWorkbench` 挂载，但后端 `ReadAWDDefenseFile` 和 `ListAWDDefenseDirectory` 仍会在配置开启时读取容器文件与目录。
- 前端 API wrapper 仍保留 `requestContestAWDDefenseDirectory` 和 `requestContestAWDDefenseFile`。

修复：

- `code/backend/internal/module/runtime/api/http/handler.go`
  - `ReadAWDDefenseFile` 直接返回 `ErrForbidden`。
  - `ListAWDDefenseDirectory` 直接返回 `ErrForbidden`。
- `code/backend/configs/config.dev.yaml`
  - `defense_workbench_readonly_enabled` 默认改为 `false`。
- `code/frontend/src/api/contest.ts`
  - 删除学生侧通用目录/文件读取 wrapper。
- `code/backend/internal/module/runtime/api/http/handler_test.go`
  - 更新为断言读文件、列目录均 forbidden，且不返回文件内容或目录项。
- `code/frontend/src/api/__tests__/contest.test.ts`
  - 删除保活目录/文件读取 API 的测试。

### 2. 阻塞：未选中服务卡点击 SSH 后看不到生成结果

结论：成立，已修复。

原问题：

- 服务卡中的 SSH 按钮只生成对应服务 ticket，不切换 `selectedServiceId`。
- SSH 展示面板绑定选中服务，导致用户可能生成了未选中服务的 SSH 却看不到复制入口。

修复：

- `ContestAWDWorkspacePanel.vue` 新增 `requestDefenseSSH(serviceId)`。
- 服务卡与操作面板的 SSH 事件都进入该函数，先选中服务，再请求 SSH。

### 3. 低风险：页签改名测试不够精确

结论：成立，已修复。

修复：

- `ContestDetail.test.ts` 运行时断言从模糊 `战场` 改为精确 `攻防战场`。
- 源码断言同步更新 `ContestDetail.vue` 的 workspace heading。

## Validation After Fix

已通过：

```bash
cd code/frontend
npm run test:run -- src/api/__tests__/contest.test.ts src/views/contests/__tests__/ContestDetail.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/components/contests/awd/__tests__/AWDDefenseConnectionPanel.test.ts
npm run typecheck
npm run check:theme-tail
```

已通过：

```bash
cd code/backend
go test ./internal/module/runtime/api/http ./internal/config
```

## Residual Risk

- 通用 files/directories 路由仍保留，但 HTTP handler 已统一 forbidden。后续二期应新增 fragments/patch API，而不是复活这些通用路由。
- `runtimeHTTPServiceAdapter` 内仍存在只读 workbench 方法，用于历史计划和测试覆盖；当前学生 HTTP 入口不会调用它们。
- SSH 作为高级入口仍具备容器访问能力；这是赛制能力边界，不应被表述为“完全禁止学生看到源码”。
