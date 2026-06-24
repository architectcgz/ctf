# challenge 模块设计

> 状态：Current
> 事实源：`code/backend/internal/module/challenge/`、`docs/contracts/challenge-pack-v1.md`
> 替代：无

## 定位

`challenge` 是题目、题包、附件、镜像、Flag、拓扑、题解和发布检查 owner，负责把题目作者侧与运行前校验相关能力收口到一个模块。

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
- `docs/contracts/challenge-pack-v1.md`：题包格式事实源。

## 已知限制

- `challenge` 仍包含多个较大的子用例族；当前通过 `application/challengecore`、`challengeimport`、`challengeselfcheck`、`challengepublishcheck`、`challengepackageexport` 降低单 service 膨胀。
- `TagHandler.CreateTag/ListTags/AttachTags/DetachTags` 在模块内存在，但当前 app 路由未公开接线；文档中标签事实以 `TagService` 和 `tags` / `challenge_tags` 数据表为准。
