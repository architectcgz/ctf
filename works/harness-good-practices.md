# CTF Harness Good Practices

这份文档整理 `ctf` 仓库里已经证明有效的 harness 归档做法，重点不是介绍目录本身，而是说明什么内容该落到哪里，避免“有文档但找不到事实源”。

## 1. 先分清四类内容

- 最终事实：已经采用并且后续实现、review、维护都应服从的结论。
- 中间设计：还在比较方案、梳理页面、推导边界的草稿。
- 实施过程：准备怎么做、按什么阶段落地、验证怎么跑。
- 评审与积累：已经做完后的 review 证据、踩坑反馈、长期规则和可复用说明。

一个文件如果先天混了这四类内容，后面一定会漂。

## 2. 好做法：按“内容类型”选目录

- 最终架构和最终设计放 `docs/architecture/`。
- 面向产品能力或业务专题的最终结论放 `docs/architecture/features/`；如果专题里包含已经固定的内部边界结论，直接并回 owning doc，不要额外保留迁移/收口专题。
- 单题题包设计、题面设计和解法设计不要塞进 `docs/architecture/`；已落地题目直接以题包目录内的 `challenge.yml`、`statement.md`、源码和测试为事实源。
- 接口、事件、题包格式和结构契约放 `docs/contracts/`。
- 还在推演的设计稿放 `docs/design/`。
- 结构性实现方案和阶段计划放 `docs/plan/impl-plan/`。
- review 结论和 findings 放 `docs/reviews/`。
- 需求基线、范围和 gap 分析放 `docs/requirements/`。
- backlog、延期项和明确未收口事项放 `docs/todos/`。
- 运维说明、联调步骤和演练记录放 `docs/operations/`。
- 阶段报告、差距报告和综合分析放 `docs/reports/`。
- 会被反复引用的问答式解释放 `docs/Q&A/`。
- 论文、周报、开题和文献材料留在 `docs/thesis/`、`docs/weekly-reports/`、`docs/开题报告/`、`docs/文献/`、`docs/毕业设计文档相关/`。

## 3. 好做法：让 harness 目录只装“长期有用的积累”

- `concepts/` 只写项目级概念和长期原则。
- `thinking/` 只写还没有变成规则或实现的判断和取舍。
- `practice/` 只写实验记录、迁移过程、历史计划索引和实践说明。
- `feedback/` 只写 agent 工作流、review、prompt、policy、协作方式的踩坑与修正。
- `works/` 只写可复用输出，例如地图、模板、教程、good practices。
- `prompts/` 只放已经验证过的提示词和工作流。
- `references/` 只放外部资料索引和研究入口。

如果内容是“当前产品怎么工作”，优先回到 `docs/` 事实源，而不是塞进 harness 目录。

## 4. 好做法：稳定后及时回收

- 中间设计一旦稳定，架构结论要回收到 `docs/architecture/`。
- 字段、接口、格式一旦稳定，契约要回收到 `docs/contracts/`。
- 页面结构和视觉方向一旦稳定，页面最终稿要回收到 `docs/architecture/frontend/pages/` 或 `docs/architecture/frontend/design-system/`。
- 单题设计如果已经落地到题包，不再额外复制一份到 `features/`；需要说明时只保留题包入口或迁移索引。
- 原中间稿不要继续冒充事实源，应该在原位置写明 `Superseded by ...`。

## 5. 好做法：旧入口停用后不要偷偷复活

当前仓库已经把一部分旧入口迁出：

- `docs/improvements/` -> `feedback/`
- `docs/superpowers/` 的历史计划/过程 -> `practice/`
- `docs/refs/` -> `references/`
- `docs/skills/` -> `prompts/`

后续如果再把同类内容写回旧目录，会同时制造两个入口，review 和后续 agent 都会误判。

## 6. 常见坏味道

- 把 `方案比较`、`竞品分析`、`实施顺序` 长期留在 `docs/architecture/features/`。
- 先写一篇“迁移收口稿”，后面又不把最终结论并回正式架构文档。
- 把某一道具体题目的题包设计、题面推导或解法说明伪装成 `features/` 架构专题。
- 把 implementation checklist 当成架构文档保存。
- 把 review 结论当作当前设计事实源。
- 把仍未决定的方案写进 `contracts` 或 `architecture`。
- 把待办和 backlog 写进 `feedback/`。
- 为了“先放一下”创建新的临时 docs 入口，结果后面没人回收。

## 7. 当前推荐动作

- 新建文档前先看 `AGENTS.md` 里的 `File Placement Rules`。
- 如果内容要跨多个目录，先决定哪一份是事实源，其他文件只保留索引或 `Superseded by ...`。
- 如果目录契约变更会影响 agent 导航或发现路径，同步更新 `works/` 索引和 `scripts/check-consistency.sh`。
