# 设计文档索引

## 读取顺序

后续判断“当前设计是什么”时，按下面顺序读取，不要直接从旧设计或 review 里下结论。

1. `docs/design/README.md`：当前事实源入口和过期规则。
2. `docs/architecture/README.md`：最终架构与设计事实源入口。
3. `docs/contracts/`、`docs/architecture/backend/`、`docs/architecture/frontend/`：接口契约、架构边界、运行时事实。
4. `docs/architecture/frontend/design-system/` 与 `docs/architecture/frontend/pages/`：UI 风格和页面结构设计。
5. `docs/architecture/features/`：专题最终设计。
6. `docs/superpowers/specs/`：阶段性专题草稿，只用于追溯设计演进。
7. `docs/superpowers/plans/`：执行计划和实施记录，只作为当时实现上下文。
6. `docs/reviews/`：历史评审快照，不作为当前设计事实源。

## 当前事实源

- 架构与设计总入口：`docs/architecture/README.md`
- 全局 UI 设计：`docs/architecture/frontend/design-system/MASTER.md`
- 页面级 UI 设计：`docs/architecture/frontend/pages/README.md`
- 专题设计索引：`docs/architecture/features/README.md`
- 前端 review 当前索引：`docs/reviews/frontend/README.md`
- 后端架构设计索引：`docs/architecture/backend/design/README.md`

## 专题归属

- 拓扑编排与环境模板：以 `docs/architecture/frontend/pages/topology-editor.md` 和 `docs/architecture/frontend/pages/env-templates.md` 为 UI 设计入口。`docs/superpowers` 当前没有独立的 topology editor 专题最终设计。
- AWD 运行态与服务模型：以 `docs/architecture/features/2026-04-17-awd-final-design.md` 和 `docs/architecture/backend/design/awd-engine-migration.md` 为准。
- 社区题解与推荐题解：以 `docs/architecture/features/2026-04-04-community-writeup-and-recommended-solutions-design.md` 为准。
- 攻防证据链、判题模式、赛事运营增强：以对应 `docs/architecture/features/2026-04-01-*.md` 为专题入口。

## 过期规则

- 如果新设计文档明确写了“替代 / 不再 / 已由”，旧文档应移除或在索引中标记为历史。
- 如果只有 review 提到旧问题，不能直接把它当成当前设计；必须回到当前代码、当前事实源和最近索引复核。
- 如果 `docs/architecture/frontend/pages/` 中的页面稿与当前代码明显冲突，先判断页面稿是否仍是目标设计；确认仍有效时改代码，确认已过期时更新或移除页面稿。
- `docs/superpowers/specs/` 只代表阶段草稿，不能直接覆盖 `docs/architecture/features/` 的最终设计。
- `docs/superpowers/plans/` 里的命令、文件清单和阶段状态只代表当时实施计划，不能覆盖后续代码事实。

## 已移除的旧设计

- `docs/superpowers/specs/2026-04-01-contestant-writeup-workflow-design.md`
  - 移除原因：其中“教师评阅学生 writeup”的部分已由 `2026-04-04-community-writeup-and-recommended-solutions-design.md` 替代，当前产品方向改为“社区题解 + 推荐题解”，教师/管理员角色从批改者变为内容运营者。
