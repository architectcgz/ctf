# 项目 Skill 到全局 Skill 的升级协议

## 触发条件

项目 skill 中的 pattern 满足以下**全部**条件时，考虑上收到全局 skill：

1. **已在至少 2 个项目中验证有效**
2. **不依赖特定项目结构**（不引用特定表名、目录、模块名）
3. **可以用通用语言描述**（不包含项目特定术语）
4. **解决的是通用问题**（不是项目历史包袱的 workaround）

## 升级流程

### Step 1: 标记候选

在项目 skill 中标记：

```markdown
### Package Split By Responsibility

（当前内容）

**升级候选**：此 pattern 已在 ctf 项目验证，待在其他项目复用后上收到全局。
```

### Step 2: 在新项目复用验证

在第二个项目中：
1. 按这个 pattern 执行
2. 记录是否需要调整
3. 如果完全适用 → 满足升级条件
4. 如果需要大量调整 → 说明还不够通用，继续留项目级

### Step 3: 泛化并上收

**在全局 skill 增加通用版本**：

```markdown
# ~/.agents/skills/backend-engineer/SKILL.md

## Refactoring Patterns

### Package Split By Responsibility

When a single file exceeds 800-1000 lines with mixed responsibilities:
- Split within the same package (avoid new package unless encapsulation needed)
- Structure: types / load / validate / defaults / domain-logic
- Write structure guard test before splitting
- Keep all exported APIs unchanged

**Origins**: Validated in ctf project (2026-06), reused in [project-2] (2026-XX)
```

**在项目 skill 降级为引用**：

```markdown
# ctf/.agents/skills/ctf-backend-patterns/SKILL.md

### Package Split By Responsibility

→ 此 pattern 已上收到全局 `backend-engineer` skill

CTF 项目特定上下文：
- 案例：internal/config 包拆分
- 详细模板：feedback/2026-06-12-package-split-by-responsibility-template.md
```

### Step 4: 更新 feedback 沉淀状态

```markdown
## 沉淀状态

- 状态：已上收到全局 skill
- Owner: ~/.agents/skills/backend-engineer/SKILL.md § Refactoring Patterns
- 升级时间：2026-XX-XX
- 验证项目：ctf, [project-2]
- 项目特定上下文：仍保留在 ctf/.agents/skills/ctf-backend-patterns/
```

## 反向路径：降级

如果全局 skill 中的 pattern 发现不够通用，可以降级：

1. 从全局 skill 移除
2. 移回各项目 skill
3. 标注"此前认为通用，但实际依赖项目上下文"

## 维护职责

**项目 skill owner**：
- 发现 pattern 可能通用时，标记"升级候选"
- 在新项目中主动尝试复用
- 收集调整成本

**全局 skill owner**：
- 收到升级请求时，检查是否满足 4 条标准
- 上收时去除项目特定细节
- 定期检查全局 pattern 是否仍然通用

## 示例：不应该上收的 Pattern

❌ **CTF AWD Topology 本地验题流程**
- 理由：依赖 CTF 项目的 AWD 题目结构
- 结论：留在 `ctf-backend-patterns`

❌ **前端 WorkspaceDirectoryToolbar 间距规则**
- 理由：特定于 CTF 项目的管理端 UI 组件
- 结论：留在 `ctf-frontend-patterns`

✅ **Package Split By Responsibility**
- 理由：纯语言层面的重构方法，不依赖 CTF 特定结构
- 结论：可以在第二个 Go 项目验证后上收

## 快速判断

**问自己**：如果把这个 pattern 给一个完全不了解 CTF 项目的人，他能直接用吗？

- ✅ 能 → 可能适合上收（还需多项目验证）
- ❌ 不能，需要解释 CTF 特定背景 → 留项目级
