# Findings

- `notification` 是 `system` 剩余能力里最重的一块，但其依赖边界是清晰的：HTTP、websocket manager、ws ticket、practice 事件总线。
- 这条链路不需要继续留在 `system`，可以直接并入 `ops`，且外部路径无需变化。
- 一旦通知迁走，后端代码中的 `system` owner 逻辑就可以整体删除。
