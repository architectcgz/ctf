# AWD Runtime Image Digest Implementation Plan

**Goal:** 修复 AWD 题和相关容器运行链路对可变 tag 的错误依赖，让实例启动、自检和 AWD checker preview 在存在镜像 digest 时稳定使用不可变的 `name@digest` 运行引用，避免“镜像已更新但仍复用本地旧 tag 缓存”的设计问题。

**Architecture:** 保持题包导入、镜像构建、镜像表和运行时接口的整体分层不变；只把“运行时实际拉起使用哪个 image ref”收口成统一 helper，并替换现有关键调用点。

**Tech Stack:** Go, Docker runtime, AWD challenge import/runtime, unit tests

---

## Objective

- 明确 runtime 的 canonical image ref 生成规则
- 当镜像记录存在 digest 时，优先使用 `name@digest`
- 保留无 digest 的兼容回退：继续使用 `name:tag`
- 覆盖三个关键运行入口：
  - 练习/比赛实例启动
  - 题目自检 runtime startup
  - AWD checker preview runtime startup

## Non-goals

- 不重写镜像构建任务、registry 同步流程或镜像表 schema
- 不改 AWD 题包 `challenge.yml` 的 author-facing 写法
- 不改现有 `images` 表按 `(name, tag)` 建唯一索引的存储设计
- 不处理已经在运行中的旧容器热更新；本次只修后续运行引用与新实例拉起

## Inputs

- `code/backend/internal/model/image.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- `code/backend/internal/module/contest/application/commands/awd_service_test.go`
- `code/backend/internal/module/runtime/infrastructure/engine.go`
- `docs/contracts/challenge-pack-v1.md`

## Ownership Boundary

- `model.Image` 相关 helper
  - 负责：从镜像元数据推导运行时应使用的 image ref
  - 不负责：镜像构建、推送、导入事务和 registry 校验
- `runtime_container_create.go`
  - 负责：实例启动时把 `image_id` 解析成真正的 runtime image ref
  - 不负责：决定镜像是否应该重建
- `challenge_service.go`
  - 负责：自检 runtime startup 时解析 image ref
  - 不负责：题目包导入策略
- `awd_preview_runtime_support.go`
  - 负责：AWD preview runtime startup 时解析 image ref
  - 不负责：preview checker 业务逻辑本身

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-15-awd-runtime-image-digest-implementation-plan.md`
- Modify: `code/backend/internal/model/image.go`
- Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_preview_runtime_support.go`
- Modify: `code/backend/internal/module/practice/application/commands/runtime_container_create_test.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service_self_check_test.go`
- Modify: `code/backend/internal/module/contest/application/commands/awd_service_test.go`

## Task 1: 收口 runtime image ref helper

- [ ] 在 `model.Image` 或等价 owning surface 中提供统一 helper
- [ ] helper 规则：
  - 有 `Digest` 时返回 `name@digest`
  - 否则返回 `name:tag`
  - 输入缺字段时给出最小安全回退

验证：

- `cd /home/azhi/workspace/projects/ctf && go test ./code/backend/internal/module/...`

Review focus：

- helper 不要偷偷改变非 runtime 场景下的镜像标识语义
- digest 优先逻辑必须是显式、可复用、无副作用的

## Task 2: 替换关键运行入口

- [ ] 实例启动改为通过 helper 解析 image ref
- [ ] 题目自检 runtime startup 改为通过 helper 解析 image ref
- [ ] AWD preview runtime startup 改为通过 helper 解析 image ref

验证：

- 定向 `go test` 覆盖 practice / challenge / contest 三条链路

Review focus：

- 不要只修 AWD 实例启动而漏掉 preview 或 self-check
- 不能影响无 digest 镜像的旧行为

## Task 3: 补契约测试

- [ ] practice runtime test 断言有 digest 时节点镜像使用 `name@digest`
- [ ] challenge self-check test 断言 startup 使用 `name@digest`
- [ ] AWD preview test 断言 runtime probe 使用 `name@digest`

验证：

- `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/practice/application/commands ./internal/module/challenge/application/commands ./internal/module/contest/application/commands -run 'Test(CreateSingleAWDContainerUsesPrivateTopology|ChallengeSelfCheckSingleContainerSuccess|AWDServicePreviewCheckerUsesCheckerTokenForHTTPStandard)' -count=1 -timeout 300s`

Review focus：

- 测试必须覆盖 digest 存在和不存在两类路径中的至少关键一类
- 断言点要锁在实际传给 runtime 的 image ref，而不是只看数据库字段

## Risks

- 如果只改实例启动，不改自检和 preview，会继续出现“平台自检 / preview 看到的是另一个版本”的分叉
- 如果 helper 无条件返回 `name@digest`，而 digest 字段为空或格式异常，可能导致原本可运行的镜像被错误拒绝
- 如果继续依赖 raw `image_ref` 字符串而不走 `image_id` 解析，旧 contest snapshot 仍可能绕过修复

## Verification Plan

1. `cd /home/azhi/workspace/projects/ctf && git diff --check`
2. `cd /home/azhi/workspace/projects/ctf/code/backend && go test ./internal/module/practice/application/commands ./internal/module/challenge/application/commands ./internal/module/contest/application/commands -count=1 -timeout 300s`
3. 如有必要，补跑当前 AWD 场景的定向测试用例

## Architecture-Fit Evaluation

- owner 明确：runtime image ref 的不可变性由镜像元数据 helper 统一负责，而不是散落在多个服务各自拼 `name:tag`
- reuse point 明确：实例启动、自检、preview 共用同一规则，不再各自复制一套 tag 解析
- 结构收敛明确：既不依赖“记得手动删本地旧镜像”，也不依赖“latest 刚好被 pull 到新内容”，而是直接消费镜像记录里已经存在的 digest
