# 教学复盘训练数据扩量 Implementation Plan

## Objective

把现有 `code/backend/cmd/seed-teaching-review-data` 从“7 个手工样本场景”扩成“能覆盖大题库的训练数据生成器”，重点解决当前约 80 道题场景下：

- 推荐题池过窄，难以判断推荐是否真的贴合弱项
- 复盘样本覆盖面不足，班级复盘和个人归档很容易只落到少数固定分支
- 命令输出缺少题库覆盖信息，无法快速判断这批数据到底用了多少题、还能推荐多少题

目标不是改推荐/复盘算法，而是给现有算法补一批更适合评估的输入数据。

## Non-goals

- 不修改 `assessment`、`teaching_readmodel`、`teaching/advice` 的业务规则
- 不新增 HTTP API、数据库 schema 或新的 seed 命令
- 不把样本数据从专用班级 `信安2401` 扩散到其他测试班级
- 不引入新的长期事实源目录；本次只补 plan / reuse evidence

## Inputs

- `docs/plan/impl-plan/2026-05-08-teaching-review-seed-data-implementation-plan.md`
- `docs/reviews/backend/2026-05-08-teaching-review-seed-data-review.md`
- `docs/reviews/backend/2026-05-09-teaching-review-seed-data-enrichment-review.md`
- `code/backend/cmd/seed-teaching-review-data/main.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/{class_insight_service.go,student_review_service.go}`

## Current Baseline

- 命令当前固定写死 7 个学生样本场景，练习会话和题目索引也基本是手工指定
- 这批样本能验证“有数据时不是空态”，但无法覆盖 80 题级别的推荐候选池
- 班级复盘与个人归档已经能产出结果，但题库覆盖范围与推荐可选余量都不可见
- 现有 AWD 迁移样本只有 1 场、1 个 round，且只写入 `teams` 与 `awd_attack_logs`
- 教师 AWD 复盘 archive 依赖的 `team_members`、`awd_team_services`、`awd_traffic_events` 目前没有对应 seed，导致服务运行、流量审计和成员画像层的数据仍然偏空
- 之前 review 已明确：当公开题目变多时，现有样本量不足以判断推荐分布和复盘质量

## Chosen Direction

沿现有 `seed-teaching-review-data` 扩展，而不是新建并行命令：

1. 保留现有 7 个手工 archetype 样本，继续承担“稳定闭环 / 高试错 / 低活跃 / AWD 迁移”等明确叙事
2. 当已发布题目总量达到“大题库阈值”时，基于当前题库自动生成一批扩展覆盖学生
3. 自动样本按类别生成弱项学生和支撑训练会话，复用现有 `instance / audit_logs / submissions / writeups / skill_profiles` 写入链路
4. 命令输出补充题库覆盖摘要，明确：
   - 已发布题目总数
   - 本次训练样本实际触达的题目数
   - 各类别覆盖情况
   - 有推荐结果的学生数与推荐结果去重情况
5. 保留“一人一场 AWD contest”的现有主流程，但把 AWD 迁移样本扩成可复盘的完整证据集：
   - 单场 contest 内补齐多 round、多队伍
   - 每个 round 补齐 `team_members`、`awd_team_services`、`awd_attack_logs`、`awd_traffic_events`
   - 让教师 AWD 复盘页的 rounds / teams / services / attacks / traffic 都有可判断的样本

## Ownership Boundary

- `cmd/seed-teaching-review-data`
  - 负责：如何生成训练样本、如何组织样本规模、如何补齐 AWD 复盘证据、如何打印覆盖摘要
  - 不负责：改变推荐理由规则、改变复盘归档结构、改变教师查询算法
- `assessment` / `teaching_readmodel`
  - 负责：消费已有事实并生成推荐、归档和班级复盘
  - 不负责：为了这次补量任务引入新的专用分支或测试后门

## Change Surface

- Add: `docs/plan/impl-plan/2026-05-13-teaching-review-training-data-expansion-implementation-plan.md`
- Add: `.harness/reuse-decisions/teaching-review-training-data-expansion.md`
- Add: `code/backend/cmd/seed-teaching-review-data/main_test.go`
- Modify: `code/backend/cmd/seed-teaching-review-data/main.go`

## Task Slices

### Slice 1: 建立自动扩量规则与覆盖约束

目标：

- 题库较小时保留现有基线行为
- 题库达到阈值时自动补量
- 自动样本覆盖六个能力维度，不再只集中在少数题目

Validation:

- `cd code/backend && go test ./cmd/seed-teaching-review-data -run 'TestBuildCoverageStudentScenarios|TestBuildStudentScenarios' -count=1 -timeout 120s`

Review focus:

- 扩量规则是否仍然复用现有样本写入链路，而不是再造一套数据模型
- 自动样本是否既扩大覆盖面，又不会把每个学生的推荐池全部解空

### Slice 2: 接入命令主流程并补输出摘要

目标：

- `seedTeachingReviewData` 使用新的场景构建逻辑
- 输出中能直接看到题库覆盖和推荐可观察性

Validation:

- `cd code/backend && go test ./cmd/seed-teaching-review-data ./internal/module/assessment/... ./internal/module/teaching_readmodel/... -count=1 -timeout 120s`
- `cd code/backend && go run ./cmd/seed-teaching-review-data`

Review focus:

- 输出摘要是否真的能帮助判断推荐/复盘质量，而不是只堆计数
- 新增扩量逻辑是否仍保持命令幂等边界和专用班级隔离

### Slice 3: 补齐 AWD 复盘证据层 seed

目标：

- 把现有 AWD 样本从“只有攻击记录”补成“教师复盘 archive 可直接消费”的完整样本
- 至少覆盖 3 支队伍、3 个 round，并让每个 round 都有服务状态、攻击记录和流量事件
- 队伍成员数不再恒为 0，教师复盘分析区的 `member_count / service_count / attack_count / traffic_count` 都能读到真实值

Validation:

- `cd code/backend && go test ./cmd/seed-teaching-review-data -run 'TestBuildBaseStudentScenariosIncludeRichAWDReviewData|TestSeedStudentAWDScenarioBuildsTeacherReviewArchiveEvidence' -count=1 -timeout 120s`
- `cd code/backend && go test ./internal/module/assessment/application/queries -run TestTeacherAWDReviewService -count=1 -timeout 120s`

Review focus:

- AWD seed 是否继续复用现有 teacher review 查询链路，而不是为了种子数据改查询逻辑
- 新增 evidence 是否语义自洽：服务状态、攻击成功/失败、流量路径与 round 时间线能对应上
- 补齐 `awd_team_services` 后，清理逻辑是否也同步删除，避免 seed 命令重复运行留下脏数据

## Integration Checks

- 当题库足够大时，样本学生数明显高于原始 7 人基线
- 六个维度都至少有扩展样本参与
- 学生推荐仍然能返回未解题，而不是因为样本解题过多导致推荐为空
- 班级复盘和个人归档仍然能基于真实事件生成可读结果
- AWD 教师复盘 archive 在选择 round 后，队伍、服务、攻击、流量四个区块都不是空数组

## Risks

- 如果自动样本把某些学生在弱项维度上解题铺得过满，推荐会被误伤成空
- 如果扩量完全依赖手工标签而不带训练痕迹，复盘质量仍然不可判断
- 如果输出摘要只统计总量、不暴露类别覆盖，用户仍然无法判断“80 题是否真的被用起来”
- 如果 AWD 样本只补攻击日志、不补服务和流量，教师复盘页仍会停留在“能打开但判断信息不够”

## Verification Plan

1. `cd code/backend && go test ./cmd/seed-teaching-review-data -count=1 -timeout 120s`
2. `cd code/backend && go test ./internal/module/assessment/... ./internal/module/teaching_readmodel/... -count=1 -timeout 120s`
3. `cd code/backend && go run ./cmd/seed-teaching-review-data`
4. 如果本地 dev 数据库可用，再用命令输出核对题库覆盖、推荐学生数和班级复盘摘要
5. `bash scripts/check-consistency.sh`

## Architecture-Fit Evaluation

- reuse point 明确：继续复用现有 `seed-teaching-review-data`、`createSession`、`seedStudentAWDScenario`、推荐与归档查询链路
- owner 明确：只扩训练数据输入，不侵入 recommendation / report / teaching readmodel 规则层
- 这刀同时解决“样本规模”“可观测性”和 “AWD 证据完整度” 三个问题：不仅补数据，还补对覆盖质量与教师复盘质量的直接反馈
