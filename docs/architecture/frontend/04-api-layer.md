# 前端 API 层设计

> 对应：../backend/04-api-design.md
>
> **统一接口契约（强制）**：前后端联调时，接口契约以 `ctf/docs/contracts/openapi-v1.yaml`（机器可读）与 `ctf/docs/contracts/api-contract-v1.md`（说明性文档）保持一致为准；本文只描述前端实现策略与模块划分。

---

## 1. Axios 实例与拦截器

### 1.1 基础配置

```ts
// api/request.ts
const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 15000,
  // 若浏览器与 API 不同源，需要开启 withCredentials 以携带 session cookie
  // withCredentials: true,
  headers: { 'Content-Type': 'application/json' }
})
```

### 1.2 请求拦截器

```
请求发出前:
  ├─ 浏览器自动携带 session cookie
  └─ 公开接口与受保护接口共用同一请求实例
```

### 1.3 响应拦截器

```
响应到达后:
  ├─ HTTP 2xx + code === 0 → 返回 data 字段
  ├─ HTTP 401 → 清理登录态并跳转 /login
  ├─ HTTP 429 → Toast "请求过于频繁" + 读取 Retry-After
  ├─ HTTP 4xx/5xx → 通过 errorMap 映射错误码为中文提示 → Toast
  └─ 网络错误 → Toast "网络连接失败"
```

### 1.4 401 处理

关键实现：认证已切到服务端 session，前端不再维护 token，也不做 `/auth/refresh` 重试链路；401 统一按登录态失效处理。

```ts
// 401 处理
if (error.response?.status === 401) {
  authStore.logout()
  router.push('/login')
  return Promise.reject(error)
}
```

> 约束：前端禁止将后端返回的 `message` 当作 HTML 渲染；Toast/提示只显示文本，并优先使用 `errorMap`/通用文案（必要时附带 `request_id` 便于排障）。

---

## 2. API 模块划分

每个模块对应一个文件，导出具名函数，与后端模块一一对应。

### 2.1 认证模块 `api/auth.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `login(data)` | POST | `/auth/login` |
| `register(data)` | POST | `/auth/register` |
| `logout()` | POST | `/auth/logout` |
| `getProfile()` | GET | `/auth/profile` |
| `changePassword(data)` | PUT | `/auth/password` |
| `getWsTicket()` | POST | `/auth/ws-ticket` |

说明：认证采用 HttpOnly session cookie；同源无需额外配置；不同源需 `withCredentials` + 后端 CORS 允许。

### 2.2 靶场模块 `api/challenge.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getChallenges(params)` | GET | `/challenges` |
| `getChallengeDetail(id)` | GET | `/challenges/:id` |
| `getChallengeWriteup(id)` | GET | `/challenges/:id/writeup` |
| `upsertMyChallengeWriteup(id, data)` | POST | `/challenges/:id/writeup-submissions` |
| `getMyChallengeWriteup(id)` | GET | `/challenges/:id/writeup-submissions/me` |
| `submitFlag(id, flag)` | POST | `/challenges/:id/submissions` |
| `unlockHint(id, level)` | POST | `/challenges/:id/hints/:level/unlock` |
| `createInstance(id)` | POST | `/challenges/:id/instances` |

### 2.3 实例模块 `api/instance.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getMyInstances()` | GET | `/instances` |
| `destroyInstance(id)` | DELETE | `/instances/:id` |
| `extendInstance(id)` | POST | `/instances/:id/extend` |

### 2.4 竞赛模块 `api/contest.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getContests(params)` | GET | `/contests` |
| `getContestDetail(id)` | GET | `/contests/:id` |
| `registerContest(id)` | POST | `/contests/:id/register` |
| `getContestChallenges(id)` | GET | `/contests/:id/challenges` |
| `submitContestFlag(contestId, contestChallengeId, flag)` | POST | `/contests/:id/challenges/:cid/submissions` |
| `getScoreboard(id, params)` | GET | `/contests/:id/scoreboard` |
| `getAnnouncements(id)` | GET | `/contests/:id/announcements` |
| `createTeam(cid, data)` | POST | `/contests/:id/teams` |
| `joinTeam(cid, tid)` | POST | `/contests/:id/teams/:tid/join` |
| `getMyProgress(id)` | GET | `/contests/:id/my-progress` |

### 2.5 评估模块 `api/assessment.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getSkillProfile()` | GET | `/users/me/skill-profile` |
| `getRecommendations()` | GET | `/users/me/recommendations` |
| `getMyProgress()` | GET | `/users/me/progress` |
| `getMyTimeline()` | GET | `/users/me/timeline` |
| `exportPersonalReport(data)` | POST | `/reports/personal` |

### 2.6 通知模块 `api/notification.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getNotifications(params)` | GET | `/notifications` |
| `markAsRead(id)` | PUT | `/notifications/:id/read` |

### 2.7 教师模块 `api/teacher.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getClasses()` | GET | `/teacher/classes` |
| `getClassStudents(name)` | GET | `/teacher/classes/:name/students` |
| `getStudentProgress(id)` | GET | `/teacher/students/:id/progress` |
| `getStudentEvidence(id)` | GET | `/teacher/students/:id/evidence` |
| `getTeacherWriteupSubmissions(params)` | GET | `/teacher/writeup-submissions` |
| `getTeacherWriteupSubmission(id)` | GET | `/teacher/writeup-submissions/:id` |
| `reviewTeacherWriteupSubmission(id, data)` | PUT | `/teacher/writeup-submissions/:id/review` |
| `exportClassReport(data)` | POST | `/reports/class` |

### 2.8 管理后台 `api/admin.ts`

| 函数 | 方法 | 路径 |
|------|------|------|
| `getDashboard()` | GET | `/admin/dashboard` |
| `getUsers(params)` | GET | `/admin/users` |
| `createUser(data)` | POST | `/admin/users` |
| `updateUser(id, data)` | PUT | `/admin/users/:id` |
| `deleteUser(id)` | DELETE | `/admin/users/:id` |
| `importUsers(file)` | POST | `/admin/users/import` |
| `getChallenges(params)` | GET | `/admin/challenges` |
| `createChallenge(data)` | POST | `/admin/challenges` |
| `updateChallenge(id, data)` | PUT | `/admin/challenges/:id` |
| `deleteChallenge(id)` | DELETE | `/admin/challenges/:id` |
| `getImages(params)` | GET | `/admin/images` |
| `createImage(data)` | POST | `/admin/images` |
| `deleteImage(id)` | DELETE | `/admin/images/:id` |
| `getAuditLogs(params)` | GET | `/admin/audit-logs` |
| `getContests(params)` | GET | `/admin/contests` |
| `createContest(data)` | POST | `/admin/contests` |
| `updateContest(id, data)` | PUT | `/admin/contests/:id` |
| `deleteContest(id)` | DELETE | `/admin/contests/:id` |

---

## 3. 错误码映射

```ts
// utils/errorMap.ts
// 后端错误码 → 前端用户友好提示
// 仅覆盖需要特殊提示的错误码，其余使用通用失败文案（可附带 request_id）
const ERROR_MAP = {
  11001: '用户名或密码错误',
  11010: '登录失败次数过多，请稍后再试',
  13002: '实例数量已达上限，请先销毁已有实例',
  13003: 'Flag 错误，请检查后重试',
  13004: '提交过于频繁，请稍后再试',
  14003: '队伍人数已满',
  14008: '邀请码无效',
}
```

优先使用 `ERROR_MAP` 中的提示；未命中则使用通用失败文案（可附带 `request_id` 便于定位）。后端返回的 `message` 仅作为调试信息输出到控制台，不作为用户提示直接透传（避免泄露内部细节）。
