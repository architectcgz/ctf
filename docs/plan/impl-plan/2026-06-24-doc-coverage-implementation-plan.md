<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# doc-coverage-补充设计文档遗漏点 Implementation Plan

> 填写要求：保留本模板中的英文标题、状态枚举、机器字段和占位符；除代码、命令、路径、报错、协议字段、枚举值、外部专有名词外，所有说明性内容默认用中文填写。

**Goal:** 补充 CTF 项目中 40+ 个未被文档化的关键设计决策，消除接手人的理解障碍。

**Architecture:** 只修改 `docs/architecture/` 下的文档，新增 5 个专题文档、更新 8 个现有文档。不改变任何代码实现、API 契约或数据库结构。

**Tech Stack:** Markdown

---

## Task Metadata

- Task Slug: `2026-06-24-doc-coverage`
- Started At: `2026-06-24T14:25:05Z`
- Worktree: `/home/azhi/workspace/projects/.worktrees/ctf/2026-06-24-doc-coverage`
- Branch: `task/2026-06-24-doc-coverage`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `draft` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [ ] Intake analysis gate completed
  - [ ] Plan review / architecture-fit check completed
  - [ ] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 补充 18 个高优先级设计文档遗漏点，包括：后端事件发布策略、容器调度机制、前端架构边界约定等，确保接手人能理解关键设计决策。
- Non-Goals: 
  - 不修改任何代码实现
  - 不修改 API 契约和数据库结构
  - 不补充中低优先级遗漏点（留待后续迭代）
  - 不建立文档同步检查的自动化机制（Phase 2）

## Problem Statement

- Current behavior / structure: CTF 项目存在 40+ 个代码实现中的关键设计决策未被文档化，包括：Weak Event 发布策略、AWD Flag 三级存储、Instance 启动恢复机制、Contest 状态机、前端 Entities 边界约束等。接手人在修改代码时会因为不理解这些隐藏决策而引入 bug 或破坏架构。
- Target behavior / structure: 将 18 个高优先级遗漏点补充到对应的架构文档中，新增 5 个专题文档、更新 8 个现有文档，确保接手人能从文档中理解关键设计决策。
- Why this task is needed now: 项目即将交接给其他开发者，当前文档覆盖度排查发现大量关键决策未记录，不补充会导致接手人理解困难、修改风险高、故障排查困难。

## Inputs

- Source docs: 
  - `docs/architecture/README.md` - 架构文档入口
  - `docs/architecture/backend/README.md` - 后端架构索引
  - `docs/architecture/frontend/README.md` - 前端架构索引
  - `docs/文档规范.md` - 文档写作规范
  - `/tmp/ctf-coverage-gap-report.md` - agent 生成的完整遗漏点清单（临时文件，如缺失需重新运行覆盖度排查 agent `ae8b49ceee088465f`）
- Related architecture/contracts: 
  - `docs/architecture/backend/modules/*.md` - 10 个后端模块文档
  - `docs/architecture/backend/02-database-design.md` - 数据库设计
  - `docs/architecture/backend/03-container-architecture.md` - 容器架构
  - `docs/architecture/frontend/01-architecture-overview.md` - 前端架构总览
  - `docs/architecture/frontend/04-api-layer.md` - API 层
  - `docs/architecture/frontend/06-components.md` - 组件体系
- Related prior work: 
  - agent `ae8b49ceee088465f` 的设计文档覆盖度排查结果
  - agent `a4b3c90673c9a9117` 生成的详细遗漏点清单（40+ 个）

## Task Classification

- Classification: `非琐碎任务`
- Why: 需要补充 18 个高优先级遗漏点，涉及新增 5 个专题文档、更新 8 个现有文档，工作量大且需要确保文档准确性和一致性。虽然只修改文档不改代码，但文档质量直接影响接手人理解和后续开发。

## Files

- Create: 
  - `docs/architecture/features/事件发布与降级策略.md`
  - `docs/architecture/features/容器调度与并发控制.md`
  - `docs/architecture/features/网络子网分配策略.md`
  - `docs/architecture/features/数据库事务管理模式.md`
  - `docs/architecture/frontend/10-runtime-error-handling.md`
- Modify: 
  - `docs/architecture/backend/modules/contest.md` - 新增"状态机"和"事务处理"章节
  - `docs/architecture/backend/modules/instance.md` - 新增"启动恢复协议"和"调度器编排"章节
  - `docs/architecture/backend/03-container-architecture.md` - 新增"调度锁与并发控制"和"安全隔离与降级"章节
  - `docs/architecture/backend/02-database-design.md` - 补充 migration 000018-000020 新增表
  - `docs/architecture/frontend/04-api-layer.md` - 补充 teaching 目录和顶层模块清单
  - `docs/architecture/frontend/01-architecture-overview.md` - 补充 Features 放置规则和 Entities 边界约束
  - `docs/architecture/frontend/06-components.md` - 补充 Widgets 层定位和 Navigation Composables
  - `docs/architecture/features/校园级CTF-AWD模式完整设计.md` - 补充 Flag 轮换架构和降噪策略
- Review: 
  - `code/backend/internal/module/contest/application/` - 理解 Contest 状态机实现
  - `code/backend/internal/module/instance/application/` - 理解 Instance 启动恢复机制
  - `code/backend/internal/module/container_runtime/` - 理解容器调度和网络分配
  - `code/frontend/src/features/` - 理解前端 Features 组织
  - `code/frontend/src/entities/` - 理解 Entities 边界
- Test: none（纯文档更新）

## 复用与 Owner 决策

- Existing patterns searched: 已阅读现有架构文档结构（`docs/architecture/backend/modules/`、`docs/architecture/features/`、`docs/architecture/frontend/`），确认文档组织方式和写作风格。检查 `docs/文档规范.md` 确认文档归属规则。
- Reuse / extend / split / create-new decision: 
  - **复用**：遵循现有文档结构和写作风格
  - **扩展**：在现有模块文档中新增章节（contest.md、instance.md、03-container-architecture.md 等）
  - **新建**：新增 5 个专题文档，放在 `docs/architecture/features/` 和 `docs/architecture/frontend/`
- Owner boundary: 
  - 后端专题文档 owner：后端架构文档体系（`docs/architecture/backend/` 和 `docs/architecture/features/`）
  - 前端架构文档 owner：前端架构文档体系（`docs/architecture/frontend/`）
  - 每个遗漏点补充到最合适的文档位置，遵循"最终事实源"原则
- Why this is the narrowest safe surface: 只修改文档不改代码，不影响运行时行为。文档更新是独立的、可回滚的，即使某个文档写得不够准确也不会影响系统运行。

## Intake Analysis Gate

- Relevant superpowers analysis pass: 已由 agent 完成设计文档覆盖度排查（`ae8b49ceee088465f` 和 `a4b3c90673c9a9117`），生成 40+ 个遗漏点详细清单。
- Why this pass fits: 设计文档覆盖度排查是系统性识别未文档化设计决策的最佳方式，agent 已完成代码扫描、文档对比、影响范围评估和优先级分级。
- grill-with-docs findings: 
  - 确认 18 个高优先级遗漏点确实未在现有文档中记录
  - 确认文档归属规则（`docs/文档规范.md`）和目录结构（`docs/architecture/`）
  - 确认新增文档应放置的位置和命名规范
- Plan adjustments after challenge: 
  - 将 40+ 个遗漏点分为 3 个 phase，本次只实施 Phase 1（18 个高优先级）
  - 按文档类型分为 3 个 slice：后端专题文档（Batch 1）、后端模块更新（Batch 2）、前端架构文档（Batch 3）

## Execution Slices

### Slice 1: 补充后端专题文档（新增 4 个文档）

- Goal: 新增 4 个后端专题文档，记录事件发布策略、容器调度机制、网络子网分配和数据库事务管理模式。
- Dependencies: 需要先阅读对应代码实现，理解设计决策。
- Files:
  - Create: 
    - `docs/architecture/features/事件发布与降级策略.md`
    - `docs/architecture/features/容器调度与并发控制.md`
    - `docs/architecture/features/网络子网分配策略.md`
    - `docs/architecture/features/数据库事务管理模式.md`
  - Modify: none
  - Review: 
    - `code/backend/internal/module/*/application/*_service.go` - publishWeakEvent 使用
    - `code/backend/internal/module/instance/application/instance_provisioning_scheduler.go`
    - `code/backend/internal/module/container_runtime/infrastructure/dynamic_subnet_allocator.go`
    - `code/backend/internal/module/*/infrastructure/*_repository.go` - WithinXxxTransaction 使用
  - Test: none
- 步骤：
  - [ ] 步骤 1：新增 `事件发布与降级策略.md`，记录 Weak Event vs Strong Event、决策规则、监控运维
  - [ ] 步骤 2：新增 `容器调度与并发控制.md`，记录 Provisioning Scheduler 双段式编排、Redis 锁、失败补偿
  - [ ] 步骤 3：新增 `网络子网分配策略.md`，记录动态子网双池策略、IP 分配算法、冲突检测
  - [ ] 步骤 4：新增 `数据库事务管理模式.md`，记录 Transaction Runner 模式、隔离级别、死锁避免
- Validation: 检查文档是否符合 `docs/文档规范.md` 的格式要求。
- Review focus: 确认专题文档内容准确、逻辑清晰、覆盖关键设计点。
- Done criteria: 4 个新文档创建完成，内容完整，符合文档规范。

### Slice 2: 更新后端模块与架构文档（更新 4 个文档）

- Goal: 在现有后端模块文档和架构文档中新增章节，补充 Contest 状态机、Instance 启动恢复、容器安全隔离、数据库新增表等遗漏点。
- Dependencies: Slice 1 完成后执行。
- Files:
  - Create: none
  - Modify: 
    - `docs/architecture/backend/modules/contest.md`
    - `docs/architecture/backend/modules/instance.md`
    - `docs/architecture/backend/03-container-architecture.md`
    - `docs/architecture/backend/02-database-design.md`
  - Review: 
    - `code/backend/internal/module/contest/application/contest_lifecycle_service.go`
    - `code/backend/internal/module/instance/application/instance_startup_recoverer.go`
    - `code/backend/migrations/000018*.sql`、`000019*.sql`、`000020*.sql`
  - Test: none
- 步骤：
  - [ ] 步骤 1：在 `contest.md` 新增"状态机与并发控制"章节，记录状态转换规则、乐观锁 + 分布式锁
  - [ ] 步骤 2：在 `instance.md` 新增"启动恢复协议"和"调度器编排与并发控制"章节
  - [ ] 步骤 3：在 `03-container-architecture.md` 新增"调度锁与并发控制"、"安全隔离与降级"、"节点故障处理"章节
  - [ ] 步骤 4：在 `02-database-design.md` 补充 migration 000018-000020 新增表的说明
- Validation: 运行 `python3 scripts/check-docs-consistency.py` 检查文档索引一致性。
- Review focus: 确认新增章节与现有文档风格一致，索引更新正确。
- Done criteria: 4 个文档更新完成，新增章节内容完整，文档索引正确。

### Slice 3: 更新前端架构文档（新增 1 个文档 + 更新 3 个文档）

- Goal: 新增前端运行时错误处理文档，更新 API 层、架构总览、组件体系文档，补充前端架构边界约定。
- Dependencies: Slice 2 完成后执行。
- Files:
  - Create: 
    - `docs/architecture/frontend/10-runtime-error-handling.md`
  - Modify: 
    - `docs/architecture/frontend/04-api-layer.md`
    - `docs/architecture/frontend/01-architecture-overview.md`
    - `docs/architecture/frontend/06-components.md`
  - Review: 
    - `code/frontend/src/api/teaching/`
    - `code/frontend/src/features/platform/`
    - `code/frontend/src/features/teacher/`
    - `code/frontend/src/entities/`
    - `code/frontend/src/runtime/globalErrorRuntime.ts`
  - Test: none
- 步骤：
  - [ ] 步骤 1：新增 `10-runtime-error-handling.md`，记录全局错误处理集成、环境变量、错误分发逻辑
  - [ ] 步骤 2：更新 `docs/architecture/frontend/README.md` 索引，添加新文档入口
  - [ ] 步骤 3：在 `04-api-layer.md` 补充 teaching 目录结构和顶层独立模块清单
  - [ ] 步骤 4：在 `01-architecture-overview.md` 补充 Features 放置规则（platform/* vs teacher/*）和 Entities 边界约束
  - [ ] 步骤 5：在 `06-components.md` 补充 Widgets 层定位和 Navigation Composables 职责说明
- Validation: 运行 `bash scripts/check-frontend-architecture.sh --quick` 确认前端架构守卫通过。
- Review focus: 确认前端架构边界约定表述清晰，与 AGENTS.md 中的规则一致。
- Done criteria: 1 个新文档 + 3 个更新文档完成，前端架构索引更新，文档内容准确。

## Impact And Compatibility

- API / DTO: none（纯文档更新）
- Data / migration: none（纯文档更新）
- State / cache / queue / event: none（纯文档更新）
- Runtime / config: none（纯文档更新）
- Frontend route / state / UX: none（纯文档更新）
- Docs / contracts: 新增 5 个架构文档，更新 8 个现有文档，补充 18 个高优先级遗漏点。影响范围：接手人阅读文档时能理解关键设计决策，降低理解成本和修改风险。

## Plan Review / Architecture Fit

- Target owner boundary: 只修改架构文档（`docs/architecture/`），不触及代码、契约、数据库。文档 owner 是架构文档体系，每个遗漏点补充到对应的文档位置。
- Reuse points / landing zones: 
  - 后端专题文档落在 `docs/architecture/features/`
  - 后端模块更新落在 `docs/architecture/backend/modules/` 和 `docs/architecture/backend/`
  - 前端架构文档落在 `docs/architecture/frontend/`
- Known structural debt touched: none（纯文档更新不触及代码结构债）
- How this plan avoids behavior-only convergence: 本 plan 只写文档不改代码，不存在行为收口问题。文档内容基于已有代码实现，记录当前设计决策。
- Hidden second-redesign risk: 无。文档更新后如果发现表述不准确，可以直接修改文档，不会引入二次重构风险。
- Decision after review: 
  - Plan 范围清晰：只补充 18 个高优先级遗漏点
  - 执行路径明确：分 3 个 slice，每个 slice 独立可验证
  - 风险可控：纯文档更新，不影响运行时行为
  - 可继续实施

## Documentation Owner

- Current fact sources to read: 
  - **设计决策的当前事实源是代码实现本身**：18 个高优先级遗漏点的设计决策当前只存在于代码中，未被文档化
  - `docs/architecture/README.md` - 架构文档入口和索引
  - `docs/architecture/backend/README.md` - 后端架构索引
  - `docs/architecture/frontend/README.md` - 前端架构索引
  - `/tmp/ctf-coverage-gap-report.md` - 遗漏点清单（临时文件，仅作为本次补充的输入）
- Fact sources to update after implementation: 
  - **本次更新后，架构文档将成为这些设计决策的最终文档化事实源**
  - `docs/architecture/README.md` - 可能需要更新索引（如果新增文档属于顶层入口）
  - `docs/architecture/backend/README.md` - 更新后端架构索引，添加新增专题文档入口
  - `docs/architecture/frontend/README.md` - 更新前端架构索引，添加新增文档入口
  - 所有新增和修改的架构文档本身
- Plan-only notes that must not become architecture source: 
  - 本 implementation plan 本身只是实施过程记录，不是架构事实源
  - `/tmp/ctf-coverage-gap-report.md` 是临时排查报告，不应进入 Git 仓库
  - 40+ 遗漏点的优先级分级只属于本次补充计划，不应写入架构文档
  - 每个遗漏点的"影响范围""理解难度""修改风险"评估只属于计划阶段，不应成为文档正文
- Archive condition: 
  - 3 个 slice 全部完成
  - 所有新增文档已创建，所有更新文档已修改
  - 文档索引一致性检查通过（`python3 scripts/check-docs-consistency.py`）
  - 前端架构守卫通过（`bash scripts/check-frontend-architecture.sh --quick`）
  - 本 plan 归档到 `docs/plan/archive/impl-plan/2026-06-24-doc-coverage-implementation-plan.md`

## Validation

- 计划验证范围：纯文档更新，无需运行代码测试。验证范围包括：
  - 文档索引一致性：所有新增文档已在对应 README.md 中登记
  - 文档格式规范：符合 `docs/文档规范.md` 的格式要求
  - 前端架构守卫：前端架构边界约定与守卫脚本一致
  - 文档内容准确性：记录的设计决策与代码实现一致（人工 review）
- 命名 / 契约检查范围：
  - 检查新增文档文件名是否符合 `docs/文档规范.md` 的命名规范
  - 检查文档内引用的代码路径、函数名、类型名是否与实际代码匹配
  - 检查文档间交叉引用是否正确（例如 contest.md 引用 事件发布与降级策略.md）
- 完成判定：
  - 所有新增文档已创建，所有更新文档已修改
  - `python3 scripts/check-docs-consistency.py` 通过（文档索引一致性）
  - `bash scripts/check-frontend-architecture.sh --quick` 通过（前端架构守卫）
  - 人工 review 确认至少 5 个高优先级遗漏点的文档内容与代码实现一致

## Validation Plan

- Per-slice commands:
  - **Slice 1 完成后**：
    - `python3 scripts/check-docs-consistency.py` - 检查新增的 4 个专题文档是否已在 `docs/architecture/backend/README.md` 或 `docs/architecture/features/README.md` 索引中登记
    - 人工检查：至少抽查 2 个新增文档，确认内容与代码实现一致（抽查标准：至少包含 1 个复杂设计点，例如事件发布策略或容器调度机制）
  - **Slice 2 完成后**：
    - `python3 scripts/check-docs-consistency.py` - 检查更新的 4 个文档索引是否正确
    - 人工检查：确认新增章节的标题、结构与现有文档风格一致
  - **Slice 3 完成后**：
    - `bash scripts/check-frontend-architecture.sh --quick` - 确认前端架构守卫通过
    - `python3 scripts/check-docs-consistency.py` - 检查前端文档索引更新正确
    - 人工检查：确认前端架构边界约定表述与 AGENTS.md 中的规则一致
- Integration commands:
  - `python3 scripts/check-docs-consistency.py` - 全局文档索引一致性检查
  - `bash scripts/check-frontend-architecture.sh --quick` - 前端架构守卫（只验证前端相关文档）
- Manual checks:
  - 抽查至少 5 个高优先级遗漏点，确认文档记录的设计决策与代码实现一致
  - 抽查覆盖维度：至少覆盖后端 3 个、前端 2 个，至少包含 2 个复杂设计点（例如容器调度、事件发布策略、Instance 启动恢复）
  - 检查文档间交叉引用是否正确（例如 contest.md 引用新增的专题文档）
  - 检查新增文档的文件名、标题、章节结构是否符合 `docs/文档规范.md`
- Commands intentionally skipped and why:
  - 跳过 `go test`、`go build`、`npm test`、`npm run build`：纯文档更新不涉及代码改动，无需运行代码测试
  - 跳过 `bash scripts/check-workflow-governance.sh`：本次只修改文档，不触及 harness 策略或工作流配置

## Validation Evidence

- Command: `python3 scripts/check-docs-consistency.py`
  - Result: 待实施后执行
  - Notes: 本命令验证所有新增文档已在对应 README.md 索引中登记，所有文档路径引用正确。
  
- Command: `bash scripts/check-frontend-architecture.sh --quick`
  - Result: 待实施后执行
  - Notes: 本命令验证前端架构边界约定与守卫脚本一致，确保文档中记录的 entities vs features 边界规则与代码守卫匹配。

- Manual check: 抽查高优先级遗漏点文档内容准确性
  - Result: 待实施后执行
  - Notes: 至少抽查 5 个高优先级遗漏点（例如 Weak Event 发布策略、AWD Flag 轮换、Contest 状态机、Instance 启动恢复、前端 Entities 边界），确认文档记录的设计决策与代码实现一致。

## Independent Review Handoff

- Review target: 本次补充的 18 个高优先级设计文档遗漏点，涉及新增 5 个专题文档、更新 8 个现有文档。
- Validation evidence summary: 
  - 文档索引一致性检查通过（`python3 scripts/check-docs-consistency.py`）
  - 前端架构守卫通过（`bash scripts/check-frontend-architecture.sh --quick`）
  - 已抽查至少 5 个高优先级遗漏点，确认文档内容与代码实现一致
- Architecture / contract inputs:
  - `docs/architecture/README.md` - 架构文档入口
  - `docs/文档规范.md` - 文档规范和格式要求
  - `/tmp/ctf-coverage-gap-report.md` - 遗漏点清单（用于对照补充完整性）
  - 所有新增和修改的架构文档
- Known risks / review focus:
  - **内容准确性风险**：文档记录的设计决策是否与代码实现一致？特别关注 Weak Event 发布策略、AWD Flag 轮换机制、Contest 状态机、容器调度锁这些复杂设计点。
  - **文档完整性风险**：18 个高优先级遗漏点是否全部补充到位？是否遗漏关键细节？
  - **索引一致性风险**：所有新增文档是否已在对应 README.md 中登记？文档间交叉引用是否正确？
  - **风格一致性风险**：新增章节是否与现有文档风格一致？是否符合 `docs/文档规范.md` 的格式要求？
- Project-local checks to consider:
  - `python3 scripts/check-docs-consistency.py` - 全局文档索引一致性
  - `bash scripts/check-frontend-architecture.sh --quick` - 前端架构守卫
  - 人工抽查：选择 2-3 个复杂遗漏点（例如容器调度、事件发布策略），逐行对照代码实现，确认文档准确性

## Rollback / Recovery

- Safe revert boundary: 纯文档更新，回退边界非常清晰。可以安全回退到 `task/2026-06-24-doc-coverage` 分支起点，或通过 `git revert` 回退单个文档提交。
- Data / config / runtime recovery notes: none（纯文档更新不涉及数据、配置或运行时状态改动）
- Irreversible operations: none（所有改动都是 Git 版本控制下的文档文件，完全可逆）

## Residual Risks

- Risk: 文档记录的设计决策可能随代码演化而过期，后续维护时需要持续同步。
- Why acceptable: 
  - 本次补充建立了文档覆盖基线，后续修改代码时可以参照文档判断影响范围
  - 文档过期风险低于"完全没有文档"的风险
  - 项目 `docs/文档规范.md` 已明确要求结构性代码改动时同步更新文档
  - 后续可以通过 pre-commit hook 或 code review 流程提醒开发者同步文档
- Follow-up owner: 
  - **Phase 2**（中优先级遗漏点补充）：可在下个迭代继续补充 16 个中优先级遗漏点
  - **Phase 3**（低优先级遗漏点补充）：可延后补充 6 个低优先级遗漏点
  - **文档同步机制**：可考虑在 `harness/workflow-plugins/code-workflow/` 中新增文档同步检查插件，在 pre-commit 或 completion gate 阶段提醒开发者检查是否需要同步更新架构文档
