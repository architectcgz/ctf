# Challenge Package Registry Delivery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现“编写题包 -> 上传 -> 平台构建 -> 推送 registry -> 校验 -> 落库/发布”的可追踪闭环。

**Architecture:** 现有题包 parser 继续只负责解析和静态校验；新增镜像交付层负责生成平台 image ref、创建构建任务、执行 build/push/verify，并把结果写回 `images` 与导入记录。Jeopardy 与 AWD 导入入口共享镜像交付能力，但保持各自题目落库语义；运行时只消费已验证 `available` 镜像。

**Tech Stack:** Go、Gin、GORM、PostgreSQL migrations、Docker CLI/Engine、Registry HTTP v2、Vue 3、Vite、Vitest。

---

## 输入文档

- `docs/architecture/features/challenge-package-registry-delivery-design.md`
- `docs/contracts/challenge-pack-v1.md`
- `docs/architecture/features/package-topology-sync-design.md`
- `docs/architecture/backend/03-container-architecture.md`

## 当前实现事实

- Jeopardy 上传入口：
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - `PreviewChallengeImport` 解压 zip、解析 `challenge.yml`、保存本地 preview JSON。
  - `CommitChallengeImport` 重新解析 preview source，并在事务里调用 `resolveImportedImageID`。
- AWD 上传入口：
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - `PreviewImport` 与 Jeopardy 类似。
  - `CommitImport` 当前要求 `runtime.image.ref`，并把 `image_ref/image_id` 写入 `runtime_config`。
- 当前最大问题：
  - `resolveImportedImageID` 只按字符串拆 `name/tag`，创建或更新 `images.status=available`，没有 build、push、manifest、pull、inspect 或试跑证据。
  - AWD parser 当前强制 `runtime.image.ref`，和目标“平台构建模式可省略最终镜像名”冲突。
  - `scripts/registry/build-and-push-challenge-image.sh` 仍从 `runtime.image.ref` 反推 repository，默认 `ctf/<slug>`，和目标 `<registry>/<mode>/<slug>:<tag>` 不一致。
  - `images` 表只有 `name/tag/description/size/status`，无法表达 digest、source type、build job、错误摘要和验证时间。
  - 后端只有运行时 registry 拉取认证，缺少平台 build/push/manifest client 与任务系统。
- 前端现状：
  - Jeopardy 上传预览显示 `preview.runtime.image_ref`，位置包括 `ChallengePackageImportReview.vue`。
  - AWD 上传列表显示 `runtime_config` JSON，位置包括 `AWDChallengeLibraryPage.vue`。
  - 镜像管理仍是手工登记镜像，位置包括 `ImageManage.vue` 与 `features/image-management`。

## 文件结构

### 后端模型与迁移

- Modify: `code/backend/internal/model/image.go`
  - 增加 `Digest`、`SourceType`、`BuildJobID`、`LastError`、`VerifiedAt`。
  - 增加状态常量 `pushed`、`verifying`，保留 `pending/building/available/failed`。
- Create: `code/backend/internal/model/image_build_job.go`
  - 定义 `ImageBuildJob` 和状态/source 常量。
- Create: `code/backend/migrations/000005_create_image_build_jobs.up.sql`
- Create: `code/backend/migrations/000005_create_image_build_jobs.down.sql`

### 后端领域与服务

- Modify: `code/backend/internal/module/challenge/domain/package_manifest.go`
  - 增加 parsed 字段：镜像来源、建议 tag、构建上下文、目标 ref 预览。
- Modify: `code/backend/internal/module/challenge/domain/package_parser.go`
  - Jeopardy 平台构建模式允许省略 `runtime.image.ref`。
  - 从 `runtime.image.tag/ref` 提取 tag suggestion。
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
  - AWD 平台构建模式允许省略最终 `runtime.image.ref`，但必须有 `docker/Dockerfile`。
  - 引用已有镜像模式仍要求完整 `runtime.image.ref`。
- Create: `code/backend/internal/module/challenge/domain/image_delivery.go`
  - 纯函数：`ResolvePackageImageSource`、`BuildPlatformImageRef`、`SplitImageRef`、`NormalizeImageTag`。
- Create: `code/backend/internal/module/challenge/application/commands/image_build_service.go`
  - 创建 build job、生成 image 记录、调度 worker、更新状态。
- Create: `code/backend/internal/module/challenge/application/commands/registry_client.go`
  - Registry v2 manifest check 和错误归一化。
- Create: `code/backend/internal/module/challenge/application/commands/docker_image_builder.go`
  - 最小 Docker CLI adapter：`docker build/tag/push/pull/inspect`，第一阶段可同步调用但必须在后台 worker 内执行。
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
  - preview 展示平台最终 image ref。
  - commit 不再直接把 image 标为 available。
  - 平台构建路径创建 build job；外部引用路径做 verify job 或阻断 commit。
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
  - 同步接入 image delivery。
  - AWD 初始 `readiness_status=pending`，镜像未 available 时不允许通过 readiness。
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_service.go`
  - 发布自检继续通过 `resolveAvailableImageRef` 阻断非 available 镜像。
- Modify: `code/backend/internal/module/challenge/ports/ports.go`
  - 增加 `ImageBuildJobRepository`、`RegistryVerifier`、`DockerImageBuilder` 或等价端口。
- Modify: `code/backend/internal/module/challenge/infrastructure/image_repository.go`
  - 增加 build job CRUD、按状态拉取 pending job、CAS 启动任务。
- Modify: `code/backend/internal/app/composition/challenge_module.go`
  - 装配 image build service，并随 app context 启动 worker。

### DTO 与 HTTP

- Modify: `code/backend/internal/dto/challenge_import*.go`
  - preview 增加 `image_delivery` 或扩展 `runtime`：`source_type`、`suggested_tag`、`target_image_ref`、`build_status`、`digest`、`last_error`。
- Modify: `code/backend/internal/dto/image*.go`
  - image response 增加 digest/source/build/error/verified 字段。
- Modify: `code/backend/internal/module/challenge/api/http/handler.go`
  - commit response 需要返回题目 + 镜像交付状态；若保持旧响应，也至少前端可通过 import detail 查询到状态。
- Modify: `code/backend/internal/module/challenge/api/http/image_handler.go`
  - 增加 rebuild/reverify endpoint 时再扩展；第一阶段可只暴露 image list 字段。

### 脚本与配置

- Modify: `code/backend/internal/config/config.go`
  - `container.registry` 下增加 build 相关配置：`build_enabled`、`build_timeout`、`build_concurrency`、`default_tag_policy`。
- Modify: `code/backend/configs/config.yaml`
- Modify: `code/backend/configs/config.prod.yaml`
- Modify: `scripts/registry/build-and-push-challenge-image.sh`
  - 默认 repository 改为 `<mode>/<slug>`。
  - 平台构建语义下不再要求把最终 image ref 写回 manifest。

### 前端

- Modify: `code/frontend/src/api/admin/authoring.ts`
  - 补齐 Jeopardy import preview 的 image delivery 类型。
- Modify: `code/frontend/src/api/admin/awd-authoring.ts`
  - 补齐 AWD import preview 的 image delivery 类型。
- Modify: `code/frontend/src/api/contracts.ts`
  - 同步类型字段。
- Modify: `code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`
  - 显示镜像来源、最终 image ref、构建状态、错误摘要。
  - 平台构建模式不展示“上传者填写完整镜像名”的要求。
- Modify: `code/frontend/src/features/challenge-package-import/model/useChallengePackageImport.ts`
  - 上传/commit 后轮询 import 状态或 image 状态。
- Modify: `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
  - AWD 上传队列显示平台生成 image ref 与构建状态，不把 `runtime_config.image_ref` 当成作者必填项。
- Modify: `code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportFlow.ts`
  - commit 后刷新构建状态。
- Modify: `code/frontend/src/components/platform/images/ImageDirectoryPanel.vue`
  - 显示 digest/source/build job/verified/error。

### 文档

- Modify after implementation: `docs/contracts/challenge-pack-v1.md`
- Modify after implementation: `challenges/awd/challenge-package-contract.md`
- Modify after implementation: `docs/architecture/features/challenge-package-registry-delivery-design.md` if implementation intentionally differs.

## 目标边界

- 第一轮实现只支持单构建上下文：`docker/Dockerfile` + `docker/` context -> 单镜像。
- 多节点拓扑先把所有节点归一到同一个平台 image ref；节点级多镜像构建作为后续任务，不能在第一轮半实现。
- 平台构建任务必须异步执行，不能在上传 HTTP 请求中同步 build。
- 外部引用镜像路径必须显式选择，并通过 manifest/pull/inspect 后才能变为 `available`。
- `images.status=available` 必须只代表镜像已验证。

## Task 1: 数据模型和迁移

**Files:**

- Modify: `code/backend/internal/model/image.go`
- Create: `code/backend/internal/model/image_build_job.go`
- Create: `code/backend/migrations/000005_create_image_build_jobs.up.sql`
- Create: `code/backend/migrations/000005_create_image_build_jobs.down.sql`
- Modify tests using AutoMigrate where needed:
  - `code/backend/internal/module/challenge/testsupport/test_helper.go`
  - selected import tests that create image/build job records

- [x] **Step 1: 写迁移文件**

`000005_create_image_build_jobs.up.sql` 应完成：

```sql
ALTER TABLE public.images
  ADD COLUMN IF NOT EXISTS digest text,
  ADD COLUMN IF NOT EXISTS source_type text NOT NULL DEFAULT 'manual',
  ADD COLUMN IF NOT EXISTS build_job_id bigint,
  ADD COLUMN IF NOT EXISTS last_error text,
  ADD COLUMN IF NOT EXISTS verified_at timestamp without time zone;

CREATE TABLE IF NOT EXISTS public.image_build_jobs (
  id bigserial PRIMARY KEY,
  source_type text NOT NULL,
  challenge_mode text NOT NULL,
  package_slug text NOT NULL,
  source_dir text NOT NULL,
  dockerfile_path text NOT NULL,
  context_path text NOT NULL,
  target_ref text NOT NULL,
  target_digest text,
  status text NOT NULL,
  log_path text,
  error_summary text,
  created_by bigint,
  started_at timestamp without time zone,
  finished_at timestamp without time zone,
  created_at timestamp without time zone,
  updated_at timestamp without time zone
);

CREATE INDEX IF NOT EXISTS idx_image_build_jobs_status_created_at ON public.image_build_jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_image_build_jobs_package_slug ON public.image_build_jobs(package_slug);
CREATE INDEX IF NOT EXISTS idx_images_source_type ON public.images(source_type);
CREATE INDEX IF NOT EXISTS idx_images_build_job_id ON public.images(build_job_id);
```

- [x] **Step 2: 写 down migration**

Down migration 删除新增 index、`image_build_jobs` 表和 `images` 新增列。

- [x] **Step 3: 更新 Go model**

`Image` 增加：

```go
Digest    string     `gorm:"column:digest"`
SourceType string    `gorm:"column:source_type"`
BuildJobID *int64    `gorm:"column:build_job_id"`
LastError string     `gorm:"column:last_error"`
VerifiedAt *time.Time `gorm:"column:verified_at"`
```

新增 `ImageBuildJob` model，状态至少包含：

```go
const (
    ImageBuildJobStatusPending = "pending"
    ImageBuildJobStatusBuilding = "building"
    ImageBuildJobStatusPushed = "pushed"
    ImageBuildJobStatusVerifying = "verifying"
    ImageBuildJobStatusAvailable = "available"
    ImageBuildJobStatusFailed = "failed"
)
```

- [x] **Step 4: 跑模型相关测试**

Run:

```bash
cd code/backend
go test ./internal/model ./internal/app -run 'Migration|AutoMigrate|FullRouter' -count=1
```

Expected: PASS，或没有测试文件；如 FullRouter 太重，至少跑 `./internal/app -run TestMigrationFiles`。

- [x] **Step 5: Commit**

```bash
git add code/backend/internal/model code/backend/migrations code/backend/internal/module/challenge/testsupport
git commit -m "feat(image): 增加镜像构建任务模型"
```

## Task 2: 镜像命名与题包解析归一

**Files:**

- Create: `code/backend/internal/module/challenge/domain/image_delivery.go`
- Modify: `code/backend/internal/module/challenge/domain/package_manifest.go`
- Modify: `code/backend/internal/module/challenge/domain/package_parser.go`
- Modify: `code/backend/internal/module/challenge/domain/awd_package_parser.go`
- Test:
  - `code/backend/internal/module/challenge/domain/package_parser_test.go`
  - `code/backend/internal/module/challenge/domain/awd_package_parser_test.go`

- [x] **Step 1: 写命名规则测试**

新增测试覆盖：

- Jeopardy slug `web-demo` + tag `v1` -> `registry.example.edu/jeopardy/web-demo:v1`
- AWD slug `awd-demo` + tag `c1` -> `registry.example.edu/awd/awd-demo:c1`
- 平台构建模式忽略 `registry.example.edu/team/web-demo:v1` 的 repository，只复用 `v1`
- 空 tag 生成可预测 fallback，测试可传入固定 import id 或 clock

- [x] **Step 2: 实现纯函数**

在 `image_delivery.go` 中实现：

```go
func BuildPlatformImageRef(registry, mode, slug, tag string) (string, error)
func ExtractImageTagSuggestion(ref, tag string) string
func SplitImageRef(ref string) (name, tag string, err error)
```

规则：

- mode 只允许 `jeopardy` / `awd`
- slug 复用题包 slug 校验
- tag 为空时由 caller 传入导入批次 tag
- registry 为空时只返回 `<mode>/<slug>:<tag>`，便于本地测试

- [x] **Step 3: 扩展 parsed 结构**

给 Jeopardy/AWD parsed 结果加入：

```go
ImageSourceType string
SuggestedImageTag string
BuildContextPath string
DockerfilePath string
```

第一轮 source type 可由规则决定：

- 有 `docker/Dockerfile` 且未显式选择 external -> `platform_build`
- 显式 external 模式再使用 `external_ref`

- [x] **Step 4: 调整 AWD parser**

AWD 平台构建路径不再因为 `runtime.image.ref` 为空失败；但如果既没有 `runtime.image.ref` 又没有 `docker/Dockerfile`，仍失败。

- [x] **Step 5: 跑 domain 测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/domain -run 'Package|ImageDelivery|AWD' -count=1
```

Expected: PASS。

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge/domain
git commit -m "feat(challenge): 归一题包镜像命名"
```

## Task 3: 构建任务仓储和服务骨架

**Files:**

- Modify: `code/backend/internal/module/challenge/ports/ports.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/image_repository.go`
- Create: `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- Test:
  - `code/backend/internal/module/challenge/infrastructure/image_repository_test.go`
  - `code/backend/internal/module/challenge/application/commands/image_build_service_test.go`

- [x] **Step 1: 写 repository 测试**

覆盖：

- create job
- find job by id
- list pending jobs with limit
- try start job uses `WHERE status='pending'`
- update job status/error/digest

- [x] **Step 2: 扩展 repository**

在 `ImageRepository` 上实现 build job 方法，避免新建另一个 GORM repo 后造成 composition 分散。

- [x] **Step 3: 写 service 测试**

测试 `CreatePlatformBuildJob`：

- 生成 target ref。
- 创建 `images` 行时 status=`pending`，source_type=`platform_build`。
- 创建 `image_build_jobs` 行 status=`pending`。
- 返回 image id、job id、target ref。

- [x] **Step 4: 实现 service 骨架**

`ImageBuildService` 先只负责创建 job，不执行 Docker。

- [x] **Step 5: 跑测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/infrastructure ./internal/module/challenge/application/commands -run 'ImageBuild|BuildJob|ImageRepository' -count=1
```

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge/ports code/backend/internal/module/challenge/infrastructure code/backend/internal/module/challenge/application/commands
git commit -m "feat(image): 增加构建任务服务"
```

## Task 4: Docker builder 与 registry verifier

**Files:**

- Create: `code/backend/internal/module/challenge/application/commands/docker_image_builder.go`
- Create: `code/backend/internal/module/challenge/application/commands/registry_client.go`
- Modify: `code/backend/internal/config/config.go`
- Modify: `code/backend/configs/config.yaml`
- Modify: `code/backend/configs/config.prod.yaml`
- Test:
  - `code/backend/internal/module/challenge/application/commands/docker_image_builder_test.go`
  - `code/backend/internal/module/challenge/application/commands/registry_client_test.go`
  - `code/backend/internal/config/config_test.go` if config tests exist

- [x] **Step 1: 定义端口接口**

接口应表达最小能力：

```go
type DockerImageBuilder interface {
    Build(ctx context.Context, contextPath, dockerfilePath, localRef string) error
    Tag(ctx context.Context, sourceRef, targetRef string) error
    Push(ctx context.Context, targetRef string) error
    Pull(ctx context.Context, targetRef string) error
    Inspect(ctx context.Context, targetRef string) (ImageInspectResult, error)
}

type RegistryVerifier interface {
    CheckManifest(ctx context.Context, imageRef string) (digest string, err error)
}
```

- [x] **Step 2: 写 fake 测试**

先用 fake builder/verifier 测 service 状态流转，不在单元测试里真实跑 Docker。

- [x] **Step 3: 实现 Docker CLI adapter**

使用 `exec.CommandContext`，命令超时由 caller context 控制。日志输出写入 job log file，不进入 HTTP response。

- [x] **Step 4: 实现 Registry v2 manifest check**

支持：

- HTTP/HTTPS scheme
- basic auth / identity token
- Docker content digest 读取
- 401/404/5xx 错误摘要归一化

- [x] **Step 5: 跑测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/application/commands -run 'DockerImageBuilder|RegistryClient|ImageBuild' -count=1
go test ./internal/config -count=1
```

- [x] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge/application/commands code/backend/internal/config code/backend/configs
git commit -m "feat(image): 增加 registry 校验与 docker builder"
```

## Task 5: 构建 worker 与状态流转

**Files:**

- Modify: `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- Modify: `code/backend/internal/app/composition/challenge_module.go`
- Test:
  - `code/backend/internal/module/challenge/application/commands/image_build_service_test.go`
  - `code/backend/internal/app/composition/challenge_module_test.go` or existing composition tests

- [ ] **Step 1: 写状态流转测试**

覆盖成功路径：

```text
pending -> building -> pushed -> verifying -> available
```

断言：

- job status 最终 available
- image status 最终 available
- image digest/verified_at 被写入

- [ ] **Step 2: 写失败路径测试**

覆盖：

- build 失败 -> image/job failed + last_error
- push 失败 -> failed
- manifest 失败 -> failed
- inspect 失败 -> failed

- [ ] **Step 3: 实现 worker loop**

实现：

- `StartBackgroundTasks(ctx)`
- poll interval
- batch size
- CAS start job
- max concurrency
- graceful stop

- [ ] **Step 4: 装配到 app**

`buildChallengeImageHandler` 或单独 builder 中创建并启动 `ImageBuildService`。不要把 build worker 挂在 HTTP handler 生命周期之外。

- [ ] **Step 5: 跑测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/application/commands -run 'ImageBuild' -count=1
go test ./internal/app -run 'TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouter' -count=1
```

- [ ] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge/application/commands code/backend/internal/app/composition
git commit -m "feat(image): 启动镜像构建 worker"
```

## Task 6: Jeopardy 导入接入 image delivery

**Files:**

- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/dto/challenge_import*.go`
- Modify: `code/backend/internal/module/challenge/api/http/handler.go`
- Test:
  - `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
  - `code/backend/internal/module/challenge/api/http/challenge_import_handler_test.go`
  - `code/backend/internal/app/challenge_import_integration_test.go`

- [ ] **Step 1: 写 preview 测试**

上传不带 `runtime.image.ref`、但带 `docker/Dockerfile` 的题包，preview 应返回：

- `source_type=platform_build`
- `target_image_ref=127.0.0.1:5000/jeopardy/<slug>:<tag>` 或测试注入 registry
- 不报缺少镜像名

- [ ] **Step 2: 写 commit 测试**

commit 后：

- challenge 保持 draft
- image status=pending/building，不是 available
- challenge.image_id 指向该 image
- build job 创建成功

- [ ] **Step 3: 替换 `resolveImportedImageID`**

保留 `SplitImageRef` 用于 external ref，但平台构建路径调用 `ImageBuildService.CreatePlatformBuildJob`。

- [ ] **Step 4: 拓扑快照使用目标 image ref**

第一轮把拓扑所有节点映射到同一 image id；如果节点有独立 `image.ref`，preview 加 warning，commit 不再按节点创建外部 available image。

- [ ] **Step 5: 跑测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/application/commands -run 'ChallengeImport|ImageBuild' -count=1
go test ./internal/module/challenge/api/http -run 'ChallengeImport' -count=1
go test ./internal/app -run 'ChallengeImport' -count=1
```

- [ ] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge/application/commands code/backend/internal/dto code/backend/internal/module/challenge/api/http code/backend/internal/app
git commit -m "feat(challenge): 导入题包创建平台构建任务"
```

## Task 7: AWD 导入接入 image delivery

**Files:**

- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Modify: `code/backend/internal/dto/awd*.go`
- Modify: `code/backend/internal/module/challenge/api/http/awd_challenge_handler.go`
- Test:
  - `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service_test.go`
  - `code/backend/internal/module/challenge/api/http/awd_challenge_handler_test.go`

- [ ] **Step 1: 写 AWD 无 image ref 测试**

AWD 包包含 `docker/Dockerfile`、省略 `runtime.image.ref`，preview 通过并返回平台目标 ref。

- [ ] **Step 2: 写 commit 测试**

commit 后：

- `runtime_config.image_ref` 使用平台生成 ref
- `runtime_config.image_id` 指向 pending/building image
- readiness 仍为 pending
- 不把 image 标为 available

- [ ] **Step 3: 实现 AWD 接入**

`AWDChallengeImportService` 注入或持有 image delivery service。为避免构造函数破坏面过大，可以先用可选参数扩展 `NewAWDChallengeImportService(db, repo, imageDelivery...)`。

- [ ] **Step 4: readiness gate 检查**

确认 AWD 加入赛事或 readiness 校验路径读取 image status；若只看 runtime_config，需要补检查。

- [ ] **Step 5: 跑测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/application/commands -run 'AWDChallengeImport|ImageBuild' -count=1
go test ./internal/module/challenge/api/http -run 'AWDChallenge' -count=1
go test ./internal/module/contest/application/commands -run 'AWDService|Readiness|Image' -count=1
```

- [ ] **Step 6: Commit**

```bash
git add code/backend/internal/module/challenge/application/commands code/backend/internal/module/challenge/api/http code/backend/internal/dto code/backend/internal/module/contest
git commit -m "feat(awd): 导入题包创建平台构建任务"
```

## Task 8: 外部镜像引用路径

**Files:**

- Modify: `code/backend/internal/module/challenge/application/commands/image_build_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- Modify: `code/backend/internal/module/challenge/application/commands/awd_challenge_import_service.go`
- Test:
  - image build service tests
  - challenge import tests
  - awd import tests

- [ ] **Step 1: 明确 external source 输入**

第一轮后端可以通过 manifest 标记或 commit request 字段选择 external；如果前端还没提供，先保留 manifest compatible path。

- [ ] **Step 2: 写 external verify 测试**

外部 ref 成功：

- image source_type=`external_ref`
- 状态经过 verifying -> available
- digest/verified_at 写入

外部 ref 失败：

- import commit 返回错误，或 image failed 且题目不可发布。推荐第一轮 commit 返回错误，避免创建不可运行题目。

- [ ] **Step 3: 实现 external verify**

调用 manifest check + pull + inspect，不执行 build/push。

- [ ] **Step 4: 跑测试**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/application/commands -run 'External|ChallengeImport|AWDChallengeImport|ImageBuild' -count=1
```

- [ ] **Step 5: Commit**

```bash
git add code/backend/internal/module/challenge/application/commands
git commit -m "feat(image): 支持外部镜像校验路径"
```

## Task 9: 前端上传与状态展示

**Files:**

- Modify: `code/frontend/src/api/admin/authoring.ts`
- Modify: `code/frontend/src/api/admin/awd-authoring.ts`
- Modify: `code/frontend/src/api/contracts.ts`
- Modify: `code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`
- Modify: `code/frontend/src/features/challenge-package-import/model/useChallengePackageImport.ts`
- Modify: `code/frontend/src/components/platform/awd-service/AWDChallengeLibraryPage.vue`
- Modify: `code/frontend/src/features/platform-awd-challenges/model/useAwdChallengeImportFlow.ts`
- Modify: `code/frontend/src/components/platform/images/ImageDirectoryPanel.vue`
- Tests:
  - `code/frontend/src/components/platform/__tests__/ChallengePackageImportReview.test.ts`
  - `code/frontend/src/views/platform/__tests__/ChallengeImportManage.test.ts`
  - `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
  - `code/frontend/src/views/platform/__tests__/ImageManage.test.ts`

- [ ] **Step 1: 更新 API 类型测试**

确保 import preview 包含：

```ts
image_delivery?: {
  source_type: 'platform_build' | 'external_ref'
  suggested_tag?: string
  target_image_ref?: string
  status?: string
  digest?: string
  last_error?: string
}
```

- [ ] **Step 2: Jeopardy 上传 UI**

预览页显示平台生成 image ref 和构建状态；不要再把 `runtime.image_ref` 文案写成上传者必须提供。

- [ ] **Step 3: AWD 上传 UI**

AWD import queue 卡片显示 source、target ref、status、error。`runtime_config` JSON 保留给高级信息，但不作为主要镜像入口。

- [ ] **Step 4: 镜像管理 UI**

镜像列表显示 source/status/digest/verified_at/last_error，失败状态要能被管理员快速看到。

- [ ] **Step 5: 跑前端测试**

Run:

```bash
cd code/frontend
npm run test:run -- src/components/platform/__tests__/ChallengePackageImportReview.test.ts src/views/platform/__tests__/ChallengeImportManage.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/views/platform/__tests__/ImageManage.test.ts
npm run typecheck
```

- [ ] **Step 6: Commit**

```bash
git add code/frontend/src
git commit -m "feat(frontend): 展示题包镜像交付状态"
```

## Task 10: 脚本、契约文档和端到端验证

**Files:**

- Modify: `scripts/registry/build-and-push-challenge-image.sh`
- Modify: `docs/contracts/challenge-pack-v1.md`
- Modify: `challenges/awd/challenge-package-contract.md`
- Modify: `docs/architecture/features/challenge-package-registry-delivery-design.md` if needed
- Test:
  - shell script dry run or bats if available
  - backend focused tests
  - frontend focused tests

- [ ] **Step 1: 更新脚本默认 repository**

脚本按 `meta.mode` 生成：

```text
jeopardy/<slug>:<tag>
awd/<slug>:<tag>
```

不再默认 `ctf/<slug>`，也不再提示必须写回 manifest 才能上传。

- [ ] **Step 2: 更新契约文档**

只有代码实现后才把 `challenge-pack-v1` 从“当前不支持平台构建”改成：

- 平台构建模式可省略最终 image ref。
- 上传者只提供源码、Dockerfile、可选 tag 建议。
- 外部镜像引用模式才要求完整 `runtime.image.ref`。

- [ ] **Step 3: 本地闭环验证**

准备一个最小 Jeopardy 题包和一个最小 AWD 题包，使用本地 registry：

```bash
scripts/registry/deploy-private-registry.sh --port 5000 --server 127.0.0.1:5000
```

启动后端和 worker，上传题包，确认：

- image job 从 pending 到 available。
- registry 中存在 `127.0.0.1:5000/jeopardy/<slug>:<tag>`。
- registry 中存在 `127.0.0.1:5000/awd/<slug>:<tag>`。
- 发布自检或 AWD readiness 在镜像失败时阻断。

- [ ] **Step 4: 最小全量验证**

Run:

```bash
cd code/backend
go test ./internal/module/challenge/... ./internal/module/runtime/... ./internal/app -run 'Image|ChallengeImport|AWDChallengeImport|PublishCheck|PrivateRegistry|BuildRoot|NewRouter' -count=1

cd ../frontend
npm run typecheck
npm run test:run -- src/components/platform/__tests__/ChallengePackageImportReview.test.ts src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts src/views/platform/__tests__/ImageManage.test.ts
```

- [ ] **Step 5: Review 记录**

本任务属于跨模块结构性改动，完成后必须做 review，记录到：

```text
docs/reviews/backend/2026-05-05-challenge-package-registry-delivery-review.md
```

Review 必查：

- 是否还有导入路径直接创建 `available` 镜像。
- 平台构建模式是否仍要求上传者填写完整镜像名。
- build/push/verify 失败是否阻断发布或 readiness。
- registry 凭据是否没有进入题包、日志和 response。
- worker 是否可停止，测试是否不会留下后台进程。

- [ ] **Step 6: Commit**

```bash
git add scripts/registry docs/contracts challenges/awd docs/architecture/features docs/reviews
git commit -m "docs(challenge): 同步题包 registry 交付契约"
```

## 架构适配检查

- 目标边界明确：parser 只做解析，image delivery 负责 build/push/verify，runtime 只消费 available image。
- 共享层落点明确：命名规则在 `domain/image_delivery.go`，构建任务在 `application/commands/image_build_service.go`，Docker/registry 操作通过端口隔离。
- 结构收敛没有只做表面行为：计划显式移除 `resolveImportedImageID` 直接 available 的导入路径，并把 Jeopardy/AWD 都接到同一 image delivery service。
- 第一轮有意延后节点级多镜像构建，并已写入目标边界；完成标准是不再散落作者 repository，所有节点先归一到平台生成 image ref。
- 发布/readiness gate 仍依赖 `images.status=available`，不会让未验证镜像进入运行时。

## 风险与回退

- 迁移风险：`images` 表新增字段默认值必须兼容旧数据，旧手工镜像 source_type 用 `manual`。
- 运行风险：Docker build 需要宿主机 Docker 权限；默认配置应允许关闭平台构建，关闭时只允许 external verify 或返回清晰错误。
- 网络风险：registry manifest/pull 可能因 insecure registry、TLS 或 auth 失败；错误摘要必须保留给管理员。
- 安全风险：构建上下文必须限制在题包根目录，日志要脱敏 registry password。
- 回退策略：如果 worker 不稳定，可以保留外部镜像引用路径；平台构建通过配置开关禁用，不影响已有手工 image 管理。

## Plan Review Note

`writing-plans` skill 建议派发 plan-document-reviewer subagent。但当前会话的系统级规则只允许在用户明确要求 sub-agent/并行代理时使用 `spawn_agent`，本次用户只要求写执行方案，因此未派发 subagent。执行前仍应由实现者按本计划的 Review Gate 做独立 review。
