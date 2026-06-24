# Package Split By Responsibility - Good Taste Template

日期：2026-06-12  
触发：review `task/2026-06-12-backend-config-package-split` 发现这是一次教科书级的同包分文件重构

## Pattern

当单个文件过长且混合多种职责时，按职责拆分成同包内的多个文件，而不是创建新 package。

## Why This Worked

### 1. 保持 package boundary 不变

- 所有文件仍然是 `package config`
- 对外 API 完全不变：`config.Load()`, `config.Config`, `config.PostgresConfig.DSN()` 等
- 调用方无需修改任何 import 或类型名
- 避免了广泛的 package 迁移风险

### 2. 职责拆分清晰

```
types.go               → 所有类型定义和导出的方法
load.go                → Load() 函数和配置加载逻辑
defaults.go            → setDefaults() 和所有默认值
validate.go            → Validate() 方法和所有校验辅助函数
container_flag_secret.go → container flag secret 持久化与解析
config.go              → 只保留一行注释说明拆分意图
```

### 3. 完整性验证到位

**实施前：**
- 新增结构护栏测试 `config_structure_test.go`，明确要求必须存在这 5 个职责文件
- 先运行测试确认失败（红）

**实施后：**
- 运行结构护栏测试确认通过（绿）
- 运行原有包测试确认行为不变
- `go build ./internal/config` 确认编译通过

### 4. 同步更新文档引用

所有活动文档中对 `internal/config/config.go` 的事实源引用，统一改为 `internal/config/` 目录。

## Checklist for Future Package Splits

- [ ] 只在同包内拆文件，不创建新 package（除非有明确的 encapsulation 需求）
- [ ] 先写结构护栏测试，要求新的文件布局存在
- [ ] 按职责拆分：types / load / validate / defaults / domain-specific-logic
- [ ] 保持所有导出 API 不变
- [ ] 保持所有 package-level 变量和函数
- [ ] 运行包测试确认行为等价
- [ ] 运行 `go build` 确认编译通过
- [ ] 同步更新活动文档中的事实源引用
- [ ] 如果原文件还有剩余逻辑，保留；如果完全空了，只留一行注释说明拆分意图

## Anti-patterns to Avoid

❌ **不要为了拆而拆成新 package**  
→ 会引入新的 import cycle 风险、visibility 问题和调用方迁移成本

❌ **不要只拆一半就停**  
→ 要么不拆，要么拆透，避免出现"主文件 + 一堆 `_helper.go`"的尴尬局面

❌ **不要拆完不写结构护栏测试**  
→ 没有测试锁住新布局，后续改动会退回单文件聚集

❌ **不要忘记同步文档**  
→ 文档仍然指向 `config.go` 单文件会误导后续维护

## Example Commit Structure

```
refactor(backend): 拆分 internal/config 同包职责文件

- 新增 types.go: 所有类型定义
- 新增 load.go: Load() 函数
- 新增 defaults.go: setDefaults() 和默认值
- 新增 validate.go: Validate() 和校验逻辑
- 新增 container_flag_secret.go: flag secret 持久化
- 新增 config_structure_test.go: 结构护栏测试
- 修改 config.go: 只保留包说明注释
- 更新活动文档事实源引用

对外 API 与行为保持完全不变。

验证：
- go test ./internal/config -count=1
- go build ./internal/config
- bash scripts/run-workflow-stage.sh completion-full
- bash scripts/run-workflow-stage.sh workflow-governance
```

## Related

- 这次重构属于 `非琐碎任务`，走了完整的 code-workflow：plan → implement → validate → review
- implementation plan: `docs/plan/archive/impl-plan/2026-06/2026-06-12-backend-config-package-split-implementation-plan.md`
- 没有新增或删除任何配置字段，纯结构性收口
- 触达了已知结构债（单文件过长），在本次切片内完整收口，而不是给 follow-up

## When to Apply

- 单个文件超过 800-1000 行且职责混合
- 多人频繁触碰同一文件导致 merge conflict
- 明确可以按 load / validate / types / defaults / domain-logic 等维度拆分
- **不适用**：如果拆分后需要引入循环依赖，说明应该考虑 package 级重构而不是文件级拆分

## 沉淀状态

- 状态：archived
- Owner: `.agents/skills/ctf-backend-patterns/SKILL.md` § Refactoring Patterns
- 更新时间：2026-06-12
- 说明：在项目 backend patterns skill 增加轻量索引；本 feedback 保留详细模板和案例上下文
