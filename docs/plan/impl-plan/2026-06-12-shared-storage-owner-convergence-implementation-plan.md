<!-- Managed by code-workflow package (version: 2026-06-10.1) -->
# 共享文件与共享密钥 Owner 收口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `subagent-driven-development` (recommended) or `executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking; flip each checkbox immediately after the expected result is reached.

**Goal:** 消除 API 多副本对本地磁盘和副本私有密钥材料的隐式依赖，让报告下载、题目附件下载、动态 Flag secret、AWD SSH host key 都有明确的共享来源和业务 owner。

**Architecture:** 新增平台级 shared storage abstraction，先落地 `shared_fs` adapter，供 assessment report、challenge attachment、runtime secret / host key 复用；module 内保留业务 port，不让 handler 或 application service 继续自己读 env / 拼本地路径。短期仍允许 report renderer 写 shared filesystem local path，但下载链路要从业务 store 返回可跨副本读取的对象或 reader，为后续 object storage 留接口余量。

**Tech Stack:** Go, shared filesystem adapter, GORM-backed report metadata, Gin streaming response, existing config loader, code-workflow

---

## Task Metadata

- Task Slug: `2026-06-12-shared-storage-owner-convergence`
- Started At: `2026-06-12T00:00:00Z`
- Worktree: `后续实现时运行 scripts/start-implementation.sh 2026-06-12-shared-storage-owner-convergence 生成`
- Branch: `task/2026-06-12-shared-storage-owner-convergence`

## Objective And Non-Goals

- Objective:
  - 新增 shared storage 配置与 `shared_fs` adapter，作为多 API 副本共用文件/密钥材料的统一入口。
  - 把 assessment report output 从 `report.storage_dir` 本地路径隐式 owner 收口为 `ReportOutputStore` 显式 shared storage owner。
  - 把 challenge imported attachment persist/download 从 store + handler 双方各读本地 env 收口为同一个 attachment store / download port。
  - 让 dynamic flag secret 与 AWD SSH host key 文档和配置明确要求多副本共源；实现侧优先复用 shared_fs 文件来源。
- Non-Goals:
  - 不在本任务内实现 S3 / MinIO / object storage adapter；只保留 port 和 config 余量。
  - 不迁移所有 challenge package export/source/checker artifact LocalFS surface，只把 report output、imported attachment、flag secret、SSH host key 列为本阶段 scope。
  - 不改变报告格式生成器本身；PDF / Excel writer 可以先继续写 shared_fs local path。
  - 不实现 SSH host key 在线轮换；T3 仍以 restart 后加载稳定 key 为前提。

## Inputs

- Source docs:
  - `docs/plan/impl-plan/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Related architecture/contracts:
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/api/http/report_handler.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `code/backend/internal/config/config.go`
- Related prior work:
  - `docs/plan/impl-plan/2026-06-07-awd-defense-ssh-gateway-split-implementation-plan.md`
  - `docs/plan/impl-plan/2026-06-02-runtime-control-plane-agent-split-plan.md`

## Task Classification

- Classification: `结构性改动 / 非琐碎任务`
- Why:
  - 同时触达 `assessment`、`challenge`、`app/composition`、`config` 和部署文档。
  - 当前 report 下载和 challenge attachment 下载都直接使用 `c.FileAttachment(localPath)`，API 多副本 behind LB 会出现请求落到不持有文件的副本。
  - Secret / host key 若继续每副本本地生成，会导致动态 Flag 不一致或 SSH host key fingerprint 抖动。

## Files

- Create:
  - `code/backend/internal/platform/storage/storage.go`
  - `code/backend/internal/platform/storage/sharedfs/store.go`
  - `code/backend/internal/platform/storage/sharedfs/store_test.go`
  - 如需要，新增 `code/backend/internal/module/challenge/application/queries/attachment_download_service.go`
- Modify:
  - `code/backend/internal/config/config.go`
  - `code/backend/internal/config/config_test.go`
  - `code/backend/internal/module/assessment/ports/ports.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_file_output.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/api/http/report_handler.go`
  - `code/backend/internal/module/challenge/ports/ports.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store_test.go`
  - `code/backend/internal/module/challenge/api/http/handler.go`
  - `code/backend/internal/module/challenge/runtime/wiring.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway.go`
  - `docs/architecture/backend/03-container-architecture.md`
  - `docs/operations/awd-host-reboot-recovery-drill.md`
- Review:
  - `code/backend/internal/module/challenge/infrastructure/challenge_import_preview_store.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_package_storage.go`
  - `code/backend/internal/module/challenge/infrastructure/awd_checker_artifact_store.go`
- Test:
  - `code/backend/internal/platform/storage/sharedfs/store_test.go`
  - `code/backend/internal/module/assessment/infrastructure/report_output_store_test.go`
  - `code/backend/internal/module/assessment/application/commands/report_service_test.go`
  - `code/backend/internal/module/assessment/api/http/report_handler_test.go`
  - `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store_test.go`
  - `code/backend/internal/module/challenge/api/http/handler_test.go`
  - `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`

## 复用与 Owner 决策

- Existing patterns searched:
  - assessment `ReportOutputStore` 当前已有业务 port，但 contract 是 prepare/resolve 本地 path。
  - challenge `ChallengeAttachmentStore` 只负责 persist imported attachment，download handler 自己读 env 和 local path。
  - config 已有 `container.flag_global_secret_file`、`container.defense_ssh_host_key_path`，但默认都指向本地 `storage/runtime`。
  - `clustersecret` 已能校验 Flag secret fingerprint，但不提供 secret 原文共源。
- Reuse / extend / split / create-new decision:
  - 新增 `internal/platform/storage` 做横切 shared storage port；不把 shared storage 放进 assessment 或 challenge 单个模块。
  - assessment / challenge 保留各自业务 port，内部依赖 platform storage adapter。
  - challenge attachment 下载 handler 改成依赖 store / service，不再直接读取 `CHALLENGE_ATTACHMENT_STORAGE_DIR`。
  - flag secret 和 SSH host key 先收口为 shared filesystem 文件来源；object storage secret provider 另起任务。
- Owner boundary:
  - `platform/storage`：共享存储 key、reader/writer、path containment、shared_fs adapter owner。
  - `assessment`：报告生成、报告下载元数据、报告文件业务命名 owner。
  - `challenge`：导入附件持久化、URL path 到 storage key 映射、下载权限/错误映射 owner。
  - `config` / `app/composition`：runtime secret 与 host key 共享来源接线 owner。
- Why this is the narrowest safe surface:
  - 先以 shared_fs 交付多副本 correctness，不把 object storage、预签名 URL、全部 challenge package artifact 一次性纳入。
  - 只修当前 HA umbrella plan 标出的最直接 blocker：report download、attachment download、flag secret、SSH host key。

## Intake Analysis Gate

- Relevant superpowers analysis pass:
  - `brainstorming`
  - `dispatching-parallel-agents`
- Why this pass fits:
  - 这是跨模块 owner 收口，不只是把本地路径换一个目录；需要先分清平台共享能力和业务下载语义。
- grill-with-docs findings:
  - `docs/plan/impl-plan/2026-06-12-true-ha-control-plane-and-runtime-recovery-implementation-plan.md` 明确报告和附件 LocalFS 是 API 多副本 correctness blocker。
  - T3 SSH gateway 多副本依赖本任务提供稳定 host key 共源，否则 TCP LB 后客户端会看到 host key mismatch。
  - Flag secret 已有 DB fingerprint guard，但不提供 secret 原文共源；多副本必须显式配置同源 secret / shared file。
- Plan adjustments after challenge:
  - 明确 challenge package export/source/checker artifact 不在 T2 第一阶段，以免 scope 发散。
  - 不把 handler 继续作为 storage root owner；download 统一走 module port / service。

## Ordered Task Slices

### Slice 1: platform shared_fs substrate

- [ ] **Step 1: 写 shared_fs adapter path safety 测试**
  - Create: `code/backend/internal/platform/storage/sharedfs/store_test.go`
  - 覆盖：合法 key 写入/读取、path traversal 拒绝、missing 映射明确错误、两个 store instance 指向同一 root 可读同一对象。

- [ ] **Step 2: 定义 shared storage port**
  - Create: `code/backend/internal/platform/storage/storage.go`
  - 最小接口包含 `Open(ctx, key)`、`Put(ctx, key, reader)`、`Stat(ctx, key)`；shared_fs 可额外提供 `PrepareLocalWrite(ctx, key)`。

- [ ] **Step 3: 实现 shared_fs adapter**
  - Create: `code/backend/internal/platform/storage/sharedfs/store.go`
  - 统一 key containment、目录创建、regular file 校验、reader close 责任说明。

- [ ] **Step 4: 增加 shared_storage 配置**
  - Modify: `code/backend/internal/config/config.go`
  - 增加 `SharedStorageConfig{Type, SharedFS.Root}`，默认 `type=shared_fs`、root 可指向 `storage/shared`。
  - 生产环境要求 shared_fs root 明确配置或文档明确 mount contract。

- [ ] **Step 5: 运行 storage/config 测试**
  - Run: `cd code/backend && go test ./internal/platform/storage/... ./internal/config -run 'Test.*SharedStorage|Test.*Config' -count=1`

### Slice 2: assessment report output convergence

- [ ] **Step 6: 写 ReportOutputStore 跨副本读测试**
  - Modify/Create: `code/backend/internal/module/assessment/infrastructure/report_output_store_test.go`
  - 两个 store instance 共用 shared_fs root：A prepare/write，B resolve/open 成功。

- [ ] **Step 7: 扩展 assessment report port**
  - Modify: `code/backend/internal/module/assessment/ports/ports.go`
  - 从只返回 local path 扩成 storage ref / download reader；若短期保留 `PrepareReportOutput`，命名和注释必须声明它是 shared_fs writable path。

- [ ] **Step 8: 改 report output store 依赖 shared storage**
  - Modify: `code/backend/internal/module/assessment/infrastructure/report_output_store.go`
  - 不再直接拥有裸 `storageDir`；通过 shared_fs namespace 生成 report key / shared path。

- [ ] **Step 9: 改报告下载 handler 流式返回**
  - Modify: `code/backend/internal/module/assessment/api/http/report_handler.go`
  - 优先使用 store 返回 reader / size / content-type，避免只依赖 `c.FileAttachment(localPath)`。

- [ ] **Step 10: 运行 assessment focused tests**
  - Run: `cd code/backend && go test ./internal/module/assessment/... -run 'Report(Output|Service|Handler|Download)' -count=1`

### Slice 3: challenge imported attachment convergence

- [ ] **Step 11: 写 challenge attachment 下载 handler 测试**
  - Create/Modify: `code/backend/internal/module/challenge/api/http/handler_test.go`
  - 覆盖：handler 使用 injected attachment download port；not found 返回 404；traversal 拒绝；成功响应不依赖 env local path。

- [ ] **Step 12: 扩展 challenge attachment port**
  - Modify: `code/backend/internal/module/challenge/ports/ports.go`
  - 增加 download/open 语义，例如 `OpenAttachment(ctx, relativePath)`，保持 URL contract `/api/v1/challenges/attachments/imports/...`。

- [ ] **Step 13: 改 ChallengeAttachmentStore 统一 persist/download owner**
  - Modify: `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`
  - store 使用 shared storage namespace，不再和 handler 各自读取 `CHALLENGE_ATTACHMENT_STORAGE_DIR`。

- [ ] **Step 14: 改 challenge handler wiring**
  - Modify: `code/backend/internal/module/challenge/api/http/handler.go`
  - Modify: `code/backend/internal/module/challenge/runtime/wiring.go`
  - `DownloadAttachment` 从 store/service open reader，不再计算 local base dir。

- [ ] **Step 15: 运行 challenge focused tests**
  - Run: `cd code/backend && go test ./internal/module/challenge/... -run 'Attachment|DownloadAttachment|Import' -count=1`

### Slice 4: shared secret / host key contract

- [ ] **Step 16: 写 host key shared source 测试**
  - Modify: `code/backend/internal/app/composition/awd_defense_ssh_gateway_test.go`
  - 覆盖两个 gateway 使用同一 shared_fs key 时 fingerprint 相同；非法 shared key fail fast。

- [ ] **Step 17: 收口 flag secret / host key 配置说明**
  - Modify: `code/backend/internal/config/config.go`
  - 保持现有字段兼容，但注释和校验中明确生产多副本不能依赖每副本本地 auto-generate。
  - 必要时把默认 path 指向 shared storage root 下的 runtime namespace。

- [ ] **Step 18: 更新 docs 事实源与操作说明**
  - Modify: `docs/architecture/backend/03-container-architecture.md`
  - Modify: `docs/operations/awd-host-reboot-recovery-drill.md`
  - 说明 report / attachment / flag secret / host key 的共享 owner 和部署 mount 要求。

- [ ] **Step 19: 运行最小验证**
  - Run: `cd code/backend && go test ./internal/platform/storage/... ./internal/module/assessment/... ./internal/module/challenge/... ./internal/app/composition -run 'Shared|Report|Attachment|HostKey|Config' -count=1`

- [ ] **Step 20: Commit**
  - Run: `git add code/backend/internal/platform/storage code/backend/internal/config/config.go code/backend/internal/config/config_test.go code/backend/internal/module/assessment code/backend/internal/module/challenge code/backend/internal/app/composition/awd_defense_ssh_gateway.go docs/architecture/backend/03-container-architecture.md docs/operations/awd-host-reboot-recovery-drill.md && git commit -m "feat(backend): 收口共享文件与密钥来源" -m "新增 shared_fs 共享存储 owner，并让报告输出与题目导入附件下载不再依赖副本本地路径。" -m "明确动态 Flag secret 与 AWD SSH host key 的多副本共源要求，为 SSH gateway HA 提供前置契约。" -m "Task: 2026-06-12-shared-storage-owner-convergence"`

## Validation

- Commands:
  - `cd code/backend && go test ./internal/platform/storage/... -count=1`
  - `cd code/backend && go test ./internal/module/assessment/... -run 'Report(Output|Service|Handler|Download)' -count=1`
  - `cd code/backend && go test ./internal/module/challenge/... -run 'Attachment|DownloadAttachment|Import' -count=1`
  - `cd code/backend && go test ./internal/app/composition -run 'TestAWDDefenseSSH.*HostKey|Test.*Shared' -count=1`
  - `git diff --check -- code/backend/internal/platform/storage code/backend/internal/module/assessment code/backend/internal/module/challenge code/backend/internal/app/composition/awd_defense_ssh_gateway.go docs/architecture/backend/03-container-architecture.md docs/operations/awd-host-reboot-recovery-drill.md`
- Manual checks:
  - 两个 API 副本挂载同一 shared_fs root：副本 A 生成 report，副本 B 下载成功。
  - 题包导入附件由副本 A commit 后，副本 B 处理 `/api/v1/challenges/attachments/...` 下载成功。
  - 多个 gateway 副本读取同一 host key，`ssh-keygen -lf` fingerprint 一致。
  - 多个 API 副本启动时动态 Flag secret fingerprint 一致，不因本地文件缺失生成不同 secret。
- Review focus:
  - handler 是否不再拥有 storage root / env owner。
  - shared_fs path containment 是否避免 traversal 与 symlink 绕过。
  - business port 是否仍表达 report / attachment 语义，而不是把 platform storage 细节泄露到 handler。
  - object storage 是否只是保留扩展点，没有在本任务半实现。
