# Package Topology Sync Implementation Plan

> 状态：已合并到 `main`。  
> 合并提交：`3501135c 合并(main): 整理题包拓扑同步分支`

**Goal:** 让标准题包通过 `docker/topology.yml` 描述拓扑；导入时写入 `challenge_topologies`，保存完整题包源码树和修订记录，并在拓扑工作台展示题包来源、基线、漂移状态、文件列表和导出入口。

**Architecture:** 题包解析器读取 `extensions.topology.source` 指向的拓扑文件，解析 package-native topology 后在导入提交阶段映射到平台拓扑模型。后端通过题包修订表保存导入/导出基线，拓扑保存只更新当前状态并保留 provenance。前端在导入预览和拓扑工作台展示题包拓扑摘要、文件树、修订历史和导出操作。

**Tech Stack:** Go、Gin、GORM、SQLite test DB、Vue 3、Vitest

---

### Task 1: Lock backend shape with tests

**Files:**
- `code/backend/internal/module/challenge/application/commands/challenge_import_service_test.go`
- `code/backend/internal/app/full_router_state_matrix_integration_test.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`

- [x] 覆盖导入预览返回拓扑摘要和题包文件树。
- [x] 覆盖导入提交从 `docker/topology.yml` 创建 `challenge_topologies`。
- [x] 覆盖题包源码快照和 package revision 持久化。
- [x] 覆盖题包导出生成完整 zip，并重写 `challenge.yml` 与 `docker/topology.yml`。

### Task 2: Add package topology parsing and revision persistence

**Files:**
- `code/backend/internal/module/challenge/domain/package_manifest.go`
- `code/backend/internal/module/challenge/domain/package_parser.go`
- `code/backend/internal/module/challenge/domain/package_topology_parser.go`
- `code/backend/internal/model/topology.go`
- `code/backend/internal/model/challenge_package_revision.go`
- `code/backend/internal/module/challenge/ports/ports.go`
- `code/backend/internal/module/challenge/infrastructure/writeup_repository.go`
- `code/backend/internal/module/challenge/testsupport/test_helper.go`

- [x] 增加支持 image ref 的 package-native topology 解析模型。
- [x] 解析 `extensions.topology.source` 和 `docker/topology.yml`。
- [x] 增加拓扑 provenance 和题包修订持久化模型。
- [x] 拆分 `ChallengeTopologyRepository` 与 `ChallengePackageRevisionRepository`，避免普通拓扑仓储被题包修订能力污染。

### Task 3: Import topology, preserve source tree, and export full packages

**Files:**
- `code/backend/internal/module/challenge/application/commands/challenge_import_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_package_revision_service.go`
- `code/backend/internal/module/challenge/application/commands/challenge_service.go`
- `code/backend/internal/module/challenge/api/http/handler.go`
- `code/backend/internal/app/router_routes.go`
- `code/backend/internal/dto/challenge_import.go`
- `code/backend/internal/dto/topology.go`
- `code/backend/internal/module/challenge/domain/topology_codec.go`

- [x] 导入提交时保存题包源码树并创建 package revision。
- [x] 导入拓扑时将 image ref 映射到平台镜像并保存题包基线。
- [x] 增加题包导出和下载接口。
- [x] 拓扑保存保留题包来源，并按当前拓扑与基线对比更新 drift 状态。

### Task 4: Upgrade frontend contracts and import preview

**Files:**
- `code/frontend/src/api/contracts.ts`
- `code/frontend/src/api/admin.ts`
- `code/frontend/src/components/platform/challenge/ChallengePackageImportReview.vue`

- [x] 增加题包拓扑摘要、文件树、provenance、revision、export response 类型。
- [x] 导入预览展示拓扑摘要和题包文件树。
- [x] 导入审查区说明自动导入拓扑和保存源码树的状态。

### Task 5: Redesign topology workspace around package provenance

**Files:**
- `code/frontend/src/composables/useChallengeTopologyStudioPage.ts`
- `code/frontend/src/components/platform/topology/ChallengeTopologyStudioPage.vue`
- `code/frontend/src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`

- [x] 增加题包来源、基线、漂移状态、修订历史和导出状态。
- [x] 保留 `main` 已拆分的拓扑编辑器组件结构。
- [x] 在右侧上下文栏展示题包来源、题包文件和修订历史。
- [x] 保留模板库行为，挑战模式增加 package-aware workflow。

### Task 6: Verification and cleanup

- [x] `go test ./internal/module/challenge/... -count=1`
- [x] `npm run test:run -- src/views/platform/__tests__/ChallengeTopologyStudio.test.ts`
- [x] `npm run typecheck`

**Notes:**
- `go test ./internal/app -count=1` 在本地环境中仍会受到端口占用和部分历史测试 schema 缺表影响；本次合并已补齐题包修订相关测试 schema。
- 分支合并时以 `main` 的前端组件拆分和 Go context 约定为基线，再补回题包拓扑能力。
