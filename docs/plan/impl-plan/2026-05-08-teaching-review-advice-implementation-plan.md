# Teaching Review Advice Implementation Plan

## Plan Summary

### Objective

基于 `docs/architecture/features/教学复盘建议生成架构.md`，把当前教学复盘中的班级建议、个人观察和推荐题理由，从分散的硬编码阈值文案升级为共享的规则型建议层，并用现有样本数据验证建议结果是否更贴近训练事实。

### Non-goals

- 不重做教师端页面布局
- 不重写攻击证据链或攻击会话读模型
- 不新增数据库事实表
- 不引入 LLM 或外部推理服务
- 不改 `solved / score / rank` 的统计口径

### Source architecture or design docs

- `docs/design/教学复盘建议优化方案.md`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/architecture/features/攻击证据链与教学复盘架构.md`
- `docs/architecture/features/教学复盘优化设计.md`
- `docs/design/AWD能力画像回流方案.md`

### Dependency order

1. 先补教学信号与弱项置信度所需的读取模型
2. 再落共享建议规则层，先接管班级建议
3. 再把个人观察与推荐理由迁入共享规则层
4. 最后用样本数据补验证和回归测试

### Expected specialist skills

- `backend-engineer`
- `code-reviewer`
- `test-engineer`
- `doc-admin-agent`

### Architecture-fit evaluation

这次实现显式触达一块已知结构债：建议逻辑目前散落在：

- `teaching_readmodel.QueryService.GetClassReview`
- `assessment.ReportService.buildReviewArchiveObservations`
- `assessment.RecommendationService`

本计划要求在同一流水线内把这块 debt 收口为共享 owner `internal/teaching/advice`。这里的“收口”包括：

- 弱项维度识别
- 弱项证据充足性判断
- 统一严重度枚举
- 班级建议 / 个人观察 / 推荐理由的规则码

不允许出现“功能先补了，但三处规则还各写各的”的结果；如果实现发现这一收口超出当前切片，需要回到计划阶段重切，而不是把 debt 留成后续任务。

## Task 1

### Goal

为建议层补齐稳定的原始教学事实输入，解决当前 `weak_dimension` 退化成最低分或字典序命中的前置问题。

### Touched modules or boundaries

- `code/backend/internal/module/teaching_readmodel/infrastructure`
- `code/backend/internal/module/teaching_readmodel/ports`
- `code/backend/internal/module/assessment/infrastructure`
- `code/backend/internal/module/assessment/domain`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`

### Dependencies

- 无前置代码任务；直接依赖现有画像、证据与提交事实

### Validation

- `go test ./internal/module/teaching_readmodel/... -count=1`
- `go test ./internal/module/assessment/... -count=1`
- 必要时补针对弱项置信度的 focused tests

### Review focus

- 原始事实模型里是否没有提前写入 `confidence`、`is_weak` 这类派生判断
- 是否为后续统一弱项评估提供了足够原始输入
- 是否引入额外的 N+1 查询或无界扫描

### Risk notes

- 如果原始事实模型里已经写入 `confidence` / `is_weak`，结构债会在输入层继续分叉
- 如果只补字段不修 recommendation 的 owner，样本数据之外仍会出现随机弱项

## Task 2

### Goal

引入共享建议规则层 `internal/teaching/advice`，先接管弱项评估、统一严重度和班级复盘建议生成。

### Touched modules or boundaries

- `code/backend/internal/teaching/advice`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- `code/backend/internal/dto/teacher.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`

### Dependencies

- 依赖 Task 1 提供的教学信号输入

### Validation

- `go test ./internal/module/teaching_readmodel/... -count=1`
- `go test ./internal/module/assessment/application/queries -count=1`
- 新增 class review focused tests，覆盖：
  - 低活跃风险
  - 高置信度薄弱维度
  - 闭环不足学生
  - 高试错成本学生

### Review focus

- `GetClassReview` 是否只保留装配与访问控制，而不再持有规则 owner
- `RecommendationService` 是否只消费 advice 的弱项评估，而不再自己判弱项
- 规则输出是否能追溯到明确指标
- advice DTO 是否已经直接替换旧的 `Accent`、`Level` 语义
- 前后端消费是否已经统一改为 `severity` 主语义

### Risk notes

- 如果 advice 层只返回大段文案，后续仍难以复用到归档和推荐理由
- 如果 `GetWeakDimensions` 仍保留原 owner，结构债就没有真正收口
- 如果 class review 继续保留旧 helper 作为兜底，结构债就没有真正收口

## Task 3

### Goal

把个人复盘观察和推荐题理由迁入共享建议规则层，统一建议码、严重度和证据摘要。

### Touched modules or boundaries

- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/dto/recommendation.go`
- `code/backend/internal/module/assessment/domain/report.go`
- `code/backend/internal/teaching/advice`

### Dependencies

- 依赖 Task 2 中 advice 层的核心模型与规则框架

### Validation

- `go test ./internal/module/assessment/... -count=1`
- focused tests 覆盖：
  - `training_closure`
  - `submission_stability`
  - `hands_on_depth`
  - `dimension_focus`
  - `severity` 直接驱动三条链路输出
  - 推荐题理由包含维度、难度和证据解释

### Review focus

- `hint_usage` 固定文案是否被真正替换掉，而不是继续保留占位逻辑
- 推荐理由是否仍然只是回显维度名
- report archive 与 recommendation 是否共用同一套 advice rule code
- 三条链路是否共用同一套 evidence-insufficient 负向判断

### Risk notes

- 如果 recommendation service 只在文本层补文案，没有引入共享 reason model，后续仍会再次分叉
- 如果 report observation 没有显式区分“过程问题”和“维度问题”，教师建议仍会混杂

## Task 4

### Goal

用教学复盘样本数据做回归验证，确认不同学生场景确实得到不同建议，并补齐本轮最小充分测试证据。

### Touched modules or boundaries

- `code/backend/cmd/seed-teaching-review-data`
- 相关测试文件
- 必要时微调 `docs/architecture/features/教学复盘建议生成架构.md`

### Dependencies

- 依赖 Task 1~3 全部完成

### Validation

- `go test ./internal/module/teaching_readmodel/... ./internal/module/assessment/... ./cmd/seed-teaching-review-data -count=1`
- `go run ./cmd/seed-teaching-review-data`
- 必要时用 `docker exec ctf-postgres psql ...` 核对输出所依赖的关键事实

### Review focus

- 建议结果是否能区分高闭环、低活跃、连续错误提交和 AWD 实战参与学生
- 推荐题理由是否比现有模板更具解释性
- 证据不足维度是否在 class review、archive、recommendation 三条链路里都不会被判成明确弱项
- 本轮是否对样本数据做了为“让建议好看”而违背真实训练逻辑的硬编码

### Risk notes

- 如果样本数据只能验证 happy path，规则很容易在真实数据上退化
- 如果为了通过验证继续手写 `skill_profiles` 而不修弱项判定，设计目标没有完成
- 如果实现采用按学生 fan-out 的事实查询，性能风险会直接回流到建议层

## Integration Checks

- `class review`、`review archive`、`recommendation` 三条链路是否都只通过共享 advice 层输出建议
- 现有教师端和学生端消费是否已经全部切到统一 advice 契约
- 同一学生在班级建议、个人观察和推荐题理由中的弱项维度与过程判断是否一致
- 查询方式是否仍然保持聚合读取或有界批量读取，而不是按学生逐个 fan-out

## Rollback / Recovery Notes

- `internal/teaching/advice` 可以独立回滚，不影响底层事实表
- 当前处于开发期，DTO 与前端消费按最终契约一次性重构，不保留旧字段降级路径
- 本计划不涉及 schema 迁移或持久化数据回填，无需数据库恢复步骤

## Residual Risks

- 第一阶段仍然是规则型建议，不会覆盖所有教师个体化判断
- 若后续需要年级级聚合或长期趋势画像，应另起专题，不在本轮建议层里继续堆积
