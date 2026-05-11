# 2026-05-10 reuse-first harness implementation plan

## 目标

为 `ctf` 仓库补一套 `reuse-first` harness，强制 agent 在新增或修改页面、组件、hook、API wrapper、store、表单、表格、modal、layout、schema 等受保护实现前，先完成分类、检索、复用决策，再允许编码。

## 非目标

- 不在这一轮重构现有前端页面或批量改造历史实现。
- 不引入 AST 级别的复杂静态分析。
- 不把所有 backend/service 复用检测一次性做成完整语义分析器；这一轮先把前端高频重复面和通用决策门槛接上。

## 输入依据

- `AGENTS.md`：当前 harness 入口、结构性改动流程、前端 guardrails、hook/check 约束。
- `scripts/check-consistency.sh`、`.githooks/pre-commit`、`.githooks/README.md`：现有机械检查和 hook 接线点。
- `code/frontend/src/components/common/WorkspaceDirectoryToolbar.vue`
- `code/frontend/src/components/common/WorkspaceDataTable.vue`
- `code/frontend/src/features/student-directory/model/useStudentListQuery.ts`
- `code/frontend/src/views/notifications/NotificationList.vue`
- `code/frontend/src/components/teacher/student-management/StudentManagementPage.vue`

## 方案摘要

### reuse-first 闭环

1. `harness/policies/reuse-first.yaml`
   - 定义受保护创建面、强制检索目录、禁止模式和决策枚举。
2. `harness/policies/project-patterns.yaml`
   - 用当前仓库真实页面、表格、toolbar、hook、API wrapper 建立模式索引。
3. `.harness/reuse-decision.md`
   - 作为每次受保护改动前必须更新的当前任务复用决策记录，可以被下一次任务覆盖。
4. `.harness/reuse-index.yaml` + `.harness/reuse-history.md`
   - 长期保存可复用模式索引和 append-only 历史摘要，避免当前任务覆盖历史复用线索。
5. `harness/checks/*.py` + `scripts/check-reuse-first.sh`
   - 机械化检查 reuse decision、相似页面、重复 hook、重复 API wrapper。
6. `.githooks/pre-commit`
   - 本地提交前执行 reuse-first 检查。
7. 可选的远端 workflow
   - 若仓库需要额外兜底，可再接 CI；但它不是本轮约束成立的前提。
8. `AGENTS.md` + `harness/prompts/coding-agent-system-prompt.md`
   - 把 Step 1 Classify / Step 2 Search / Step 3 Decide / Step 4 Implement 固化进当前 agent 工作流。

### 结构适配

- 本仓库不是 `src/pages` 结构，前端真实复用面以 `code/frontend/src/views`、`components`、`features/*/model`、`composables`、`api`、`stores` 为主。
- 页面相似度检查以 `*Page.vue`、`views/**/*.vue` 为主，关键词使用当前仓库已有的 filter / table / pagination / modal / form / query / detail 等结构特征。
- hook/API 检查优先用文件名、导入来源、请求路径和结构关键词做轻量相似度判定，避免引入第三方依赖。

## 任务切片

### Slice 1：策略与模板

- 新增 `harness/policies/reuse-first.yaml`
- 新增 `harness/policies/project-patterns.yaml`
- 新增 `harness/templates/reuse-decision.md`
- 新增 `harness/templates/pattern-index-example.yaml`
- 新增 `.harness/reuse-decision.md`
- 新增 `.harness/reuse-index.yaml`
- 新增 `.harness/reuse-history.md`
- 新增 `harness/prompts/coding-agent-system-prompt.md`

验证：

- 文件结构存在且路径可被 `scripts/check-consistency.sh` 检查。
- `project-patterns.yaml` 示例全部指向仓库现有真实文件。

### Slice 2：机械检查

- 新增 `harness/checks/common.py`
- 新增 `harness/checks/check-reuse-decision.py`
- 新增 `harness/checks/check-similar-pages.py`
- 新增 `harness/checks/check-duplicate-hooks.py`
- 新增 `harness/checks/check-api-wrapper-duplication.py`
- 新增 `scripts/check-reuse-first.sh`

验证：

- `bash scripts/check-reuse-first.sh --staged` 在无受保护改动时通过。
- 对模拟的受保护新增文件，脚本会因为缺少/未更新 `.harness/reuse-decision.md` 或未引用相似实现而失败。

### Slice 3：本地接线与工作流

- 更新 `AGENTS.md`
- 更新 `.githooks/pre-commit`
- 更新 `.githooks/README.md`
- 更新 `scripts/install-githooks.sh`
- 更新 `scripts/check-consistency.sh`
- 更新 `scripts/doctor-local-harness.sh`
- 更新 `feedback/AGENTS.md`
- 新增 `feedback/2026-05-10-reuse-first-harness.md`

验证：

- `bash scripts/check-consistency.sh`
- `bash scripts/check-reuse-first.sh --staged`
- `bash scripts/doctor-local-harness.sh`

## Plan Review

- ownership：reuse-first 的事实源分成四层，`harness/policies/*` 管规则，`.harness/reuse-decision.md` 管本轮决策证据，`.harness/reuse-index.yaml` 和 `.harness/reuse-history.md` 管长期复用线索，`harness/checks/*` 管执行。
- reuse point：页面模式、hook/API 复用点都落到 `project-patterns.yaml`，避免把“先搜一搜”只写在 prompt 里。
- hidden redesign risk：这一轮不动历史页面结构，只新增 guardrail；不会形成“先上规则、马上再重构规则本身”的第二轮立刻重设计。
- touched debt：当前前端已有 `components/*Page.vue` 与 `views/*.vue` 并存，这是已知历史形态。这一轮不新增第三种页面入口，而是把两类现有入口都纳入 reuse-first 检查。

## 回退方式

- 移除 `harness/`、`.harness/`、`scripts/check-reuse-first.sh`
- 回退 `AGENTS.md`、`.githooks/pre-commit`、`.githooks/README.md`、`scripts/check-consistency.sh`、`scripts/doctor-local-harness.sh` 的接线
