# 前端 API 层设计

> 状态：Current
> 事实源：`code/frontend/src/api/`、`code/frontend/src/runtime/globalErrorRuntime.ts`、`docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`
> 替代：无

## 本文档范围

| 覆盖 | 不覆盖 |
|------|--------|
| `request.ts` 统一请求入口、Axios 实例配置、envelope 解包、`ApiError` 构造 | 页面如何展示错误提示、页面如何组织 loading 状态 |
| 各业务 API 模块（`auth.ts`、`challenge.ts`、`contest.ts`、`instance.ts` 等） | 后端接口语义本身（见 `docs/contracts/`） |
| `teacher/` 子目录（教师工作区）、`teaching/` 子目录（教学分析领域）、`admin/` 子目录（平台工作区）拆分 | 页面级权限判断和 feature-owned 错误边界 |
| 顶层独立模块（`assessment.ts`、`instances.ts`、`awd-reviews.ts`、`contracts.ts`）和角色聚合入口（`teacher.ts`、`teaching.ts` re-export） | API 层之外的错误日志上报与 APM 集成 |
| 错误模型（`ApiError`）、环境变量（`VITE_API_BASE_URL`、`VITE_API_TIMEOUT`） | 页面业务状态机中的显式 error state |

## 定位

本文档只说明前端请求层的封装方式、模块边界、错误回退和数据归一化规则。

- 覆盖：`request.ts`、各业务 API 模块、teacher/teaching/admin 子目录拆分、顶层角色聚合入口、错误模型和环境变量。
- 不覆盖：页面如何展示错误提示、页面如何组织 loading 状态、后端接口语义本身；接口契约仍以 `docs/contracts/` 为准。

## 当前设计

- `code/frontend/src/api/request.ts`
  - 负责：创建统一 `Axios` 实例、注入 `baseURL / timeout / withCredentials`、解包响应 envelope、构造 `ApiError`、暴露 `request<T>()`
  - 不负责：直接弹 Toast、直接决定页面重试策略、直接跳错误状态页，或实现 token refresh 链路

- `code/frontend/src/runtime/globalErrorRuntime.ts`
  - 负责：接管 truly global 的错误导航 owner，包括 HTTP 401 会话失效、WebSocket 鉴权关闭、Vue runtime error 与 router runtime error
  - 不负责：页面内 429 / 5xx / 业务错误 / 网络错误的 toast、inline fallback、retry 或 draft 保留

- `code/frontend/src/api/auth.ts`、`challenge.ts`、`contest.ts`、`instance.ts`、`notification.ts`、`scoreboard.ts`
  - 负责：按领域封装学生侧和共享能力接口，并在 API 边界完成 ID、可空字段和响应结构归一化
  - 不负责：在页面层重复写 URL、手动解 envelope，或把同类接口继续散落到多个 view

- `code/frontend/src/api/teacher/`、`code/frontend/src/api/teaching/`、`code/frontend/src/api/admin/`
  - 负责：按教师工作区、教学分析领域和平台工作区拆分接口 owner，避免旧的 `api/teacher.ts`、`api/admin.ts` 大文件继续膨胀
  - 不负责：要求所有后台接口都落在同一个 URL 前缀；当前 `authoring`、`reports` 等接口仍按后端契约分组

- `code/frontend/src/api/assessment.ts`、`instances.ts`、`awd-reviews.ts`、`teacher.ts`、`teaching.ts`、`contracts.ts`
  - 负责：保留学生画像 / 报告、按角色分发的实例目录和 AWD 复盘入口、teacher / teaching 子目录 re-export、共享 DTO 类型。
  - 不负责：继续承载新的大型后台接口实现；新增平台或教师工作区接口优先进入对应子目录。

## 1. 请求入口与统一契约

`request.ts` 当前固定使用以下基线：

| 配置项 | 当前值 | 来源 |
| --- | --- | --- |
| `baseURL` | `import.meta.env.VITE_API_BASE_URL || '/api/v1'` | `request.ts` |
| `timeout` | `Number(import.meta.env.VITE_API_TIMEOUT) || 15000` | `request.ts` |
| `withCredentials` | `true` | `request.ts` |
| 头部 | `Content-Type: application/json` | `request.ts` |

统一响应包结构：

```ts
interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
  request_id: string
  errors?: Array<{ field: string; message: string }>
}
```

处理规则：

1. HTTP 2xx 且 `code === 0` 时，`request<T>()` 返回 `data`
2. HTTP 2xx 但 `code !== 0` 时，构造 `ApiError`
3. HTTP 非 2xx 时，按状态码和错误码映射构造 `ApiError`

## 2. 错误模型与回退

`ApiError` 当前暴露这些字段：

- `message`
- `code`
- `requestId`
- `status`
- `errors`
- `requestUrl`

当前回退策略：

| 场景 | 当前行为 |
| --- | --- |
| `401` | `request.ts` 只构造 `ApiError`；`runtime/globalErrorRuntime.ts` 读取标准化 `401` 后，再决定是否执行 `logout()` 并跳到 `/401` |
| `429` | 读取 `Retry-After`，构造提示文案并返回 `ApiError`；页面 / feature owner 自己决定如何提示或重试 |
| `5xx` / `502` / `503` / `504` | 构造通用失败文案并返回 `ApiError`；不再由请求层直接跳状态页 |
| 后端业务错误码 | 通过 `mapErrorCode()` 生成用户可读文案 |
| 网络错误 | 返回 `网络连接失败` 的 `ApiError` |
| 取消请求 | 直接透传，不进入错误状态页 |

补充说明：

- `/500` 这类状态页当前主要用于 Vue runtime error 与 router runtime error，不再把普通 HTTP `5xx` 一律升级成全局跳页。
- `shouldRedirectToErrorStatusPage()` 仍保留在 `errorStatusPage.ts`，但“何时调用它”已经从请求层挪到 runtime owner。

相关代码：

- 错误页映射：`code/frontend/src/utils/errorStatusPage.ts`
- 错误码文案：`code/frontend/src/utils/errorMap.ts`

说明：

- 当前前端不做 refresh token 重试；认证模式已经切到 HttpOnly session cookie。
- 请求层只返回标准化错误对象，不在这里直接展示 Toast，也不直接决定全局导航。

## 3. 模块边界

### 3.1 共享与学生侧模块

| 文件 | 当前负责 |
| --- | --- |
| `api/auth.ts` | 登录、注册、登出、读取 profile、修改密码、获取 ws ticket |
| `api/challenge.ts` | 题目列表、详情、题解、社区解法、Flag 提交、实例创建 |
| `api/contest.ts` | 竞赛列表、详情、队伍、公告、排行榜、AWD 工作区与相关数据 |
| `api/instance.ts` | 我的实例、实例续期、销毁、访问入口 |
| `api/notification.ts` | 通知列表、标记已读 |
| `api/scoreboard.ts` | 练习排行榜等独立排行入口 |

### 3.1.1 顶层独立模块

| 文件 | 当前负责 |
| --- | --- |
| `api/assessment.ts` | 学生个人进度、画像、推荐、timeline 和报告导出 |
| `api/instances.ts` | 按角色分发教师 / 平台实例目录和销毁能力 |
| `api/awd-reviews.ts` | 按角色分发教师 / 平台 AWD 复盘、归档和报告导出 |
| `api/contracts.ts` | 前端共享 DTO、分页、ID、枚举与接口响应类型 |

边界：

- `assessment.ts` 承载学生画像与报告能力，语义上服务学生侧和教师侧
- `instances.ts` 和 `awd-reviews.ts` 作为按角色分发的目录入口，内部调用 `teacher/` 或 `admin/` 子目录的具体实现
- `contracts.ts` 只承载前端共享 DTO 类型，不承载接口实现；新增平台或教师工作区接口应进入对应子目录

### 3.2 教师工作区模块

`code/frontend/src/api/teacher/index.ts` 当前重导出以下子模块：

| 文件 | 当前负责 |
| --- | --- |
| `teacher/classes.ts` | 班级目录、班级学生、班级摘要、趋势、复盘、洞察 |
| `teacher/students.ts` | 学生进度、画像、建议、证据、时间线、复盘归档 |
| `teacher/writeups.ts` | 题解审核、社区题解推荐/隐藏、人工评审流 |
| `teacher/instances.ts` | 教师视角实例目录、销毁、班级报告导出 |
| `teacher/awd-reviews.ts` | AWD 复盘、轮次、攻击记录、归档导出 |

`code/frontend/src/api/teacher.ts` 当前只 re-export `teacher/index.ts`，作为历史顶层入口；新增教师工作区接口应进入 `api/teacher/*` 或已采用的 teaching 子目录。

### 3.2.1 教学分析领域模块

`code/frontend/src/api/teaching/index.ts` 当前重导出以下子模块：

| 文件 | 当前负责 |
| --- | --- |
| `teaching/classes.ts` | 教师 overview、班级目录、班级学生、班级洞察、班级复盘 |
| `teaching/students.ts` | 学生目录、个人分析、证据、建议、时间线和 review archive |
| `teaching/writeups.ts` | 教学视角题解审核与推荐 |
| `teaching/instances.ts` | 教学视角实例目录和销毁能力 |
| `teaching/awd-reviews.ts` | 教学视角 AWD 复盘归档和报告 |

`code/frontend/src/api/teaching.ts` 当前只 re-export `teaching/index.ts`。`admin/teaching.ts` 复用 `teaching/instances` 和 `teaching/classes`，为平台工作区提供实例目录与学生目录兼容入口。

边界：

- `teaching/` 子目录承载教学分析领域接口，语义上服务 `/academy/*` 教师后台页面，但部分接口也被平台工作区复用（例如实例目录、班级学生目录）
- `teaching/` 与 `teacher/` 的区分：`teacher/` 承载教师工作区能力（班级管理、学生管理、题解审核、实例管理、AWD 复盘），`teaching/` 承载教学分析领域能力（班级洞察、学生分析、时间线、复盘归档）
- 当前 `teaching/` 子目录与 `teacher/` 子目录存在部分重叠（例如 `classes.ts`、`students.ts`、`writeups.ts`、`instances.ts`、`awd-reviews.ts` 在两个子目录都存在）；后续可考虑收口成单一子目录或按接口语义明确划分

### 3.3 平台工作区模块

`code/frontend/src/api/admin/index.ts` 当前重导出以下子模块：

| 文件 | 当前负责 |
| --- | --- |
| `admin/platform.ts` | 平台概览、审计、镜像、通知发布等平台侧能力 |
| `admin/users.ts` | 用户目录、创建、更新、删除、导入 |
| `admin/authoring.ts` | 题目创作、题包导入、拓扑、题解管理、镜像相关创作接口 |
| `admin/awd-authoring.ts` | AWD 题目库和导入管理 |
| `admin/contests.ts` | 竞赛管理、公告、队伍、AWD 运维与导出 |
| `admin/teaching.ts` | 平台工作区复用 teaching 实例目录、销毁和学生目录能力 |

说明：

- 当前事实已经不再是“一个 `api/teacher.ts`、一个 `api/admin.ts` 总表”。
- 平台工作区接口并不都落在 `/admin/*` 下；例如题目创作走 `/authoring/*`，导出能力也会走 `/reports/*`。

## 4. 数据归一化规则

前端当前把“接口返回值清洗”放在 API 边界，而不是让页面自己兜底。

主要规则：

- 统一把数字 ID 转成字符串，例如：
  - `auth.ts` 的 `getProfile()`
  - `challenge.ts` 的 `normalizeChallengeDetail()`
  - `contest.ts` 的 `normalizeContest()`、`normalizeTeam()`
  - `instance.ts` 的 `normalizeInstanceData()`
- 对可空字段补默认值，例如：
  - `challenge.ts` 给 `tags`、`hints`、`need_target`、`instance_sharing` 补默认值
  - `instance.ts` 统一计算 `remaining_extends`
- 对页面更好处理的 404 语义，在 API 边界改写成 `null`，例如：
  - `getChallengeWriteup()`
  - `getMyChallengeWriteupSubmission()`
- 上传类接口在 API 边界构造 `FormData`，例如：
  - `admin/users.ts` 的 `importUsers()`

这层的直接目的，是让 feature model 读取到的都是已经收口过的业务数据，而不是把“ID 可能是 number”“字段可能缺省”继续传播到页面。

## 5. 接口或数据影响

当前请求层依赖这些长期约定：

- 认证依赖浏览器自动携带的 session cookie，因此 `withCredentials` 必须保持开启
- 所有业务 API 默认从 `/api/v1` 起步，除非 `VITE_API_BASE_URL` 覆盖
- 错误对象允许携带 `request_id`，页面可在必要时展示给用户或埋点系统
- 表单校验失败可通过 `errors` 字段传递字段级错误

## 6. 边界与已知例外

- route page 不直接 import 非 contract API 模块；业务调用应下沉到 feature model。
- 请求层当前没有统一的自动重试机制；竞赛实时刷新、导出轮询等重试逻辑继续留在 feature/composable。
- `getAxiosInstance()` 只作为少量特殊场景的逃生口，默认调用方仍应使用 `request<T>()`。
- truly global 错误导航只允许通过 `runtime/globalErrorRuntime.ts` 进入；页面 / feature 对可恢复错误保留自己的 UX owner。

## 7. Guardrail

- 前端分层边界，防止低层 UI 和页面随意穿透到 API：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- route page 不直接依赖业务 API：`code/frontend/src/__tests__/routePageArchitectureBoundary.test.ts`
- 请求层与 runtime error owner 边界：`code/frontend/src/api/__tests__/request.test.ts`、`code/frontend/src/runtime/__tests__/globalErrorRuntime.test.ts`
- 长期接口契约：`docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`
