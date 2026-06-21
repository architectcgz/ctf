# feedback/ — 反馈记录

实践中的踩坑、修正、迭代心得。把失败变成可复用经验。

## 入口规则

- 本项目只使用 `feedback/` 记录 agent 工作流、review、prompt、skill、policy 和项目协作方式相关反馈。
- `feedback/` 本身不是项目行为入口；只有当结论被明确沉到 `AGENTS.md`、`harness/`、脚本、hook、事实文档或全局 skill / harness 后，才算真正进入项目或全局规则面。
- 历史 `docs/improvements/` 条目迁入 `feedback/` 后，不再新增 `docs/improvements/`。
- 如果外部 skill 仍提示写入 `docs/improvements/`，在本项目内以本文件规则覆盖。
- 可执行的业务任务、需求拆分或待办不写入 `feedback/`，按性质进入 `docs/plan/` 或 `docs/todo/`。

## 文件约定

- 文件名：`{日期}-{简述}.md`
- 结构：问题描述 → 原因分析 → 解决方案 → 收获 → 沉淀状态
- 写正文时先透过现象抽象问题类别，再把本次事故作为反例或示例；不要把一次 review 现象直接写成只适用于本次的规则。
- `沉淀状态` 必须说明状态、owner 和链接，避免 feedback 只积累、不归位。
- 如果反馈导致 `harness/prompts/`、concepts、脚本、AGENTS 或 skill 更新，必须交叉链接。
- 新增或修改的 `feedback/*.md` 会被 `scripts/check-workflow-governance.sh` 检查是否包含 `## 沉淀状态`。

## 沉淀状态字段

推荐格式：

```text
## 沉淀状态

- 状态：已沉淀 / 仅项目保留 / 待同步 skill / 已机械化 / archived / 已废弃
- Owner：frontend-engineer skill / scripts/check-workflow-governance.sh / ctf AGENTS.md / .harness/reuse-index/index.yaml
- 链接：...
```

归属判断：

- 只是事故记录、背景信息或一次性复盘：留在 `feedback/`，不默认外溢到项目行为。
- 只影响 CTF 项目事实或路径：留在项目 `AGENTS.md`、项目 `harness/` 或对应文档。
- 能机械化检查：沉到 `harness/checks/`、`scripts/check-*.sh` 或 hook。
- 项目复用模式：沉到 `.harness/reuse-index/index.yaml` 或对应源码镜像目录下的 `README.md`。
- 项目 prompt 入口和项目特化补充：沉到 `harness/prompts/`；跨项目共享正文：沉到 `/home/azhi/.agents/harness/prompts/`。
- 跨项目通用方法、坏做法或检查清单：沉到对应全局 skill。
- 已经完成吸收但仍希望保留原始复盘上下文：把文件切到 `archived` 状态，明确 owner 与链接，不再把它当成活动反馈。

## 已迁移积累

- `improvements-index.md`：历史 improvement 条目迁移索引。
- `2026-05-03-independent-review-loop.md`：实现后独立 review 闭环。
- `2026-05-03-directory-contract-needs-static-check.md`：目录契约变更联动静态校验。
- `2026-05-06-feedback-replaces-docs-improvements.md`：feedback 取代 docs/improvements。
- `2026-05-09-do-not-bend-tests-to-fit-broken-architecture.md`：测试不能迁就坏架构，测试失败时先回到 owner / contract / 架构分析。
- `2026-05-10-top-tab-panels-should-not-repeat-eyebrow.md`：带顶部 tab 的页面，tab 面板内默认不再重复渲染分区 eyebrow。
- `2026-05-10-workspace-grid-only-for-real-layout.md`：`workspace-grid` 只用于真实多列或区域布局，普通单列页面直接使用 `content-pane`。
- `2026-05-10-workspace-page-header-should-be-shared.md`：工作区首屏标题区默认使用共享 `workspace-page-header`，不要继续新增页面局部 `workspace-hero` 布局和分隔线。
- `2026-05-10-awd-topology-local-readiness.md`：AWD topology 题本地验题默认使用 `healthcheck + service_healthy + up --wait`，不要把容器 running 误当成服务 ready。
- `2026-05-10-student-button-dark-mode-token-bridge.md`：学生侧按钮的 primary / secondary / outline 语义必须落到共享按钮变体，并在 light / dark 下保留可见边框。
- `2026-05-10-error-pages-use-ui-btn.md`：错误页、空状态和恢复动作按钮默认使用通用 `ui-btn`，页面不再私有实现按钮 hover / dark mode。
- `2026-05-23-frontend-review-snapshots-can-be-pruned-after-audit-absorption.md`：前端 review 单轮快照在主索引吸收后应及时清理，避免旧“未修复”状态继续污染活动目录。
- `2026-06-04-frontend-state-surface-owner-and-ui-test-layering.md`：前端 loading / empty / loaded 状态 surface owner 应收口到同一视觉区域，同时 UI 测试需要从大量源码字符串断言转向分层覆盖。
- `2026-06-21-constructors-should-not-hide-external-effects.md`：构造函数不隐藏 hostname / env / file / network / time / random / client dial 这类外部依赖读取或副作用，外部事实由 bootstrap / composition / 显式 factory owner 读取并处理错误。

已沉淀到全局 skill、全局 AGENTS 或项目机械检查中的旧反馈，不应继续以活动状态停留在本目录；默认切到 `archived` 状态保留上下文，需要追溯更早版本时再使用 Git 历史。
