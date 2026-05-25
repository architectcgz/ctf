# CTF 前端 Review 文档索引

## 当前事实源

- 当前前端专项审查主索引：`ctf-frontend-audit-20260422.md`
- 后续判断“是否仍需修复”时，以主索引的 `当前结论`、最近一轮修复进展、`后续技术债 Backlog` 为准。
- 早期 `ctf-frontend-code-review-*` 单轮快照的有效结论已经吸收到主索引和本文件，原始快照不再作为活动事实源保留；需要追溯原文时直接使用 Git 历史。

## 快速核查入口

- 当用户只是在问“前端技术债现在还在不在”时，默认先读：
  - `docs/reviews/frontend/README.md`
  - `docs/reviews/frontend/ctf-frontend-audit-20260422.md` 的 `后续技术债 Backlog`
  - `docs/reviews/architecture/2026-05-24-frontend-architecture-review.md`
- 默认先把结论分成三类：仍然成立、已经收口、历史 finding 已被后续代码覆盖，避免重新把旧 review 全量扫一遍。
- 需要最小代码证据时，优先跑这一组定向护栏，而不是做全量前端扫描：
  - `npm run test:run -- src/__tests__/architectureBoundaries.test.ts src/views/__tests__/routeViewArchitectureBoundary.test.ts src/views/__tests__/sharedThemeTokenAdoption.test.ts src/views/__tests__/duplicateActionGuardAudit.test.ts src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts src/views/teacher/__tests__/studentAnalysisPanelExtraction.test.ts`
- 2026-05-25 当前快速结论：
  - 仍成立：`TD-1` 超大组件专题拆分、`TD-3` 性能监控接入、`TD-4` i18n 预留、allowlist 驱动的边界结构债、请求层错误 owner 未完全回到页面层。
  - 已收口：`TD-2` 主题 token / Tailwind 任意值尾项、`TD-5` 历史 review 快照清理、`StudentInsightPanel.vue` 的当前 `TD-1` touched surface。
  - 需单独跟踪但不属于主索引既有 `TD-1/3/4/5`：`duplicateActionGuardAudit.test.ts` 当前暴露的图片管理页重复提交 owner 缺口。

## 已清理快照的回读方式

- 已删除快照中的行号、提交范围、测试数量和“未修复”状态，只代表对应审查日期的代码状态。
- 如果历史结论与 `ctf-frontend-audit-20260422.md` 冲突，以主索引为准；必要时再回到代码和测试验证。
- 需要追溯原始措辞、旧行号或审查上下文时，直接查看 Git 历史，不回流活动目录里的单轮快照。
- 已合并到主线的修复应继续优先写回主索引；低优先级体验建议若已被后续产品方向覆盖，应在主索引标记为“过期 / 已由后续实现覆盖”。

## 已吸收的过期结论

- 早期全量 review 中“页面组件大量占位未实现”“缺少单元测试”等结论已经过期；当前前端已有完整路由页面、组件拆分与全量测试门禁。
- 竞赛详情早期关于“题目选中状态未持久化”“弹窗关闭逻辑不完整”的结论，已经被后续 `ContestDetail` 的 query 同步与交互收口覆盖；若重新审查，应按当前组件结构复核。
- 实例管理早期关于“即将过期弹窗缺少 ESC / 关闭按钮 / ARIA 语义”“延时时长常量未使用”“状态色值硬编码”的结论，已经被后续实例页修复覆盖。
- 管理员题目管理早期剩余项以低优先级体验优化为主；如果重新处理，应先核对当前题目管理页面与共享目录组件状态。

## 仍然保留为当前 Backlog 的事项

- `TD-1` 超大组件专题拆分：拓扑页已按小切片持续收口，模板侧栏、摘要指标、状态说明、网络分段编辑区、节点编辑区、拓扑连线与链路策略编辑区、画布工作区、入口节点卡片、题包上下文区都已从 `ChallengeTopologyStudioPage.vue` 抽到独立组件；`StudentInsightPanel.vue` 当前 touched surface 已在 2026-05-25 切片里收口，历史 `AWDChallengeConfigDialog.vue` 也已退场。后续继续按父页面 owner 边界处理 `ChallengeTopologyStudioPage.vue` 剩余页面编排壳，以及 `ContestAWDWorkspacePanel.vue` 等高复杂度组件。
- `TD-3` 性能监控接入：需要先确定指标、上报端点、隐私边界和生产开关。
- `TD-4` i18n 预留：取决于产品是否需要多语言。

## 已完成的技术债

- `TD-2` 主题 token 与任意值尾项：第七十五轮已完成真实产品路径收口；`mock / reference / UI lab / ThemePreview / refs / SVG / 主题 token 定义文件` 不作为当前未完成项。后续用 `npm run check:theme-tail` 复核真实产品路径，该脚本不会把 Vue `#default` 插槽误判为十六进制色值。

## 维护约定

- 新增前端 review 时，先判断它是“当前事实源更新”还是“单次审查快照”。
- 当前事实源更新应写入 `ctf-frontend-audit-20260422.md` 或后续新的主索引文档。
- 单次审查快照只有在后续轮次尚未吸收、仍需独立作为活动证据时才临时保留；一旦主索引已吸收结论，就直接清理，不在活动目录长期堆积。
