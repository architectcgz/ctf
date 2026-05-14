# AWD 能力画像回流方案

> 状态：Draft
> 事实源：中间方案、方案比较与落地取舍，不覆盖当前 `docs/architecture/` / `docs/contracts/` / 代码事实
> 替代：当前已落地的最终事实见 `docs/architecture/features/教学复盘建议生成架构.md`

## 定位

本文档保留 AWD 攻击证据如何回流教学事实快照、推荐目标层和班级复盘维度事实的中间设计判断。

- 当前文档保留问题重述、方案比较和落地取舍，不覆盖 `docs/architecture/`、`docs/contracts/` 和当前代码事实。
- 已稳定的最终边界、snapshot owner 和评估口径已经回收到 `docs/architecture/features/教学复盘建议生成架构.md`。

当前仓库已经具备两项 AWD 回流前置能力：

- `awd_attack_logs.submitted_by_user_id` 已经存在，可把攻击成功归因到个人提交人
- `contest.awd.attack_accepted` 事件及 assessment 侧 consumer 已经存在

但毕业设计里“基于实训数据生成能力画像，推荐针对性靶场练习”的缺口没有完全关闭，因为 recommendation snapshot 和 class teaching fact snapshot 仍长期主要按练习侧事实组装维度结论。

这意味着 AWD 会影响“是否活跃、是否有实操、是否有 AWD 成功”，但不稳定影响“当前哪个维度该补样本、该推荐什么维度和难度带的题”。

## 非目标

- 不重做 `SkillProfile.vue` 页面结构
- 不把队伍级防守分直接映射到个人画像
- 不在这一轮补教师端新的画像面板
- 不修改 `AssessmentDimensionScoreRepository` 或 `skill_profiles` 持久化公式
- 不调整 `ListSolvedChallengeIDs` / `excludeSolved` owner，本轮只补 teaching snapshot 与 recommendation target 层
- 不重算历史报告模板，只补 teaching snapshot 与 recommendation / class review 的事实源

## 关键问题

当前缺口已经不再是“有没有个人提交人字段”或“有没有 AWD 事件链”，而是 teaching snapshot owner 仍以练习侧维度事实为主。

另一个边界约束是：`Challenge` 与 `AWDChallenge` 是两套表，普通练习画像当前使用 points 驱动的持久化公式，而 `AWDChallenge` 没有同一套 points 口径。为了“统一画像”强行把 AWD 成功塞进 `skill_profiles`，会制造伪统一分值。

## 方案比较

### 方案 A：直接改 `skill_profiles` 持久化公式，把 AWD 成功塞进画像分值

优点：

- 看起来最像“真正把 AWD 写回画像”
- 理论上 `/skill-profile` 页面也能直接看到变化

缺点：

- `AWDChallenge` 没有普通题 `Challenge.points` 的统一口径
- `AssessmentDimensionScoreRepository` 目前是 points / published-total 逻辑，硬塞 AWD 会发明不存在的分值模型
- 会把本轮 owner 从 teaching snapshot 扩大成持久化画像重构，超出当前 thesis gap 的最小收口范围

不采用。

### 方案 B：只保留 `AWDSuccessCount` 辅助信号，不改维度事实

做法：

- 继续让 AWD 只影响活跃度、实操深度和 `AWDSuccessCount`
- recommendation / class review 的 `DimensionFact` 仍主要来自普通练习

优点：

- 改动最小
- 不需要触碰 teaching snapshot owner

缺点：

- 这正是当前 thesis gap 的问题本身
- AWD 仍然只是“附加证据”，不会改变 recommendation target 和 class review 的维度判断

不采用。

### 方案 C：在 teaching fact snapshot owner 合并 AWD 个人正向成功证据

做法：

- 在 `assessment.Repository.GetStudentTeachingFactSnapshot(...)` 中补 AWD 维度成功覆盖、difficulty 覆盖和 `profile_score` 补充信号
- 在 `teaching_query.Repository.ListClassTeachingFactSnapshots(...)` 中补同一套 AWD 维度成功覆盖
- `profile_score` 使用保守合并：`max(skill_profiles.score, awd_success_coverage_ratio)`
- `success_count`、`evidence_count`、`solved_difficulty_counts` 补 AWD challenge 的个人去重成功覆盖
- `AWDSuccessCount` 收紧到统一口径：`submitted_by_user_id + source=submission + is_success=true + score_gained>0`

优点：

- owner 清晰，正好收口 thesis gap 提到的 recommendation / class review 维度事实断层
- 不需要发明 AWD points，也不需要改 `skill_profiles` 持久化表
- recommendation 和 class review 会同时看到同一套 AWD 维度事实

缺点：

- `/skill-profile` 页面持久化画像不会直接显示这次 snapshot 合并结果
- 推荐排除集合 owner 保持现状，这一轮解决的是 recommendation target，而不是所有推荐 owner

推荐采用，并已按该方向落地。

## 目标

本轮完成后：

1. recommendation snapshot 和 class teaching fact snapshot 都能吸收 AWD 个人正向成功证据
2. `internal/teaching/advice` 在 recommendation / class review 上不再只看到练习侧维度成功样本
3. `AWDSuccessCount` 统一收紧到个人正向成功证据口径
4. 保持 `skill_profiles` 持久化公式不变，只在 snapshot 层合并 AWD 覆盖率

## 当前落地设计

采用方案 C，把 AWD 回流 owner 收口在 teaching fact snapshot，而不是持久化画像公式。

前置已存在但不在本轮改动内：

- `awd_attack_logs.submitted_by_user_id`
- `contest.awd.attack_accepted`
- assessment 侧基于 AWD 事件的画像更新 / 推荐缓存 consumer

本轮新增事实合并点：

- `assessment.Repository.GetStudentTeachingFactSnapshot(...)`
- `teaching_query.Repository.ListClassTeachingFactSnapshots(...)`

## 数据与口径

### AWD 个人正向成功证据

个人 AWD 能力证据只统计满足以下条件的记录：

- `source = submission`
- `submitted_by_user_id = 当前用户`
- `is_success = true`
- `score_gained > 0`

这里显式要求 `score_gained > 0`，避免把同队重复提交、复用已知 flag 的行为记成新的个人能力证明。

### Teaching snapshot 合并口径

对于每个维度，snapshot owner 追加三类 AWD 聚合：

- 已发布 AWD 题总数：
  - 来自 `awd_challenges`
  - 仅统计 `status = published`
- 个人去重成功覆盖：
  - 按 `awd_challenge_id` 去重
  - 同一题打多个受害队只记一次
- difficulty 覆盖：
  - 按 `category + difficulty` 聚合个人去重成功的 AWD 题

合并规则：

- `profile_score = max(skill_profiles.score, awd_success_coverage_ratio)`
- `success_count += distinct awd_challenge_id`
- `evidence_count += distinct awd_challenge_id`
- `solved_difficulty_counts[difficulty] += distinct awd_challenge_id`

这意味着 recommendation 和 class review 读取到的 `DimensionFact`，会把普通练习基础和个人 AWD 正向成功覆盖一起纳入评估。

### 推荐层效果

- `internal/teaching/advice` 读取的是合并后的 teaching snapshot，而不是单独的 AWD 展示字段
- 当学生在某一维度已经完成 `easy + medium` 的 AWD 成功覆盖，即使 `skill_profiles.score` 仍偏低，也可以把该维度从“弱项”转成 `progression-ready`
- 当前推荐改进体现在 recommendation target 和 difficulty band，不在本轮调整 `excludeSolved` owner

## 边界与取舍

### 为什么这轮不改 `skill_profiles` 持久化公式

- `AWDChallenge` 没有与普通练习一致的 `points` 语义
- 当前 `AssessmentDimensionScoreRepository` 是持久化画像 owner，不适合在这一刀里硬塞新的伪分值模型
- thesis gap 当前真正需要收口的是 teaching fact snapshot 对 recommendation / class review 的维度事实断层

### 为什么这轮不纳入防守分

当前 `awd_team_services` 是队伍级结果：

- 无法区分是哪个成员完成修复、防守或运维动作
- 直接均摊到个人会污染画像

因此这轮只纳入“可明确归因到个人提交行为”的攻击侧证据。

### 为什么这轮不改前端和推荐排除 owner

- 现有推荐卡片、学生首页和教师复盘接口已经能消费 recommendation / class review 结果
- 这次改动发生在 snapshot owner 和 advice 输入层，前端不需要新增字段
- `ListSolvedChallengeIDs` / `excludeSolved` 仍是独立 owner；本轮 closing 的是“维度画像与推荐目标层回流”，不是所有推荐行为都改成 AWD 并集

## 测试策略

至少覆盖：

1. `assessment.Repository.GetStudentTeachingFactSnapshot(...)`
   - AWD 个人正向成功证据会抬升 `profile_score`
   - `success_count`、`evidence_count`、`solved_difficulty_counts` 会合并 AWD 覆盖
   - `AWDSuccessCount` 会忽略 `score_gained = 0` 与 `source != submission`
2. `teaching_query.Repository.ListClassTeachingFactSnapshots(...)`
   - 班级 snapshot 与个人 snapshot 使用同一套 AWD 合并口径
3. `RecommendationService`
   - 当普通画像低但 AWD `easy + medium` 覆盖充分时，推荐链路会切到 progression-ready，并优先返回更高一档候选题

本轮对应测试：

- `code/backend/internal/module/assessment/infrastructure/repository_test.go`
- `code/backend/internal/module/teaching_query/infrastructure/repository_test.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service_test.go`

## 与毕业设计要求的对应

- “技能评估，基于实训数据生成能力画像，推荐针对性靶场练习”
  - recommendation snapshot 和 class teaching snapshot 现在会把 AWD 个人正向成功覆盖并回维度事实
  - recommendation target 会直接读取这套合并后的维度事实，而不是只看练习侧样本
- “攻防演练，学员选择靶场后获取攻击目标，系统自动记录攻击步骤与漏洞利用过程”
  - 个人 AWD 提交成功证据现在不再只停留在日志与活跃度，而会进入教学事实快照的维度判断
- “导出实训报告供教学复盘”
  - 这轮先补 recommendation / class review 的 teaching fact owner；班级时间段和导出结构仍是下一切片

## 结论

这次 AWD 回流不再试图“一步统一所有画像 owner”，而是先把 thesis gap 真正缺的 owner 收口在 teaching fact snapshot。这样 recommendation 和 class review 终于能把个人 AWD 正向成功样本稳定带进维度事实，同时又不需要在当前切片里重写 `skill_profiles` 的持久化公式。
