# ops / assessment 事件与读模型边界复核与收口方案

> 状态：Draft
> 事实源：`code/backend/internal/module/ops`、`internal/module/assessment` 当前 import 事实、`docs/todos/2026-05-17-project-tech-debt-from-migrations.md` P2、`docs/design/backend-module-boundary-target.md`
> 替代：无

## 定位与一条重要更正

本方案原本针对"ops / assessment 事件化边界未收口"。**深入根因后需要更正这个判断**：

- `ops -> contest / challenge / practice` 这几条 baseline 边，经核查引用的全部是**事件类型**（`contestcontracts.EventAnnouncementCreated` / `AnnouncementCreatedEvent` / `ScoreboardUpdatedEvent` / `AWDPreviewProgressEvent`、`practicecontracts.EventFlagAccepted` / `FlagAcceptedEvent`、`challengecontracts.EventPublishCheckFinished` / `PublishCheckFinishedEvent`）。ops 已经是**事件消费者**（`contest_realtime_service.go` 做 WebSocket relay、`notification_service.go` 发通知），不是硬依赖业务模块的 service / repository。这在目标版图里本就是合规的事件消费模式。
- `assessment -> contest / challenge / identity / practice` 主要是**同步读各 owner 的数据契约**（`AWDAttackLog` / `Submission` / `Contest` / `User` / `AWDChallenge` 等）来生成画像、推荐、报告、复盘——这是分析产物 owner 的本性，类似 `teaching_query` 的跨 owner 只读聚合。

结论：**第二梯队在"模块边界"层面基本已经收口**，不存在第一梯队那种方向反了的边。真正剩余的是两个更轻、性质不同的项，优先级明显低于第一梯队。本方案如实记录，不把已收口的事件消费包装成"待解耦"。

## 项 1：事件契约 owner 是否中性化（评估，推荐暂不改）

- 现状：ops 为拿事件类型定义，import 了 contest / challenge / practice 三个模块的 `contracts`。
- 评估：事件消费方依赖**事件发布方的 contract** 是合理的——事件类型本就应放在 owner 模块的 `contracts`（boundary-target 明确"事件类型应放在 owner 模块的 contracts 或稳定事件包"）。这不是耦合异味。
- 推荐：**保持现状**。仅在 baseline / 文档里写明这几条边是"事件消费边"，与硬依赖区分。
- 仅当未来出现"多个消费者 + 需要独立演进事件契约"时，再评估把事实事件抽到中性事件包；当前抽离收益低于新增一个跨模块契约包的成本。

## 项 2：assessment 画像 practice / contest 语义统一（实质项，P2 数据语义）

这是第二梯队真正值得做的项，但它是**数据 / 读模型语义工程**，不是模块边界重排。

- 现状（来自 P2 技术债与代码符号）：个人 recommendation 与班级 class review 的 teaching fact snapshot 已吸收 AWD 成功覆盖、profile score 补充信号与 solved difficulty 覆盖，`difficulty_band` 已进入推荐候选查询；但推荐与复盘主链路里的 `attempt / approved review evidence` 仍主要沿用 `contest_id IS NULL` 的练习侧语义。
- 问题：当前画像不是一份完全统一的训练画像——练习侧事实与竞赛 / AWD 侧事实在主链路里仍未完全对齐。
- 目标：让推荐与复盘主链路的 attempt / evidence 语义同时覆盖练习与竞赛 / AWD 事实，形成统一训练画像。
- 依据文档：`docs/reviews/architecture/2026-05-14-teaching-review-thesis-gap-review.md`、`docs/architecture/features/教学复盘建议生成架构.md`。

### 方向框架（详细切片待细化）

1. 盘点 assessment 画像 / 推荐 / 复盘主链路里仍以 `contest_id IS NULL` 表达"练习侧"的查询与事实读取点。
2. 定义统一的"训练事实"读取口径，让 attempt / approved review evidence 同时纳入练习与竞赛 / AWD 来源。
3. 优先经事件 / 既有维度事实表收口，不在提交计分同步写路径上叠加画像更新。
4. 保留教师可见复盘结果的兼容，迁移期对旧口径与新口径做一致性核对。

说明：项 2 的可执行切片需要先深入读 `assessment` 画像 / 推荐 / 复盘的具体生成逻辑与上述架构文档，再拆成 reviewable 切片。本方案先确立方向与边界，不预先锁定实现细节。

## 优先级与建议

- 本梯队整体优先级**低于第一梯队**（`2026-06-10-module-reverse-dependency-convergence-plan.md`）。
- 项 1 推荐暂不动，只补 baseline / 文档的"事件消费边"说明。
- 项 2 是真实剩余债，但属数据语义工程；建议在第一梯队与 runtime 残余拆分推进后，单独安排一轮深入再细化切片。
- 若需要，项 2 可直接沿用并细化 `docs/todos/2026-05-17-project-tech-debt-from-migrations.md` 的对应 P2 条目，不必另起炉灶。

## 验证计划

- 项 1（若最终选择中性化）：`go test ./internal/module -run TestModuleDependencyBaselineIsCurrent`、`bash scripts/check-backend-architecture.sh --full`。
- 项 2：assessment 模块测试 + 画像 / 推荐 / 复盘回归 + 新旧口径一致性核对；具体命令随切片细化补充。

## 完成判定

- 项 1：baseline 中 `ops -> *` 事件消费边有明确"事件消费"标注（或在确有收益时迁移到中性事件包并消除边）。
- 项 2：推荐与复盘主链路的训练事实统一覆盖练习与竞赛 / AWD，形成一致训练画像；P2 技术债条目收口。
