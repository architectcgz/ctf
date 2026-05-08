# 题包 Registry 交付架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`backend`、`frontend`、`contracts`
- 关联模块：
  - `internal/module/challenge/domain`
  - `internal/module/challenge/application/commands`
  - `internal/config`
  - `frontend/src/components/platform/challenge`
  - `frontend/src/components/platform/images`
- 最后更新：`2026-05-07`

## 1. 背景与问题

题目导入要解决的不是“把镜像地址保存下来”，而是把题包从上传预览、镜像生成或验证，到最终可运行镜像这一整段链路稳定收口。若没有统一规则，会出现三个问题：

- 题包作者自定义镜像命名空间，平台无法保证不同题型和不同环境下的镜像归属一致。
- 导入成功不代表镜像真正可拉取、可校验、可运行。
- 长时间的 `docker build/push/pull/inspect` 若直接挂在同步 HTTP 链路里，导入体验和失败恢复都会很差。

当前需要收口的是：题包镜像交付的最终事实源不是“作者写了什么镜像名”，而是平台导入服务和镜像构建服务共同维护的 `images + image_build_jobs` 状态。

## 2. 架构结论

- 默认交付路径是 `platform_build`：平台根据题型、slug 和 tag 生成最终镜像地址，再异步构建、推送和校验。
- 兼容交付路径是 `external_ref`：平台接受题包显式声明的完整镜像引用，并在导入提交时立即做 manifest / pull / inspect 校验。
- 管理员手工维护镜像目录仍然存在，但它是独立的 `manual` 路径，不参与题包导入时的内部镜像记录创建。
- 平台生成的目标镜像命名规则统一为 `<registry>/<mode>/<slug>:<tag>`，其中 `mode` 当前为 `jeopardy` 或 `awd`。
- 题包导入服务在事务内直接创建或更新 `images`、`image_build_jobs`；不会通过 HTTP 自调用镜像管理接口。
- 镜像只有在 `manifest check + docker pull + docker image inspect` 全部通过后，才进入 `available`。

## 3. 模块边界与职责

### 3.1 模块清单

- `internal/module/challenge/domain/image_delivery`
  - 负责：生成平台镜像地址、拆分镜像引用、提取 tag 建议
  - 不负责：执行 build、push 或 registry 校验

- `internal/module/challenge/application/commands/challenge_import_service`
  - 负责：普通题包预览、导入提交时解析镜像来源，并在事务内创建平台构建任务或校验外部镜像
  - 不负责：后台异步执行镜像构建

- `internal/module/challenge/application/commands/awd_challenge_import_service`
  - 负责：AWD 题包沿用同一套镜像交付语义，但使用 `awd` 命名空间
  - 不负责：定义另一套独立镜像状态机

- `internal/module/challenge/application/commands/image_build_service`
  - 负责：创建 `image_build_jobs`、处理构建队列、推进镜像状态机、执行 push 后校验
  - 不负责：解析题包结构

- `internal/module/challenge/application/commands/docker_image_builder`
  - 负责：执行 `docker build/push/pull/inspect`
  - 不负责：决定目标镜像命名规则

- `internal/module/challenge/application/commands/registry_client`
  - 负责：对配置的 registry 做 manifest 检查并返回 digest
  - 不负责：本地镜像构建

- `frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`
  - 负责：在导入预览中展示镜像来源、建议 tag、目标镜像和当前状态
  - 不负责：前端自行生成目标镜像名

- `frontend/src/components/platform/images/ImageDirectoryPanel.vue`
  - 负责：展示镜像目录中的 `source_type / status / digest / verified_at`
  - 不负责：代替导入服务推进构建

### 3.2 事实源与所有权

- 平台 registry 配置事实源：`container.registry`
- 导入预览镜像交付事实源：`ChallengeImportImageDeliveryResp`
- 镜像记录事实源：`images`
- 构建任务事实源：`image_build_jobs`

## 4. 关键模型与不变量

### 4.1 镜像来源类型

- `manual`
- `platform_build`
- `external_ref`

### 4.2 镜像状态模型

`images.status` 与 `image_build_jobs.status` 当前共享同一条主状态流转：

```text
pending -> building -> pushed -> verifying -> available
                               -> failed
```

### 4.3 关键字段

- `images`
  - `name`
  - `tag`
  - `status`
  - `digest`
  - `source_type`
  - `build_job_id`
  - `last_error`
  - `verified_at`

- `image_build_jobs`
  - `source_type`
  - `challenge_mode`
  - `package_slug`
  - `source_dir`
  - `dockerfile_path`
  - `context_path`
  - `target_ref`
  - `target_digest`
  - `status`
  - `error_summary`
  - `created_by`
  - `started_at`
  - `finished_at`

### 4.4 不变量

- 平台构建路径下，最终镜像地址只能由服务端生成，题包中的 `runtime.image.ref` 最多提供 tag 建议。
- `external_ref` 校验当前要求镜像属于已配置的 registry server，平台不会对外部引用静默改名。
- 导入服务创建内部镜像记录时不经过 `/authoring/images` HTTP 接口。
- Docker registry 认证通过临时 `DOCKER_CONFIG` 注入，不把密码或 identity token 拼进 `docker` 命令参数。
- 当前实现按“一个题包对应一个最终镜像记录”组织交付；更复杂的多镜像拓扑命名还不属于本专题当前事实源。

## 5. 关键链路

### 5.1 导入预览链路

1. 上传题包后，解析器判断镜像来源是 `platform_build` 还是 `external_ref`。
2. 预览接口返回 `image_delivery.source_type`、`suggested_tag`。
3. 若是 `platform_build`，服务端提前计算 `target_image_ref`，并在预览中显示 `build_status=pending`。

### 5.2 平台构建链路

1. 普通题或 AWD 题提交导入时，导入服务在事务内调用 `CreatePlatformBuildJobInTx`。
2. 平台生成目标镜像名：
   - `registry/<mode>/<slug>:<tag>`
3. 事务内创建或更新 `images` 与 `image_build_jobs`，初始状态为 `pending`。
4. `ImageBuildService` 后台轮询待处理 job，并依次执行：
   - `docker build`
   - `docker push`
   - registry manifest check
   - `docker pull`
   - `docker image inspect`
5. 全部通过后，镜像状态更新为 `available`；任一环节失败则更新为 `failed` 并记录错误摘要。

### 5.3 外部镜像引用链路

1. 题包显式声明完整 `runtime.image.ref`。
2. 导入提交时调用 `VerifyExternalImageRefInTx`。
3. 事务内完成 manifest check、pull、inspect，并更新镜像记录为 `available` 或 `failed`。
4. 失败时阻断导入提交，不生成“看起来成功但不可用”的运行镜像。

## 6. 接口与契约

### 6.1 预览契约

`ChallengeImportImageDeliveryResp` 当前稳定字段：

- `source_type`
- `suggested_tag`
- `target_image_ref`
- `build_status`
- `digest`
- `last_error`

### 6.2 管理侧镜像目录契约

镜像目录当前暴露的关键信息包括：

- `name`
- `tag`
- `status`
- `digest`
- `source_type`
- `build_job_id`
- `last_error`
- `verified_at`

## 7. 兼容与迁移

- `external_ref` 仍然保留，适合已有 CI 或已有 registry 资产的题包迁移。
- `manual` 镜像目录仍然保留，适合管理员手工维护镜像记录，但它不再是题包导入默认路径。
- 未来如果扩展为多节点或多镜像构建，应继续由平台生成最终命名空间，而不是退回作者自定义 repository。

## 8. 代码落点

- `code/backend/internal/model/image.go`
- `code/backend/internal/model/image_build_job.go`
- `code/backend/internal/module/challenge/domain/image_delivery.go`
- `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- `code/backend/internal/module/challenge/application/commands/docker_image_builder.go`
- `code/backend/internal/module/challenge/application/commands/registry_client.go`
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- `code/backend/internal/config/config.go`
- `code/frontend/src/api/admin/authoring.ts`
- `code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`
- `code/frontend/src/components/platform/images/ImageDirectoryPanel.vue`

## 9. 验证标准

- 导入预览可以稳定展示镜像来源、建议 tag 和平台目标镜像地址。
- 提交导入时，平台构建路径会创建 `images` 与 `image_build_jobs`，外部镜像路径会完成即时校验。
- 只有通过 `push + manifest + pull + inspect` 的镜像才会进入 `available`。
- 管理员镜像目录能看到镜像的来源、状态、摘要和验证时间。
   - 屏蔽 HTTP registry、TLS registry、认证失败等细节。

4. **Challenge import service**
   - 负责把题包预览、构建任务、镜像记录和题目实体串起来。
   - commit 时只接受已验证镜像，或创建构建任务后进入等待确认状态。
   - 保存题包 revision、拓扑基线和最终 image ref 的对应关系。

运行时 `runtime` 模块只消费已验证 image ref，不参与题包解析和镜像命名决策。

## 前端交互

上传题目时显示镜像来源：

```text
镜像来源
- 平台构建 Dockerfile（默认）
- 引用已有镜像
```

平台构建模式字段：

- 题包文件
- 镜像 tag
- 构建资源限制或超时（可后置）

引用已有镜像模式字段：

- `runtime.image.ref`
- registry 凭据来源提示
- 拉取校验结果

导入预览应展示：

- 最终 image ref
- 构建状态
- registry digest
- 最近错误摘要
- build 日志入口
- 重新构建 / 重新检测操作

AWD 题目加入赛事前，readiness 必须检查镜像状态；镜像不可用时不允许进入服务配置通过状态。

## 失败处理

### build 失败

状态：`failed`

保存：

- Dockerfile 阶段
- stderr 摘要
- 完整日志路径
- 失败时间

动作：

- 不写 `images.status=available`
- 不允许发布
- 允许上传者修正题包后重新上传

### push 失败

常见原因：

- registry auth failed
- registry unavailable
- network timeout
- denied: requested access to the resource is denied

动作：

- build artifact 可保留短期缓存
- job 状态为 `failed`
- 管理员修复 registry 后可重试 push

### pull / manifest / inspect 失败

常见原因：

- manifest not found
- TLS handshake timeout
- no matching platform linux/amd64
- registry auth failed

动作：

- `images.status=failed` 或保持 `verifying`
- 题目不可发布
- AWD service readiness 不通过

### 运行预览失败

动作：

- 镜像可以保留为 `pushed` 或 `verifying_failed`
- 题目不可发布
- 保存 runtime failure summary
- 支持重新试跑

## 安全与资源限制

平台构建必须在隔离 builder 中运行，不能直接在 API 进程里执行不受控命令。

最低要求：

- build job 有超时。
- build 并发受限。
- 构建上下文限制在题包根目录内。
- 禁止读取宿主机敏感路径。
- registry 凭据只由平台配置注入，不进入题包。
- 构建日志脱敏，不输出 registry password。
- 可配置是否允许访问外网依赖源。

后续可以增加：

- 镜像扫描
- SBOM
- 基础镜像白名单
- 构建缓存策略
- 多架构构建策略

## 与现有文档的关系

`docs/contracts/challenge-pack-v1.md` 当前仍写着“AWD 题目包只声明运行镜像引用，不承担镜像构建职责”。这是当前实现边界，不是目标架构。

本文落地后，需要同步更新：

- `docs/contracts/challenge-pack-v1.md`
- `challenges/awd/challenge-package-contract.md`
- registry 构建脚本说明
- 管理端上传页面说明
- API 契约和 OpenAPI

在实现完成前，不应把题包契约改成“已支持平台构建”，避免文档提前承诺未落地能力。

当前本地开发约定已经补充为：

- `docker/ctf/infra/registry/ctf-platform-registry.env` 是平台后端与 `ctf-api` 唯一读取的 registry 配置事实源。
- `scripts/registry/build-and-push-challenge-image.sh` 同时支持 Jeopardy 的 `docker/Dockerfile` 和 AWD 的 `docker/runtime/Dockerfile`。
- `scripts/registry/sync-challenge-registry-state.py` 负责把 registry 回灌结果同步到 `images`、`challenges`、`awd_challenges`、`contest_awd_services`。

## 实施阶段

### 阶段 1：命名与引用统一

- AWD 使用 `awd/<slug>:<tag>`。
- Jeopardy 使用 `jeopardy/<slug>:<tag>`。
- 迁移开发库和示例镜像。
- 调整本地构建脚本，让平台生成 repository。

### 阶段 2：导入默认补全

- 普通题包默认 `mode=jeopardy`。
- AWD 题包默认 `mode=awd`。
- 平台构建模式下生成最终 image ref。
- 拓扑节点 image ref 统一按平台命名空间重写。
- 引用已有镜像模式下保留外部 ref。

### 阶段 3：构建任务系统

- 新增 `image_build_jobs`。
- 后台上传 commit 后进入 build job。
- 保存日志、digest、失败原因。
- 支持重试。

### 阶段 4：registry 校验与 readiness gate

- push 后 manifest check。
- pull/inspect 校验。
- Jeopardy 发布和 AWD readiness 阻断不可用镜像。
- 管理端提供重新检测。

### 阶段 5：安全加固

- builder 隔离。
- 资源限制。
- 日志脱敏。
- 依赖源策略。
- 镜像扫描。

## 验收标准

- 教师上传 Jeopardy 题包后，平台可构建并推送到 `registry/jeopardy/<slug>:<tag>`。
- 管理员上传 AWD 题包后，平台可构建并推送到 `registry/awd/<slug>:<tag>`。
- registry manifest、pull、inspect、发布自检或 checker preview 失败时，题目不能发布或加入 AWD 可用服务。
- 数据库中的 `images.status=available` 只表示镜像已验证可运行。
- 管理端能看到构建日志、最终 image ref、digest 和失败摘要。
- 引用已有镜像路径仍可使用，但必须显式选择并通过拉取校验。
