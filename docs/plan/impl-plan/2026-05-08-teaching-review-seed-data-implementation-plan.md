# 教学复盘样本数据 Implementation Plan

## Objective

为本地 `ctf` 开发库补一组可重复执行的“教师教学复盘”样本数据，覆盖：

- 新的教师账号与独立班级
- 多名具有真实姓名/学号风格的学生
- 练习实例、访问日志、代理请求、提交记录、复盘材料
- 稳定可控的 `skill_profiles`，用于驱动班级复盘结论与个人推荐题

目标是让教师侧班级复盘、学生分析页和复盘归档在有数据时能展示出可评估的建议，而不是只出现空态或被随机弱项主导。

## Non-goals

- 不修改教师复盘算法本身
- 不引入新的业务 API
- 不重构现有 assessment / teaching_readmodel 边界
- 不把这批样本数据混入现有 `CTF-1` 演示班级，避免旧脏数据干扰评估

## Inputs

- `docs/architecture/features/教学复盘优化设计.md`
- `code/backend/internal/module/assessment/application/commands/report_service.go`
- `code/backend/internal/module/assessment/application/queries/recommendation_service.go`
- `code/backend/internal/module/teaching_readmodel/application/queries/service.go`
- 当前开发库中已发布题目分布（`crypto/forensics/misc/pwn/reverse/web` 各至少 1 题，`crypto/web` 各 2 题）

## Change Surface

- 新增一个可重复执行的后端 seed 命令
- 通过命令向开发库写入：
  - `users`
  - `user_roles`
  - `instances`
  - `audit_logs`
  - `submissions`
  - `submission_writeups`
  - `skill_profiles`
- 运行命令后实际把样本数据种入本地开发库

## Data / Contract Notes

- 班级复盘结论依赖近 7 天活跃度、最近事件数、已解题数、`skill_profiles` 弱项与推荐题。
- 个人复盘归档依赖实例生命周期、访问/代理日志、练习提交、writeup 与能力画像。
- 当前开发库题量太稀疏，若完全依赖自动画像计算，多个维度会长期同分为 `0`，导致 `weak_dimension` 退化成按维度字典序随机命中。
- 因此本次样本数据会“真实写训练事件 + 显式写 skill_profiles”组合，保证复盘建议稳定且可解释。

## Task Slices

### Slice 1: 建立独立样本班级与账号

- 创建独立教师账号与班级，例如 `信安2401`
- 创建 6 名学生账号，使用真实风格姓名、学号、邮箱与默认密码
- 保证命令可重跑，重复执行不会生成重复用户

Files / modules:

- `code/backend/cmd/seed-teaching-review-data/main.go`

Validation:

- 运行 seed 命令
- 查询库内班级/账号是否已落库

Review focus:

- 用户 upsert 是否幂等
- 是否只影响专用样本班级

### Slice 2: 写入训练事件与复盘材料

- 为学生写入近 7 天内分布的实例、访问日志、代理请求、提交记录、writeup
- 保证不同学生呈现不同训练画像：闭环完成、反复试错、低活跃、只做实操未复盘等
- 保证推荐题目类别仍存在未解已发布题

Files / modules:

- `code/backend/cmd/seed-teaching-review-data/main.go`

Validation:

- 运行 seed 命令
- 查询 `instances/audit_logs/submissions/submission_writeups` 计数

Review focus:

- 时间分布是否落在近 7 天内
- 是否覆盖复盘归档依赖的核心事件类型

### Slice 3: 稳定驱动建议的画像与缓存清理

- 为每名学生 upsert 明确的 `skill_profiles`
- 清理推荐缓存，避免复跑后读取旧结果
- 在命令输出中打印每名学生的弱项维度与推荐题摘要，便于肉眼校验

Files / modules:

- `code/backend/cmd/seed-teaching-review-data/main.go`

Validation:

- 运行 seed 命令
- 命令输出应展示每名学生的弱项和推荐题

Review focus:

- 画像与事件是否具备基本可解释性
- 推荐题是否与弱项一致且不是已解题

## Verification Plan

1. `go test ./internal/module/assessment/... ./internal/module/teaching_readmodel/...`
2. `go test ./cmd/seed-teaching-review-data`
3. `go run ./cmd/seed-teaching-review-data`
4. 使用 `docker exec ctf-postgres psql ...` 核对新增班级、学生、事件和复盘材料数量

## Rollback

- 删除样本班级 `信安2401` 下的教师和学生账号
- 同步删除这些用户关联的 `user_roles / instances / audit_logs / submissions / submission_writeups / skill_profiles`
- 本次命令会将数据收口在专用班级内，回退时无需影响现有 `CTF-1` 或其他测试班级
