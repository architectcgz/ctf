# CTF 前端试探型彩蛋 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不影响真实业务流程的前提下，为登录页、错误页、通知页和题目详情页补齐一组会话内本地触发的轻量试探型彩蛋。

**Architecture:** 新增一个仅负责本地计数、阈值判断和一次性信号消费的 `useProbeEasterEggs` composable，各页面只接入自己的触发器和局部展示，不把彩蛋状态混入业务 store 或 API 流程。所有交互都走前端本地同步链路，异常时静默降级，不阻断主功能。

**Tech Stack:** Vue 3, TypeScript, Vue Test Utils, Vitest, sessionStorage, existing ui-btn / workspace shell styles

---

## Execution Notes

- 先写 RED 测试，再写最小实现，不预写生产代码。
- 本轮默认直接在当前 `main` 工作区执行；`ctf/AGENTS.md` 已明确前端 UI 类工作优先直接在主工作区修改。
- 当前工作区已有未提交改动：
  - `code/frontend/src/components.d.ts`
  - `code/frontend/src/components/auth/AuthEntryShell.vue`
  - `code/frontend/src/views/auth/LoginView.vue`
- 实现时必须在理解这些现有改动的基础上继续追加，不回退、不覆盖。
- 验证命令全部串行执行，并为每条命令设置明确超时。

## Planned File Map

### Docs

- Create: `docs/superpowers/plans/2026-04-22-probe-easter-eggs-implementation.md`
- Reference: `docs/architecture/features/2026-04-22-probe-easter-eggs-design.md`

### Frontend

- Create: `code/frontend/src/composables/useProbeEasterEggs.ts`
- Create: `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- Modify: `code/frontend/src/views/auth/LoginView.vue`
- Modify if needed: `code/frontend/src/components/auth/AuthEntryShell.vue`
- Modify: `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
- Modify: `code/frontend/src/components/errors/ErrorStatusShell.vue`
- Modify: `code/frontend/src/views/errors/__tests__/NotFoundView.test.ts`
- Modify: `code/frontend/src/views/notifications/NotificationList.vue`
- Modify: `code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- Modify: `code/frontend/src/views/notifications/NotificationDetail.vue`
- Modify: `code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts`
- Modify: `code/frontend/src/views/challenges/ChallengeDetail.vue`
- Modify: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

## Task 1: 为本地彩蛋状态模型补红测

**Files:**
- Create: `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`
- Test target: `code/frontend/src/composables/useProbeEasterEggs.ts`

- [ ] **Step 1: 写 session 计数与阈值红测**

新增至少这些用例：

- `达到阈值前不应产生彩蛋信号`
- `达到阈值后应返回一次性彩蛋信号`
- `同一会话重复触发同一彩蛋时不应重复返回未消费信号`

- [ ] **Step 2: 写存储与降级红测**

新增至少这些用例：

- `应从 sessionStorage 恢复已记录计数`
- `sessionStorage 不可用时应静默降级到内存态`

- [ ] **Step 3: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/composables/__tests__/useProbeEasterEggs.test.ts
```

Expected:

- 因 `useProbeEasterEggs.ts` 尚不存在或行为未实现而失败

## Task 2: 为登录页与错误页彩蛋补红测

**Files:**
- Modify: `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
- Modify: `code/frontend/src/views/errors/__tests__/NotFoundView.test.ts`

- [ ] **Step 1: 给登录页补控制台与品牌区触发红测**

新增至少这些用例：

- `应继续输出基础控制台提示并追加审计口吻提示`
- `连续点击品牌区后应出现轻提示且不影响表单提交`

至少断言：

- 控制台输出包含新增试探型文案
- 轻提示默认不可见，达到阈值后出现
- 触发彩蛋后提交登录仍只调用一次登录逻辑

- [ ] **Step 2: 给 404 页面补点击序列红测**

新增：

- `连续点击状态区域后应显示路径试探附注`

至少断言：

- 初始不渲染附注
- 达到阈值后附注出现
- 原有返回按钮和回首页链接仍存在

- [ ] **Step 3: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/views/auth/__tests__/LoginView.test.ts
timeout 120s npm run test:run -- src/views/errors/__tests__/NotFoundView.test.ts
```

Expected:

- 新增彩蛋断言失败，其余现有登录与错误页行为保持可跑

## Task 3: 为通知页与题目页彩蛋补红测

**Files:**
- Modify: `code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- Modify: `code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts`
- Modify: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

- [ ] **Step 1: 给通知列表补高频刷新红测**

新增：

- `短时间内连续刷新后应显示试探型提示且仍执行真实刷新`

至少断言：

- `getNotifications` 仍按真实流程被调用
- 达到阈值后出现轻提示

- [ ] **Step 2: 给通知详情补 ID 连击红测**

新增：

- `连续点击 ID 卡片后应短暂显示值守备注`

至少断言：

- 默认显示真实 ID
- 达到阈值后出现值守备注
- 正文与返回入口不受影响

- [ ] **Step 3: 给题目详情补非关键区触发红测**

新增：

- `点击分值侧栏达到阈值后应显示试探提示`

至少断言：

- 反馈只出现在局部区域
- 题目主内容、附件按钮、tab 文案仍正常存在

- [ ] **Step 4: 运行定向命令确认 RED**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/views/notifications/__tests__/NotificationList.test.ts
timeout 120s npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts
timeout 120s npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts
```

Expected:

- 新增彩蛋断言失败，现有通知与题目详情基础行为断言仍提供回归保护

## Task 4: 实现本地彩蛋 composable

**Files:**
- Create: `code/frontend/src/composables/useProbeEasterEggs.ts`
- Test: `code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts`

- [ ] **Step 1: 写最小可用的彩蛋状态模型**

至少实现：

- 彩蛋 key 的计数
- 阈值判定
- 一次性信号消费
- `sessionStorage` 持久化
- 存储失败时的内存降级

- [ ] **Step 2: 重跑 composable 测试至 GREEN**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/composables/__tests__/useProbeEasterEggs.test.ts
```

- [ ] **Step 3: 仅在绿灯后做必要收口**

允许的收口：

- 抽常量
- 压缩重复 key 判断
- 补充小注释说明状态结构

## Task 5: 接入登录页与错误页实现

**Files:**
- Modify: `code/frontend/src/views/auth/LoginView.vue`
- Modify if needed: `code/frontend/src/components/auth/AuthEntryShell.vue`
- Modify: `code/frontend/src/components/errors/ErrorStatusShell.vue`
- Test: `code/frontend/src/views/auth/__tests__/LoginView.test.ts`
- Test: `code/frontend/src/views/errors/__tests__/NotFoundView.test.ts`

- [ ] **Step 1: 在登录页接入控制台扩展与品牌区轻提示**

要求：

- 基于当前已存在的控制台文案继续扩展
- 品牌区可点击区域只放在视觉区，不侵入输入框和提交按钮
- 轻提示不遮挡表单，不参与布局主流

- [ ] **Step 2: 在错误页壳接入点击序列与附注展示**

要求：

- 点击目标限制在状态标题区域
- 附注为纯只读局部文案
- 不能影响返回与跳转动作

- [ ] **Step 3: 重跑登录页与错误页测试至 GREEN**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/views/auth/__tests__/LoginView.test.ts
timeout 120s npm run test:run -- src/views/errors/__tests__/NotFoundView.test.ts
```

## Task 6: 接入通知页与题目页实现

**Files:**
- Modify: `code/frontend/src/views/notifications/NotificationList.vue`
- Modify: `code/frontend/src/views/notifications/NotificationDetail.vue`
- Modify: `code/frontend/src/views/challenges/ChallengeDetail.vue`
- Test: `code/frontend/src/views/notifications/__tests__/NotificationList.test.ts`
- Test: `code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts`
- Test: `code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts`

- [ ] **Step 1: 给通知列表接入高频刷新计数**

要求：

- 真实 `refresh` 流程保持不变
- 彩蛋提示只在本地阈值达成后显示
- 不新增全局错误处理

- [ ] **Step 2: 给通知详情接入 ID 卡片翻面备注**

要求：

- 真实 ID 仍是默认视图
- 值守备注只短暂替换卡片局部内容
- 不影响 `link` 关联入口和返回列表

- [ ] **Step 3: 给题目详情接入非关键区轻反馈**

要求：

- 只挂在分值侧栏或同等级非关键区域
- 不接入下载、tab 切换或题解加载逻辑
- 提示自动消失，不持久占位

- [ ] **Step 4: 重跑通知页与题目页测试至 GREEN**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/views/notifications/__tests__/NotificationList.test.ts
timeout 120s npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts
timeout 120s npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts
```

## Task 7: 做最小充分回归与提交

**Files:**
- Review: `code/frontend/src/composables/useProbeEasterEggs.ts`
- Review: `code/frontend/src/views/auth/LoginView.vue`
- Review: `code/frontend/src/components/errors/ErrorStatusShell.vue`
- Review: `code/frontend/src/views/notifications/NotificationList.vue`
- Review: `code/frontend/src/views/notifications/NotificationDetail.vue`
- Review: `code/frontend/src/views/challenges/ChallengeDetail.vue`

- [ ] **Step 1: 串行跑本轮最小充分回归**

```bash
cd /home/azhi/workspace/projects/ctf/code/frontend
timeout 120s npm run test:run -- src/composables/__tests__/useProbeEasterEggs.test.ts
timeout 120s npm run test:run -- src/views/auth/__tests__/LoginView.test.ts
timeout 120s npm run test:run -- src/views/errors/__tests__/NotFoundView.test.ts
timeout 120s npm run test:run -- src/views/notifications/__tests__/NotificationList.test.ts
timeout 120s npm run test:run -- src/views/notifications/__tests__/NotificationDetail.test.ts
timeout 120s npm run test:run -- src/views/challenges/__tests__/ChallengeDetail.test.ts
timeout 120s npm run typecheck
```

- [ ] **Step 2: 手工检查用户工作区边界**

确认：

- 不回退 `components.d.ts`、`AuthEntryShell.vue`、`LoginView.vue` 中原有未提交改动
- 提交范围只包含本轮相关文件

- [ ] **Step 3: 提交本轮实现**

```bash
cd /home/azhi/workspace/projects/ctf
git add code/frontend/src/composables/useProbeEasterEggs.ts \
  code/frontend/src/composables/__tests__/useProbeEasterEggs.test.ts \
  code/frontend/src/views/auth/LoginView.vue \
  code/frontend/src/components/auth/AuthEntryShell.vue \
  code/frontend/src/views/auth/__tests__/LoginView.test.ts \
  code/frontend/src/components/errors/ErrorStatusShell.vue \
  code/frontend/src/views/errors/__tests__/NotFoundView.test.ts \
  code/frontend/src/views/notifications/NotificationList.vue \
  code/frontend/src/views/notifications/NotificationDetail.vue \
  code/frontend/src/views/notifications/__tests__/NotificationList.test.ts \
  code/frontend/src/views/notifications/__tests__/NotificationDetail.test.ts \
  code/frontend/src/views/challenges/ChallengeDetail.vue \
  code/frontend/src/views/challenges/__tests__/ChallengeDetail.test.ts \
  docs/superpowers/plans/2026-04-22-probe-easter-eggs-implementation.md
git commit -m "feat(前端): 增加试探型彩蛋"
```
