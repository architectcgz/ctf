# CTF 平台接口契约（v1）

> 目的：把 **CTF 前端（`ctf/frontend`）当前会调用的接口** 与 **后端 API 设计（`ctf/docs/architecture/backend/04-api-design.md`）** 的返回结构统一成一份“契约”，作为联调与实现的唯一参考，避免后期因字段名/类型不一致返工。
>
> 机器可读版本：`ctf/docs/contracts/openapi-v1.yaml`（OpenAPI 3.0），应与本文保持一致。
>
> 最后更新：2026-03-03

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

### 3.4 DELETE `/api/v1/instances/:id`

`data`：`null`

### 3.5 POST `/api/v1/instances/:id/extend`

`data`：

```ts
export interface InstanceExtendData {
  id: ID
  expires_at: ISODateTime
  remaining_extends: number
}
```

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

### 4.6 POST `/api/v1/contests/:id/challenges/:cid/instances`

`data`：同 `InstanceData`（但应额外校验竞赛状态/报名/队伍关系）。

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

### 8.14 POST `/api/v1/admin/contests` / PUT `/api/v1/admin/contests/:id`

`data`（缺少示例，需确认）：

```ts
export interface AdminContestUpsertData {
  contest: ContestDetailData
}
```

### 8.15 DELETE `/api/v1/admin/contests/:id`

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

