# AWD Challenge Picker Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化 AWD 赛事从 AWD 题库选题的前端链路，让管理员能在无“模板”概念的语义下高效筛选、批量关联和识别题目状态。

**Architecture:** 保留现有 AWD 题库与赛事 AWD service 契约，不改 `/authoring/awd-challenges` 与 `/admin/contests/:id/awd/services` 后端接口。前端把选题弹层从一次性加载 100 条改为可分页/搜索的题库选择器，并把批量创建、默认配置和 UI 命名从旧 `template` 语义迁移到 `awdChallenge` 语义。

**Tech Stack:** Vue 3, TypeScript, Vite, Vitest, Vue Test Utils, existing admin API layer, existing workspace UI components

---

## Requirement Summary

确认需求：

- AWD 题库继续保留，AWD 赛事必须能从 AWD 题库中选题。
- AWD 题目是每道题独立设计的资源，不再表达为“模板”。
- 已移除环境模板库页面入口，本计划不恢复该入口。

本计划不做：

- 不移除 `/platform/awd-challenges`。
- 不把 AWD 选题改回普通 `getChallenges`。
- 不改后端 API 契约，除非实施时发现后端确实缺少分页字段或搜索参数支持。
- 不重写 Checker 结构化配置面板。

## Acceptance Criteria

- AWD 赛事新增服务弹层可以按关键词、分类、难度、服务类型、部署方式和就绪状态筛选 AWD 题库。
- AWD 选题弹层不再只依赖前 100 条数据；至少支持分页加载。
- 选题列表展示 `slug`、就绪状态、最近验证时间或等价状态信息，帮助判断每道题是否可用于赛事。
- 批量关联多个 AWD 题目时，失败提示能说明哪些成功、哪些失败；不会表现成全量成功。
- 新增 AWD service 时可以统一设置分值、起始顺序和可见性。
- 前端用户可见文案、DOM id、CSS class、测试命名中与 AWD 题库相关的旧 `template` 语义被替换为 `awdChallenge` / `AWD 题目`。
- 现有 AWD 题库页面、AWD 赛事题目池、AWD 配置、赛前检查相关定向测试通过。

## Planned File Map

### AWD 题库选择状态

- Modify: `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
  - 管理 AWD 题库查询参数、分页状态、批量创建结果和新增服务默认设置。
- Modify or Create: `code/frontend/src/composables/useContestAwdChallengePicker.ts`
  - 推荐新建。收敛 AWD 题库加载、筛选、分页、去重、选择状态和错误处理，避免继续扩大面板文件。

### AWD 选题弹层

- Modify: `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
  - 把 AWD 新增模式升级为题库选择器：筛选栏、分页、状态列、默认分值/顺序/可见性设置、批量选择。
  - 将 `contest-template-*` 相关 DOM id/class 改为 `contest-awd-challenge-*`。

### AWD 题库页面语义

- Modify: `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
  - 保持页面能力不变，只清理 `awd-template-*` CSS class / id / 变量名，迁移为 `awd-challenge-*`。
- Modify: `code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue`
  - 清理 `awd-template-*` 表单 id/class，改为 `awd-challenge-*`。

### Tests

- Modify: `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
  - 覆盖分页搜索、筛选参数、默认设置、批量部分失败。
- Modify: `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
  - 覆盖从 AWD 配置阶段跳回题目池选题时仍能使用新选择器。
- Modify: `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
  - 清理旧 template 命名断言。
- Modify: `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts`
  - 清理旧 template id 断言。
- Modify: `code/frontend/src/views/__tests__/sharedThemeTokenAdoption.test.ts`
  - 若 raw source 检查受 class 改名影响，同步更新。

## Task 1: 为 AWD 题库选择器补 RED 测试

**Files:**
- Modify: `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- Modify: `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`

- [ ] **Step 1: 写失败测试，要求 AWD 新增弹层按筛选条件加载题库**

在 `ContestChallengeOrchestrationPanel.test.ts` 增加用例：

```ts
it('应该按关键词和筛选条件加载 AWD 题库候选', async () => {
  contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
    list: [],
    total: 0,
    page: 1,
    page_size: 20,
  })

  const wrapper = mountPanel({ contestMode: 'awd' })
  await flushPromises()
  await wrapper.get('#contest-challenge-add').trigger('click')
  await wrapper.get('#contest-awd-challenge-keyword').setValue('bank')
  await wrapper.get('#contest-awd-challenge-service-type').setValue('web_http')
  await wrapper.get('#contest-awd-challenge-readiness').setValue('passed')
  await flushPromises()

  expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenLastCalledWith({
    page: 1,
    page_size: 20,
    keyword: 'bank',
    service_type: 'web_http',
    readiness_status: 'passed',
    status: 'published',
  })
})
```

- [ ] **Step 2: 写失败测试，要求分页加载不是固定 100 条**

同文件增加用例：

```ts
it('应该支持 AWD 题库分页翻页', async () => {
  contestApiMocks.listAdminAwdChallenges.mockResolvedValue({
    list: [buildAwdCatalogItem({ id: '21', name: 'Page 1' })],
    total: 45,
    page: 1,
    page_size: 20,
  })

  const wrapper = mountPanel({ contestMode: 'awd' })
  await flushPromises()
  await wrapper.get('#contest-challenge-add').trigger('click')
  await flushPromises()
  await wrapper.get('#contest-awd-challenge-next-page').trigger('click')
  await flushPromises()

  expect(contestApiMocks.listAdminAwdChallenges).toHaveBeenLastCalledWith(
    expect.objectContaining({ page: 2, page_size: 20 })
  )
})
```

- [ ] **Step 3: 写失败测试，要求列表展示独立题目信息而不是模板语义**

断言：

```ts
expect(wrapper.text()).toContain('upload-http')
expect(wrapper.text()).toContain('已就绪')
expect(wrapper.text()).not.toContain('模板')
expect(wrapper.find('#contest-template-option-11').exists()).toBe(false)
expect(wrapper.find('#contest-awd-challenge-option-11').exists()).toBe(true)
```

- [ ] **Step 4: 在 `ContestEdit.test.ts` 补跨阶段入口断言**

从 AWD 配置阶段点击“关联新资源”后：

```ts
expect(wrapper.get('#contest-workbench-stage-tab-pool').attributes('aria-selected')).toBe('true')
expect(wrapper.find('#contest-awd-challenge-option-999').exists()).toBe(true)
expect(wrapper.find('#contest-template-option-999').exists()).toBe(false)
```

- [ ] **Step 5: 运行定向测试确认 RED**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120 npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts
```

预期：FAIL，失败点集中在新筛选控件、新分页控件、新 DOM id 尚不存在。

## Task 2: 抽出 AWD 题库选择状态

**Files:**
- Create: `code/frontend/src/composables/useContestAwdChallengePicker.ts`
- Create: `code/frontend/src/composables/__tests__/useContestAwdChallengePicker.test.ts`
- Modify: `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`

- [ ] **Step 1: 写 composable RED 测试**

测试覆盖：

- 初始加载使用 `page=1&page_size=20&status=published`
- keyword 改变后回到第一页
- service type / readiness 改变后回到第一页
- 已关联题目不出现在 selectable 列表
- 加载失败保留上次成功数据并记录错误

运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120 npm run test:run -- src/composables/__tests__/useContestAwdChallengePicker.test.ts
```

预期：FAIL。

- [ ] **Step 2: 实现 `useContestAwdChallengePicker`**

接口建议：

```ts
export interface ContestAwdChallengePickerFilters {
  keyword: string
  category: ChallengeCategory | ''
  difficulty: ChallengeDifficulty | ''
  serviceType: AWDServiceType | ''
  deploymentMode: AWDDeploymentMode | ''
  readinessStatus: AWDReadinessStatus | ''
}

export function useContestAwdChallengePicker(options: {
  existingChallengeIds: Readonly<Ref<string[]>>
  pageSize?: number
}) {
  return {
    filters,
    page,
    pageSize,
    total,
    list,
    selectableList,
    loading,
    loadError,
    refresh,
    changePage,
    reset,
  }
}
```

实现要点：

- `keyword` debounce 250ms。
- `refresh` 捕获异常并用 `loadError` 返回给 UI owner。
- 不在 composable 内 toast；由面板决定是否提示。
- 参数传给 `listAdminAwdChallenges` 时过滤空字符串。

- [ ] **Step 3: 在 `ContestChallengeOrchestrationPanel.vue` 接入 composable**

替换当前 `awdChallengeCatalog`、`loadingAwdChallengeCatalog`、`ensureAwdChallengeCatalogLoaded` 局部状态。

保留面板职责：

- 打开弹层时调用 `picker.refresh()`。
- 捕获 `loadError` 后 toast 一次。
- 保存后刷新赛事 service 列表。

- [ ] **Step 4: 跑 composable 与面板测试转 GREEN**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120 npm run test:run -- src/composables/__tests__/useContestAwdChallengePicker.test.ts src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts
```

预期：PASS。

## Task 3: 升级 AWD 选题弹层 UI

**Files:**
- Modify: `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
- Modify: `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`

- [ ] **Step 1: 给弹层 props/emits 扩展筛选与分页契约**

新增 props：

```ts
awdChallengeTotal?: number
awdChallengePage?: number
awdChallengePageSize?: number
awdChallengeFilters?: ContestAwdChallengePickerFilters
awdChallengeLoadError?: string
```

新增 emits：

```ts
'update-awd-challenge-keyword': [value: string]
'update-awd-challenge-service-type': [value: AWDServiceType | '']
'update-awd-challenge-deployment-mode': [value: AWDDeploymentMode | '']
'update-awd-challenge-readiness': [value: AWDReadinessStatus | '']
'change-awd-challenge-page': [page: number]
'refresh-awd-challenge-catalog': []
```

- [ ] **Step 2: 添加筛选栏和分页控件**

控件 id 固定：

- `contest-awd-challenge-keyword`
- `contest-awd-challenge-service-type`
- `contest-awd-challenge-deployment-mode`
- `contest-awd-challenge-readiness`
- `contest-awd-challenge-prev-page`
- `contest-awd-challenge-next-page`

要求：

- loading 时保留旧列表但显示同步状态。
- 空态区分“没有 AWD 题目”和“筛选无结果”。
- 加载失败显示错误文本和重试按钮。

- [ ] **Step 3: 增强表格列**

列建议：

- 名称 + slug
- 分类
- 难度
- 服务类型
- 部署方式
- 就绪状态
- 最近验证
- 选择

就绪状态映射：

```ts
pending -> 待验证
passed -> 已就绪
failed -> 未通过
```

- [ ] **Step 4: 新增默认设置区**

AWD create 模式下显示：

- `contest-awd-service-points`
- `contest-awd-service-order`
- `contest-awd-service-visibility`

这三个字段复用现有 `points/order/is_visible` 表单状态。多选创建时顺序按起始顺序递增。

- [ ] **Step 5: 跑面板测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120 npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts
```

预期：PASS。

## Task 4: 改善批量关联的部分成功处理

**Files:**
- Modify: `code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue`
- Modify: `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`
- Modify: `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`

- [ ] **Step 1: 写失败测试，覆盖批量关联部分失败**

场景：

- 选择 3 个 AWD 题目。
- 第 1、3 个创建成功，第 2 个创建失败。
- 期望 toast 提示部分成功，弹层保持打开或明确提示失败项。
- 期望成功项已经刷新进题目池。

断言示例：

```ts
expect(contestApiMocks.createContestAWDService).toHaveBeenCalledTimes(3)
expect(toastMocks.error).toHaveBeenCalledWith(expect.stringContaining('1 个关联失败'))
expect(wrapper.text()).toContain('Bank Service')
```

- [ ] **Step 2: 实现 `Promise.allSettled` 批量保存**

替换串行 `for await`：

```ts
const results = await Promise.allSettled(
  awdChallengeIds.map((awdChallengeId, index) =>
    createContestAWDService(props.contestId, {
      awd_challenge_id: awdChallengeId,
      points: payload.points,
      order: payload.order + index,
      is_visible: payload.is_visible,
    })
  )
)
```

处理策略：

- 全部成功：toast success，关闭弹层，刷新列表。
- 部分成功：toast warning/error，刷新列表，弹层保留并取消已成功项选择。
- 全部失败：toast error，弹层保留。

- [ ] **Step 3: 为失败项生成可读摘要**

根据 `awdChallengeOptions` 映射名称：

```ts
const failedNames = failedIds.map((id) => catalogById.get(String(id))?.name || `AWD #${id}`)
```

- [ ] **Step 4: 跑相关测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120 npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts
```

预期：PASS。

## Task 5: 清理 AWD 前端旧 template 命名

**Files:**
- Modify: `code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue`
- Modify: `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
- Modify: `code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue`
- Modify: `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- Modify: `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts`
- Modify: `code/frontend/src/views/platform/__tests__/ContestEdit.test.ts`
- Modify: `code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts`

- [ ] **Step 1: 批量前先列出旧命名，不直接全局替换**

运行：

```bash
cd /home/azhi/workspace/projects/ctf
rg -n "awd-template|contest-template|templateTable|template-list|模板" code/frontend/src/components/platform/awd-service code/frontend/src/components/platform/contest code/frontend/src/views/platform/__tests__/ContestEdit.test.ts code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts
```

检查每个命中是否属于 AWD 题库旧命名。不要改环境拓扑相关 `TopologyTemplate*`，那是另一套拓扑草稿能力。

- [ ] **Step 2: 改 `ContestChallengeEditorDialog.vue` 的 id/class**

映射建议：

- `contest-template-list` -> `contest-awd-challenge-list`
- `contest-template-option-*` -> `contest-awd-challenge-option-*`
- `contest-template-name-*` -> `contest-awd-challenge-name-*`
- `contest-template-table-*` -> `contest-awd-challenge-table-*`

- [ ] **Step 3: 改 AWD 题库页面 class/id**

映射建议：

- `awd-template-library-*` -> `awd-challenge-library-*`
- `awd-template-table-*` -> `awd-challenge-table-*`
- `awd-template-import-*` -> `awd-challenge-import-*`
- `awd-template-dialog-*` -> `awd-challenge-dialog-*`

用户可见文案保留“AWD 题目库 / AWD 题目包”，不写“模板”。

- [ ] **Step 4: 更新测试断言**

替换所有旧 id/class 断言，保留业务断言：

```ts
expect(wrapper.text()).toContain('AWD 题目库')
expect(wrapper.text()).not.toContain('创建模板')
expect(wrapper.find('#awd-challenge-dialog-submit').exists()).toBe(true)
```

- [ ] **Step 5: 跑命名清理相关测试**

运行：

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120 npm run test:run -- src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts
```

预期：PASS。

## Task 6: 最小全链路验证

**Files:**
- Modify: `docs/superpowers/plans/2026-04-29-awd-challenge-picker-optimization.md`

- [ ] **Step 1: 运行 AWD 题库与竞赛编辑相关测试**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 180 npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts src/composables/__tests__/useContestAwdChallengePicker.test.ts
```

预期：PASS。

- [ ] **Step 2: 运行前端类型检查**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 180 npm run typecheck
```

预期：PASS。

- [ ] **Step 3: 搜索确认 AWD 题库旧 template 命名已清理**

```bash
cd /home/azhi/workspace/projects/ctf
rg -n "awd-template|contest-template|AWD.*模板|模板.*AWD" code/frontend/src/components/platform/awd-service code/frontend/src/components/platform/contest code/frontend/src/views/platform/__tests__/ContestEdit.test.ts code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts
```

预期：没有 AWD 题库相关残留。若命中 `TopologyTemplate*`，不在本计划范围内。

- [ ] **Step 4: 更新计划执行状态**

把已完成步骤勾选，记录实际验证命令结果。

- [ ] **Step 5: 提交**

```bash
cd /home/azhi/workspace/projects/ctf
git add code/frontend/src/components/platform/contest/ContestChallengeOrchestrationPanel.vue code/frontend/src/components/platform/contest/ContestChallengeEditorDialog.vue code/frontend/src/composables/useContestAwdChallengePicker.ts code/frontend/src/composables/__tests__/useContestAwdChallengePicker.test.ts code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue code/frontend/src/components/platform/awd-service/AWDChallengeEditorDialog.vue code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts code/frontend/src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts code/frontend/src/views/platform/__tests__/ContestEdit.test.ts docs/superpowers/plans/2026-04-29-awd-challenge-picker-optimization.md
git commit -m "feat(awd): 优化赛事 AWD 题库选题"
```

## Risks

- 如果后端 `listAdminAwdChallenges` 不支持 `readiness_status`、`category`、`difficulty`、`deployment_mode` 查询参数，前端只能先发参数，后端忽略时筛选无效。实施时需要用 API 测试或联调确认。
- 批量部分成功策略可能需要产品取舍：保留弹层并取消成功项选择是当前推荐，但如果后端不允许重复创建，刷新后过滤已关联项也能避免重复提交。
- `template` 命名清理不要误伤拓扑环境模板相关组件，例如 `TopologyTemplateSidePanel`。

## Verification Commands

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 180 npm run test:run -- src/components/platform/__tests__/ContestChallengeOrchestrationPanel.test.ts src/views/platform/__tests__/ContestEdit.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts src/composables/__tests__/useContestAwdChallengePicker.test.ts
timeout 180 npm run typecheck
```
