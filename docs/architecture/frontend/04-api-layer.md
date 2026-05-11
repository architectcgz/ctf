# 前端 API 层设计

> 状态：Current
> 事实源：`code/frontend/src/api/`、`docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`
> 替代：无

## 定位

本文档只说明前端请求层的封装方式、模块边界、错误回退和数据归一化规则。

- 覆盖：`request.ts`、各业务 API 模块、teacher/admin 子目录拆分、错误模型和环境变量。
- 不覆盖：页面如何展示错误提示、页面如何组织 loading 状态、后端接口语义本身；接口契约仍以 `docs/contracts/` 为准。

## 当前设计

- `code/frontend/src/api/request.ts`
  - 负责：创建统一 `Axios` 实例、注入 `baseURL / timeout / withCredentials`、解包响应 envelope、构造 `ApiError`、处理 401/429/5xx 状态页回退、暴露 `request<T>()`
  - 不负责：直接弹 Toast、直接决定页面重试策略，或实现 token refresh 链路

- `code/frontend/src/api/auth.ts`、`challenge.ts`、`contest.ts`、`instance.ts`、`notification.ts`、`scoreboard.ts`
  - 负责：按领域封装学生侧和共享能力接口，并在 API 边界完成 ID、可空字段和响应结构归一化
  - 不负责：在页面层重复写 URL、手动解 envelope，或把同类接口继续散落到多个 view

- `code/frontend/src/api/teacher/`、`code/frontend/src/api/admin/`
  - 负责：按教师工作区和平台工作区拆分接口 owner，避免旧的 `api/teacher.ts`、`api/admin.ts` 大文件继续膨胀
  - 不负责：要求所有后台接口都落在同一个 URL 前缀；当前 `authoring`、`reports` 等接口仍按后端契约分组

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

当前回退策略：

| 场景 | 当前行为 |
| --- | --- |
| `401` | 构造 `ApiError`；非登录/注册请求会执行 `authStore.logout()` 并跳到 `/401` |
| `429` | 读取 `Retry-After`，构造提示文案，跳到 `/429` |
| `5xx` / `502` / `503` / `504` | 构造通用失败文案，跳到对应错误页 |
| 后端业务错误码 | 通过 `mapErrorCode()` 生成用户可读文案 |
| 网络错误 | 返回 `网络连接失败` 的 `ApiError` |
| 取消请求 | 直接透传，不进入错误状态页 |

相关代码：

- 错误页映射：`code/frontend/src/utils/errorStatusPage.ts`
- 错误码文案：`code/frontend/src/utils/errorMap.ts`

说明：

- 当前前端不做 refresh token 重试；认证模式已经切到 HttpOnly session cookie。
- 请求层只返回错误对象和状态页跳转，不在这里直接展示 Toast。

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

### 3.2 教师工作区模块

`code/frontend/src/api/teacher/index.ts` 当前重导出以下子模块：

| 文件 | 当前负责 |
| --- | --- |
| `teacher/classes.ts` | 班级目录、班级学生、班级摘要、趋势、复盘、洞察 |
| `teacher/students.ts` | 学生进度、画像、建议、证据、时间线、复盘归档 |
| `teacher/writeups.ts` | 题解审核、社区题解推荐/隐藏、人工评审流 |
| `teacher/instances.ts` | 教师视角实例目录、销毁、班级报告导出 |
| `teacher/awd-reviews.ts` | AWD 复盘、轮次、攻击记录、归档导出 |

### 3.3 平台工作区模块

`code/frontend/src/api/admin/index.ts` 当前重导出以下子模块：

| 文件 | 当前负责 |
| --- | --- |
| `admin/platform.ts` | 平台概览、审计、镜像、通知发布等平台侧能力 |
| `admin/users.ts` | 用户目录、创建、更新、删除、导入 |
| `admin/authoring.ts` | 题目创作、题包导入、拓扑、题解管理、镜像相关创作接口 |
| `admin/awd-authoring.ts` | AWD 题目库和导入管理 |
| `admin/contests.ts` | 竞赛管理、公告、队伍、AWD 运维与导出 |

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

- 页面 view 不直接 import 非 contract API 模块；业务调用应下沉到 feature model。
- 请求层当前没有统一的自动重试机制；竞赛实时刷新、导出轮询等重试逻辑继续留在 feature/composable。
- `getAxiosInstance()` 只作为少量特殊场景的逃生口，默认调用方仍应使用 `request<T>()`。

## 7. Guardrail

- 前端分层边界，防止低层 UI 和页面随意穿透到 API：`code/frontend/src/__tests__/architectureBoundaries.test.ts`
- route view 不直接依赖业务 API：`code/frontend/src/views/__tests__/routeViewArchitectureBoundary.test.ts`
- 长期接口契约：`docs/contracts/openapi-v1.yaml`、`docs/contracts/api-contract-v1.md`
