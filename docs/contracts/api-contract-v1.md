# CTF 平台接口契约（v1）

> 目的：把 **CTF 前端（`ctf/frontend`）当前会调用的接口** 与 **后端 API 设计（`ctf/docs/architecture/backend/04-api-design.md`）** 的返回结构统一成一份“契约”，作为联调与实现的唯一参考，避免后期因字段名/类型不一致返工。
>
> 机器可读版本：`ctf/docs/contracts/openapi-v1.yaml`（OpenAPI 3.0），应与本文保持一致。
>
> 最后更新：2026-04-18

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
  user: AuthUser
}
```

> 说明：
> - 登录成功后，服务端会通过 `Set-Cookie` 写入 HttpOnly session cookie。
> - 账号在连续输错密码达到阈值后会被临时锁定；触发锁定的那次登录返回 `429`，锁定期内再次尝试返回 `403`。

### 2.2 POST `/api/v1/auth/register`

`data`：同 `LoginData`（注册成功即登录）。

### 2.3 POST `/api/v1/auth/logout`

`data`：`null`

### 2.4 GET `/api/v1/auth/profile`

`data`：`AuthUser`

### 2.5 PUT `/api/v1/auth/password`

`data`：`null`

### 2.6 POST `/api/v1/auth/ws-ticket`

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
  need_target: boolean      // false 表示该题目无需启动靶机
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
export type FlagType = 'static' | 'dynamic' | 'regex' | 'manual_review'

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

### 3.7 POST `/api/v1/challenges/:id/submit`

`data`：

```ts
export interface SubmitFlagData {
  is_correct: boolean
  status: 'correct' | 'incorrect' | 'pending_review'
  message: string
  points?: number
  submitted_at: ISODateTime
}
```

> 说明：
> - `status='pending_review'` 仅用于 `manual_review` 判题模式。
> - 人工审核题在教师审核通过前保持 `is_correct=false`。
> - 学员题目页收到 `pending_review` 后应展示“等待教师审核”的中性/警告态反馈，且不能把题目标记为已通过。

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

### 3.12 POST `/api/v1/challenges/:id/writeup-submissions`

请求体：

```ts
export type SubmissionWriteupStatus = 'draft' | 'submitted'
export type SubmissionWriteupReviewStatus = 'pending' | 'reviewed' | 'excellent' | 'needs_revision'

export interface UpsertSubmissionWriteupReq {
  title: string
  content: string
  submission_status: SubmissionWriteupStatus
}
```

`data`：

```ts
export interface SubmissionWriteupData {
  id: ID
  user_id: ID
  challenge_id: ID
  contest_id?: ID
  title: string
  content: string
  submission_status: SubmissionWriteupStatus
  review_status: SubmissionWriteupReviewStatus
  submitted_at?: ISODateTime
  reviewed_by?: ID
  reviewed_at?: ISODateTime
  review_comment?: string
  created_at: ISODateTime
  updated_at: ISODateTime
}
```

> 说明：
> - 同一学生对同一题目只有一条 `submission_writeup` 记录，接口语义为 upsert。
> - 当 `submission_status='submitted'` 时，后端会回填 `submitted_at`，并将 `review_status` 置为 `pending`。

### 3.13 GET `/api/v1/challenges/:id/writeup-submissions/me`

`data`：`SubmissionWriteupData`

> 说明：当前学生尚未提交时返回 `404`。

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
> - 该入口仅用于非 AWD 竞赛的题目实例启动。
> - 非 AWD 竞赛仍按用户维度创建竞赛实例，不影响练习模式实例。

### 4.6.1 POST `/api/v1/contests/:id/awd/services/:sid/instances`

`data`：同 `InstanceData`

> 说明：
> - 当前要求竞赛已进入 `running/frozen` 且当前用户报名状态为 `approved`。
> - `mode=awd` 时必须已加入队伍，并按 `contest_id + team_id + service_id` 复用同队共享实例；同队成员再次启动会直接拿到同一实例。
> - AWD 学生端运行链路统一以 `contest_awd_services.id` 作为实例启动主键，旧的 `challenge_id` 入口不再作为 AWD 实例启动入口。

### 4.6.2 POST `/api/v1/admin/contests/:id/awd/current-round/check`

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

### 6.4 GET `/api/v1/teacher/students/:id/evidence`

`data`：

```ts
export interface TeacherEvidenceSummaryData {
  total_events: number
  proxy_request_count: number
  submit_count: number
  success_count: number
  challenge_id: ID
}

export interface TeacherEvidenceEventData {
  type: string
  challenge_id: ID
  title: string
  detail: string
  timestamp: ISODateTime
  meta?: Record<string, unknown>
}

export interface TeacherEvidenceData {
  summary: TeacherEvidenceSummaryData
  events: TeacherEvidenceEventData[]
}
```

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
  status: 'ready' | 'processing' | 'failed'
  download_url?: string     // status=ready 时返回
  expires_at?: ISODateTime
  error_message?: string    // status=failed 时返回
}
```

### 7.4 POST `/api/v1/reports/class`

`data`：同 `ReportExportData`（大班级建议异步，status=processing；完成后通过通知推送下载地址）。

### 7.5 POST `/api/v1/admin/contests/:id/export`

请求体：

```ts
export interface CreateContestExportReq {
  format?: 'json'
}
```

`data`：同 `ReportExportData`。

说明：

- 第一版仅支持 `json`
- 导出文件内容包含赛事基础信息、榜单结果、题目解出情况和队伍成员信息
- 下载文件为 JSON，根对象当前包含：
  - `generated_at`
  - `contest`
  - `scoreboard`
  - `challenges`
  - `teams`

### 7.6 GET `/api/v1/teacher/students/:id/review-archive`

`data`：

```ts
export interface ReviewArchiveData {
  generated_at: ISODateTime
  student: {
    id: ID
    username: string
    name?: string
    class_name?: string
  }
  summary: {
    total_challenges: number
    total_solved: number
    total_score: number
    rank: number
    total_attempts: number
    timeline_event_count: number
    evidence_event_count: number
    writeup_count: number
    manual_review_count: number
    hint_unlock_count: number
    correct_submission_count: number
    last_activity_at?: ISODateTime
  }
  skill_profile: Array<{ dimension: string; score: number }>
  timeline: Array<{ type: string; challenge_id: ID; title: string; timestamp: ISODateTime }>
  evidence: Array<{ type: string; challenge_id: ID; title: string; timestamp: ISODateTime }>
  writeups: Array<{ id: ID; challenge_id: ID; challenge_title: string; title: string }>
  manual_reviews: Array<{ id: ID; challenge_id: ID; challenge_title: string; review_status: string }>
  teacher_observations: {
    items: Array<{
      key: string
      label: string
      level: string
      summary: string
      evidence?: string
    }>
  }
}
```

说明：

- 返回与导出 JSON 同源的学生复盘归档聚合数据
- 供教师/管理员在平台内直接查看完整复盘页
- 教师仅可查看自己班级学生，管理员可查看任意学生
- `teacher_observations` 为基于事件与评阅记录生成的可解释教学观察

### 7.7 POST `/api/v1/teacher/students/:id/review-archive/export`

请求体：

```ts
export interface CreateStudentReviewArchiveReq {
  format?: 'json'
}
```

`data`：同 `ReportExportData`。

说明：

- 第一版仅支持 `json`
- 导出文件内容包含学生基础信息、训练摘要、时间线证据、writeup 评阅状态和人工审核记录
- 教师仅可导出自己班级学生，管理员可导出任意学生
- 下载文件为 JSON，根对象当前包含：
  - `generated_at`
  - `student`
  - `summary`
  - `skill_profile`
  - `timeline`
  - `evidence`
  - `writeups`
  - `manual_reviews`
  - `teacher_observations`

---

## 8. 管理与出题后台（`/admin/*` + `/authoring/*`）

> 仅覆盖当前前端 `ctf/frontend/src/api/admin.ts` 中实际调用的接口；其中用户与审计等“仅管理员”接口保留在 `/admin/*`，题目、镜像、拓扑、环境模板、题目包导入等共享出题接口统一收敛到 `/authoring/*`。

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

### 8.5.1 POST `/api/v1/authoring/challenge-imports`

请求体：`multipart/form-data`，字段 `file`

`data`：

```ts
export interface AdminChallengeImportAttachment {
  name: string
  path: string
}

export interface AdminChallengeImportFlag {
  type: 'static' | 'dynamic'
  prefix?: string
}

export interface AdminChallengeImportRuntime {
  type?: 'container'
  image_ref?: string
}

export interface AdminChallengeImportExtensions {
  topology: {
    source?: string
    enabled: boolean
  }
}

export interface AdminChallengeImportPreview {
  id: ID
  file_name: string
  slug: string
  title: string
  description: string
  category: ChallengeCategory
  difficulty: ChallengeDifficulty
  points: number
  attachments?: AdminChallengeImportAttachment[]
  hints?: AdminChallengeHint[]
  flag: AdminChallengeImportFlag
  runtime: AdminChallengeImportRuntime
  extensions: AdminChallengeImportExtensions
  warnings?: string[]
  created_at: ISODateTime
}
```

### 8.5.2 GET `/api/v1/authoring/challenge-imports/:id`

`data`：同 `AdminChallengeImportPreview`

### 8.5.3 POST `/api/v1/authoring/challenge-imports/:id/commit`

`data`：

```ts
export interface AdminChallengeImportCommitData {
  challenge: AdminChallengeListItem
}
```

### 8.6 GET `/api/v1/authoring/challenges`（分页）

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

### 8.7 POST `/api/v1/authoring/challenges` / PUT `/api/v1/authoring/challenges/:id`

`data`（缺少示例，需确认）：

```ts
export interface AdminChallengeUpsertData {
  challenge: AdminChallengeListItem
}
```

### 8.8 DELETE `/api/v1/authoring/challenges/:id`

`data`：`null`

### 8.8.1 GET `/api/v1/authoring/challenges/:id/flag` / PUT `/api/v1/authoring/challenges/:id/flag`

GET `data`：

```ts
export interface AdminChallengeFlagConfigData {
  configured: boolean
  flag_type?: FlagType
  flag_regex?: string
  flag_prefix?: string
}
```

PUT 请求体：

```ts
export interface ConfigureChallengeFlagReq {
  flag_type: FlagType
  flag?: string
  flag_regex?: string
  flag_prefix?: string
}
```

PUT `data`：

```ts
export interface ConfigureChallengeFlagData {
  message: string
}
```

> 说明：
> - `GET` 不回传静态 Flag 明文，后台详情页仅基于 `flag_type / flag_regex / flag_prefix / configured` 展示当前摘要。
> - `PUT` 时：
>   - `static` 必填 `flag`
>   - `dynamic` 必填 `flag_prefix`
>   - `regex` 必填 `flag_regex`，可选 `flag_prefix`
>   - `manual_review` 无额外字段
> - 当前前端在出题后台题目详情页以内联配置卡直接消费该接口。

### 8.8.2 POST `/api/v1/authoring/challenges/:id/self-check`

`data`：

```ts
export interface ChallengeSelfCheckStepData {
  name: string
  passed: boolean
  message: string
}

export interface ChallengeSelfCheckPhaseData {
  passed: boolean
  started_at: ISODateTime
  ended_at: ISODateTime
  steps: ChallengeSelfCheckStepData[]
}

export interface ChallengeSelfCheckRuntimeData extends ChallengeSelfCheckPhaseData {
  access_url?: string
  container_count: number
  network_count: number
}

export interface ChallengeSelfCheckData {
  challenge_id: ID
  precheck: ChallengeSelfCheckPhaseData
  runtime: ChallengeSelfCheckRuntimeData
}
```

### 8.8.3 POST `/api/v1/authoring/challenges/:id/publish-requests`

状态码：`202 Accepted`

`data`：

```ts
export type AdminChallengePublishRequestStatus = 'queued' | 'running' | 'succeeded' | 'failed'

export interface AdminChallengePublishRequestData {
  id: ID
  challenge_id: ID
  requested_by: ID
  status: AdminChallengePublishRequestStatus
  active: boolean
  request_source: 'admin_publish' | string
  failure_summary?: string
  started_at?: ISODateTime
  finished_at?: ISODateTime
  published_at?: ISODateTime
  created_at: ISODateTime
  updated_at: ISODateTime
  result?: ChallengeSelfCheckData
}
```

说明：

- 若当前题目已有 `queued` 或 `running` 的发布自检任务，再次提交会返回当前活动任务，而不是创建第二条并行任务。
- 通过后题目会自动发布，并通知请求人。
- 失败后题目保持 `draft`，并返回失败摘要供后台展示。

### 8.8.4 GET `/api/v1/authoring/challenges/:id/publish-requests/latest`

状态码：

- `200 OK`：返回最近一次发布自检记录
- `404 Not Found`：该题目还没有发布自检记录

`data`：同 `AdminChallengePublishRequestData`

### 8.9 GET `/api/v1/authoring/images`（分页）

`data`：

```ts
export type ImageStatus = 'pending' | 'building' | 'available' | 'failed'

export interface AdminImageListItem {
  id: ID
  name: string
  tag: string
  description: string
  size: number
  size_formatted: string
  status: ImageStatus
  created_at: ISODateTime
  updated_at: ISODateTime
}

export type AdminImageListData = PageResult<AdminImageListItem>
```

### 8.10 POST `/api/v1/authoring/images`

`data`：

```ts
export type AdminImageCreateData = AdminImageListItem
```

### 8.10.1 PUT `/api/v1/authoring/images/:id`

`data`：`null`

### 8.11 DELETE `/api/v1/authoring/images/:id`

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

### 8.14 GET `/api/v1/authoring/challenges/:id/writeup` / PUT `/api/v1/authoring/challenges/:id/writeup` / DELETE `/api/v1/authoring/challenges/:id/writeup`

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

### 8.15 GET `/api/v1/teacher/writeup-submissions`

查询参数：

```ts
export interface TeacherSubmissionWriteupQuery {
  student_id?: ID
  challenge_id?: ID
  class_name?: string
  submission_status?: SubmissionWriteupStatus
  review_status?: SubmissionWriteupReviewStatus
  page?: number
  page_size?: number
}
```

`data`：

```ts
export interface TeacherSubmissionWriteupItem {
  id: ID
  user_id: ID
  student_username: string
  student_name?: string
  class_name?: string
  challenge_id: ID
  challenge_title: string
  title: string
  content_preview: string
  submission_status: SubmissionWriteupStatus
  review_status: SubmissionWriteupReviewStatus
  submitted_at?: ISODateTime
  reviewed_at?: ISODateTime
  updated_at: ISODateTime
}

export type TeacherSubmissionWriteupListData = PageResult<TeacherSubmissionWriteupItem>
```

> 说明：
> - `teacher` 角色查询时，后端会自动收敛到教师自己的 `class_name`，不会跨班查看。
> - `admin` 角色可跨班检索。

### 8.16 GET `/api/v1/teacher/writeup-submissions/:id`

`data`：

```ts
export interface TeacherSubmissionWriteupDetail extends SubmissionWriteupData {
  student_username: string
  student_name?: string
  class_name?: string
  challenge_title: string
  reviewer_name?: string
}
```

### 8.17 PUT `/api/v1/teacher/writeup-submissions/:id/review`

请求体：

```ts
export interface ReviewSubmissionWriteupReq {
  review_status: SubmissionWriteupReviewStatus
  review_comment?: string
}
```

`data`：`SubmissionWriteupData`

### 8.18 GET `/api/v1/teacher/manual-review-submissions`

`query`：

```ts
export interface TeacherManualReviewSubmissionQuery {
  student_id?: ID
  challenge_id?: ID
  class_name?: string
  review_status?: 'pending' | 'approved' | 'rejected'
  page?: number
  page_size?: number
}
```

`data`：

```ts
export interface TeacherManualReviewSubmissionItemData {
  id: ID
  user_id: ID
  student_username: string
  student_name?: string
  class_name?: string
  challenge_id: ID
  challenge_title: string
  answer_preview: string
  review_status: 'pending' | 'approved' | 'rejected'
  submitted_at: ISODateTime
  reviewed_at?: ISODateTime
  updated_at: ISODateTime
}
```

### 8.19 GET `/api/v1/teacher/manual-review-submissions/:id`

`data`：

```ts
export interface TeacherManualReviewSubmissionDetailData {
  id: ID
  user_id: ID
  student_username: string
  student_name?: string
  class_name?: string
  challenge_id: ID
  challenge_title: string
  answer: string
  is_correct: boolean
  score: number
  review_status: 'pending' | 'approved' | 'rejected'
  reviewed_by?: ID
  reviewed_at?: ISODateTime
  review_comment?: string
  submitted_at: ISODateTime
  updated_at: ISODateTime
  reviewer_name?: string
}
```

### 8.20 PUT `/api/v1/teacher/manual-review-submissions/:id/review`

请求体：

```ts
export interface ReviewManualReviewSubmissionReq {
  review_status: 'approved' | 'rejected'
  review_comment?: string
}
```

`data`：`TeacherManualReviewSubmissionDetailData`

### 8.21 GET `/api/v1/authoring/challenges/:id/topology` / PUT `/api/v1/authoring/challenges/:id/topology` / DELETE `/api/v1/authoring/challenges/:id/topology`

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

### 8.16 GET `/api/v1/authoring/environment-templates` / POST `/api/v1/authoring/environment-templates` / GET `/api/v1/authoring/environment-templates/:id` / PUT `/api/v1/authoring/environment-templates/:id` / DELETE `/api/v1/authoring/environment-templates/:id`

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

### 8.18.1 GET `/api/v1/admin/contests/:id/challenges` / POST `/api/v1/admin/contests/:id/challenges` / PUT `/api/v1/admin/contests/:id/challenges/:cid` / DELETE `/api/v1/admin/contests/:id/challenges/:cid`

`data`：

```ts
export interface AdminContestChallengeData {
  id: ID
  contest_id: ID
  challenge_id: ID
  title?: string
  category?: ChallengeCategory
  difficulty?: ChallengeDifficulty
  points: number
  order: number
  is_visible: boolean
  created_at: ISODateTime
}

export type AdminContestChallengeListData = AdminContestChallengeData[]
```

`POST` 请求体：

```ts
export interface CreateAdminContestChallengeReq {
  challenge_id: ID
  points?: number
  order?: number
  is_visible?: boolean
}
```

`PUT` 请求体：

```ts
export interface UpdateAdminContestChallengeReq {
  points?: number
  order?: number
  is_visible?: boolean
}
```

> 说明：
> - `contest_challenges` 只负责赛事题目关系、分值、顺序与可见性等编排字段。
> - 当赛事模式为 `awd` 时，管理员侧的服务关联、checker / SLA / 防守分 / 最近试跑结果等运行态字段统一通过 `contest_awd_services` 读写；关系接口不再承载这些字段。

### 8.18.1.1 GET `/api/v1/admin/contests/:id/awd/services` / POST `/api/v1/admin/contests/:id/awd/services` / PUT `/api/v1/admin/contests/:id/awd/services/:sid` / DELETE `/api/v1/admin/contests/:id/awd/services/:sid`

`data`：

```ts
export interface AdminContestAWDServiceData {
  id: ID
  contest_id: ID
  challenge_id: ID
  template_id?: ID
  display_name: string
  order: number
  is_visible: boolean
  score_config?: Record<string, unknown>
  runtime_config?: Record<string, unknown>
  validation_state?: 'pending' | 'passed' | 'failed' | 'stale'
  last_preview_at?: ISODateTime
  last_preview_result?: AWDCheckerPreviewData
  created_at: ISODateTime
  updated_at: ISODateTime
}

export type AdminContestAWDServiceListData = AdminContestAWDServiceData[]
```

`POST` 请求体：

```ts
export interface CreateAdminContestAWDServiceReq {
  challenge_id: ID
  template_id: ID
  points: number
  display_name?: string
  order?: number
  is_visible?: boolean
  checker_type?: 'legacy_probe' | 'http_standard'
  checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}
```

`PUT` 请求体：

```ts
export interface UpdateAdminContestAWDServiceReq {
  template_id?: ID
  points?: number
  display_name?: string
  order?: number
  is_visible?: boolean
  checker_type?: 'legacy_probe' | 'http_standard'
  checker_config?: Record<string, unknown>
  awd_sla_score?: number
  awd_defense_score?: number
  awd_checker_preview_token?: string
}
```

> 说明：
> - `contest_awd_services` 是 AWD 运行态配置的显式服务层。
> - `checker_type / checker_config / awd_sla_score / awd_defense_score / awd_checker_preview_token` 统一通过该接口写入。
> - `points` 写入 `score_config.points`，用于服务在赛事内的展示分值，不参与每轮 SLA/防御累计。
> - AWD 计分契约：新建服务默认 `awd_sla_score = 1`、`awd_defense_score = 2`；二者单项取值范围均为 `0-5`。轮次默认 `attack_score = 30`、`defense_score = 3`；`attack_score` 范围为 `0-100`，`defense_score` 范围为 `0-10`。
> - 返回体中的 `runtime_config` 仅保留正式 runtime 配置字段；新写入记录也不再持久化兼容影子字段 `challenge_id`。
> - `validation_state / last_preview_at / last_preview_result` 反映最近一次保存到 service 层的 checker 校验状态。

### 8.18.2 GET `/api/v1/admin/contests/:id/awd/readiness`

`data`：

```ts
export type AWDReadinessAction = 'create_round' | 'run_current_round_check' | 'start_contest'
export type AWDReadinessBlockingReason =
  | 'missing_checker'
  | 'invalid_checker_config'
  | 'pending_validation'
  | 'last_preview_failed'
  | 'validation_stale'
export type AWDReadinessGlobalReason = 'no_challenges'

export interface AWDReadinessItemData {
  challenge_id: ID
  title: string
  checker_type?: 'legacy_probe' | 'http_standard'
  validation_state: 'pending' | 'passed' | 'failed' | 'stale'
  last_preview_at?: ISODateTime
  last_access_url?: string
  blocking_reason: AWDReadinessBlockingReason
}

export interface AWDReadinessData {
  contest_id: ID
  ready: boolean
  total_challenges: number
  passed_challenges: number
  pending_challenges: number
  failed_challenges: number
  stale_challenges: number
  missing_checker_challenges: number
  blocking_count: number
  blocking_actions: AWDReadinessAction[]
  global_blocking_reasons: AWDReadinessGlobalReason[]
  items: AWDReadinessItemData[]
}
```

### 8.18.3 POST `/api/v1/admin/contests/:id/awd/checker-preview`

`data`：

```ts
export interface AWDCheckerPreviewContextData {
  access_url: string
  preview_flag: string
  round_number: number
  team_id: ID
  challenge_id: ID
}

export interface AWDCheckerPreviewData {
  checker_type?: 'legacy_probe' | 'http_standard'
  service_status: 'up' | 'down' | 'compromised'
  check_result: Record<string, unknown>
  preview_context: AWDCheckerPreviewContextData
  preview_token?: string
}
```

请求体：

```ts
export interface PreviewAwdCheckerReq {
  challenge_id: ID
  checker_type: 'legacy_probe' | 'http_standard'
  checker_config?: Record<string, unknown>
  access_url: string
  preview_flag?: string
}
```

> 说明：
> - 试跑接口会返回 `preview_token`，后续管理员保存赛事题目配置时可通过 `awd_checker_preview_token` 把最近一次试跑结果与配置草稿绑定。
> - 当前 readiness 与 checker preview 统一读取 `contest_awd_services.runtime_config + score_config + validation`。

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

export type AwdTrafficStatusGroup = 'success' | 'redirect' | 'client_error' | 'server_error'

export interface AwdTrafficTrendBucketData {
  bucket_start_at: ISODateTime
  request_count: number
  error_count: number
}

export interface AwdTrafficTopTeamData {
  team_id: ID
  team_name: string
  request_count: number
  error_count: number
}

export interface AwdTrafficTopChallengeData {
  challenge_id: ID
  challenge_title: string
  request_count: number
  error_count: number
}

export interface AwdTrafficTopPathData {
  path: string
  request_count: number
  error_count: number
  last_status_code: number
}

export interface AwdTrafficSummaryData {
  round: AwdRoundData
  contest_id: ID
  round_id: ID
  total_request_count: number
  active_attacker_team_count: number
  victim_team_count: number
  error_request_count: number
  unique_path_count: number
  latest_event_at?: ISODateTime
  trend_buckets: AwdTrafficTrendBucketData[]
  top_attackers: AwdTrafficTopTeamData[]
  top_victims: AwdTrafficTopTeamData[]
  top_challenges: AwdTrafficTopChallengeData[]
  top_paths: AwdTrafficTopPathData[]
  top_error_paths: AwdTrafficTopPathData[]
}

export interface AwdTrafficEventData {
  id: ID
  contest_id: ID
  round_id: ID
  attacker_team_id: ID
  attacker_team_name: string
  victim_team_id: ID
  victim_team_name: string
  service_id?: ID
  challenge_id: ID
  challenge_title: string
  method: string
  path: string
  status_code: number
  status_group: AwdTrafficStatusGroup
  is_error: boolean
  source: string
  request_id?: string
  occurred_at: ISODateTime
}

export type AwdTrafficEventPageData = PageResult<AwdTrafficEventData>
```

> 说明：当前已补齐 AWD 轮次调度、最小版实例健康 Checker、真实容器 Flag 注入、基于当前轮 Flag 的真实攻击提交流程、短窗口的上一轮 Flag 兼容，以及队伍级共享实例模型。

### 8.19.1 GET `/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/summary`

`data`：`AwdTrafficSummaryData`

> 说明：
> - 数据源为 `awd_traffic_events` 轻量事实表，仅覆盖平台代理链路下的 AWD 共享实例访问流量。
> - `top_paths` 表示本轮热点路径，`top_error_paths` 表示错误请求最集中的路径。
> - 该接口展示的是“代理请求流量摘要”，不等价于“已确认攻击成功”；管理员需要结合攻击日志一起判断。

### 8.19.2 GET `/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/events`

`data`：`AwdTrafficEventPageData`

查询参数：

```ts
export interface ListAwdTrafficEventsQuery {
  attacker_team_id?: ID
  victim_team_id?: ID
  challenge_id?: ID
  status_group?: AwdTrafficStatusGroup
  path_keyword?: string
  page?: number
  page_size?: number
}
```

> 说明：
> - 按 `created_at DESC, id DESC` 返回，适合管理员查看最新代理流量。
> - 事件会显式返回 `service_id`，用于和 AWD 赛事服务建立一一对应的运行态归因。
> - 第一版 `source` 固定为 `runtime_proxy`；`request_id` 目前预留，实际明细返回中可以省略。

### 8.19.3 POST `/api/v1/contests/:id/awd/challenges/:cid/submissions`

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

### 8.19.4 POST `/api/v1/contests/:id/awd/services/:sid/targets/:team_id/access`

`data`：

```ts
export interface InstanceAccessData {
  access_url: string
}
```

> 说明：
> - 该接口用于获取 AWD 跨队攻击代理入口，`team_id` 表示受害队伍 ID，`sid` 表示 AWD 服务 ID。
> - 当前登录用户必须属于该 AWD 赛事中的攻击队伍，且目标队伍不能是自己的队伍。
> - 比赛必须处于 `running` / `frozen`，当前必须存在 `running` 轮次，目标服务必须可见，目标队伍服务实例必须处于 `running` 且有 `access_url`。
> - 返回的 `access_url` 指向平台代理路径，调用方不应直接使用实例真实端口作为攻击入口。

### 8.19.5 ANY `/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath`

`data`：目标实例原始响应。

> 说明：
> - 该代理入口接受 `access` 接口签发的短期 ticket，并把请求转发到目标队伍服务实例。
> - ticket 首次出现在 query 中时，后端会设置 HttpOnly cookie 并重定向到去除 ticket 的 URL。
> - 代理请求会写入 `awd_traffic_events`，归因字段来自 ticket 中的 `contest_id`、`attacker_team_id`、`victim_team_id`、`service_id` 和 `challenge_id`。
> - 该入口只覆盖 HTTP 层代理审计，不等价于 TCP 抓包或 L4 流量审计。
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

- `/ws/notifications?ticket=...`
- `/ws/contests/:id/announcements?ticket=...`
- `/ws/contests/:id/scoreboard?ticket=...`

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
export type SystemConnected = WsMessage<{
  user_id: ID
  heartbeat_interval_seconds: number
  retry: {
    strategy: 'exponential_backoff'
    initial_delay_ms: number
    max_delay_ms: number
  }
}> & { type: 'system.connected' }

export type ScoreboardUpdated = WsMessage<{ contest_id: ID }> & { type: 'scoreboard.updated' }

export type NotificationCreated = WsMessage<NotificationItem> & { type: 'notification.created' }
export type NotificationRead = WsMessage<NotificationItem> & { type: 'notification.read' }
export type AnnouncementCreated = WsMessage<{
  contest_id: ID
  announcement: ContestAnnouncement
}> & { type: 'contest.announcement.created' }
export type AnnouncementDeleted = WsMessage<{
  contest_id: ID
  announcement_id: ID
}> & { type: 'contest.announcement.deleted' }

export type InstanceStatusMsg = WsMessage<{ instance_id: ID; status: InstanceStatus; access_url?: string; expires_at?: ISODateTime }> & {
  type: 'instance.status'
}
```

---
