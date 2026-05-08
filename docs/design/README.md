# 设计文档索引

## 读取顺序

后续判断“当前设计是什么”时，按下面顺序读取，不要直接从旧设计或 review 里下结论。

1. `docs/design/README.md`：当前事实源入口和过期规则。
2. `docs/architecture/README.md`：最终架构与设计事实源入口。
3. `docs/contracts/`、`docs/architecture/backend/`、`docs/architecture/frontend/`：接口契约、架构边界、运行时事实。
4. `docs/architecture/frontend/`：UI 风格、页面结构、组件体系和间距规则。
5. `docs/architecture/frontend/pages/`：页面参考稿、截图和可复用设计样本。
6. `docs/architecture/features/`：专题最终设计。
7. `docs/superpowers/plans/`：执行计划和实施记录，只作为当时实现上下文。
8. `docs/reviews/`：历史评审快照，不作为当前设计事实源。

## 当前事实源

- 架构与设计总入口：`docs/architecture/README.md`
- 全局 UI 设计：`docs/architecture/frontend/01-architecture-overview.md` 到 `docs/architecture/frontend/09-spacing-system.md`
- 页面级参考稿：`docs/architecture/frontend/pages/`
- 专题设计索引：`docs/architecture/features/专题架构索引.md`
- 前端 review 当前索引：`docs/reviews/frontend/README.md`
- 后端架构设计索引：`docs/architecture/backend/design/README.md`

## 当前中间方案

- `AWD题目配置面板方案.md`
  - 说明：后台 AWD service 配置面板的中间设计稿，仍包含方案比较和待落地交互。
- `AWD能力画像回流方案.md`
  - 说明：AWD 个人攻击证据接入能力画像与推荐链路的中间方案，仍包含数据归因和事件设计取舍。
- `教学复盘建议优化方案.md`
  - 说明：围绕毕设主线整理“教学证据 -> 能力画像补充信号 -> 推荐理由 -> 教学复盘建议”的中间方案，当前仍包含范围取舍与实现阶段拆分。

## 专题归属

- 拓扑编排与环境模板：以 `docs/architecture/frontend/` 中的前端事实源和对应功能专题为入口；历史页面草稿不再作为活动事实源。
- AWD 运行态与服务模型：以 `docs/architecture/features/校园级CTF-AWD模式完整设计.md` 和 `docs/architecture/backend/design/awd-engine-migration.md` 为准。
- 社区题解与推荐题解：以 `docs/architecture/features/社区题解与推荐题解设计.md` 为准。
- 攻防证据链、判题模式、赛事运营增强：以 `docs/architecture/features/攻击证据链与教学复盘架构.md`、`docs/architecture/features/判题模式扩展架构.md`、`docs/architecture/features/赛事导出与复盘归档架构.md` 为专题入口。

## 过期规则

- 如果新设计文档明确写了“替代 / 不再 / 已由”，旧文档应移除或在索引中标记为历史。
- 如果只有 review 提到旧问题，不能直接把它当成当前设计；必须回到当前代码、当前事实源和最近索引复核。
- 如果 `docs/architecture/frontend/pages/` 中的页面稿与当前代码明显冲突，先判断页面稿是否仍是目标设计；确认仍有效时改代码，确认已过期时更新或移除页面稿。
- `docs/superpowers/plans/` 里的命令、文件清单和阶段状态只代表当时实施计划，不能覆盖后续代码事实。

## 已移除的旧设计

- `docs/superpowers/specs/2026-04-01-contestant-writeup-workflow-design.md`
  - 移除原因：其中“教师评阅学生 writeup”的部分已由 `社区题解与推荐题解设计.md` 替代，当前产品方向改为“社区题解 + 推荐题解”，教师/管理员角色从批改者变为内容运营者。
- `docs/superpowers/specs/*.md`
  - 迁移原因：仍有效的专题设计已经统一迁入 `docs/architecture/features/`，`docs/superpowers` 不再保存最终设计副本。
