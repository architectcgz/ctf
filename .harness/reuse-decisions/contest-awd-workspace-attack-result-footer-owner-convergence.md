# Reuse Decision

## Change type
component

## Existing code searched
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
- `code/frontend/src/components/contests/awd/AWDAttackToolbar.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## Similar implementations found
- `code/frontend/src/components/contests/awd/AWDAttackTargetGrid.vue`
- `code/frontend/src/components/contests/awd/AWDWorkspaceIntelColumn.vue`
- `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
- `code/frontend/src/components/common/AppAlert.vue`

## Decision
refactor_existing

## Reason
`ContestAWDWorkspacePanel.vue` 的中区目前只剩攻击结果 footer 还是纯展示块。它只消费 `submitResult` 和父组件已经格式化好的消息，不拥有提交动作、异步状态或路由逻辑，适合继续按“父组件保留 workflow owner，子组件承接展示壳”的模式抽出，作为进入左侧服务编排壳前的最后一刀低风险切片。

## Files to modify
- `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
- `code/frontend/src/components/contests/awd/AWDAttackResultFooter.vue`
- `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
- `docs/reviews/frontend/README.md`
- `docs/reviews/frontend/ctf-frontend-audit-20260422.md`

## After implementation
- 如果后续继续拆左侧服务编排壳，再补 `.harness/reuse-index/` 镜像索引，记录“page workflow owner + footer display child” 模式。
