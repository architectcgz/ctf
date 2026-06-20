# Feedback 快速沉淀协议

## 目标

feedback 写入后立即判断归属，让下次相关任务能看到。

## 沉淀决策树

```
新增 feedback
  ↓
质量够高？（有 pattern / checklist / anti-pattern）
  ↓ Yes
适用范围？
  ├─ 通用方法（跨项目验证） → 沉淀到全局 skill (~/.agents/skills/)
  ├─ 项目特定 pattern → 沉淀到项目 skill (.agents/skills/)
  ├─ 项目规则/约束 → 沉淀到项目 AGENTS.md 或 harness/prompts/
  └─ 可机械化 → 沉淀到 harness/checks/ 或 scripts/check-*.sh
  ↓ No
只是事故复盘 → 留在 feedback/，标记 "仅项目保留"
```

## Skill 沉淀规则

### 项目 skill vs 全局 skill

**优先沉淀到项目 skill**：
- 路径：`.agents/skills/ctf-backend-patterns/` 或 `.agents/skills/ctf-frontend-patterns/`
- 通过 `.claude/skills` 软链接被 Claude 发现
- 可以直接引用项目 feedback 路径
- 例：`feedback/2026-06-12-package-split-by-responsibility-template.md`

**上收到全局 skill** 的条件：
- Pattern 在多个项目验证有效
- 不依赖特定项目结构
- 可以泛化成通用方法
- 例：TDD 方法、测试分层策略

**格式**：
```markdown
## Known Patterns

### Pattern Name
Brief description (1-2 lines)
- Key point 1
- Key point 2
- Details: `project/feedback/YYYY-MM-DD-pattern-name.md`
```

**原则**：
- skill 里只放轻量索引（5-10 行）
- feedback 保留完整模板、案例和上下文
- 交叉链接：skill → feedback，feedback § 沉淀状态 → skill

### 适用 skill 映射

| Pattern 类型 | 项目 Skill | 全局 Skill (跨项目验证后) |
|------------|-----------|----------------------|
| 后端重构、分层、owner | `ctf-backend-patterns` | `backend-engineer` |
| 前端组件、状态、UI | `ctf-frontend-patterns` | `frontend-engineer` |
| 测试策略、分层、fixture | `ctf-test-patterns` | `test-engineer` |
| Review 方法、finding 分类 | `code-reviewer` (项目已有) | `code-reviewer` |
| 架构决策、边界、trade-off | 项目 AGENTS.md | `architect-agent` |
| Harness、workflow、检查 | `harness-engineering` (项目已有) | `harness-engineering` |

## 沉淀检查

### 手工检查

新增 feedback 时：
1. 是否包含 `## 沉淀状态` 章节？
2. 状态是否明确（已沉淀 / 仅项目保留 / 待同步 / archived）？
3. 如果状态是"已沉淀"，是否有 Owner 和链接？

### 机械化检查（待实施）

```bash
# scripts/check-feedback-settlement.sh
# 检查最近 7 天新增的 feedback 是否包含沉淀状态
```

## 示例

### ✅ Good - 立即沉淀到项目 skill

```markdown
# feedback/2026-06-12-package-split-by-responsibility-template.md

## 沉淀状态

- 状态：已沉淀到项目 skill
- Owner: .agents/skills/ctf-backend-patterns/SKILL.md § Refactoring Patterns
- 更新时间：2026-06-12
- 说明：在项目 backend patterns skill 增加轻量索引；本 feedback 保留详细模板
```

对应在 `.agents/skills/ctf-backend-patterns/SKILL.md` 增加：
```markdown
## Refactoring Patterns

### Package Split By Responsibility
单个文件超过 800-1000 行且职责混合时的拆分方法：
- 同包内拆分，不创建新 package
- 按职责拆分：types / load / validate / defaults / domain-logic
- 详细模板：feedback/2026-06-12-package-split-by-responsibility-template.md
```

### ❌ Bad - 只写不沉淀

```markdown
# feedback/2026-06-12-some-pattern.md

（没有沉淀状态章节）
```

→ 下次相关任务，agent 不知道这条 feedback 存在

### ⚠️ Acceptable - 明确不沉淀

```markdown
## 沉淀状态

- 状态：仅项目保留
- 说明：一次性数据库迁移事故复盘，不泛化成通用 pattern
```

## 迁移计划

**Phase 1（已完成）**：
- ✅ 在 `backend-engineer` skill 增加 "Known Patterns" 章节
- ✅ 将 package split pattern 沉淀为首个示例
- ✅ 编写本协议文档

**Phase 2（按需）**：
- 扫描 `feedback/` 中质量高的条目，判断是否应该沉淀
- 在对应 skill 中增加轻量索引
- 更新 feedback 的沉淀状态

**Phase 3（如果 feedback 继续增长）**：
- 建立 `.harness/feedback-index.yaml` 分类索引
- 在 `scripts/check-task-intake.sh` 集成关键词检索
- 机械化检查新增 feedback 是否包含沉淀状态

## 维护责任

- **写 feedback 时**：立即判断沉淀归属，不拖到"以后再说"
- **更新 skill 时**：同步更新 feedback 的沉淀状态链接
- **review 时**：检查 feedback 是否有沉淀状态章节
