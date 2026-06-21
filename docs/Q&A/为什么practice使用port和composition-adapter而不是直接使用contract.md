# Q&A：为什么 practice 使用 port 和 composition adapter，而不是直接使用 contract？

本文记录 AWD runtime node placement 实现中的一次边界判断：`practice` 为什么定义自己的 port，并由 `app/composition` 用 adapter 接入 `contest` 与 `container_runtime`，而不是直接使用其他模块提供的 contract。

关联文档：

- `ctf/docs/architecture/backend/01-system-architecture.md`
- `ctf/docs/architecture/backend/03-container-architecture.md`
- `ctf/docs/architecture/backend/05-key-flows.md`
- `ctf/docs/plan/impl-plan/2026-06-21-awd-runtime-node-identity-and-placement-implementation-plan.md`

---

## 1. 先给结论

这里使用 port + composition adapter 是对的，因为问题本质不是“读取某个模块暴露的数据结构”，而是“`practice` 的实例生命周期需要一个可调用能力”。

`practice` 需要表达的是：

> 启动实例时，给定 `InstanceScope`，选择一个平台 `runtime_node_id`。如果是 AWD scope，必须复用 contest runtime placement；绑定 node 不可用时等待，不得 fallback 到默认 selector。

这个能力横跨两个 provider：

- `contest`：拥有 `contest_runtime_placements` 的持久化事实。
- `container_runtime`：拥有 runtime node registry、默认 selector 和健康查询。

决策入口却在 `practice` 的实例生命周期里。因此它不是某个 provider 天然能单独暴露的 contract，而是 app composition 组合出来的跨模块能力。

---

## 2. contract 和 port 在这里分别适合表达什么

### contract 适合 provider 暴露稳定事实

`contracts` 更适合表达 provider 对外稳定承诺，例如：

- 某个跨模块 DTO 的字段。
- 某个错误类型或状态枚举。
- 某个 provider 愿意长期暴露的数据事实。

它回答的是：

> 我这个模块能稳定告诉外部什么？

例如“contest runtime placement 记录长什么样”，可以是 contest 对外 contract 的候选。

### port 适合 consumer 声明用例需要的能力

`ports` 更适合由 consumer 定义自己用例需要什么能力，例如：

- `practice` 启动实例时需要 `RuntimeNodeSelector`。
- `practice` 的 AWD scope 需要 `AWDRuntimePlacementStore`。
- `practice` 需要按 `runtime_node_id` 做健康查询。

它回答的是：

> 我这个用例为了完成业务决策，需要外部提供什么能力？

这次 `practice` 不是只要一个 placement DTO，而是要一个“选择并校验 runtime node”的能力。

---

## 3. 为什么 practice 不直接依赖其他模块 contract

如果让 `practice` 直接使用 `contest` 和 `container_runtime` 的 contract，仍然解决不了核心问题：

1. contract 通常只描述数据形状，不描述跨模块决策顺序。
2. AWD placement 规则不是单个 provider 的事实，而是组合规则。
3. `practice.InstanceScope` 是 practice 的生命周期语义，`contest` 和 `container_runtime` 不应该反向理解它。

这次真正需要锁住的是行为：

1. 非 AWD scope 走默认 runtime node selector。
2. AWD scope 已有 placement 时，只校验绑定的 `runtime_node_id` 是否健康。
3. AWD scope 已有 placement 但 node 不健康时，返回 `ErrRuntimeNodeUnavailable`，不 fallback。
4. AWD scope 没有 placement 时，首次使用默认 selector 选 node，然后持久化 active placement。
5. 并发 ensure 发生唯一冲突时，以 repository 返回的 active placement 为准。
6. AWD scope 的 placement 依赖没有接好时，返回 `ErrRuntimeNodeUnavailable`，不静默 fallback。

这些是 `practice` 实例生命周期需要的能力语义，不是某个 provider contract 单独能承载的内容。

---

## 4. 为什么 adapter 放在 app/composition

`app/composition` 是跨模块装配层，它可以同时看见多个 provider 的具体实现。

当前接线是：

- `practice/ports.RuntimeNodeSelector`
  - 由 `app/composition.practiceRuntimeNodeSelectorAdapter` 实现。
- `practice/ports.AWDRuntimePlacementStore`
  - 由 `app/composition.contestRuntimePlacementStoreAdapter` 接到 `contest/infrastructure.ContestRuntimePlacementRepository`。
- `practice/ports.RuntimeNodeHealthLookup`
  - 由 `app/composition.runtimeNodeHealthLookupAdapter` 接到 `container_runtime/infrastructure.RuntimeNodeRepository`。

这样做的效果是：

- `practice` 不 import `contest/infrastructure`。
- `practice` 不 import `container_runtime/infrastructure`。
- `contest` 不需要理解 `practice.InstanceScope`。
- `container_runtime` 不需要知道 AWD contest placement 规则。
- 跨模块组合策略集中在 composition 层，符合当前模块化单体的装配边界。

---

## 5. 如果不用 adapter，会出现什么问题

### 方案 A：practice 直接 import contest repository

问题：

- `practice` 会知道 `contest_runtime_placements` 的持久化细节。
- `contest` 的 repository 结构变化会直接影响 practice。
- practice 生命周期逻辑会混入 contest persistence owner。

判断：不合适。

### 方案 B：practice 直接 import container_runtime repository

问题：

- `practice` 会知道 runtime node 健康查询的 repository 实现。
- `schedulable`、`health_status`、`last_seen_at` 等 container_runtime 细节会泄漏进 practice。
- 后续 runtime node 健康策略变化会扩大改动面。

判断：不合适。

### 方案 C：contest 暴露一个“给 practice 用”的高层 contract

问题：

- contest 必须理解 `practice.InstanceScope` 或 practice 的实例生命周期。
- contest 会被迫知道何时选择默认 runtime node、何时回 pending、何时不 fallback。
- 这会让 provider 反向依赖 consumer 用例。

判断：不合适。

### 方案 D：container_runtime 暴露 AWD-aware selector

问题：

- container_runtime 会知道 contest runtime placement。
- 底层 runtime module 被迫理解 AWD 业务策略。
- `container_runtime` 的职责会从“运行节点与执行通道”膨胀到“比赛 placement 策略”。

判断：不合适。

---

## 6. 当前边界判断

| 关注点 | Owner |
| --- | --- |
| runtime node registry / health / default selector | `container_runtime` |
| contest-level AWD placement persistence | `contest` |
| instance lifecycle / pending scheduler / desired reconcile | `practice` |
| per-instance runtime attempt identity | `instance` |
| 跨模块接线与 provider -> consumer port 适配 | `app/composition` |

因此，`practice` 定义自己需要的 port，`app/composition` 负责把 `contest` 和 `container_runtime` 的具体能力适配进去。

---

## 7. 代码落点

关键代码：

- `code/backend/internal/module/practice/ports/ports.go`
  - `RuntimeNodeSelector`
  - `AWDRuntimePlacementStore`
  - `RuntimeNodeHealthLookup`
- `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter.go`
  - 将 `InstanceScope` 映射到默认 selector 或 AWD placement 选择逻辑。
- `code/backend/internal/app/composition/contest_runtime_placement_adapter.go`
  - 将 contest placement repository 适配为 practice 需要的 placement store。
- `code/backend/internal/module/contest/infrastructure/contest_runtime_placement_repository.go`
  - contest-owned persistence。

关键测试：

- `code/backend/internal/app/composition/instance_practice_runtime_node_selector_adapter_test.go`
  - 已有 placement 不 fallback。
  - placement node 不健康不 fallback。
  - 无 placement 首次创建 active placement。
  - AWD placement 依赖缺失不 fallback。
  - 非 AWD scope 继续走默认 selector。
- `code/backend/internal/module/practice/application/commands/instance_provisioning_test.go`
  - pending AWD instance 会把完整 AWD scope 传给 selector。
- `code/backend/internal/module/practice/application/commands/awd_desired_runtime_reconciler_test.go`
  - desired reconcile 创建 AWD instance 时持久化 selector 返回的 `runtime_node_id`。

---

## 8. 复盘时的判断准则

以后遇到类似问题，可以按这几句判断：

- 如果是 provider 对外稳定数据形状，优先考虑 `contracts`。
- 如果是 consumer 用例需要的可调用能力，优先由 consumer 定义 `ports`。
- 如果这个能力组合了多个 provider，放在 `app/composition` 适配。
- 如果某个底层模块开始理解上层业务 scope，通常说明边界反了。
- 如果某个业务模块开始 import 其他模块 `infrastructure`，通常说明缺了 port 或 adapter。

这次选择 port + composition adapter 的核心理由是：AWD runtime placement 是 `practice` 用例消费的跨模块能力，不是 `contest` 或 `container_runtime` 单独提供的数据契约。
