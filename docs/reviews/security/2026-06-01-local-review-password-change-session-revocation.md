# 2026-06-01 本地 Review：改密后会话撤销修复

- Review target：`/home/azhi/workspace/projects/ctf` 当前工作区未提交改动；审查范围限定为 `code/backend/internal/module/auth/{api/http,contracts,infrastructure}` 与 `code/frontend/src/pages/profile`、`code/frontend/src/features/profile/model`
- Classification check：独立本地 review；未命中额外 leader / pipeline 分级说明
- Gate verdict：blocked

## 当前状态

- 2026-06-02：当前工作区已按本 review 修复两个 blocker。
- 修复摘要：
  - `tokenService` 改为以用户级 `session version` 作为鉴权失效语义，覆盖没有反向索引的历史 session。
  - 改密 handler 在无法完成用户级会话撤销时不再返回成功。
- 已执行验证：
  - `cd code/backend && go test ./internal/module/auth/...`
  - `cd code/frontend && npm exec -- vitest run src/pages/profile/__tests__/SecuritySettings.test.ts`
- 仍待补充：按当前 pipeline 规则，这次只完成了同上下文自查，独立 review gate 还没有补齐。

## Findings

### Blocker 1：旧会话索引没有回填，发布后已存在的其他设备会话不会被改密流程撤销

- Location：`code/backend/internal/module/auth/infrastructure/token_service.go:75-84`, `code/backend/internal/module/auth/infrastructure/token_service.go:158-183`
- Issue：`RevokeAllUserSessions()` 只依赖新引入的 `user_sessions` 反向索引，但这个索引只会在本次补丁之后新建会话时写入。发布前已经存在的 Redis 会话不会被补索引，改密时也没有 fallback 扫描、懒回填或版本戳校验。
- Why it matters：这次修复的目标是“改密后所有设备重新登录”。按当前实现，当前请求所在浏览器会因为清 cookie 退出，但同一用户在其他设备上、且登录时间早于这次发布的会话仍会继续有效，修复在真实上线切换窗口内并不成立。
- Suggestion：不要把安全语义建立在“部署后新写入的辅助索引”上。可选方向包括：为现有会话做回填迁移 / 在 `GetSession` 或登录链路里懒回填索引 / 改成基于用户级 `session_version` 或 `password_changed_at` 的服务端校验，让旧会话天然失效。

### Blocker 2：会话撤销链路被实现成 best-effort，失败时接口仍返回成功

- Location：`code/backend/internal/module/auth/infrastructure/token_service.go:81-84`, `code/backend/internal/module/auth/api/http/handler.go:221-228`
- Issue：`CreateSession()` 对索引写入失败直接吞掉，`ChangePassword()` 对 `RevokeAllUserSessions()` 失败只记日志不回错。于是“密码已改且所有设备已退出”这个结果并不是强保证，而是 best-effort。
- Why it matters：这是认证安全动作，不是普通体验优化。任何一次索引写失败、索引丢失或 Redis 读写异常，都可能让其他设备继续持有有效会话，但前端会提示“密码修改成功，请重新登录”，页面文案也写着“修改后会同步退出其他设备”，这会把安全失败伪装成成功。
- Suggestion：这里应当 fail closed。要么把“改密 + 撤销所有会话”收敛成一个必须整体成功的服务端语义；要么改成不依赖可选索引的失效机制。至少在撤销失败时不能返回成功响应。

## Material findings

- `Blocker 1`：补齐旧会话失效策略，并增加覆盖“发布前已有会话”的测试路径
- `Blocker 2`：让会话撤销失败对改密结果可见，并增加失败分支测试

## Senior implementation assessment

当前方向把“单 session 删除”扩展成“用户级批量撤销”，思路本身是对的，前后端联动也保持了较小改动面。但对安全场景来说，依赖辅助索引且允许静默失败，会让语义从“强制退出所有设备”退化成“尽量退出一部分设备”。更稳妥的实现通常会把失效判断收口到服务端主鉴权路径，例如用户级版本号、密码更新时间或可验证的全局失效戳，而不是只依赖额外索引是否恰好完整。

## Required re-validation

- `cd code/backend && go test ./internal/module/auth/...`
- 为 `token_service` 增加至少一组测试：已有 session 缺少 `user_sessions` 索引时，改密撤销仍会使该 session 失效
- 为 HTTP 改密链路增加失败分支验证：当 `RevokeAllUserSessions()` 失败时，接口不应返回成功
- `cd code/frontend && npm exec -- vitest run src/pages/profile/__tests__/SecuritySettings.test.ts`

## Residual risk

- 本次 review 只覆盖了密码修改后的 cookie/session 语义，没有继续扩展到已建立的 WebSocket 连接是否需要主动断开。
- 前端测试当前只断言跳转登录页，没有同时断言 `authStore` 已清空；这不是当前 blocker，但修正后可以顺手补强。

## Touched known-debt status

- 依据 `references/ctf-current-review-status-checks.md` 和当前 `docs/reviews/*` 入口，本次 touched surface 未命中需要强制一起收口的既有结构债清单。
