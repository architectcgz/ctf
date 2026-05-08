# Batch Modern Challenge Packs Review

## Scope

- Jeopardy: `web-mcp-manifest-ssrf`
- Jeopardy: `pwn-passkey-recovery-relay`
- Jeopardy: `reverse-agent-cache-key`
- Jeopardy: `crypto-stream-backup-ticket`
- Jeopardy: `misc-zero-width-briefing`
- Jeopardy: `forensics-ci-preview-artifact`
- AWD: `awd-webhook-inspector`
- AWD: `awd-passkey-sync-gateway`

## Finding 1: `pwn-passkey-recovery-relay` 附件二进制与运行时二进制不一致

- 严重性：blocker
- 现象：题解按附件 `relay.bin` 取到 `unlock_relay=0x4011d6`，但打服务拿不到 flag。
- 根因：附件是另一份编译产物；容器构建时重新用 `docker/src/challenge.c` 编译，运行时真实地址变成了 `0x401176`。
- 影响：选手分析到的附件和真实服务不是同一份程序，题解与题目不一致。

## Resolution

- 将容器运行时改为直接使用固定二进制 `docker/relay.bin`。
- 用当前 registry 中真实运行镜像里的 `/opt/chal/challenge` 回灌：
  - `attachments/relay.bin`
  - `docker/relay.bin`
- 更新题解，不再写死 `unlock_relay` 地址，而是先从附件里用 `nm` 动态取符号地址。

## Finding 2: 运行中实例如果已被后台清理，再走手动销毁会因为缺失 network 报 500

- 严重性：high
- 现象：`DELETE /api/v1/instances/120` 返回 `10009`。
- 根因：`RuntimeCleanupService.removeNetwork(...)` 把 Docker `network ... not found` 当成硬错误；而 `container not found` 已经做了幂等跳过，两边行为不一致。
- 影响：实例已经被后台维护任务清理过一次时，用户手动销毁会错误失败。

## Resolution

- `runtime_cleanup_service.go`
  - 新增缺失 network 的幂等判断。
  - `network ... not found` 时记录日志并跳过。
- `service_test.go`
  - 新增 `TestServiceCleanupRuntimeIgnoresMissingNetwork` 回归测试。

## Verification

### Registry

- 本地 registry `127.0.0.1:5000` 已确认存在并带 `20260508` tag：
  - `jeopardy/web-mcp-manifest-ssrf`
  - `jeopardy/pwn-passkey-recovery-relay`
  - `jeopardy/reverse-agent-cache-key`
  - `jeopardy/crypto-stream-backup-ticket`
  - `jeopardy/misc-zero-width-briefing`
  - `jeopardy/forensics-ci-preview-artifact`
  - `awd/awd-webhook-inspector`
  - `awd/awd-passkey-sync-gateway`

### Real API

- `crypto-stream-backup-ticket` 已走通真实平台链路：
  - authoring import preview
  - import commit
  - image build / push / registry verify
  - publish check succeeded
  - `POST /api/v1/challenges/16/instances`
  - 访问运行中实例并兑换动态 flag
  - `POST /api/v1/challenges/16/submit` 返回 `is_correct=true`

### Local solve replay

- Jeopardy 已逐题从 registry 镜像启动并按题解完成兑换：
  - `web-mcp-manifest-ssrf`
  - `reverse-agent-cache-key`
  - `crypto-stream-backup-ticket`
  - `misc-zero-width-briefing`
  - `forensics-ci-preview-artifact`
  - `pwn-passkey-recovery-relay`
- `pwn-passkey-recovery-relay` 修复后再次验证：
  - 附件地址与运行时地址一致：`0x401176`
  - `72-byte padding + ret2win` 可稳定打印动态 flag
- AWD 已完成本地运行验证：
  - `awd-webhook-inspector`：checker 通过，攻击链 `/preview?url=http://2130706433:8080/internal/snapshot` 能读到 flag
  - `awd-passkey-sync-gateway`：checker 通过，攻击链 `EXPORT sync-support-2026` 能读到 flag

### Runtime cleanup regression

- 已通过：
  - `go test ./internal/module/runtime -run 'TestServiceCleanupRuntimeIgnoresMissingNetwork|TestServiceCleanupRuntimeHonorsCancellation|TestServiceCleanupRuntimeFailsWhenRuntimeEngineUnavailable' -count=1`
  - 修复版后端 live 重放 `DELETE /api/v1/instances/120`，返回 `code=0`
  - 随后 `GET /api/v1/instances` 返回空列表

## Residual Notes

- AWD 这次没有重跑完整赛事编排链路，但 authoring import commit 已成功，本地镜像启动、checker、攻击链三项都已通过。
- 本 review 中的 blocker 与高优先级问题都已修复，当前没有新增未收口 findings。
