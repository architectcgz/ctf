# assessment 六维推荐简化实现方案

## Objective

把能力画像推荐逻辑收口到 `SkillProfile` 六维得分本身：

- `weak_dimensions` 只根据六维画像低分结果生成
- 推荐目标维度只取当前最低分维度
- 推荐难度只由该维度当前覆盖率推导
- 学生端与教师端文案同步改成“六维学习画像 / 六维训练分布”口径

## Non-goals

- 不改动 `SkillProfile` 的六维计算公式
- 不删除教学复盘、证据链、题解审核等教师侧功能
- 不重构 recommendation HTTP contract、缓存结构或前端路由

## Inputs

- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`
- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/frontend/src/views/profile/SkillProfile.vue`
- `code/frontend/src/components/teacher/StudentInsightPanel.vue`
- `code/frontend/src/views/profile/__tests__/SkillProfile.test.ts`

## Ownership Evaluation

- owner 明确：六维弱项与推荐目标 owner 在 `assessment/application/queries/recommendation_service.go`
- 复用点明确：继续复用既有 `SkillProfile`、challenge recommendation repository 和前端能力画像页面
- 结构收敛明确：不再让 `teaching/advice` 决定六维弱项与推荐主目标，避免 AWD 证据补分改变六维图口径
- 影响面可控：后端推荐查询、对应测试，以及前端画像页与教师侧图表说明文案

## Task slices

1. 修复并收口后端 recommendation service
2. 按六维低分口径更新 assessment 推荐测试
3. 同步学生端与教师端画像文案
4. 跑最小充分验证并记录结果

## Validation

- `go test ./internal/module/assessment/application/queries -count=1`
- `pnpm vitest run src/utils/__tests__/skillProfile.test.ts src/views/profile/__tests__/SkillProfile.test.ts`
- `bash scripts/check-consistency.sh`

## Review focus

- `weak_dimensions` 是否只来自 `SkillProfile`
- 推荐目标是否只针对最低分维度，而不是健康维度进阶或 AWD 补分结果
- 推荐摘要与证据文案是否使用真实覆盖率分数，而不是反向置信度
- 前端“六维画像”文案是否与后端口径一致

## Rollback

如有回归，可回退 recommendation service 到原先的 teaching advice 驱动逻辑；前端文案回退不涉及数据迁移或契约变更。
