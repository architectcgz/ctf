# AWD 报告与复盘归档口径对齐设计

phase11 已经把 AWD 学生成功攻击接入能力画像与推荐，但报告与复盘归档还停留在旧口径：

- `report_repository.go` 的个人统计、班级统计仍只看 `submissions.contest_id IS NULL`
- 学生复盘时间线只有练习实例与练习提交，没有 AWD 个人攻击事件
- 学生证据链只有实例访问、代理流量、练习提交，没有 AWD 攻击日志
- 复盘摘要里的 `CorrectSubmissionCount` 和教学观察也只识别普通 `flag_submit` / `challenge_submission`

这会导致同一个学生在 phase11 之后出现三套互相不一致的数据视图：

1. 能力画像已经认定“该题在 AWD 里证明过能力”
2. 个人报告和班级排名仍显示“没解过”
3. 教师复盘页看不到对应的 AWD 攻击过程

这和毕业设计课题里“系统自动记录攻击步骤与漏洞利用过程，实时反馈得分”“基于实训数据生成能力画像，推荐针对性靶场练习，导出实训报告供教学复盘”的要求仍然有一段断层。

## 非目标

- 不改教师 AWD 复盘目录与比赛级归档结构
- 不把队伍级防守分映射进个人报告
- 不重做教师学生复盘页 UI
- 不引入新的报表快照表或离线统计任务

## 当前问题

### 个人与班级统计口径滞后

以下查询仍只统计普通练习：

- `GetPersonalStats`
- `ListPersonalDimensionStats`
- `GetClassAverageScore`
- `ListClassTopStudents`

结果是：

- `TotalSolved / TotalScore / Rank` 漏掉 AWD 个人成功攻击
- 班级平均分与 Top 学生榜不会反映 AWD 实战结果
- 画像、推荐、报告三者不再共用同一事实源

### 复盘过程缺少 AWD 个人证据

以下查询还没有接入 `awd_attack_logs`：

- `GetStudentTimeline`
- `GetStudentEvidence`

结果是：

- 教师打开学生复盘页时，能看到练习路径，却看不到学生在 AWD 里真实打过哪些题、攻击了哪支队伍、是否成功、拿了多少分
- 导出的复盘归档也无法满足“记录攻击步骤与漏洞利用过程”的要求

### 摘要与观察没有识别 AWD 事件

`report_service.go` 里的：

- `countCorrectSubmissions`
- `hasRepeatedWrongSubmissions`
- `hasHandsOnExploit`

当前只识别练习事件类型。即使 phase12 把 AWD 事件写进时间线和证据链，摘要统计与教师观察仍会漏算。

## 方案比较

### 方案 A：分别在每个 SQL 查询里内联补 AWD 条件

优点：

- 改动直接
- 代码最少

缺点：

- 个人 solved 口径会在多个查询里重复
- 后续 phase13+ 再扩展时容易继续漂移
- 复盘事件与统计查询很难保证长期一致

不采用。

### 方案 B：在 `report_repository` 内统一“个人已证明题目”和“个人 AWD 攻击事件”两类事实源，再让各查询复用

做法：

- 个人 solved 集合统一为：
  - 普通练习正确提交
  - `awd_attack_logs` 中 `submitted_by_user_id = 当前用户`、`source = submission`、`is_success = true`、`score_gained > 0` 的成功攻击
- 个人 attempts 集合统一为：
  - 普通练习提交
  - 个人 AWD 攻击提交日志
- 时间线和证据链把 AWD 攻击日志作为新的个人事件流接入
- 复盘摘要与观察同步识别新增 AWD 事件类型

优点：

- 和 phase11 使用同一事实源
- 后端改动集中，前端契约基本保持稳定
- 能直接满足“训练过程记录 + 评估闭环 + 报告复盘”三段要求

缺点：

- `report_repository.go` 会新增一些公共 SQL 片段或辅助查询约定
- 需要补一层仓库集成测试，避免多查询口径再次分叉

推荐采用。

### 方案 C：新增报表中间表，先把 AWD 数据离线汇总后再查

优点：

- 查询层可读性更高

缺点：

- 会引入额外同步链路与回填成本
- 现阶段报表量级不值得为此增加复杂度

不采用。

## 目标

本轮完成后：

1. 个人报告里的 `Solved / Score / Rank` 与 phase11 的画像事实源一致
2. 班级平均分和 Top 学生榜会计入个人 AWD 成功攻击
3. 维度统计会把练习正确题与 AWD 成功攻击题按 `challenge_id` 去重后合并
4. 学生复盘时间线能看到 AWD 攻击提交事件
5. 学生证据链能看到 AWD 攻击日志证据
6. 复盘摘要与教学观察能识别 AWD 成功/失败攻击，不再只看练习提交流

## 总体设计

phase12 继续沿用 phase11 的个人归因边界：

- 只统计 `submitted_by_user_id` 可明确归因到学生本人的 AWD 攻击日志
- 只把 `source = submission` 的学生真实提交纳入个人画像/报告事实源
- 只有 `is_success = true AND score_gained > 0` 的 AWD 记录进入 solved / score / rank

在此基础上把报表拆成两条统一事实流：

### 1. 个人已证明题目事实流

用途：

- `GetPersonalStats`
- `ListPersonalDimensionStats`
- `GetClassAverageScore`
- `ListClassTopStudents`

口径：

- 普通练习正确提交
- 个人 AWD 成功攻击

合并规则：

- 按 `user_id + challenge_id` 去重
- 再关联 `challenges.points`

这样能保证：

- 同一题在练习和 AWD 都成功时只算一次
- 同一题在 AWD 里多次命中不同目标队伍时也只算一次
- 个人报告、班级统计、画像、推荐的 solved 集合保持一致

### 2. 个人攻击活动事实流

用途：

- `GetPersonalStats` 中的 `total_attempts`
- `GetStudentTimeline`
- `GetStudentEvidence`

口径：

- 普通练习提交事件
- `awd_attack_logs` 中由学生本人发起的提交日志，不区分成功或失败

这样能保证：

- “总提交”能覆盖实际攻击尝试
- 复盘页能完整反映学生的攻防行为轨迹
- 教师可以从证据链直接回看 AWD 攻击结果

## 查询与事件设计

### 个人统计

`GetPersonalStats` 调整为：

- `solved` 使用统一 solved 集合
- `user_scores` 对全体用户使用统一 solved 集合算总分，再产出 rank
- `total_attempts` 改为 `普通练习提交数 + 个人 AWD 攻击提交日志数`

### 维度统计

`ListPersonalDimensionStats` 改为：

- 先得到该用户 solved challenge 集合
- 再按 `challenges.category` 聚合 solved / total

### 班级统计

`GetClassAverageScore` 与 `ListClassTopStudents` 改为：

- 对班级内学生使用统一 solved 集合算个人总分
- 平均分与排行榜都基于相同的用户总分结果

`ListClassDimensionAverages` 不需要调整，因为它读取的 `skill_profiles` 已在 phase11 接入 AWD 事实源。

### 复盘时间线

`GetStudentTimeline` 新增 AWD 事件：

- `type = awd_attack_submit`
- `is_correct = awd_attack_logs.is_success`
- `points = score_gained`，失败时为空或 0
- `detail` 包含目标队伍与得分结果

排序仍保持全局时间倒序。

### 复盘证据链

`GetStudentEvidence` 新增 AWD 证据：

- `type = awd_attack_submission`
- `detail` 描述目标队伍、成功/失败、得分
- `meta` 至少带上：
  - `event_stage = exploit`
  - `is_success`
  - `score_gained`
  - `round_id`
  - `victim_team_id`
  - `victim_team_name`

排序继续按时间正序，便于教师顺着过程阅读。

### 摘要与教学观察

`report_service.go` 同步识别新增事件类型：

- `CorrectSubmissionCount`
  - 统计 `flag_submit` 与 `awd_attack_submit` 的成功事件
  - 证据回退统计时也识别 `challenge_submission` 与 `awd_attack_submission`
- `hasRepeatedWrongSubmissions`
  - 识别普通提交失败与 AWD 攻击失败
- `hasHandsOnExploit`
  - 识别 `instance_access`、`instance_proxy_request`，以及 `awd_attack_submission`

这样教师看到的摘要、观察、证据链三者就不会互相打架。

## 边界与取舍

### 为什么 solved 只纳入成功且有分的 AWD 攻击

这是和 phase11 保持一致的核心边界：

- 同队重复提交已知 flag 时，`score_gained` 可能为 0
- 这种行为不能再当作新的能力证明

因此报告里的 solved / score / rank 仍然只接收 `score_gained > 0` 的成功日志。

### 为什么 attempts 与证据链要保留失败攻击

复盘归档的目标不只是统计结果，还要呈现训练过程。

学生在 AWD 中发起过但失败的攻击：

- 不应该计入 solved
- 但应该进入提交次数、时间线与证据链

否则老师看不到真实尝试过程，毕业设计里“记录攻击步骤与漏洞利用过程”的要求就仍然不成立。

### 为什么这轮不改前端契约结构

现有教师复盘页：

- 时间线只依赖 `title / created_at / detail / type`
- 证据链只依赖 `title / timestamp / detail / type / meta`

phase12 只是在现有数组里追加 AWD 事件，不需要前端改字段结构，也不会破坏已上线页面风格。

## 测试策略

至少覆盖：

1. `GetPersonalStats`
   - AWD 成功攻击进入 `total_score / total_solved / rank`
   - AWD 攻击日志进入 `total_attempts`
2. `ListPersonalDimensionStats`
   - 练习与 AWD 同题去重后 solved 不重复累计
3. `GetClassAverageScore` / `ListClassTopStudents`
   - 班级汇总使用统一 solved 集合
4. `GetStudentTimeline`
   - 返回 AWD 攻击事件，且能区分成功/失败
5. `GetStudentEvidence`
   - 返回 AWD 攻击证据及对应 meta
6. `buildReviewArchiveSummary` / 教学观察
   - AWD 事件会影响 `CorrectSubmissionCount`、错误提交识别和实操判断

## 与毕业设计要求的对应

- “系统自动记录攻击步骤与漏洞利用过程，实时反馈得分”
  - 学生复盘时间线与证据链将直接纳入 AWD 攻击日志与得分结果
- “基于实训数据生成能力画像，推荐针对性靶场练习”
  - 报告统计口径与 phase11 画像事实源完全对齐
- “导出实训报告供教学复盘”
  - 导出的个人报告与学生复盘归档能同时反映练习与 AWD 实战数据

## 结论

phase12 的目标不是再做一个新的 AWD 页面，而是把 phase11 已经建立的个人事实源彻底贯穿到报告与复盘归档。这样“攻防演练”“技能评估”“教学复盘”三块数据才真正共用同一套个人证据。
