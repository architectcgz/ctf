# CTF 前端代码 Review（contest-detail 第 1 轮）：竞赛详情页功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | contest-detail |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 2b3fb49，2 个文件，+301 行 / -5 行 |
| 变更概述 | 实现竞赛详情页核心功能：队伍管理、倒计时、题目列表、Flag 提交 |
| 审查基准 | 项目 CLAUDE.md 前端规范、Vue Composition API 最佳实践 |
| 审查日期 | 2026-03-04 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 定时器未正确清理，存在内存泄漏风险
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:175-177`
- **问题描述**：`startCountdown()` 使用 `window.setInterval` 创建定时器，但在组件卸载时只清理了 `timer` 变量，未考虑以下场景：
  1. 如果 `contest.value` 为 null，函数提前返回，但旧的 `timer` 未清理
  2. 多次调用 `startCountdown()` 会创建多个定时器，但只保留最后一个引用
  3. 倒计时结束后虽然清理了定时器，但如果用户在倒计时结束前离开页面，定时器仍会继续运行
- **影响范围/风险**：内存泄漏、CPU 占用、可能导致页面卡顿
- **修正建议**：
```typescript
function startCountdown() {
  // 先清理旧定时器
  if (timer) {
    clearInterval(timer)
    timer = null
  }

  if (!contest.value) return

  timer = window.setInterval(() => {
    if (!contest.value) {
      if (timer) clearInterval(timer)
      timer = null
      return
    }

    const now = Date.now()
    const start = new Date(contest.value.starts_at).getTime()
    const end = new Date(contest.value.ends_at).getTime()

    if (now < start) {
      countdown.value = `距离开始: ${formatDuration(start - now)}`
    } else if (now < end) {
      countdown.value = `距离结束: ${formatDuration(end - now)}`
    } else {
      countdown.value = ''
      if (timer) {
        clearInterval(timer)
        timer = null
      }
    }
  }, 1000)
}
```

#### [H2] joinTeam API 调用参数错误
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:252`
- **问题描述**：`joinTeam(contest.value.id, '', inviteCode.value.trim())` 第二个参数 `teamId` 传入空字符串，但根据 API 定义（`contest.ts:53-59`），该参数应该是有效的 team ID，用于构造 URL `/contests/{contestId}/teams/{teamId}/join`
- **影响范围/风险**：API 调用失败，用户无法加入队伍，URL 会变成 `/contests/xxx/teams//join`（双斜杠）
- **修正建议**：
  - 方案 1：如果后端支持通过邀请码直接加入（不需要 teamId），应修改 API 定义为 `joinTeam(contestId: string, code: string)`，URL 改为 `/contests/{contestId}/join`
  - 方案 2：如果必须提供 teamId，前端需要先通过邀请码查询 teamId，再调用 joinTeam

#### [H3] 缺少错误提示，用户体验差
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:217-221, 237-241, 252-257, 264-269`
- **问题描述**：所有异步操作（submitFlag、createTeamAction、joinTeamAction、kickMember）的 catch 块只有 `console.error(err)`，用户看不到任何错误提示
- **影响范围/风险**：
  - Flag 提交失败时用户不知道原因（网络错误、权限不足、格式错误等）
  - 创建/加入队伍失败时没有反馈
  - 踢出成员失败时没有提示
- **修正建议**：使用 Toast/Message 组件显示错误信息
```typescript
import { ElMessage } from 'element-plus' // 或项目使用的 UI 库

async function submitFlag() {
  // ...
  try {
    const result = await submitContestFlag(...)
    // ...
  } catch (err) {
    console.error(err)
    ElMessage.error(err instanceof Error ? err.message : '提交失败，请稍后重试')
  } finally {
    submitting.value = false
  }
}
```

#### [H4] 缺少输入校验，可能导致无效请求
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:210, 245, 252`
- **问题描述**：
  1. `flagInput` 只检查 `trim()` 后非空，未校验格式（如是否符合 `flag{...}` 格式）
  2. `teamName` 未限制长度，可能超出后端限制
  3. `inviteCode` 未校验格式（如长度、字符集）
- **影响范围/风险**：无效请求浪费网络资源，后端返回错误但前端未提前拦截
- **修正建议**：
```typescript
async function submitFlag() {
  const flag = flagInput.value.trim()
  if (!flag) {
    ElMessage.warning('请输入 Flag')
    return
  }
  if (flag.length < 5 || flag.length > 200) {
    ElMessage.warning('Flag 长度应在 5-200 字符之间')
    return
  }
  // ...
}

async function createTeamAction() {
  const name = teamName.value.trim()
  if (!name) {
    ElMessage.warning('请输入队伍名称')
    return
  }
  if (name.length < 2 || name.length > 50) {
    ElMessage.warning('队伍名称长度应在 2-50 字符之间')
    return
  }
  // ...
}
```

### 🟡 中优先级

#### [M1] 类型定义不完整，缺少 null 安全检查
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:145`
- **问题描述**：`isCaptain` computed 依赖 `authStore.user`，但未检查 `authStore.user` 是否为 null/undefined，如果用户未登录会导致运行时错误
- **影响范围/风险**：未登录用户访问页面时可能报错
- **修正建议**：
```typescript
const isCaptain = computed(() =>
  team.value &&
  authStore.user &&
  team.value.captain_user_id === authStore.user.id
)
```

#### [M2] API 返回类型与实际使用不一致
- **文件**：`code/frontend/src/api/contest.ts:65`
- **问题描述**：`getMyTeam` 返回类型为 `TeamData | null`，但在 `ContestDetail.vue:158` 中使用 `.catch(() => null)` 处理错误，这意味着：
  1. 如果后端返回 404（用户未加入队伍），会被 catch 捕获并返回 null
  2. 如果后端返回 200 + null，也会得到 null
  3. 两种情况无法区分，且 catch 会吞掉所有错误（包括网络错误）
- **影响范围/风险**：网络错误被静默处理，用户看不到错误提示
- **修正建议**：
```typescript
// API 层：明确返回 null 表示未加入队伍
export async function getMyTeam(contestId: string): Promise<TeamData | null> {
  try {
    return await request<TeamData>({
      method: 'GET',
      url: `/contests/${encodeURIComponent(contestId)}/my-team`
    })
  } catch (err: any) {
    if (err.response?.status === 404) {
      return null // 用户未加入队伍
    }
    throw err // 其他错误继续抛出
  }
}

// 组件层：只处理预期的 null，其他错误让外层处理
const teamData = await getMyTeam(contestId)
team.value = teamData
```

#### [M3] 硬编码的样式类名，缺少主题变量一致性
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:88, 228`
- **问题描述**：
  1. 第 88 行使用 `text-red-500` 硬编码红色
  2. 第 228 行使用 `bg-red-500/10 text-red-500` 硬编码错误提示颜色
  3. 其他地方使用 CSS 变量 `var(--color-primary)`，风格不统一
- **影响范围/风险**：主题切换时颜色不一致，维护困难
- **修正建议**：统一使用 CSS 变量或 Tailwind 主题配置
```vue
<!-- 使用 CSS 变量 -->
<button class="text-[var(--color-danger)] hover:underline">踢出</button>

<!-- 或在 tailwind.config.js 中配置 -->
<button class="text-danger-500 hover:underline">踢出</button>
```

#### [M4] 缺少加载状态，用户体验不佳
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:237-241, 245-250, 252-257, 264-269`
- **问题描述**：
  1. `createTeamAction`、`joinTeamAction`、`kickMember` 没有 loading 状态
  2. 用户点击按钮后无法知道请求是否正在进行
  3. 可能导致重复点击，发送多个请求
- **影响范围/风险**：用户体验差，可能产生重复请求
- **修正建议**：
```typescript
const creatingTeam = ref(false)
const joiningTeam = ref(false)

async function createTeamAction() {
  if (creatingTeam.value) return
  creatingTeam.value = true
  try {
    // ...
  } finally {
    creatingTeam.value = false
  }
}

// 模板中
<button @click="createTeamAction" :disabled="creatingTeam" class="...">
  {{ creatingTeam ? '创建中...' : '创建' }}
</button>
```

#### [M5] 弹窗关闭逻辑不完整
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:233-242, 245-258`
- **问题描述**：
  1. 弹窗使用 `@click.self` 关闭，但未清空输入框内容
  2. 用户输入一半后点击外部关闭，再次打开弹窗时仍显示上次的内容
  3. 创建/加入成功后清空了输入框，但失败时未清空
- **影响范围/风险**：用户体验不一致
- **修正建议**：
```typescript
function closeCreateTeam() {
  showCreateTeam.value = false
  teamName.value = ''
}

function closeJoinTeam() {
  showJoinTeam.value = false
  inviteCode.value = ''
}

// 模板中
<div @click.self="closeCreateTeam" class="...">
  <button @click="closeCreateTeam">取消</button>
</div>
```

#### [M6] 题目选中状态未持久化
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:203-207`
- **问题描述**：用户选中题目后，如果刷新页面或切换路由再返回，选中状态丢失
- **影响范围/风险**：用户体验不佳，需要重新选择题目
- **修正建议**：使用 URL query 参数或 sessionStorage 保存选中的题目 ID
```typescript
// 从 URL 恢复选中状态
onMounted(async () => {
  // ...
  const selectedId = route.query.challenge as string
  if (selectedId) {
    selectedChallenge.value = challenges.value.find(c => c.id === selectedId) || null
  }
})

// 选中题目时更新 URL
function selectChallenge(chal: ContestChallengeItem) {
  selectedChallenge.value = chal
  router.replace({ query: { ...route.query, challenge: chal.id } })
  // ...
}
```

### 🟢 低优先级

#### [L1] 时间格式化函数可复用性差
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:289-291`
- **问题描述**：`formatTime` 函数硬编码了 `zh-CN` 和特定的日期格式，如果其他组件需要不同格式需要重复编写
- **影响范围/风险**：代码重复，维护成本高
- **修正建议**：提取到 `@/utils/date.ts` 工具文件
```typescript
// utils/date.ts
export function formatDateTime(time: string | Date, locale = 'zh-CN'): string {
  return new Date(time).toLocaleString(locale, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
```

#### [L2] 魔法数字未提取为常量
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:175`
- **问题描述**：倒计时间隔 `1000` 毫秒硬编码在代码中
- **影响范围/风险**：如果需要调整刷新频率，需要搜索代码修改
- **修正建议**：
```typescript
const COUNTDOWN_INTERVAL = 1000 // 1 秒

timer = window.setInterval(() => {
  // ...
}, COUNTDOWN_INTERVAL)
```

#### [L3] 状态标签映射可优化
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:272-287`
- **问题描述**：`getStatusLabel`、`getModeLabel`、`getStatusBadgeClass` 三个函数都使用 Record 映射，但每次调用都会创建新对象
- **影响范围/风险**：轻微的性能损耗
- **修正建议**：提取为模块级常量
```typescript
const STATUS_LABELS: Record<ContestStatus, string> = {
  draft: '草稿', published: '已发布', registering: '报名中',
  running: '进行中', frozen: '已冻结', ended: '已结束',
  cancelled: '已取消', archived: '已归档'
}

function getStatusLabel(status: ContestStatus): string {
  return STATUS_LABELS[status] || status
}
```

#### [L4] 缺少空状态提示优化
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:104`
- **问题描述**：题目列表为空时只显示"暂无题目"，可以提供更友好的提示（如"竞赛尚未发布题目，请稍后查看"）
- **影响范围/风险**：用户体验可优化
- **修正建议**：根据竞赛状态显示不同提示
```vue
<div v-if="challenges.length === 0" class="mt-4 text-center text-[var(--color-text-muted)]">
  <template v-if="contest.status === 'draft'">竞赛尚未发布</template>
  <template v-else-if="contest.status === 'registering'">题目将在竞赛开始后公布</template>
  <template v-else>暂无题目</template>
</div>
```

#### [L5] 缺少键盘快捷键支持
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue:119`
- **问题描述**：Flag 输入框支持 Enter 提交，但弹窗中的输入框（队伍名称、邀请码）不支持
- **影响范围/风险**：用户体验可优化
- **修正建议**：
```vue
<input
  v-model="teamName"
  @keyup.enter="createTeamAction"
  placeholder="队伍名称"
  class="..."
/>
```

#### [L6] 组件可拆分，提升可维护性
- **文件**：`code/frontend/src/views/contests/ContestDetail.vue`（整体）
- **问题描述**：单文件组件 294 行，包含竞赛信息、队伍管理、题目列表、Flag 提交、两个弹窗，职责较多
- **影响范围/风险**：后续维护和测试难度增加
- **修正建议**：拆分为子组件
  - `ContestInfo.vue`：竞赛信息和倒计时
  - `TeamManagement.vue`：队伍管理
  - `ChallengeList.vue`：题目列表
  - `FlagSubmission.vue`：Flag 提交
  - `TeamDialog.vue`：创建/加入队伍弹窗

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 6 |
| 🟢 低 | 6 |
| 合计 | 16 |

## 总体评价

本次变更实现了竞赛详情页的核心功能，代码结构清晰，使用了 Vue 3 Composition API 和 TypeScript，整体质量良好。主要问题集中在：

1. **错误处理和用户反馈不足**：所有异步操作缺少错误提示，用户体验较差（H3、M4）
2. **资源管理存在风险**：定时器清理逻辑不完整，可能导致内存泄漏（H1）
3. **API 调用存在明显错误**：joinTeam 参数错误会导致功能无法使用（H2）
4. **输入校验缺失**：未对用户输入进行前端校验，可能产生无效请求（H4）

建议优先修复 4 个高优先级问题，确保功能可用性和系统稳定性。中优先级问题主要影响用户体验和代码质量，建议在下一轮迭代中修复。低优先级问题可在后续重构时优化。
