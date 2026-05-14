# AWD 教学事实快照回流实施计划

> 状态：Draft
> 事实源：`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`、`docs/design/AWD能力画像回流方案.md`、`docs/architecture/features/教学复盘建议生成架构.md`
> 替代：无

## 1. 目标

- 让 recommendation snapshot 和 class teaching fact snapshot 吸收 AWD 个人正向成功证据
- 让共享 `advice` 层在 recommendation / class review 上不再只看到练习侧维度成功样本
- 保持当前 `skill_profiles` 持久化公式和 `solved / score / rank` 统计口径不变

## 2. 范围

### 包含

- 在 `assessment` 个人 teaching fact snapshot 中补 AWD 维度成功覆盖、difficulty 覆盖和 profile score 补充信号
- 在 `teaching_query` 班级 teaching fact snapshot 中补同一套 AWD 维度成功覆盖、difficulty 覆盖和 profile score 补充信号
- 把 `AWDSuccessCount` 统一收紧到 `submitted_by_user_id + source=submission + score_gained>0 + is_success=true` 口径
- 补 recommendation / class review 相关测试，验证 AWD 证据会改变 teaching fact snapshot 的维度结论
- 更新专题架构事实文档与 AWD 回流设计稿

### 不包含

- 不修改 `AssessmentDimensionScoreRepository` 和 `skill_profiles` 持久化计算公式
- 不做 AWD 防守分回流
- 不做班级报告时间段参数和导出结构扩展
- 不做前端页面改动

## 3. Brainstorming 结论

- 候选方向 A：直接把 AWD 成功塞进 `skill_profiles` 持久化公式
  - 不采用：当前普通题 `Challenge` 与 `AWDChallenge` 是两套表，普通题是 points 口径，AWD 题没有同一套分值模型；硬塞进 `AssessmentDimensionScoreRepository` 会制造伪统一分值。
- 候选方向 B：只改 `AWDSuccessCount` 展示，不改维度事实
  - 不采用：这正是当前 thesis gap 的问题，AWD 仍然只是附加信号，不会影响 recommendation / class review 的维度判断。
- 选定方向：在 teaching fact snapshot owner 里补 AWD 个人正向成功证据
  - `profile_score`：取 `max(skill_profiles.score, awd_success_coverage_ratio)`
  - `success_count` / `evidence_count`：补充 AWD 个人成功覆盖
  - `solved_difficulty_counts`：补充 AWD challenge difficulty 覆盖

## 4. 实施步骤

1. 先补测试，锁定 AWD 成功证据会影响 recommendation snapshot 和 class teaching snapshot
2. 在 `assessment/infrastructure.Repository` 中补 AWD 维度成功聚合，合并到个人 `DimensionFact`
3. 在 `teaching_query/infrastructure.Repository` 中补同一套 AWD 维度成功聚合，合并到班级 `DimensionFact`
4. 收紧 `AWDSuccessCount` 到统一的个人正向成功证据口径
5. 更新架构事实文档与 AWD 回流设计稿

## 4.1 完成清单

- [x] 为个人 recommendation snapshot 补 AWD 回流测试
- [x] 为班级 teaching fact snapshot 补 AWD 回流测试
- [x] 在 `assessment` teaching snapshot 中补 AWD 维度成功覆盖与 difficulty 覆盖
- [x] 在 `teaching_query` teaching snapshot 中补 AWD 维度成功覆盖与 difficulty 覆盖
- [x] 统一 `AWDSuccessCount` 统计口径
- [x] 同步更新事实文档和中间设计稿

## 5. 影响范围

- `code/backend/internal/module/assessment/infrastructure/`
- `code/backend/internal/module/assessment/application/queries/`
- `code/backend/internal/module/teaching_query/infrastructure/`
- `code/backend/internal/module/teaching_query/application/queries/`
- `docs/architecture/features/教学复盘建议生成架构.md`
- `docs/design/AWD能力画像回流方案.md`

## 6. 契约与兼容

- 对外 API / DTO 不新增字段
- `DimensionFact.profile_score` 在 teaching snapshots 中会从“持久化 skill_profile 原值”变成“practice profile 与 AWD success coverage 的合并值”
- recommendation / class review 仍继续通过共享 `internal/teaching/advice` 输出，不新增第二套 owner
- `skill_profiles` 表本身不在本次切片内改写；`/skill-profile` 页面持久化画像仍保持当前 owner

## 7. 验证

- `go test ./internal/module/assessment/application/queries -count=1`
- `go test ./internal/module/teaching_query/application/queries -count=1`
- 如新增基础设施测试：`go test ./internal/module/assessment/infrastructure ./internal/module/teaching_query/infrastructure -count=1`
- `python3 scripts/check-docs-consistency.py`
- `bash scripts/check-consistency.sh`
- `bash scripts/check-workflow-complete.sh`

## 8. 风险

- 如果 AWD success coverage 直接覆盖 practice profile，而不是保守取 `max`，可能会把已有训练画像抹掉
- 如果把 AWD 原始 success log 全量计入 evidence，可能因为重复打多个受害队而虚增证据
- 如果 class snapshot 和 personal snapshot 的 AWD 口径不一致，会再次出现 recommendation / class review 结论分叉
- 当前仍不处理班级时间段 / 导出结构 owner，这条 thesis gap 会继续保留到下一切片
