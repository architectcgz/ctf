# 题包上传与 Registry 交付闭环设计

## 目标

把题目从“作者编写题包”到“平台可运行镜像”的链路收敛成一个可追踪、可重试、可阻断发布的闭环：

```text
编写题包 -> 上传预览 -> 平台构建 -> 推送 registry -> 拉取/试跑校验 -> 落库 -> 发布/配赛
```

这份文档定义题包 registry 交付闭环。当前实现已经具备题包上传预览、平台镜像命名、异步构建任务、Docker build/push、registry manifest 校验、pull/inspect 校验和导入落库能力；后续扩展以本文的边界继续推进。

## 设计结论

默认路径应当由平台构建镜像并上传到平台 registry。上传者可以选择引用已有镜像，但这是兼容路径，不是主路径。

平台构建路径下，最终镜像地址由平台生成，不由题包作者决定：

```text
<registry>/awd/<slug>:<tag>
<registry>/jeopardy/<slug>:<tag>
```

例如：

```text
registry.example.edu/awd/awd-demo:c1
registry.example.edu/jeopardy/web-demo:v1
```

本地开发环境对应：

```text
127.0.0.1:5000/awd/<slug>:<tag>
127.0.0.1:5000/jeopardy/<slug>:<tag>
```

容器运行归属仍然使用平台运行时 label，例如 `ctf.project=ctf`、`managed-by=ctf-platform`、`ctf-component=challenge-instance`。镜像 repository 只表达题型目录和题目名，不承担运行实例归属判断。

## 非目标

- 不依赖 Docker Compose project 作为平台实例归属依据。
- 不把老师本地 `docker compose` 的项目名、目录名或镜像 label 当作平台事实源。
- 不在上传 HTTP 请求内同步执行长时间 build。
- 不要求所有外部已有镜像都迁入平台 registry；外部引用路径保留兼容能力。
- 不在题目包中保存平台 registry 密码、Docker daemon 配置或生产密钥。

## 题包作者契约

题包必须保留可复现构建上下文：

```text
<slug>/
├── challenge.yml
├── statement.md
└── docker/
    ├── Dockerfile
    └── ...
```

`challenge.yml` 中的 `runtime.image` 在平台构建路径下只作为建议信息：

- `runtime.image.tag` 或 `runtime.image.ref` 中的 tag 可作为默认 tag 建议。
- 如果没有 tag，平台生成 `latest`、导入批次号或管理员指定 tag。
- repository/project 由平台统一生成，不读取作者本地 project。

平台构建路径下，上传者不需要填写 `team/web-demo:v1`、`registry.example.edu/team/web-demo:v1` 这类最终镜像名称。题包可以省略 `runtime.image.ref`，也可以写本地调试用 ref，但平台只读取其中的 tag 建议；repository、namespace、registry 都由平台生成。最终入库和运行使用平台生成后的 `image_ref`。

引用已有镜像路径下，题包必须显式声明完整 `runtime.image.ref`，平台不静默改写，只做拉取、inspect 和试跑校验。

多节点拓扑题遵循同一原则。`docker/topology.yml` 中的 `nodes[].image.ref` 在平台构建路径下也只作为本地调试或 tag 建议信息；最终每个节点运行镜像必须由平台生成并写回拓扑快照。第一阶段可以只支持单构建上下文单镜像，后续再扩展为节点级 `build.context` / `dockerfile`；但不能让拓扑节点继续散落使用作者自定义 repository。

## 两条导入路径

### 1. 平台构建 Dockerfile

默认路径。

上传者提交的是源码题包和 Dockerfile，不提交最终镜像名。平台在构建任务中生成最终 image ref，并把该 ref 作为后续 push、校验、落库和运行的唯一事实。

流程：

1. 上传题包。
2. 解析 `challenge.yml`，识别 `meta.mode`：
   - 空或 `jeopardy` -> `jeopardy`
   - `awd` -> `awd`
3. 校验题包结构、题面、Flag、AWD 扩展、`defense_scope`、拓扑等静态契约。
4. 创建镜像构建任务，状态为 `pending`。
5. builder 在隔离环境内执行 `docker build`。
6. 平台按题型生成目标镜像：
   - `registry/jeopardy/<slug>:<tag>`
   - `registry/awd/<slug>:<tag>`
7. 如果题包包含拓扑，使用平台生成后的 image ref 归一化入口节点和所有节点镜像引用。
8. push 到平台 registry。
9. 通过 registry manifest、docker pull、image inspect 做可用性校验。
10. 镜像变为 `available` 后，发布自检或 AWD checker preview 才允许进入后续发布/配赛。

### 2. 引用已有镜像

兼容路径。

流程：

1. 上传题包。
2. `runtime.image.ref` 必须是完整镜像引用。
3. 平台不重命名、不重新 push。
4. 平台使用配置的 registry auth 尝试 pull 或 manifest check。
5. inspect/试跑通过后，写入镜像记录。
6. 失败时保留导入预览，但阻断 commit 或发布。

这条路径用于已有 CI、外部题库迁移或临时演示。它不能成为默认体验，否则镜像命名空间、构建日志和运行一致性会继续分散。

## 镜像命名规则

平台构建路径统一使用：

```text
<registry>/<mode>/<slug>:<tag>
```

字段规则：

- `registry` 来自平台配置，例如 `container.registry.server`。
- `mode` 只允许 `awd` 或 `jeopardy`。
- `slug` 来自 `meta.slug`，必须满足题包 slug 校验。
- `tag` 优先级：
  1. 管理员上传时显式填写
  2. `runtime.image.ref` 或 `runtime.image.tag` 中的 tag
  3. 平台生成的导入批次 tag
  4. 本地开发可使用 `latest`

平台构建路径下，即使题包写了：

```text
team/web-demo:v1
registry.example.edu/team/web-demo:v1
```

最终平台生成的镜像仍然是：

```text
<registry>/jeopardy/web-demo:v1
```

因此，平台构建模式的上传表单不应要求上传者填写完整镜像名称；最多提供 tag 输入框。引用已有镜像路径才保留作者显式地址。

多镜像题目需要在同一命名空间下派生节点名：

```text
<registry>/<mode>/<slug>/<node_key>:<tag>
```

例如：

```text
registry.example.edu/jeopardy/web-demo/web:v1
registry.example.edu/jeopardy/web-demo/db:v1
```

如果题包只有一个构建上下文或只有一个入口服务，仍使用 `<registry>/<mode>/<slug>:<tag>`，避免给普通题目制造不必要的路径层级。

## 状态模型

镜像记录不能因为“创建了数据库行”就直接变为 `available`。目标状态流转：

```text
pending -> building -> pushed -> verifying -> available
                               -> failed
```

建议最小字段：

- `images.status`
- `images.name`
- `images.tag`
- `images.digest`
- `images.source_type`：`platform_build` / `external_ref`
- `images.build_job_id`
- `images.last_error`
- `images.verified_at`

构建任务建议独立表：

- `image_build_jobs.id`
- `source_type`
- `challenge_mode`
- `package_slug`
- `source_dir` 或归档路径
- `dockerfile_path`
- `context_path`
- `target_ref`
- `target_digest`
- `status`
- `started_at`
- `finished_at`
- `log_path`
- `error_summary`
- `created_by`

题目导入记录需要关联构建任务。只有 `available` 镜像可以进入发布或 AWD 服务 readiness 通过状态。

## 后端边界

建议拆成四个明确 owner：

1. **Package parser**
   - 只负责解析题包、静态校验、归一化题型和 slug。
   - 不执行 Docker build。

2. **Image build service**
   - 负责创建 build job、调度 builder、生成目标 image ref、保存日志和状态。
   - 拥有平台构建路径的最终 image ref 决策权。
   - 负责把单镜像和拓扑节点镜像都归一化到平台命名空间。

3. **Registry client**
   - 负责 login、push 后 manifest check、pull/inspect 校验。
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
