# challenge 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/challenge/`、`docs/contracts/challenge-pack-v1.md`
> 替代：无

## 定位

`challenge` 是题目、题包、附件、镜像、Flag、拓扑、题解和发布检查 owner，负责把题目作者侧与运行前校验相关能力收口到一个模块。

## 本文档范围

| 本文档负责 | 本文档不负责（见其他文档） |
|-----------|-------------------------|
| challenge 模块的职责边界、HTTP 入口和用例组织 | 题包格式规范 → `docs/contracts/challenge-pack-v1.md` |
| 题目、Flag、附件、镜像、拓扑的 owner | 题包 Registry 交付流程 → `docs/architecture/features/题包Registry交付架构.md` |
| 题包导入/导出、发布检查的实现细节 | 题包拓扑同步与导出 → `docs/architecture/features/题包拓扑同步与导出架构.md` |
| 模块内部组件协作和数据流 | 容器镜像运行时调度 → `container_runtime` 模块文档 |

## 事实来源

- HTTP 入口：`code/backend/internal/module/challenge/api/http/`
- 用例编排：`code/backend/internal/module/challenge/application/`
- 领域解析：`code/backend/internal/module/challenge/domain/`
- 对外契约：`code/backend/internal/module/challenge/contracts/`
- 持久化与文件适配：`code/backend/internal/module/challenge/infrastructure/`
- 题包契约：`docs/contracts/challenge-pack-v1.md`
- 专题事实：`docs/architecture/features/题包Registry交付架构.md`、`docs/architecture/features/题包拓扑同步与导出架构.md`

## 当前设计

- `challenge/api/http`
  - 负责：题目、AWD 题目、镜像、Flag、标签、拓扑、题解、题包导入/导出的 HTTP 入口。
  - 不负责：直接访问 GORM、LocalFS、Docker CLI、Registry HTTP 或容器运行时。
- `challenge/application/challengecore`
  - 负责：普通题创建、更新、删除、发布和发布目录变更事件。
  - 不负责：题包导入、自检、发布检查 job 或导出。
- `challenge/application/challengeimport` 与 `commands.AWDChallengeImportService`
  - 负责：普通题包和 AWD 题包 preview / commit 编排。
  - 不负责：把 zip 解包、附件复制、preview JSON 持久化写在 application surface。
- `challenge/application/challengeselfcheck`、`challengepublishcheck`
  - 负责：发布前自检、镜像可用性、Flag 校验、发布检查 job 生命周期和通过后的发布事务。
  - 不负责：直接实现容器探针、Docker runtime 或 Registry manifest 请求。
- `challenge/application/challengepackageexport`
  - 负责：题包导出和导出修订读取。
  - 不负责：承担赛事导出、评估报告或复盘归档导出。
- `challenge/infrastructure`
  - 负责：PostgreSQL 仓储、LocalFS 附件/题包/导出存储、Docker image builder、Registry verifier、artifact GC。
  - 不负责：训练实例生命周期、竞赛编排或 AWD 轮次调度。

## API 入口设计

| 路由组 | 代表路由 | Handler | Service / 用例 |
| --- | --- | --- | --- |
| 教师出题 `authoring` | `POST/GET /api/v1/authoring/challenges`、`GET/PUT/DELETE /api/v1/authoring/challenges/:id` | `challenge/api/http.Handler` | `challengecore` 创建、更新、删除和管理端查询 |
| 发布检查 / 自检 | `POST /api/v1/authoring/challenges/:id/publish-requests`、`GET /api/v1/authoring/challenges/:id/publish-requests/latest`、`POST /api/v1/authoring/challenges/:id/self-check` | `Handler.RequestPublishCheck`、`GetLatestPublishCheck`、`SelfCheckChallenge` | `challengepublishcheck`、`challengeselfcheck` |
| 普通题包导入 | `POST/GET /api/v1/authoring/challenge-imports`、`GET /api/v1/authoring/challenge-imports/:id`、`POST /api/v1/authoring/challenge-imports/:id/commit` | `Handler.PreviewChallengeImport`、`ListChallengeImports`、`GetChallengeImport`、`CommitChallengeImport` | `challengeimport`、`PackageDeliveryService` |
| 普通题包导出 | `POST /api/v1/authoring/challenges/:id/package-export`、`GET /api/v1/authoring/challenges/:id/package-export/download` | `Handler.ExportChallengePackage`、`DownloadChallengePackageExport` | `challengepackageexport` |
| 镜像 | `POST/GET /api/v1/authoring/images`、`GET/PUT/DELETE /api/v1/authoring/images/:id` | `ImageHandler` | `ImageService`、`ImageBuildService` |
| 环境模板 | `GET/POST /api/v1/authoring/environment-templates`、`GET/PUT/DELETE /api/v1/authoring/environment-templates/:id` | `TopologyHandler` | `TopologyService` template use case |
| Flag / 拓扑 | `PUT/GET /api/v1/authoring/challenges/:id/flag`、`GET/PUT/DELETE /api/v1/authoring/challenges/:id/topology` | `FlagHandler`、`TopologyHandler` | `FlagService`、`TopologyService` |
| AWD 题目 | `GET/POST /api/v1/authoring/awd-challenges`、`GET/PUT/DELETE /api/v1/authoring/awd-challenges/:id` | `AWDChallengeHandler` | `AWDChallengeService` |
| AWD 题包导入 | `POST/GET /api/v1/authoring/awd-challenge-imports`、`GET /api/v1/authoring/awd-challenge-imports/:id`、`POST /api/v1/authoring/awd-challenge-imports/:id/commit` | `AWDChallengeHandler` | `AWDChallengeImportService` |
| 学生题目 | `GET /api/v1/challenges`、`GET /api/v1/challenges/:id`、`GET /api/v1/challenges/attachments/*path` | `Handler.ListPublishedChallenges`、`GetPublishedChallenge`、`DownloadAttachment` | published challenge query、附件下载 |
| 官方题解 | `GET /api/v1/challenges/:id/writeup`、`GET/PUT/DELETE /api/v1/authoring/challenges/:id/writeup`、`POST/DELETE /api/v1/authoring/challenges/:id/writeup/recommend` | `WriteupHandler` | official writeup command/query |
| 社区题解提交 | `POST /api/v1/challenges/:id/writeup-submissions`、`GET /api/v1/challenges/:id/writeup-submissions/me`、`GET /api/v1/teacher/writeup-submissions`、`GET /api/v1/teacher/writeup-submissions/:id` | `WriteupHandler` | submission writeup command/query |
| 社区题解推荐与可见性 | `POST/DELETE /api/v1/teacher/community-writeups/:id/recommend`、`POST /api/v1/teacher/community-writeups/:id/hide`、`POST /api/v1/teacher/community-writeups/:id/restore` | `WriteupHandler` | community moderation command |

## Application / Service 设计

| Service / 子包 | 代码路径 | 负责 | 不负责 |
| --- | --- | --- | --- |
| Challenge core | `challenge/application/challengecore/service.go` | 普通题创建、更新、删除、发布和目录事件 | 导入、导出、自检、镜像构建 |
| Challenge import | `challenge/application/challengeimport/service.go` | 题包 preview / list / get / commit 编排 | LocalFS 解包和附件复制实现 |
| Self-check | `challenge/application/challengeselfcheck/service.go` | 发布前 Flag、镜像和 runtime probe 自检 | Docker concrete |
| Publish-check | `challenge/application/challengepublishcheck/service.go` | 发布检查 job 状态机、后台轮询、通过后发布 | 通知推送 |
| Package export | `challenge/application/challengepackageexport/service.go` | 题包导出和导出修订读取 | 赛事/报告导出 |
| Image service / build service | `challenge/application/commands/image_service.go`、`image_build_service.go` | 镜像 CRUD、平台构建、外部引用校验、状态写入 | Registry HTTP 细节、Docker CLI 细节 |
| AWD challenge service | `challenge/application/commands/awd_challenge_service.go`、`awd_challenge_import_service.go` | AWD 题目维护和 AWD 题包导入 | AWD 轮次和队伍运行态 |
| Flag / topology / writeup service | `challenge/application/commands/*_service.go`、`queries/*_service.go` | 已接线子资源读写与错误映射 | 训练提交判分 |
| Tag service | `challenge/application/commands/tag_service.go`、`queries/tag_service.go` | 标签与题目关联的模块内用例；`TagHandler` 当前没有在 `code/backend/internal/app/router_*.go` 接成公开 HTTP 路由 | 题目列表聚合、训练提交判分 |

## 数据设计

| 表 / 存储 | Entity / Adapter | Owner 语义 | 主要写入方 |
| --- | --- | --- | --- |
| `challenges` | `challenge/entity.Challenge` | 普通题元数据、发布状态、Flag 配置、运行入口基础字段 | challenge core / flag command |
| `challenge_hints` | `ChallengeHint` | 题目提示 | challenge command |
| `images` | `Image` | 镜像引用、构建状态、digest、source type | image command / image build |
| `image_build_jobs` | `ImageBuildJob` | 平台镜像构建任务状态和日志路径 | image build service |
| `tags`、`challenge_tags` | `Tag`、`ChallengeTag` | 标签与题目标签关联 | tag service |
| `awd_challenges` | `AWDChallenge` | AWD 题目服务定义、checker、runtime、readiness | AWD challenge service |
| `challenge_topologies`、`environment_templates` | `ChallengeTopology`、`EnvironmentTemplate` | 拓扑和环境模板 | topology service |
| `challenge_writeups`、`submission_writeups` | `ChallengeWriteup`、`SubmissionWriteup` | 官方题解、社区题解和提交题解 | writeup service |
| `challenge_package_revisions` | `ChallengePackageRevision` | 题包导入/导出修订、source/archive/topology snapshot | import / export service |
| `challenge_publish_check_jobs` | `ChallengePublishCheckJob` | 发布检查 job 状态与结果 | publish-check service |
| LocalFS `data/challenge-attachments` 等 | `challenge_attachment_store.go`、`challenge_package_storage.go`、`challenge_import_preview_store.go` | 附件、导入预览、题包源、导出包、checker artifact | challenge infrastructure |

Redis 仅用于部分缓存或 artifact 辅助状态时由 infrastructure adapter 收口；题目事实以 PostgreSQL 和 LocalFS 持久化产物为准。

## 边界

- `challenge` 拥有题目元数据、发布状态、题目附件、镜像定义、Flag 规则、拓扑定义和题解事实。
- `practice` 和 `contest` 消费 challenge catalog / image / flag validator / runtime subject，不拥有题目事实。
- `container_runtime` 只提供镜像探针、容器文件和 runtime capability，不决定题目发布语义。
- `assessment` 消费题目维度和目录信息生成画像，不写 challenge 状态。

## 主要用例

- 教师 / 管理员维护题目、标签、镜像、Flag、题解。
- 上传题包、预览导入、提交导入、导出题包修订。
- 发布前自检与发布检查 job。
- AWD 题目配置、检查器 artifact、拓扑与题包交付。
- 给 `practice`、`contest`、`assessment` 提供题目目录、运行定义和维度信息。

## Flag 配置与动态生成

### Flag 类型

`challenge` 模块支持三种 Flag 类型，覆盖不同场景：

| Flag 类型 | 代码常量 | 存储形式 | 适用场景 |
| --- | --- | --- | --- |
| **Static Flag** | `FlagTypeStatic` | 哈希 + salt 存储在 `challenges.flag_hash` / `flag_salt` | 固定 flag，所有用户提交相同值 |
| **Dynamic Flag** | `FlagTypeDynamic` | 不存储明文，运行时根据 `{{ .InstanceID }}` 等变量生成 | 每个实例独立 flag，防止直接抄袭 |
| **Regex Flag** | `FlagTypeRegex` | 正则表达式存储在 `challenges.flag_regex` | 接受符合模式的任意 flag |

### Dynamic Flag 模板变量

动态 Flag 支持以下模板变量，运行时由 `practice` 或 `contest` 模块在创建实例时替换：

- `{{ .InstanceID }}`：训练实例 ID 或竞赛实例 ID
- `{{ .UserID }}`：当前用户 ID
- `{{ .ContestID }}`：竞赛 ID（仅竞赛场景）

模板解析与变量替换时机：
- **训练场景**：`practice` 模块在 `instance_provisioning_service.go` 创建实例时调用 Flag 生成逻辑
- **竞赛场景**：`contest` 模块在分配题目实例时生成
- **AWD 场景**：使用确定性算法 `contestdomain.BuildAWDRoundFlag(contestID, roundNumber, teamID, serviceID, flagSecret)`

### Flag 配置与校验

- **Static Flag 配置**：`FlagService.ConfigureStaticFlag()` 校验格式（`prefix{content}` 模式）、长度（≤256 字符），生成随机 salt 并计算哈希存储
- **Dynamic Flag 配置**：`FlagService.ConfigureDynamicFlag()` 检查实例共享策略，不允许 `InstanceSharingShared` 策略使用动态 Flag
- **Regex Flag 配置**：`FlagService.ConfigureRegexFlag()` 编译正则表达式验证合法性
- **Flag 校验**：`practice` 和 `contest` 模块提交判分时调用 `challenge` 模块提供的 Flag validator

**代码位置**：
- `code/backend/internal/module/challenge/application/commands/flag_service.go`：Flag 配置用例
- `code/backend/internal/module/challenge/infrastructure/flag_repository.go`：Flag 持久化适配
- `code/backend/internal/shared/flagcrypto/`：Flag 哈希与动态生成算法

**相关专题**：
- 动态 Flag 生成算法和 salt 管理 → `docs/architecture/backend/03-container-architecture.md`（容器启动环境变量注入）

## 附件管理

### 附件存储路径结构

题目附件统一存储在 LocalFS，路径结构为：

```
<storage-root>/challenges/<challenge-id>/attachments/<filename>
```

- `<storage-root>`：由 `platformstorage.LocalWritableStore` 配置，默认为 `data/challenge-attachments`
- `<challenge-id>`：题目 ID 或题包 slug（导入时使用 package slug）
- `<filename>`：附件文件名，经过 `sanitizeImportedAttachmentName()` 清洗，移除路径遍历字符

### 附件元数据

附件元数据存储在 `challenge_attachments` 表，包含：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | bigint | 附件唯一标识 |
| `challenge_id` | bigint | 所属题目 ID |
| `filename` | varchar | 原始文件名 |
| `storage_key` | varchar | LocalFS 存储路径（相对路径） |
| `size_bytes` | bigint | 文件大小 |
| `content_type` | varchar | MIME 类型 |

### 附件访问控制与安全

- **访问控制**：学生端附件下载路由 `GET /api/v1/challenges/attachments/*path` 通过 `Handler.DownloadAttachment()` 处理，只允许下载已发布题目的附件
- **路径遍历防护**：
  - 存储时通过 `safePathSegment()` 和 `sanitizeImportedAttachmentName()` 清洗文件名
  - 读取时通过 `platformstorage.LocalWritableStore.Open()` 校验相对路径，拒绝包含 `..` 的路径
  - 错误时返回 `platformstorage.ErrUnsafeKey`，映射为 `apperror.ErrInvalidParams`
- **多附件打包**：题包导入时如果包含多个附件，自动打包为 `<package-slug>-attachments.zip` 单文件存储

**代码位置**：
- `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store.go`：附件存储适配器
- `code/backend/internal/module/challenge/api/http/handler_attachment_test.go`：附件下载与路径遍历防护测试
- `code/backend/internal/platform/storage/local_writable_store.go`：底层 LocalFS 安全封装

**相关专题**：
- 附件存储与题包交付 → `docs/architecture/features/题包Registry交付架构.md`

## 数据与副作用

- PostgreSQL：`challenge/entity/*.go` 对应题目、镜像、标签、拓扑、发布检查、题包修订、AWD 题目等持久化事实。
- LocalFS：附件、导入预览、题包 source/export archive、AWD checker artifact、构建 source。
- Docker / Registry：镜像构建通过 `docker_image_builder.go`，外部镜像校验通过 `registry_client.go`。
- Outbox：发布检查完成写入 `challenge.publish_check_finished`，由 `ops` 消费通知。

## 跨模块依赖

| 依赖 | 用途 | 接线 |
| --- | --- | --- |
| `container_runtime` | 镜像探针、runtime probe、容器文件能力 | `composition.BuildChallengeModule` |
| `practice` | 训练开题和提交消费题目运行定义、Flag 校验 | app composition 注入 challenge module |
| `contest` | 赛事题目编排、AWD 运行定义与检查器配置 | app composition 注入 challenge module |
| `assessment` | 技能画像和推荐消费题目维度 | `composition.BuildAssessmentModule` |
| `ops` | 发布检查事件通知 | platform outbox handler |

## Guardrail

- `code/backend/internal/module/challenge/architecture_test.go`：约束 challenge 分层、ports 粒度、contracts 不重导出 taxonomy。
- `code/backend/internal/module/challenge/ports/*_context_contract_test.go`：约束端口 context 传播。
- `code/backend/internal/app/challenge_import_integration_test.go`：覆盖题包导入装配。
- `code/backend/internal/module/challenge/application/commands/flag_service_test.go`：覆盖 Static / Dynamic / Regex Flag 配置用例。
- `code/backend/internal/module/challenge/api/http/handler_attachment_test.go`：覆盖附件下载与路径遍历防护。
- `code/backend/internal/module/challenge/infrastructure/challenge_attachment_store_test.go`：覆盖附件存储路径清洗与多附件打包。
- `docs/contracts/challenge-pack-v1.md`：题包格式事实源。

## 已知限制

- `challenge` 仍包含多个较大的子用例族；当前通过 `application/challengecore`、`challengeimport`、`challengeselfcheck`、`challengepublishcheck`、`challengepackageexport` 降低单 service 膨胀。
- `TagHandler.CreateTag/ListTags/AttachTags/DetachTags` 在模块内存在，但当前 app 路由未公开接线；文档中标签事实以 `TagService` 和 `tags` / `challenge_tags` 数据表为准。
