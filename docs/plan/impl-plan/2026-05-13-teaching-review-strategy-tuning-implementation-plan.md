# Teaching Review Strategy Tuning Implementation Plan

## Objective

修正教师侧教学复盘与推荐链路里几类已经影响判断质量的策略问题：

- 班级建议不再把任意学生的第一条个人推荐题挂到 `activity_risk`、`training_closure_gap`、`retry_cost_high` 这类流程型建议上
- `activity_risk` 不再因为“班级整体活跃健康，但有少数低活跃学生”就直接升级成 `warning`
- 推荐理由里的难度描述不再把实际 `easy` 题写成 `beginner`
- 个人复盘里的 `submission_stability` 不再把单次错误提交误报成高试错
- 个人复盘里的 `low_activity` 不再基于缺失的 7 天活跃事实全面误报
- knowledge tag 命中的推荐题不再用题目原始 `category` 错绑推荐维度
- 健康、证据稳定的学生不再被系统为了“始终给题”而强行制造推荐题或 `dimension_focus`

同时把新的策略边界同步到 `docs/architecture/features/教学复盘建议生成架构.md`，保证代码和事实源一致。

## Non-goals

- 不改学生个人推荐的“证据优先”弱项识别规则
- 不改 `challenge` 推荐候选集合查询 SQL
- 不新增教师端 API、DTO 字段或数据库 schema
- 不重做班级建议体系，也不改推荐题排序；只修正当前最明显的语义错配和阈值问题

## Inputs

- `docs/architecture/features/教学复盘建议生成架构.md`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/model/challenge.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service_test.go`

## Current Baseline

- `BuildClassReview` 会给 `activity_risk`、`training_closure_gap`、`retry_cost_high` 都带上 `RecommendationStudentID`
- `ClassInsightQueryService` 只拿该学生的第一条个人推荐题，直接附着到班级建议上
- `activity_risk` 只要出现任意低活跃学生，就会把全班降成 `warning`
- 个人推荐理由直接复用 `recommended_difficulty_band`，当题库里实际命中的候选题更高一级时，文案会把 `easy` 题说成 `beginner`
- 个人复盘侧 `submission_stability` 只要 `wrong > correct` 就会升级，导致单次错误提交也可能被判成 `warning`
- 个人复盘侧 `low_activity` 原本没有独立观察项；补观察项后又暴露出 review archive 事实快照没有近 7 天活跃统计
- challenge 推荐查询允许通过 knowledge tag 命中候选，但推荐理由仍可能退回题目原始 `category` 解释维度
- 现有 seed 输出已经证明：这些规则会在真实样本下产出不自然的教师建议

## Chosen Direction

1. 班级建议的推荐题只保留在“维度聚焦型”建议上，也就是 `weak_dimension_cluster`
2. 班级维度推荐不再盲拿某个学生的第一条推荐，而是只接受和该 class item `dimension` 匹配的推荐结果；找不到就不挂题
3. `activity_risk` 改成四档：
   - `danger`：整体活跃明显下滑或低活跃学生比例过高
   - `warning`：班级节奏开始走低，需要班级层面干预
   - `attention`：整体训练仍健康，但有少数学生节奏偏慢
   - `good`：整体稳定且没有明显掉队点
4. 个人推荐仍保持 `recommended_difficulty_band` 作为目标训练带宽，但推荐理由里的难度表述要以实际候选题 `challenge.difficulty` 为准；若两者不一致，文案需要明确这是当前最接近的候选题
5. 个人复盘里的 `submission_stability` 改成“连续错误或重复错误成本达到风险阈值”才升级，不再把单次错误提交误判成高试错。
6. 个人复盘里的 `low_activity` 由 review archive 装配层补齐近 7 天活跃事实后再判定，避免默认 0 值误报。
7. challenge 推荐查询继续保留现有 SQL owner，但要把实际命中的推荐维度沿链路传回 `RecommendationService` 与共享规则层。
8. 如果学生当前没有 evidence-backed weak / attention target，个人推荐允许为空；个人复盘也不再强行输出 `dimension_focus`。
9. 同步更新专题架构文档，明确：
  - 流程型建议默认只给动作，不给题
  - 题目推荐只属于维度补强型建议
  - 活跃风险要同时看整体活跃率和低活跃学生占比
  - 推荐理由的 `summary` 不能伪装题目实际难度
  - 个人低活跃观察依赖真实近 7 天活跃事实
  - 健康、证据稳定的学生允许没有推荐题

## Ownership Boundary

- `internal/teaching/advice`
  - 负责：班级建议的严重度分层、是否允许挂题的规则，以及个人 `submission_stability / low_activity` 的阈值语义
  - 不负责：查询候选题、决定推荐题具体 SQL
- `ClassInsightQueryService`
  - 负责：只在语义匹配的 advice item 上挂推荐题，并保证挂上的题和建议维度一致
  - 不负责：重新定义个人推荐规则或挑战查询规则
- `assessment/application/commands/report_service.go`
  - 负责：在 review archive 链路里把近 7 天活跃事实装配进个人教学事实快照
  - 不负责：绕过共享规则层直接生成教师观察文案
- `challenge/infrastructure/repository.go`
  - 负责：在现有候选题查询结果里补充实际命中的推荐维度
  - 不负责：生成推荐理由或改变候选排序策略
- `docs/architecture/features/教学复盘建议生成架构.md`
  - 负责：记录新的 class-level 策略边界与当前实现口径

## Change Surface

- Add: `.harness/reuse-decisions/teaching-review-strategy-tuning.md`
- Add: `docs/plan/impl-plan/2026-05-13-teaching-review-strategy-tuning-implementation-plan.md`
- Modify: `code/backend/internal/teaching/advice/advice.go`
- Modify: `code/backend/internal/teaching/advice/advice_test.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service.go`
- Modify: `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- Modify: `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- Modify: `code/backend/internal/model/challenge.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/repository.go`
- Modify: `code/backend/internal/module/challenge/infrastructure/repository_test.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- Modify: `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service_test.go`
- Modify: `docs/architecture/features/教学复盘建议生成架构.md`

## Task Slices

### Slice 1: 修正班级建议挂题语义

目标：

- 流程型建议不再附个人推荐题
- 维度聚焦型建议只附与该维度匹配的推荐题

Validation:

- `cd code/backend && go test ./internal/module/teaching_readmodel/application/queries -run 'TestClassInsightQueryServiceGetClassReview' -count=1 -timeout 120s`

Review focus:

- 是否真的消除了“活跃风险后挂 crypto 题”这类语义错配
- 是否保留了 `weak_dimension_cluster` 的可执行训练建议

### Slice 2: 重标 activity_risk 阈值

目标：

- 少数低活跃学生只触发 `attention`
- 班级级 warning / danger 必须由整体活跃率或比例级掉队支撑

Validation:

- `cd code/backend && go test ./internal/teaching/advice -run 'TestBuildClassReview' -count=1 -timeout 120s`

Review focus:

- 新阈值是否仍能保留真正的班级风险信号
- 是否避免把个别学生问题误报成班级 warning

### Slice 3: 收口个人复盘误报与缺失观察

目标：

- `submission_stability` 不再把单次错误提交误报成 `warning`
- `low_activity` 基于 review archive 里的真实近 7 天活跃事实生成

Validation:

- `cd code/backend && go test ./internal/teaching/advice ./internal/module/assessment/application/commands -count=1 -timeout 120s`
- `cd code/backend && go run ./cmd/seed-teaching-review-data`

Review focus:

- 是否补上了低活跃观察但没有把所有学生都误判成掉队
- 是否移除了单次错误提交的误报，同时保留真正的重复试错风险

### Slice 4: 修正推荐维度归因

目标：

- knowledge tag 命中的候选题要沿链路保留实际命中的推荐维度
- 推荐理由与教师侧附题都不能再靠题目原始 `category` 猜维度

Validation:

- `cd code/backend && go test ./internal/module/assessment/application/queries ./internal/module/challenge/infrastructure -count=1 -timeout 120s`

Review focus:

- 推荐理由是否解释了真正命中的推荐维度
- 是否保持现有候选排序与查询 owner 不变

### Slice 5: 同步专题架构文档

目标：

- 把新的 class-level 策略写成当前事实，而不是停留在实现细节
- 满足文档一致性脚本对 `## 当前设计` 的要求

Validation:

- `python3 scripts/check-docs-consistency.py`

Review focus:

- 文档是否清楚写明“负责 / 不负责”和策略证据来源
- 是否避免把计划性表述写成当前事实

## Verification Plan

1. `cd code/backend && go test ./internal/module/assessment/application/commands ./internal/teaching/advice ./internal/module/assessment/application/queries ./internal/module/challenge/infrastructure ./internal/module/teaching_readmodel/application/queries -count=1 -timeout 120s`
2. `cd code/backend && go run ./cmd/seed-teaching-review-data`
3. `python3 scripts/check-docs-consistency.py`
4. `bash scripts/check-consistency.sh`

## Risks

- 如果班级维度推荐过滤过严，可能导致 `weak_dimension_cluster` 暂时不挂题
- 如果活跃阈值放得过松，真实班级风险可能被延后暴露
- 如果推荐理由继续把目标带宽误写成实际题目难度，教师和学生会误判推荐是否合适
- 如果 review archive 继续丢失近 7 天活跃事实，`low_activity` 会再次全面误报
- 如果 knowledge tag 命中的维度没有沿链路回传，推荐解释会继续把题挂到错误维度
- 如果只改代码不改文档，后续 agent 可能继续按旧策略理解 class review

## Architecture-Fit Evaluation

- 规则 owner 仍然收口在 `internal/teaching/advice`
- `teaching_readmodel` 只负责装配和语义过滤，没有新建并行推荐规则层
- 文档更新直接落到现有专题事实源，不新增重复设计稿
