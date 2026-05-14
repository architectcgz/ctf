# 2026-05-14 教学推荐与复盘对照选题/开题要求 Gap Review

## Review Target

- 仓库：`ctf`
- worktree：`/home/azhi/workspace/projects/ctf/.worktrees/teaching-review-strategy-tuning-wt`
- 分支：`feat/teaching-review-strategy-tuning-wt`
- diff source：当前 `main` 与 worktree 上教学复盘相关实现、设计文档和学校材料的静态对照
- reviewed files:
  - `docs/毕业设计课题.md`
  - `docs/开题报告/开题报告.md`
  - `docs/architecture/features/教学复盘建议生成架构.md`
  - `docs/design/教学复盘建议优化方案.md`
  - `code/backend/internal/teaching/advice/advice.go`
  - `code/backend/internal/module/assessment/application/commands/report_service.go`
  - `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
  - `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
  - `code/backend/internal/module/assessment/infrastructure/repository.go`
  - `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
  - `code/backend/internal/module/challenge/infrastructure/repository.go`
  - `code/backend/internal/dto/recommendation.go`
  - `code/backend/internal/dto/teacher.go`

## Classification Check

- 结论：同意按 `architecture review / gap review` 归档。
- 理由：本轮不是实现 review，而是校验“当前教学推荐与复盘”是否已经满足选题与开题材料里关于能力画像、推荐、导出和教学闭环的论证要求。

## Gate Verdict

- `blocked`
- 当前实现已经明显提升可解释性，但还不能直接声称“已经满足选题/开题里关于技能评估、推荐练习、教学复盘闭环的核心要求”。

## Findings

### P1. 推荐链路还没有形成真正的“分类 + 难度”双维评价模型

- 选题和开题要求都把“按分类与难度维度”的能力画像和评价模型作为技能评估核心目标。
- 当前推荐链路在 `RecommendationService` 中只把 `targetDimensions` 传给题目查询，难度带没有进入候选筛选 owner。
- `FindPublishedForRecommendation(...)` 只按维度筛题，再按全局难度升序返回；`difficulty_band` 主要停留在 `advice` 的解释文案里，没有实质控制候选集。
- 结果是系统能解释“为什么建议补 web / pwn”，但还不能严格说明“为什么当前就该补 beginner / easy / medium”。

涉及位置：

- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/challenge/infrastructure/repository.go`
- `code/backend/internal/teaching/advice/advice.go`

影响：

- 会削弱开题报告里“按分类与难度维度建立评价模型”的论证强度。
- 推荐结果更像“按维度命中后优先给简单题”，而不是“先判定当前训练带宽，再给同维度、同带宽候选”。

修复方向：

- 把 `difficulty_band` 从解释字段提升为推荐查询契约的一部分。
- 让候选查询至少支持“维度 + 目标难度带”的筛选或排序 owner。
- 让推荐结果在结构上明确区分“当前目标难度带”和“实际题库最接近候选”的关系。

### P1. 训练数据和竞赛数据还没有在维度画像与推荐目标层真正汇合

- 选题和开题都强调要把训练与竞赛数据转成可解释、可导出的能力画像与评估报告。
- 当前个人复盘归档已经开始把 `AWD` 证据写回观察项和部分维度事实，但推荐快照与班级复盘维度事实仍主要依赖练习侧数据。
- `GetStudentTeachingFactSnapshot(...)` 与 `ListClassTeachingFactSnapshots(...)` 中，维度尝试和成功统计都明确限定 `contest_id IS NULL`；`AWD` 现在更多只是独立补到 `AWDSuccessCount`、活跃度和实操深度。
- 结果是竞赛数据会影响“是否活跃、是否有实操、是否有 AWD 成功”，但还不稳定影响“哪个维度偏弱、该推荐什么维度题”。

涉及位置：

- `code/backend/internal/module/assessment/infrastructure/repository.go`
- `code/backend/internal/module/teaching_readmodel/infrastructure/repository.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`

影响：

- 当前实现更接近“训练主画像 + 竞赛补充证据”，还不是“训练与竞赛共同形成统一能力画像”。
- 这会影响“训练-竞赛-评估一体化闭环”的论文主张。

修复方向：

- 明确竞赛数据如何映射回维度事实，而不是只作为 `AWDSuccessCount` 辅助信号。
- 统一推荐、班级复盘、个人归档三条链路的维度事实口径，避免只有归档链路吸收 AWD 事实。
- 补齐“竞赛结果回流后，推荐目标变化”的可验证样本和测试。

### P2. 班级与时间段统计口径、导出结构还不足以支撑开题里的教学评估输出

- 开题报告明确要求研究“班级与时间段的统计口径与导出结构”。
- 当前班级复盘窗口基本写死为最近 7 天；班级报告导出入参只有 `class_name` 和 `format`，没有可配置时间段。
- 当前班级报告导出内容也仍偏基础，只覆盖总人数、平均分、维度均值和 Top Students，尚未把班级弱项、分类/难度分布、训练闭环、竞赛迁移等教学评估项纳入导出。
- 相比之下，学生进度接口已经有 `by_category / by_difficulty`，说明数据维度不是完全缺失，而是还没进入班级评估导出主链路。

涉及位置：

- `code/backend/internal/module/teaching_readmodel/application/queries/class_insight_service.go`
- `code/backend/internal/module/assessment/application/commands/report_command_input.go`
- `code/backend/internal/dto/report.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/student_review_service.go`

影响：

- 当前更像教师实时看板和基础报告，不像论文里可以稳定导出的教学评估输出。
- 如果答辩时被追问“同一班级不同时间段的对比报告在哪里”，现在证据链会偏弱。

修复方向：

- 给班级报告和班级复盘补时间段参数，明确固定窗口和自定义窗口的 owner。
- 把分类/难度分布、班级共性弱项、训练闭环、竞赛迁移等结构化指标纳入导出。
- 明确导出结构与复盘结构之间的映射，避免页面看到的结论无法导出复用。

## Material Findings

- `P1` 推荐链路缺少真正的难度维度 owner。
- `P1` 竞赛数据没有在维度画像和推荐目标层统一回流。
- `P2` 班级与时间段统计口径、导出结构仍偏基础，不足以直接支撑开题中的教学评估输出论证。

## Senior Implementation Assessment

- 当前这轮优化的方向是对的，尤其是统一 `internal/teaching/advice`、收紧弱项阈值、统一 `solved / attempts / AWD success` 统计口径、减少模板化观察，这些都明显提高了推荐与复盘的可解释性。
- 但从更高一层看，现在的问题已经不是“文案还不够自然”，而是“能力画像模型与导出模型还差最后一层显式收口”。
- 更低风险的后续实现路径不是继续补更多文案分支，而是先收口三个 owner：
  - 难度带 owner
  - 竞赛数据回流到维度画像的 owner
  - 班级/时间段导出结构 owner

## Required Re-validation

- 当后续补完以上缺口后，需要重新验证：
  - 同一学生在训练样本与竞赛样本回流前后，推荐维度和推荐难度带是否会按预期变化。
  - 班级复盘与班级报告在自定义时间段下是否保持一致口径。
  - 导出产物是否能直接支撑“分类 + 难度 + 训练闭环 + 竞赛迁移”的教学评估展示。

## Residual Risk

- 本次 review 是静态核对，没有跑前端页面和真实导出文件，也没有用种子数据回放全部教师路径。
- 当前 `AWD` 数据回流设计在文档上已经有方向，但推荐链路和班级维度统计尚未完全跟进，后续实现如果只补其中一条链路，仍会出现论证断层。

## Touched Known-debt Status

- 本轮 review 触达的已知债务主要是“教学推荐/复盘离开题材料要求还有最后一层模型与导出收口”。
- 当前状态：`blocked by thesis-gap findings`
- 这批债务还不能降级为 residual risk；在后续继续触达教学推荐、班级评估导出或 AWD 回流表面时，应优先在 touched surface 内收口。
