# CTF 平台接口契约（v1）

> 目的：把 **CTF 前端（`ctf/frontend`）当前会调用的接口** 与 **后端 API 设计（`ctf/docs/architecture/backend/04-api-design.md`）** 的返回结构统一成一份“契约”，作为联调与实现的唯一参考，避免后期因字段名/类型不一致返工。
>
> 机器可读版本：`ctf/docs/contracts/openapi-v1.yaml`（OpenAPI 3.0），应与本文保持一致。
>
> 最后更新：2026-03-10

---

## 0. 范围

本文件覆盖以下前端 API 模块（均以 `baseURL = /api/v1` 为前缀）：

- `ctf/frontend/src/api/auth.ts`
- `ctf/frontend/src/api/challenge.ts`
- `ctf/frontend/src/api/instance.ts`
- `ctf/frontend/src/api/contest.ts`
- `ctf/frontend/src/api/assessment.ts`
- `ctf/frontend/src/api/notification.ts`
- `ctf/frontend/src/api/teacher.ts`
- `ctf/frontend/src/api/admin.ts`
- WebSocket：`ctf/docs/architecture/backend/04-api-design.md` §6

不在上述前端模块中的后端接口（例如更多后台管理/教师接口）不在本契约详列；需要时应新增条目后再实现。

---

## 1. 统一响应 Envelope（强制）

### 1.1 Envelope 结构

所有 HTTP JSON 接口（含 `DELETE`）统一返回 Envelope（**不使用 204 无响应体**），格式如下：

```ts
export interface ApiEnvelope<T> {
  code: number              // 0=成功，非0=业务错误码（见 04-api-design.md §3）
  message: string           // 成功固定 "success"，失败为摘要信息（前端提示优先走 errorMap）
  data: T                   // 成功返回业务数据；失败时允许为 null 或“错误上下文对象”（例如 flag 错误仍返回 remaining_attempts）
  request_id: string        // 必填：用于日志追踪
  errors?: Array<{          // 可选：字段级校验错误
    field: string
    message: string
  }>
}
```

### 1.2 通用类型约定（强制）

```ts
export type ID = string                 // 统一：所有 id/xxx_id 字段均返回 string（避免 JS 超过 2^53 精度问题）
export type ISODateTime = string        // RFC3339/ISO8601，如 "2026-03-01T03:15:22Z"
export type UserRole = 'student' | 'teacher' | 'admin'

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
```

> 说明：当前前端部分接口使用 `{ items, total }`（见 `ctf/frontend/src/api/*.ts`），本契约统一为 `PageResult<T>` 的 `{ list, total, page, page_size }`。

---

## 2. 认证（`/auth/*`）

### 2.1 POST `/api/v1/auth/login`

`data`：

```ts
export interface AuthUser {
  id: ID
  username: string
  role: UserRole
  avatar?: string
  name?: string
  class_name?: string
}

export interface LoginData {
  access_token: string
  token_type: 'Bearer'
  expires_in: number        // 秒
  user: AuthUser
}
```

> 说明：账号在连续输错密码达到阈值后会被临时锁定；触发锁定的那次登录返回 `429`，锁定期内再次尝试返回 `403`。

### 2.2 POST `/api/v1/auth/register`

`data`：同 `LoginData`（注册成功即登录）。

### 2.3 POST `/api/v1/auth/refresh`

`data`：

```ts
export interface RefreshData {
  access_token: string
  token_type: 'Bearer'
  expires_in: number
}
```

> Refresh Token 由后端写入 HttpOnly Cookie；前端不落盘。

### 2.4 POST `/api/v1/auth/logout`

`data`：`null`

### 2.5 GET `/api/v1/auth/profile`

`data`：`AuthUser`

### 2.6 PUT `/api/v1/auth/password`

`data`：`null`

### 2.7 POST `/api/v1/auth/ws-ticket`

`data`：

```ts
export interface WsTicketData {
  ticket: string
  expires_at: ISODateTime
}
```

### 2.8 GET `/api/v1/auth/cas/status`

`data`：

```ts
export interface CasStatusData {
  provider: 'cas'
  enabled: boolean
  configured: boolean
  auto_provision: boolean
  login_path: string
  callback_path: string
}
```

> 说明：一期仅预留 CAS 接口层，默认 `enabled=false`；仅用于前端发现平台是否启用了 CAS 登录能力。

### 2.9 GET `/api/v1/auth/cas/login`

`data`：

```ts
export interface CasLoginData {
  provider: 'cas'
  redirect_url: string
  callback_url: string
}
```

> 说明：当 CAS 配置完整且已启用时，后端返回 CAS 登录跳转地址；前端可据此跳转到学校统一认证入口。

### 2.10 GET `/api/v1/auth/cas/callback`

`data`：预期为 `LoginData`

> 说明：后端会使用 `ticket + auth.cas.service_url` 调用 CAS `serviceValidate` 完成票据校验；校验成功后按 CAS 用户名映射平台 `username`。
> 若用户已存在，则自动回填 `name/email/class_name/student_no/teacher_no` 等资料，并清理已过期的登录锁定状态；若用户不存在，则按 `auth.cas.auto_provision` 决定自动创建 `student` 角色账号或返回 `403`。

---

## 3. 靶场（学员）（`/challenges/*`）

### 3.1 GET `/api/v1/challenges`（分页）

`data`：

```ts
export type ChallengeCategory = 'web' | 'pwn' | 'reverse' | 'crypto' | 'misc' | 'forensics'
export type ChallengeDifficulty = 'beginner' | 'easy' | 'medium' | 'hard' | 'hell'

export interface ChallengeListItem {
  id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  tags: string[]
  solved_count: number
  total_attempts: number
  is_solved: boolean
  points: number
  created_at: ISODateTime
}

export type ChallengeListData = PageResult<ChallengeListItem>
```

### 3.2 GET `/api/v1/challenges/:id`

`data`（字段来源：DB 设计 + 前端页面数据流；缺少示例，需后端实现前最终确认）：

```ts
export interface ChallengeHint {
  id: ID
  level: number
  title?: string
  cost_points?: number
  is_unlocked: boolean
  content?: string          // 未解锁时可省略或置空
}

export interface ChallengeDetailData {
  id: ID
  title: string
  description: string       // Markdown
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  tags: string[]
  points: number
  attachment_url?: string
  is_solved: boolean
  solved_at?: ISODateTime
  hints: ChallengeHint[]
}
```

### 3.3 POST `/api/v1/challenges/:id/instances`

`data`：

```ts
export type InstanceStatus = 'pending' | 'creating' | 'running' | 'expired' | 'destroying' | 'destroyed' | 'failed' | 'crashed'
export type FlagType = 'static' | 'dynamic' | 'regex'

export interface InstanceData {
  id: ID
  challenge_id: ID
  status: InstanceStatus
  access_url?: string
  ssh_info?: { host: string; port: number; username: string }
  flag_type: FlagType
  expires_at: ISODateTime
  remaining_extends: number
  created_at: ISODateTime

  // 排队/创建中可选字段（资源不足/排队机制）
  queue_position?: number
  eta_seconds?: number
  progress?: number         // 0~100
}
```

> 说明：若挑战绑定了带 `protocol` / `ports` 的链路策略，实例启动会按 `source_node_key -> target_node_key` 方向把规则下发到宿主机 `DOCKER-USER` 链；其中细粒度 `allow` 会进入 allow-list 模式，未命中的同方向流量会被兜底拒绝。

### 3.4 DELETE `/api/v1/instances/:id`

`data`：`null`

> 说明：若该实例是 AWD 队伍共享实例，则当前队伍成员均可销毁。

### 3.5 POST `/api/v1/instances/:id/extend`

`data`：

```ts
export interface InstanceExtendData {
  id: ID
  expires_at: ISODateTime
  remaining_extends: number
}
```

> 说明：若该实例是 AWD 队伍共享实例，则当前队伍成员均可续期。

### 3.5.1 POST `/api/v1/instances/:id/access`

`data`：

```ts
export interface InstanceAccessData {
  access_url: string
}
```

> 说明：
> - 返回的是平台代理入口地址，不再直接暴露容器真实访问入口给浏览器新窗口。
> - 浏览器首次打开该地址时，会以短时票据换取平台代理 Cookie；后续同窗口内对靶机的 HTTP 请求将继续经平台代理转发。
> - 平台会记录请求方法、路径、状态码和有限请求摘要，用于训练时间线和教学复盘。

### 3.6 GET `/api/v1/instances`

`data`：

```ts
export interface InstanceListItem extends InstanceData {
  challenge_title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
}

export type InstanceListData = InstanceListItem[]
```

> 说明：列表会返回“当前用户可见”的运行中实例；除个人练习实例外，也会包含当前用户在 AWD 竞赛中所属队伍的共享实例。

### 3.7 POST `/api/v1/challenges/:id/submissions`

`data`：

```ts
export interface SubmitFlagSuccessData {
  correct: true
  points_earned: number
  first_blood: boolean
  solved_at: ISODateTime
  challenge_progress: { total_challenges: number; solved_challenges: number }
}

export interface SubmitFlagFailureData {
  correct: false
  remaining_attempts?: number
}

export type SubmitFlagData = SubmitFlagSuccessData | SubmitFlagFailureData
```

> 约束：当 `correct=false` 时，HTTP 可为 400（业务错误），但仍允许 `data.correct=false` 返回“错误上下文”。前端提示文案优先走 `errorMap`。

### 3.8 POST `/api/v1/challenges/:id/hints/:level/unlock`

`data`（缺少示例，需确认最小返回）：

```ts
export interface UnlockHintData {
  hint: ChallengeHint
  points_spent?: number
  remaining_points?: number
}
```

### 3.9 GET `/api/v1/users/me/progress`

`data`（缺少示例，需确认）：

```ts
export interface MyProgressData {
  total_challenges: number
  solved_challenges: number
  by_category?: Record<string, { total: number; solved: number }>
  by_difficulty?: Record<string, { total: number; solved: number }>
}
```

### 3.10 GET `/api/v1/users/me/timeline`

`data`（缺少示例，需确认）：

```ts
export interface TimelineEvent {
  id: ID
  type: 'solve' | 'submit' | 'instance' | 'hint'
  title: string
  created_at: ISODateTime
  meta?: Record<string, unknown>
}

export type MyTimelineData = TimelineEvent[]
```

### 3.11 GET `/api/v1/challenges/:id/writeup`

`data`：

```ts
export type ChallengeWriteupVisibility = 'private' | 'public' | 'scheduled'

export interface ChallengeWriteupData {
  id: ID
  challenge_id: ID
  title: string
  content: string
  visibility: ChallengeWriteupVisibility
  release_at?: ISODateTime
  is_released: true
  requires_spoiler_warning: boolean
  created_at: ISODateTime
  updated_at: ISODateTime
}
```

> 说明：当前后端仅在 Writeup 已公开或已到定时公开时间时返回该资源；`requires_spoiler_warning=true` 表示学员尚未解出题目，前端应先展示剧透提示。

---

## 4. 竞赛（`/contests/*`）

### 4.1 GET `/api/v1/contests`（分页）

`data`（缺少示例，需确认字段集）：

```ts
export type ContestMode = 'jeopardy' | 'awd' | 'awd_plus' | 'king_of_hill'
export type ContestStatus = 'draft' | 'published' | 'registering' | 'running' | 'frozen' | 'ended' | 'cancelled' | 'archived'

export interface ContestListItem {
  id: ID
  title: string
  mode: ContestMode
  status: ContestStatus
  starts_at: ISODateTime
  ends_at: ISODateTime
  register_ends_at?: ISODateTime
}

export type ContestListData = PageResult<ContestListItem>
```

### 4.2 GET `/api/v1/contests/:id`

`data`（缺少示例，需确认字段集）：

```ts
export interface ContestDetailData extends ContestListItem {
  description?: string
  rules?: string            // Markdown
  team_size_limit?: number
  scoreboard_frozen?: boolean
}
```

### 4.3 POST `/api/v1/contests/:id/register`

`data`：`null`

> 说明：当前实现为学员自助报名后写入 `pending` 状态；若管理员已审核通过，则重复报名按幂等成功处理并保持 `approved`。被驳回后再次报名会重新进入 `pending`。竞赛组队、入队和提交 Flag 均要求报名状态为 `approved`。

### 4.4 GET `/api/v1/contests/:id/challenges`

`data`（缺少示例，需确认）：

```ts
export interface ContestChallengeItem {
  id: ID                   // contest_challenge_id（竞赛题目实例）
  challenge_id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  points: number
  solved_count: number
  is_solved: boolean
}

export type ContestChallengeListData = ContestChallengeItem[]
```

### 4.5 POST `/api/v1/contests/:id/challenges/:cid/submissions`

`data`：同 `SubmitFlagData`

> 说明：当前后端会在正确提交时按 `contest_score -> contest_challenges.points -> contest.base_score(兜底)` 作为基础分，结合 `solve_count`、`contest.min_score`、`contest.decay` 实时结算本次得分；首血仍按当前结算分追加 `contest.first_blood_bonus`。同一队伍同一题只允许首次正确提交计分；未组队参赛者仍按个人维度去重。当后续有新的正确提交出现时，系统会按最新 `solve_count` 回溯重算该题全部历史正确提交，并同步修正队伍总分与 live 排行榜。

### 4.6 POST `/api/v1/contests/:id/challenges/:cid/instances`

`data`：同 `InstanceData`

> 说明：
> - 当前要求竞赛已进入 `running/frozen` 且当前用户报名状态为 `approved`。
> - `mode=awd` 时必须已加入队伍，并按 `contest_id + team_id + challenge_id` 复用同队共享实例；同队成员再次启动会直接拿到同一实例。
> - 非 AWD 竞赛仍按用户维度创建竞赛实例，不影响练习模式实例。

### 4.6.1 POST `/api/v1/admin/contests/:id/awd/current-round/check`

`data`：

```ts
export interface AWDCheckerRunData {
  round: AWDRound
  services: AWDTeamService[]
}
```

> 说明：
> - 仅管理员可调用。
> - 会立即对当前运行中的 AWD 轮次执行一次服务检查，并刷新该轮各队伍各题的服务状态。
> - 适用于运维修复后手动复查，不需要等待后台调度周期。

### 4.7 GET `/api/v1/contests/:id/scoreboard`（分页）

`data`：

```ts
export interface ScoreboardRow {
  rank: number
  team_id: ID
  team_name: string
  score: number
  solved_count: number
  last_submission_at?: ISODateTime
}

export interface ContestScoreboardData {
  contest: { id: ID; title: string; status: ContestStatus; started_at: ISODateTime; ends_at: ISODateTime }
  scoreboard: PageResult<ScoreboardRow>
  frozen: boolean
}
```

### 4.8 GET `/api/v1/contests/:id/announcements`

`data`（缺少示例，需确认）：

```ts
export interface ContestAnnouncement {
  id: ID
  title: string
  content?: string
  created_at: ISODateTime
}

export type ContestAnnouncementListData = ContestAnnouncement[]
```

### 4.9 POST `/api/v1/contests/:id/teams`

`data`（缺少示例，需确认）：

```ts
export interface TeamData {
  id: ID
  name: string
  invite_code?: string      // 仅队长可见（如有）
  captain_user_id: ID
  members: Array<{ user_id: ID; username: string; joined_at: ISODateTime }>
}
```

### 4.10 POST `/api/v1/contests/:id/teams/:tid/join`

`data`：`null`

### 4.11 GET `/api/v1/contests/:id/my-progress`

`data`（缺少示例，需确认）：

```ts
export interface ContestMyProgressData {
  contest_id: ID
  team_id?: ID
  solved: Array<{ contest_challenge_id: ID; solved_at: ISODateTime; points_earned: number }>
}
```

---

## 5. 通知（`/notifications/*`）

### 5.1 GET `/api/v1/notifications`（分页）

`data`（缺少示例，需确认字段集）：

```ts
export type NotificationType = 'system' | 'contest' | 'challenge' | 'team'

export interface NotificationItem {
  id: ID
  type: NotificationType
  title: string
  content?: string
  level?: 'info' | 'success' | 'warning' | 'error'
  unread: boolean
  created_at: ISODateTime
}

export type NotificationListData = PageResult<NotificationItem>
```

### 5.2 PUT `/api/v1/notifications/:id/read`

`data`：`null`

---

## 6. 教师（`/teacher/*`）

### 6.1 GET `/api/v1/teacher/classes`

`data`（缺少示例，需确认）：

```ts
export interface TeacherClassItem {
  name: string
  student_count?: number
}

export type TeacherClassListData = TeacherClassItem[]
```

### 6.2 GET `/api/v1/teacher/classes/:name/students`

`data`（缺少示例，需确认）：

```ts
export interface TeacherStudentItem {
  id: ID
  username: string
  name?: string
  progress?: MyProgressData
}

export type TeacherStudentListData = TeacherStudentItem[]
```

### 6.3 GET `/api/v1/teacher/students/:id/progress`

`data`：`MyProgressData`（或包含更细颗粒度的 challenge 列表；需确认）。

---

## 7. 技能评估与报告（`/users/*` `/reports/*`）

### 7.1 GET `/api/v1/users/me/skill-profile`

`data`（缺少示例，需确认）：

```ts
export interface SkillDimensionScore {
  key: string              // 如 "web" / "crypto"
  name: string             // 展示名
  value: number            // 0~100
}

export interface SkillProfileData {
  dimensions: SkillDimensionScore[]
  updated_at?: ISODateTime
}
```

### 7.2 GET `/api/v1/users/me/recommendations`

`data`（缺少示例，需确认）：

```ts
export interface RecommendationItem {
  challenge_id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  reason: string
}

export type RecommendationData = RecommendationItem[]
```

### 7.3 POST `/api/v1/reports/personal`

`data`（基于需求：单人 ≤30s，可同步；也允许异步统一口径）：

```ts
export interface ReportExportData {
  report_id: ID
  status: 'ready' | 'processing'
  download_url?: string     // status=ready 时返回
  expires_at?: ISODateTime
}
```

### 7.4 POST `/api/v1/reports/class`

`data`：同 `ReportExportData`（大班级建议异步，status=processing；完成后通过通知推送下载地址）。

---

## 8. 管理后台（`/admin/*`）

> 仅覆盖当前前端 `ctf/frontend/src/api/admin.ts` 中实际调用的接口；后端设计文档中存在更多管理员接口，未在此展开。

### 8.1 GET `/api/v1/admin/dashboard`

`data`（缺少示例，需确认）：

```ts
export interface AdminDashboardData {
  online_instances: number
  container_health_rate: number    // 0~1 或 0~100（需确认）
  submissions_today: number
  submissions_error_rate: number   // 0~1 或 0~100（需确认）
  alerts?: Array<{ id: ID; title: string; level: string; created_at: ISODateTime }>
  recent_audit_logs?: Array<{ id: ID; action: string; actor: string; created_at: ISODateTime }>
}
```

### 8.2 GET `/api/v1/admin/users`（分页）

`data`：

```ts
export interface AdminUserListItem {
  id: ID
  username: string
  name?: string
  class_name?: string
  status: 'active' | 'inactive' | 'locked' | 'banned'
  roles: UserRole[]
  created_at: ISODateTime
}

export type AdminUserListData = PageResult<AdminUserListItem>
```

### 8.3 POST `/api/v1/admin/users` / PUT `/api/v1/admin/users/:id`

`data`（缺少示例，需确认）：

```ts
export interface AdminUserUpsertData {
  user: AdminUserListItem
}
```

### 8.4 DELETE `/api/v1/admin/users/:id`

`data`：`null`

### 8.5 POST `/api/v1/admin/users/import`

`data`（缺少示例，需确认）：

```ts
export interface AdminUserImportData {
  created: number
  updated: number
  failed: number
  errors?: Array<{ row: number; message: string }>
}
```

### 8.6 GET `/api/v1/admin/challenges`（分页）

`data`（缺少示例，需确认）：

```ts
export interface AdminChallengeListItem {
  id: ID
  title: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  status: 'draft' | 'review' | 'active' | 'archived'
  base_score: number
  solve_count: number
  created_at: ISODateTime
}

export type AdminChallengeListData = PageResult<AdminChallengeListItem>
```

### 8.7 POST `/api/v1/admin/challenges` / PUT `/api/v1/admin/challenges/:id`

`data`（缺少示例，需确认）：

```ts
export interface AdminChallengeUpsertData {
  challenge: AdminChallengeListItem
}
```

### 8.8 DELETE `/api/v1/admin/challenges/:id`

`data`：`null`

### 8.9 GET `/api/v1/admin/images`（分页）

`data`（缺少示例，需确认）：

```ts
export type ImageSourceType = 'registry' | 'dockerfile' | 'upload'
export type ImageStatus = 'pending' | 'building' | 'ready' | 'failed' | 'deprecated'

export interface AdminImageListItem {
  id: ID
  name: string
  tag: string
  source_type: ImageSourceType
  status: ImageStatus
  size_bytes?: number
  created_at: ISODateTime
}

export type AdminImageListData = PageResult<AdminImageListItem>
```

### 8.10 POST `/api/v1/admin/images`

`data`（缺少示例，需确认）：

```ts
export interface AdminImageCreateData {
  image: AdminImageListItem
}
```

### 8.11 DELETE `/api/v1/admin/images/:id`

`data`：`null`

### 8.12 GET `/api/v1/admin/audit-logs`（分页）

`data`（缺少示例，需确认）：

```ts
export interface AuditLogItem {
  id: ID
  action: string
  resource_type: string
  resource_id?: ID
  actor_user_id: ID
  actor_username: string
  ip?: string
  user_agent?: string
  created_at: ISODateTime
  detail?: Record<string, unknown>
}

export type AuditLogListData = PageResult<AuditLogItem>
```

### 8.13 GET `/api/v1/admin/contests`（分页）

`data`：`PageResult<ContestListItem>`（管理视图可扩展更多字段；需确认）。

### 8.14 GET `/api/v1/admin/challenges/:id/writeup` / PUT `/api/v1/admin/challenges/:id/writeup` / DELETE `/api/v1/admin/challenges/:id/writeup`

`data`：

```ts
export type ChallengeWriteupVisibility = 'private' | 'public' | 'scheduled'

export interface AdminChallengeWriteupData {
  id: ID
  challenge_id: ID
  title: string
  content: string
  visibility: ChallengeWriteupVisibility
  release_at?: ISODateTime
  created_by?: ID
  created_at: ISODateTime
  updated_at: ISODateTime
}
```

### 8.15 GET `/api/v1/admin/challenges/:id/topology` / PUT `/api/v1/admin/challenges/:id/topology` / DELETE `/api/v1/admin/challenges/:id/topology`

`data`：

```ts
export type TopologyTier = 'public' | 'service' | 'internal'
export type TopologyPolicyAction = 'allow' | 'deny'
export type TopologyPolicyProtocol = 'tcp' | 'udp' | 'any'

export interface TopologyNetworkData {
  key: string
  name: string
  cidr?: string
  internal?: boolean
}

export interface TopologyNodeData {
  key: string
  name: string
  image_id?: ID
  service_port?: number
  inject_flag?: boolean
  tier?: TopologyTier
  network_keys?: string[]
  env?: Record<string, string>
  resources?: {
    cpu_quota?: number
    memory_mb?: number
    pids_limit?: number
  }
}

export interface TopologyLinkData {
  from_node_key: string
  to_node_key: string
}

export interface TopologyTrafficPolicyData {
  source_node_key: string
  target_node_key: string
  action: TopologyPolicyAction
  protocol?: TopologyPolicyProtocol
  ports?: number[]
}

export interface ChallengeTopologyData {
  id: ID
  challenge_id: ID
  template_id?: ID
  entry_node_key: string
  networks?: TopologyNetworkData[]
  nodes: TopologyNodeData[]
  links?: TopologyLinkData[]
  policies?: TopologyTrafficPolicyData[]
  created_at: ISODateTime
  updated_at: ISODateTime
}
```

> 说明：当前后端已支持在拓扑/模板配置中描述多网络分段与节点间链路策略；未显式提供 `networks` 时，后端会自动补一个默认网络并把所有节点挂到该网络。实例启动时会按 `network_keys` 创建并挂载多个 Docker Network，对“无端口/协议限定”的 `policies` 做节点级粗粒度隔离，并对带 `protocol` / `ports` 的策略按 `source_node_key -> target_node_key` 方向下发细粒度 ACL。

### 8.16 GET `/api/v1/admin/environment-templates` / POST `/api/v1/admin/environment-templates` / GET `/api/v1/admin/environment-templates/:id` / PUT `/api/v1/admin/environment-templates/:id` / DELETE `/api/v1/admin/environment-templates/:id`

`data`：

```ts
export interface EnvironmentTemplateData {
  id: ID
  name: string
  description: string
  entry_node_key: string
  networks?: TopologyNetworkData[]
  nodes: TopologyNodeData[]
  links?: TopologyLinkData[]
  policies?: TopologyTrafficPolicyData[]
  usage_count: number
  created_at: ISODateTime
  updated_at: ISODateTime
}
```

> 说明：`TopologyTrafficPolicyData.protocol` 支持 `any` / `tcp` / `udp`，`ports` 表示目标端口列表；细粒度 `allow` 会把对应方向切换到 allow-list 模式，未命中的同方向流量会被拒绝。

### 8.17 GET `/api/v1/admin/contests/:id/announcements` / POST `/api/v1/admin/contests/:id/announcements` / DELETE `/api/v1/admin/contests/:id/announcements/:aid`

`data`：

```ts
export interface ContestAnnouncementData {
  id: ID
  title: string
  content?: string
  created_at: ISODateTime
}

export type ContestAnnouncementListData = ContestAnnouncementData[]
```

### 8.18 GET `/api/v1/admin/contests/:id/registrations` / PUT `/api/v1/admin/contests/:id/registrations/:rid`

`data`：

```ts
export type ContestRegistrationStatus = 'pending' | 'approved' | 'rejected'

export interface ContestRegistrationData {
  id: ID
  contest_id: ID
  user_id: ID
  username: string
  team_id?: ID
  status: ContestRegistrationStatus
  reviewed_by?: ID
  reviewed_at?: ISODateTime
  created_at: ISODateTime
  updated_at: ISODateTime
}

export type ContestRegistrationListData = PageResult<ContestRegistrationData>
```

`PUT` 请求体：

```ts
export interface ReviewContestRegistrationReq {
  status: 'approved' | 'rejected'
}
```

### 8.19 GET `/api/v1/admin/contests/:id/awd/rounds` / POST `/api/v1/admin/contests/:id/awd/rounds` / GET `/api/v1/admin/contests/:id/awd/rounds/:rid/services` / POST `/api/v1/admin/contests/:id/awd/rounds/:rid/services/check` / GET `/api/v1/admin/contests/:id/awd/rounds/:rid/attacks` / POST `/api/v1/admin/contests/:id/awd/rounds/:rid/attacks` / GET `/api/v1/admin/contests/:id/awd/rounds/:rid/summary`

`data`：

```ts
export type AwdRoundStatus = 'pending' | 'running' | 'finished'
export type AwdServiceStatus = 'up' | 'down' | 'compromised'
export type AwdAttackType = 'flag_capture' | 'service_exploit'

export interface AwdRoundData {
  id: ID
  contest_id: ID
  round_number: number
  status: AwdRoundStatus
  started_at?: ISODateTime
  ended_at?: ISODateTime
  attack_score: number
  defense_score: number
  created_at: ISODateTime
  updated_at: ISODateTime
}

export interface AwdTeamServiceData {
  id: ID
  round_id: ID
  team_id: ID
  team_name: string
  challenge_id: ID
  service_status: AwdServiceStatus
  check_result: Record<string, unknown>
  attack_received: number
  defense_score: number
  attack_score: number
  updated_at: ISODateTime
}

> 说明：
> - `attack_received` 表示该轮该服务被成功打穿的次数。
> - `attack_score` 表示攻击方从该服务累计获取的分值。
> - 成功攻击发生后，服务状态会被回写为 `compromised`，同时该服务当前轮防守分清零；后续若管理员手动触发或后台再次执行服务检查，状态可按最新检查结果恢复。

export interface AwdAttackLogData {
  id: ID
  round_id: ID
  attacker_team_id: ID
  attacker_team: string
  victim_team_id: ID
  victim_team: string
  challenge_id: ID
  attack_type: AwdAttackType
  submitted_flag?: string
  is_success: boolean
  score_gained: number
  created_at: ISODateTime
}

export interface AwdRoundSummaryItem {
  team_id: ID
  team_name: string
  service_up_count: number
  service_down_count: number
  service_compromised_count: number
  defense_score: number
  attack_score: number
  successful_attack_count: number
  successful_breach_count: number
  unique_attackers_against: number
  total_score: number
}

export interface AwdRoundSummaryData {
  round: AwdRoundData
  items: AwdRoundSummaryItem[]
}
```

> 说明：当前已补齐 AWD 轮次调度、最小版实例健康 Checker、真实容器 Flag 注入、基于当前轮 Flag 的真实攻击提交流程、短窗口的上一轮 Flag 兼容，以及队伍级共享实例模型。

### 8.19.1 POST `/api/v1/contests/:id/awd/challenges/:cid/submissions`

`data`：`AwdAttackLogData`

请求体：

```ts
export interface SubmitAwdAttackReq {
  victim_team_id: ID
  flag: string
}
```

> 规则：
> - 仅 `mode=awd` 且 `status=running` 的竞赛允许提交。
> - 攻击方队伍从当前登录用户的已审核报名记录中解析，必须已加入队伍。
> - 默认校验当前轮 Flag；在 `contest.awd.previous_round_grace` 宽限期内，同时接受上一轮 Flag。
> - 同一轮内同一 `attacker_team -> victim_team -> challenge` 首次成功提交得分，之后同队成员重复成功提交仅记日志不重复得分。

### 8.20 POST `/api/v1/admin/contests` / PUT `/api/v1/admin/contests/:id`

`data`（缺少示例，需确认）：

```ts
export interface AdminContestUpsertData {
  contest: ContestDetailData
}
```

### 8.21 DELETE `/api/v1/admin/contests/:id`

`data`：`null`

---

## 9. WebSocket（`/ws/*`）

### 9.1 连接端点（来自 04-api-design.md）

- `/ws/scoreboard/:contest_id?ticket=...`
- `/ws/notifications?ticket=...`
- `/ws/contest/:id/announcements?ticket=...`

### 9.2 统一消息格式

```ts
export interface WsMessage<T> {
  type: string
  payload: T
  timestamp: ISODateTime
}
```

### 9.3 消息 payload

```ts
export type Ping = WsMessage<Record<string, never>> & { type: 'ping' }
export type Pong = WsMessage<Record<string, never>> & { type: 'pong' }

export type ScoreboardUpdate = WsMessage<ScoreboardRow> & { type: 'scoreboard.update' }
export type ScoreboardFrozen = WsMessage<{ contest_id: ID; frozen: boolean; frozen_at?: ISODateTime }> & { type: 'scoreboard.frozen' }

export type NotificationNew = WsMessage<NotificationItem> & { type: 'notification.new' }
export type AnnouncementNew = WsMessage<ContestAnnouncement> & { type: 'announcement.new' }

export type InstanceStatusMsg = WsMessage<{ instance_id: ID; status: InstanceStatus; access_url?: string; expires_at?: ISODateTime }> & {
  type: 'instance.status'
}
```

---
