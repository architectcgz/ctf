# review archive 三层复盘收口实现方案

## Objective

把学生复盘归档这条链路改成更容易解释的三层结构：

- 事实层：只沉淀训练、AWD、材料提交等可复核事实
- 分析层：只保留简单规则分析，不再依赖复杂画像推断
- 输出层：生成教师可直接阅读的个人训练复盘条目

同时同步最小前端文案，并在代码验证通过后把论文相关章节改成同一口径。

## Non-goals

- 不重写 `SkillProfile` 六维分值计算
- 不改推荐接口、班级复盘接口和教师学生分析页的大结构
- 不删除 `internal/teaching/advice` 中仍供推荐与班级复盘使用的其他规则

## Inputs

- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service_test.go`
- `code/backend/internal/teaching/advice/advice.go`
- `code/backend/internal/teaching/advice/advice_test.go`
- `code/frontend/src/widgets/teacher-review-archive/*`
- `docs/thesis/paper-template/tex/detailed_design_implementation.tex`

## Ownership Evaluation

- owner 明确：学生复盘归档装配 owner 仍在 `assessment/application/commands/report_service.go`
- 规则 landing zone 明确：个人复盘简单规则收口到 `internal/teaching/advice`
- 输出边界明确：后端仍返回 `teacher_observations.items`，前端只做展示口径同步
- 结构收敛明确：review archive 不再复用“复杂画像评估 -> 观察项”的旧路径，而是显式走“事实 -> 简单规则 -> 复盘输出”

## Task slices

1. 新增 review archive 简单规则分析模型与输出生成函数
2. 让 `report_service` 改走新的三层复盘链路
3. 更新后端测试，覆盖薄弱方向、稳定方向、活跃情况、表达与总结、AWD 摘要
4. 同步复盘页最小文案与前端测试
5. 跑代码验证
6. 按相同口径改论文章节并检查编译

## Validation

- `go test ./internal/teaching/... -count=1`
- `go test ./internal/module/assessment/application/commands -count=1`
- `pnpm vitest run src/widgets/teacher-review-archive/TeacherReviewArchiveWorkspace.test.ts src/widgets/teacher-review-archive/TeacherReviewArchiveSummarySection.test.ts src/views/teacher/__tests__/TeacherStudentReviewArchive.test.ts`
- `bash scripts/check-consistency.sh`
- `latexmk -synctex=1 -pdfxe -shell-escape -interaction=nonstopmode -file-line-error -outdir=tmp Thesis.tex`

## Review focus

- 复盘归档是否只基于可复核事实出结论
- 观察项是否已经改成简单规则分析，而不是复杂推断
- 六维图是否保留为展示层事实，不再被写成复盘推断公式
- 论文章节是否与代码实际输出口径一致

## Rollback

如果新规则导致复盘结论明显失真，可回退到旧的 `BuildReviewArchiveObservations` 路径；前端文案与论文改动不涉及数据迁移。
