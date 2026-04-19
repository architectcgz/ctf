# AWD 最终方案一期落地边界

## 目标

把此前确定的 AWD 最终方案拆成可渐进落地的阶段，并明确当前仓库已经交付到哪一层、仍保留在哪一层的兼容路径。

## 当前一期已落地

### 1. 独立 AWD 服务模板题库

- 新增独立 `awd_service_templates` 存储
- 后台已提供服务模板 CRUD 与查询接口
- 管理端已提供独立 AWD 服务模板库页面
- AWD 题库与普通 Jeopardy 题目保持分层，不再混在同一套题库语义里

### 2. 赛事级服务关联存储

- 新增 `contest_awd_services`
- 管理员在 AWD 赛事中新增、更新、删除题目时，会同步维护赛事服务关联
- 管理端赛事题目列表会返回：
  - `awd_service_id`
  - `awd_template_id`
  - `awd_service_display_name`
- AWD 赛事配置面板会显示服务关联状态，便于确认题目是否已经进入赛事服务层

### 3. 兼容式管理链路

- 赛事题目配置仍通过 `contest_challenges` 完成基础管理
- `contest_awd_services` 当前承担的是显式关联与后续迁移承接点
- 这一层已经满足“独立 AWD 题库 + 赛事服务映射”的一期设计目标

## 当前一期未切换

下面这些运行态能力暂未切到 `contest_awd_services`：

- 轮次调度
- checker 执行
- workspace / 实例编排
- flag 注入与运行时服务定义读取
- readiness 计算的底层事实源

当前运行态仍继续读取 legacy `contest_challenges.awd_*` 字段，目的是在完成显式服务关联落地前，先保证现有 AWD 比赛链路不被打断。

## 当前事实源划分

### 1. 模板层

- `awd_service_templates`
- 描述可复用的 AWD 服务模板、checker 草稿、访问配置与运行配置

### 2. 赛事服务层

- `contest_awd_services`
- 描述某场 AWD 赛事采用了哪个模板、映射到哪道题、显示名称与基础服务配置

### 3. 兼容运行层

- `contest_challenges`
- 当前仍保存 checker 类型、checker 配置、SLA / 防守分、校验状态
- 仍是当前运行态的兼容事实源

## 下一阶段

下一阶段的目标不是继续扩模板库，而是把运行态逐步切到显式服务模型：

1. 让 readiness、round updater、checker runner 逐步读取 `contest_awd_services`
2. 把 `contest_challenges` 从“运行态事实源”降级为“赛事题目关联兼容层”
3. 再推进 workspace、flag 注入、流量与战报链路对显式服务模型的统一

## 验证记录

本阶段已验证：

- `cd code/backend && go test ./internal/module/contest/... -count=1`
- `cd code/backend && go test ./internal/module/challenge/... ./internal/module/contest/... -count=1`
- `cd code/frontend && npm run test:run -- src/views/admin/__tests__/AWDServiceTemplateLibrary.test.ts src/composables/__tests__/useAdminAwdServiceTemplates.test.ts src/components/admin/awd-service/__tests__/AWDServiceTemplateLibraryPage.test.ts src/components/admin/__tests__/AWDChallengeConfigDialog.test.ts`
- `cd code/frontend && npm run typecheck`
- `python3` 解析 `docs/contracts/openapi-v1.yaml` 通过

已知仓库基线：

- `cd code/backend && go test ./internal/module/challenge/... ./internal/module/contest/... ./internal/app/... -count=1` 仍会因 `internal/app` 中既有容器启动用例返回 `容器启动失败` 而失败
- `cd code/backend && go test ./... -count=1` 同样会被同一类基线阻断
- 主要失败点仍集中在 `practice_flow_integration_test`、`full_router_integration_test`、`full_router_state_matrix_integration_test` 这类依赖容器启动的集成用例
- 上述失败不是本次 `ContestAWDService` 存储层改动引入的回归，本次新增的 `challenge` / `contest` 相关测试链路已经独立通过
