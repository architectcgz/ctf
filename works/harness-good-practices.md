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
- `harness/prompts/` 只放已经验证过、会复用的项目级 agent 工作流 prompt。
- `references/` 只放外部资料索引和研究入口。

如果内容是“当前产品怎么工作”，优先回到 `docs/` 事实源，而不是塞进 harness 目录。

## 4. 好做法：稳定后及时回收

- 中间设计一旦稳定，架构结论要回收到 `docs/architecture/`。
- 字段、接口、格式一旦稳定，契约要回收到 `docs/contracts/`。
- 页面结构和视觉方向一旦稳定，页面最终稿要回收到 `docs/architecture/frontend/pages/` 或 `docs/architecture/frontend/design-system/`。
- 单题设计如果已经落地到题包，不再额外复制一份到 `features/`；需要说明时只保留题包入口或迁移索引。
- 原中间稿不要继续冒充事实源，应该在原位置写明 `Superseded by ...`。

## 4.1 好做法：共享 UI 原语要收口到单一视觉 owner

- 当页面已经接入共享 UI 原语，例如 `header-btn`、`ui-btn`、`ui-badge` 或 `workspace-directory-status-pill`，页面组件默认只能增加布局类，不能继续覆盖共享原语的颜色、边框、hover、focus、尺寸变量。
- 如果确实存在可复用视觉差异，应在共享样式层新增明确 variant，例如 `header-btn--danger` 或 `ui-btn--compact`，而不是在单个页面里写 `.xxx-button { --header-btn-* }`。
- header 操作区统一使用 `header-actions` + `header-btn`，表格行操作、弹窗 footer、空态按钮继续使用 `ui-btn`；不要因为“都是按钮”把不同交互位置混成一个原语。
- 统一样式任务的完成标准不是“class 名已经换掉”，而是旧局部视觉入口也已经删除，并有静态测试防止重新出现。
- 测试应该断言共享原语边界，例如“不允许 feature 组件覆盖 `--header-btn-*`”，不要继续断言页面私有变量存在。

## 5. 好做法：旧入口停用后不要偷偷复活

当前仓库已经把一部分旧入口迁出：

- `docs/improvements/` -> `feedback/`
- `docs/superpowers/` 的历史计划/过程 -> `practice/`
- `docs/refs/` -> `references/`
- `docs/skills/` -> 全局 skill；只有仍需作为项目 prompt 复用的内容才进入 `harness/prompts/`

后续如果再把同类内容写回旧目录，会同时制造两个入口，review 和后续 agent 都会误判。

## 6. 常见坏味道

- 把 `方案比较`、`竞品分析`、`实施顺序` 长期留在 `docs/architecture/features/`。
- 先写一篇“迁移收口稿”，后面又不把最终结论并回正式架构文档。
- 把某一道具体题目的题包设计、题面推导或解法说明伪装成 `features/` 架构专题。
- 把 implementation checklist 当成架构文档保存。
- 把 review 结论当作当前设计事实源。
- 把仍未决定的方案写进 `contracts` 或 `architecture`。
- 把待办和 backlog 写进 `feedback/`。
- 同一 input contract 的 `normalize / default / validate` 在 `application` 和 `repository` 两层重复出现，最后靠双重兜底维持“看起来能工作”。
- repository 直接接收未收敛的裸字符串排序键、分页键或筛选键，再在仓储层补默认值和白名单，而不是由上游先收口成受限语义。
- application 虽然已经成为唯一 owner，但内部 contract 仍继续暴露可手工拼装的导出 enum / struct，最后只能靠 repository 的 panic 或 defensive branch 晚发现错误。
- 测试先发现了 contract / owner / 架构问题，但实现阶段为了让 CI 变绿，反过来修改测试期待值、fixture、mock 或页面文案，让测试去迁就当前错误实现。
- 分页或汇总语义明明已经切到后端真实总量，页面却继续拿当前页 `list` 现算指标；测试再跟着改成只断当前页数字，等于把错误语义固化。
- 共享 UI 原语已经接入后，页面仍保留 `challenge-manage-import-button`、`challenge-import-action` 这类私有 class，并通过 `--header-btn-*`、`--ui-btn-*` 覆盖视觉变量；这会形成“看起来复用了共享组件，实际视觉仍然分叉”的假统一。
- 为了修一个按钮不一致，只删除当前按钮覆盖，却不扫描同类原语的其它局部覆盖，也不补静态防线；这会让同一问题在下一个页面继续出现。
- 为了“先放一下”创建新的临时 docs 入口，结果后面没人回收。

## 7. 当前推荐动作

- 新建文档前先看 `AGENTS.md` 里的 `File Placement Rules`。
- 如果内容要跨多个目录，先决定哪一份是事实源，其他文件只保留索引或 `Superseded by ...`。
- 如果目录契约变更会影响 agent 导航或发现路径，同步更新 `works/` 索引和 `scripts/check-consistency.sh`。
- 如果测试一改就绿、但实现理由说不清，先停下来重查 owner、contract 和真实用户语义；不要把“测试通过”当成问题已经解决的证据。
