# 推荐链路 difficulty band 与 progression owner 实施计划

> 状态：Draft
> 事实源：`docs/architecture/features/教学复盘建议生成架构.md`、`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`
> 替代：无

## 1. 目标

- 让 `assessment.RecommendationService` 把目标 `difficulty_band` 传入候选题查询
- 让 `challenge` 推荐候选查询按“维度 + 目标难度带”决定优先级
- 让没有弱项的健康学生也能进入进阶推荐，而不是直接返回空结果
- 保留返回结果里“目标训练带宽”和“候选题实际难度”的双字段表达

## 2. 范围

### 包含

- 调整 `RecommendationChallengeRepository` 和 `ChallengeContract` 的推荐查询契约
- recommendation snapshot 补齐维度内已解难度覆盖事实
- `teaching/advice` 基于现有教学事实快照生成 progression-ready 推荐目标
- `RecommendationService` 使用 `evaluation.RecommendedDifficultyBand` 参与候选查询
- `challenge` 仓储把目标难度带纳入排序，优先返回同维度、同难度或最接近难度候选
- 同步推荐服务测试、challenge 仓储测试和当前专题架构事实文档

### 不包含

- 不重写 `internal/teaching/advice` 的弱项与证据判断主干，只在无弱项时补 progression 分支
- 不扩展新的推荐 API 字段；当前 `difficulty_band` 与 `difficulty` 继续复用
- 不处理今天同一份 review 里的另外两条 owner 缺口：竞赛数据回流、班级导出结构

## 3. 方案判断

- 推荐方向：把目标难度带提升为 challenge 推荐查询契约的一部分，由下游仓储负责候选排序 owner；上游 `advice` 继续负责决定当前是在“补弱/补样本”还是“进阶”
- 不采用“assessment 先查一批再本地二次排序”，避免把候选排序 owner 留在上游并扩大跨模块耦合
- 不采用“严格只查同难度题”，因为当前架构文档明确允许返回“最接近候选”
- 不采用“assessment 自己单独发明 progression 判断”，因为这会把 recommendation owner 又拆回双轨

## 4. 实施步骤

1. 先为 `advice` 和推荐服务补失败测试，锁定“健康学生进入进阶推荐”与“难度带影响候选选择”
2. 扩展教学事实快照，补齐维度内已解难度覆盖事实
3. 调整 `advice`：无弱项时允许生成 progression-ready target，并把 band 提升到 `hard/insane`
4. 调整 assessment/challenge 契约，让推荐查询显式接收 preferred difficulty
5. 实现 challenge 仓储排序：同维度、同目标难度优先，再按距离最近的难度带回退
6. 更新 `教学复盘建议生成架构.md` 的当前事实描述

## 4.1 完成清单

- [x] 为 `advice`、推荐服务和 challenge 仓储补齐 progression / difficulty owner 相关测试
- [x] 在 recommendation snapshot 中补齐维度内 `solved_difficulty_counts`
- [x] 让 `internal/teaching/advice` 在无补弱目标时支持 progression-ready 推荐
- [x] 让 `assessment -> challenge` 推荐契约显式传递 preferred difficulty
- [x] 让 challenge 推荐查询按目标 difficulty band 返回最接近候选
- [x] 同步更新专题架构事实文档和中间设计稿

## 5. 影响范围

- `code/backend/internal/module/assessment/application/queries/`
- `code/backend/internal/module/assessment/application/commands/`
- `code/backend/internal/module/assessment/infrastructure/`
- `code/backend/internal/module/assessment/ports/ports.go`
- `code/backend/internal/module/challenge/contracts/`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/module/teaching_query/infrastructure/`
- `code/backend/internal/teaching/advice/`
- `docs/architecture/features/教学复盘建议生成架构.md`

## 6. 契约与兼容

- `FindPublishedForRecommendation(...)` 签名会增加 preferred difficulty 参数
- 返回 DTO 保持不变，前端仍通过 `difficulty_band` 读取目标训练带宽，通过 `difficulty` 读取候选题实际难度
- 若 preferred difficulty 为空或不合法，仓储应退回当前稳定排序，不引入空结果回归
- review archive / class review 这次继续不把 progression-ready 渲染成问题项；进阶语义先限制在推荐链路

## 7. 验证

- `go test ./internal/module/assessment/application/queries -run 'TestRecommendationServiceRecommendChallenges' -count=1`
- `go test ./internal/teaching/advice -run 'TestEvaluateStudent|TestBuildRecommendationPlan' -count=1`
- `go test ./internal/module/challenge/infrastructure -run 'TestRepositoryFindPublishedForRecommendation' -count=1`
- `python3 scripts/check-docs-consistency.py`

## 8. 风险

- 如果仓储把目标难度实现成硬过滤，可能在题库样本稀疏时把推荐误降成空结果
- 如果 preferred difficulty 没有统一归一化，`difficulty_band` / `difficulty` 的排序距离会漂移
- 契约签名变化会同时触达 `assessment` port 和 `challenge` contract，需要测试确保装配侧仍能编译通过
- 当前没有独立持久化的 mastery 表，进阶推荐依赖维度内已解难度覆盖现算；如果后续要收口更精细的 mastery owner，再单独扩展 snapshot
