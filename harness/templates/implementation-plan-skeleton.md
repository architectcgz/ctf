<!-- Managed by code-workflow package (version: 2026-06-12.1) -->
# __TASK_TITLE__ Implementation Plan

> 填写要求：保留本模板中的英文标题、状态枚举、机器字段和占位符；除代码、命令、路径、报错、协议字段、枚举值、外部专有名词外，所有说明性内容默认用中文填写。

**Goal:** 用一句中文说明本任务要达成的结果。

**Architecture:** 用 2-3 句中文说明实现边界、主要 owner 和不改变的行为。

**Tech Stack:** 列出相关技术栈；技术名、框架名和命令保持原文。

---

## Task Metadata

- Task Slug: `__TASK_SLUG__`
- Started At: `__STARTED_AT__`
- Worktree: `__WORKTREE_PATH__`
- Branch: `__BRANCH_NAME__`
- Plan Type: `slice` <!-- slice | roadmap -->

## Plan Status

- Status: `draft` <!-- draft | ready-for-implementation | implemented | review-pending | review-passed | archived -->
- Coding may start only after:
  - [ ] Intake analysis gate completed
  - [ ] Plan review / architecture-fit check completed
  - [ ] Execution slices and validation plan filled

## Objective And Non-Goals

- Objective: 用中文说明本任务的目标。
- Non-Goals: 用中文列出本任务明确不做的范围。

## Problem Statement

- Current behavior / structure: 用中文说明当前行为或结构。
- Target behavior / structure: 用中文说明目标行为或结构。
- Why this task is needed now: 用中文说明现在处理这件事的原因。

## Inputs

- Source docs: 列出已读取的规则、README、测试说明或项目入口。
- Related architecture/contracts: 列出相关架构文档、契约、接口或模块边界。
- Related prior work: 列出相关历史计划、review、todo 或迁移记录。

## Task Classification

- Classification: `非琐碎任务`
- Why: 用中文说明分类依据。

## Files

- Create: 列出新增文件；没有则写 `none`。
- Modify: 列出要修改的文件。
- Review: 列出需要重点阅读但不一定修改的文件。
- Test: 列出相关测试文件或验证目标。

## 复用与 Owner 决策

- Existing patterns searched: 用中文说明搜索过的既有模式；命令保持原文。
- Reuse / extend / split / create-new decision: 用中文说明复用、扩展、拆分或新建的决策。
- Owner boundary: 用中文说明 owner 边界。
- Why this is the narrowest safe surface: 用中文说明为什么这是最小安全改动面。

## Intake Analysis Gate

- Relevant superpowers analysis pass: 写明使用的分析 skill。
- Why this pass fits: 用中文说明为什么该分析路径适用。
- grill-with-docs findings: 用中文记录对照项目文档后的结论。
- Plan adjustments after challenge: 用中文记录挑战后对计划做出的调整。

## Execution Slices

### Slice 1: 用中文概括本切片

- Goal: 用中文说明本切片目标。
- Dependencies: 用中文说明前置条件或依赖。
- Files:
  - Create: 列出新增文件；没有则写 `none`。
  - Modify: 列出要修改的文件。
  - Review: 列出要重点阅读的文件。
  - Test: 列出要执行或关注的测试。
- 步骤：
  - [ ] 步骤 1：用中文写可验证的小步骤。
  - [ ] 步骤 2：用中文写可验证的小步骤。
- Validation: 用中文说明本切片验证方式；命令保持原文。
- Review focus: 用中文说明 review 重点。
- Done criteria: 用中文说明完成判定标准。

## Impact And Compatibility

- API / DTO: 用中文说明影响；没有则写 `none`。
- Data / migration: 用中文说明影响；没有则写 `none`。
- State / cache / queue / event: 用中文说明影响；没有则写 `none`。
- Runtime / config: 用中文说明影响；没有则写 `none`。
- Frontend route / state / UX: 用中文说明影响；没有则写 `none`。
- Docs / contracts: 用中文说明影响；没有则写 `none`。

## Plan Review / Architecture Fit

- Target owner boundary: 用中文说明目标 owner 边界。
- Reuse points / landing zones: 用中文说明复用点和落点。
- Known structural debt touched: 用中文说明触达的已知结构债；没有则写 `none`。
- How this plan avoids behavior-only convergence: 用中文说明计划如何避免只改行为不收口结构。
- Hidden second-redesign risk: 用中文说明是否存在完成后立刻二次重构的风险。
- Decision after review: 用中文写计划自审结论。

## Documentation Owner

- Current fact sources to read: 列出当前事实源。
- Fact sources to update after implementation: 列出实现后需要更新的事实源；没有则写 `none`。
- Plan-only notes that must not become architecture source: 用中文说明只属于计划过程的信息。
- Archive condition: 用中文说明计划归档条件。

## Validation

- 计划验证范围：用中文说明必须跑哪些验证类别；命令保持原文。
- 命名 / 契约检查范围：用中文说明需要搜索或人工确认的边界；没有则写 `none`。
- 完成判定：用中文说明本计划通过验证的判定标准。

## Validation Plan

- Per-slice commands: 列出每个切片的验证命令。
- Integration commands: 列出集成验证命令；没有则写 `none`。
- Manual checks: 列出人工检查点；没有则写 `none`。
- Commands intentionally skipped and why: 用中文说明跳过的命令和原因。

## Validation Evidence

- Command: 写实际执行的命令。
  - Result: 写执行结果。
  - Notes: 用中文说明该命令证明了什么。

## Independent Review Handoff

- Review target: 用中文说明 review 目标。
- Validation evidence summary: 用中文概括已有验证证据。
- Architecture / contract inputs: 列出 review 应读取的架构或契约输入。
- Known risks / review focus: 用中文说明已知风险和 review 重点。
- Project-local checks to consider: 列出 reviewer 可考虑重跑的项目本地检查。

## Rollback / Recovery

- Safe revert boundary: 用中文说明安全回退边界。
- Data / config / runtime recovery notes: 用中文说明数据、配置或运行时恢复注意事项；没有则写 `none`。
- Irreversible operations: 用中文说明不可逆操作；没有则写 `none`。

## Residual Risks

- Risk: 用中文说明残余风险。
- Why acceptable: 用中文说明为什么可以接受。
- Follow-up owner, if any: 用中文说明后续 owner；没有则写 `none`。
