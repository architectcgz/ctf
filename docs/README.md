# CTF 文档入口

## 当前设计入口

- `docs/文档规范.md`：文档写作、目录归属、命名和验证规范。
- `docs/contracts/README.md`：接口、OpenAPI 和题包契约入口。
- `docs/architecture/README.md`：最终架构与设计事实源入口。
- `docs/design/README.md`：设计文档读取顺序、过期规则和迁移说明。

## 过程资料入口

- `docs/plan/README.md`：实施计划入口；区分当前活动 plan 与历史归档 plan。
- `docs/operations/`：运行、演练、联调和部署说明；runtime-agent 相关说明见 `docs/operations/runtime-agent-deployment.md`。

## Workflow 治理入口

- `../scripts/check-task-intake.sh`：任务开始前的轻量 intake reminder，会顺手提示 `docs/todos/` 中尚未收口的事项。
- `../scripts/start-implementation.sh <topic-or-slug>`：非琐碎任务和受保护实现面的统一入口；负责创建 worktree、implementation plan 和本地 startup gate。
- `../scripts/check-review-governance.sh`：review / doctor 级治理审计入口；包含入口一致性、文档导航、OpenAPI 同步和 harness 接线检查，不再作为所有提交前的无条件门禁。
- `../scripts/check-consistency.sh`：兼容别名，内部转发到 `../scripts/check-review-governance.sh`。

## 读取原则

- 新增或修改文档前，先按 `docs/文档规范.md` 判断文档类型和目录归属。
- 先读当前索引，再读具体设计文档。
- 接口与字段契约统一从 `docs/contracts/README.md` 进入；OpenAPI 变更先改拆分源，再同步 bundle。
- 页面设计、设计系统和专题设计统一从 `docs/architecture/` 进入。
- 实施计划先从 `docs/plan/README.md` 判断是当前活动 plan 还是历史归档，不直接把旧 plan 当事实源。
- 非琐碎实现任务先从 `scripts/start-implementation.sh` 创建的 active plan 进入，复用与 owner 决策默认写在 plan 的 `## Files` 和 `## 复用与 Owner 决策` 中。
- `docs/reviews/` 是历史评审快照，不是当前设计事实源。
- `practice/` 中的过程资料和历史计划索引不覆盖后续代码和架构事实。

<!-- BEGIN HARNESS ENGINEERING: docs-navigation -->
## Harness 入口

- `../concepts/`：Harness 核心概念与 CTF 项目映射。
- `../thinking/`：项目 harness 边界、取舍和质疑。
- `../practice/`：初始化和实验记录。
- `../feedback/`：踩坑、修正和可复用经验。
- `../works/`：可展示模板、报告和说明。
- `../harness/prompts/`：仓库内 prompt 入口、局部补充，以及仍然项目专属的 prompt；共享正文位于 `/home/azhi/.agents/harness/prompts/`。
- `../references/`：外部文章、仓库和工具索引。
- `../scripts/check-review-governance.sh`：严格参考 harness 的一致性检查。
<!-- END HARNESS ENGINEERING: docs-navigation -->
