# 探索性计划目录

## 定位

本目录用于存放探索性、临时性、验证性的 plan 文档，不是正式实施计划的归档位置。

- 负责：快速起草的技术探索、原型验证、临时调研计划。
- 不负责：结构性改动、跨模块重构、需要正式 review 和 task gate 的实施计划。

## 与 `impl-plan/` 的区别

| 维度 | `exploratory/` | `impl-plan/` |
|------|----------------|--------------|
| 用途 | 探索性、临时验证 | 正式实施计划 |
| 生命周期 | 短期，完成后可删除 | 长期追溯，需要归档 |
| 完整性 | 可以不完整，只记录思路 | 必须结构完整，包含验证步骤 |
| Review | 不需要正式 review | 需要 plan review 和独立 review |
| Task Gate | 不绑定 task slug / startup gate | 必须绑定 task slug / startup gate |
| 归档 | 移到 `../archive/exploratory/` 或直接删除 | 移到 `../archive/impl-plan/` |

## 适用场景

**应该放在这里**：
- 技术选型调研计划
- 快速原型验证思路
- 探索性重构方向
- 临时问题诊断计划
- "试试看"类型的小实验

**应该放在 `impl-plan/`**：
- 模块边界重构
- API 契约迁移
- 测试架构重组
- 核心业务流程改造
- 任何需要正式 code-workflow 的改动

## 命名规则

```
YYYY-MM-DD-<简短描述>.md
```

示例：
- `2026-06-12-redis-cache-strategy-exploration.md`
- `2026-06-12-frontend-state-refactor-prototype.md`

## 归档规则

执行完成后：
1. 如果结论有价值，提取到对应事实源（`docs/architecture/`、`docs/reviews/`、`feedback/`）
2. plan 本身移到 `../archive/exploratory/` 或直接删除
3. 不进入 `../archive/impl-plan/` 的正式归档流程

## 读取协议

- 本目录中的 plan 不是当前架构事实源
- 不能作为"已经这样设计"的依据
- 只能作为"曾经探索过这个方向"的历史参考
