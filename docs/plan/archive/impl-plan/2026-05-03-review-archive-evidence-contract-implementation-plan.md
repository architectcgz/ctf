# Review Archive Evidence Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 收敛教学复盘归档与教学读模型共用的 evidence query 契约，消除 `assessment` 对 `teaching_readmodel/ports` 的反向依赖，并确认前端复盘筛选回退同步遗留项已被现有实现覆盖。

**Architecture:** 保持 evidence 事件构建仍由各自仓储/读模型负责，但把跨模块共享的查询契约上提到 `internal/teaching/evidence`。这样共享点落在教学领域包，而不是落在某个具体读模型端口中。

**Tech Stack:** Go 1.24, Vue 3, TypeScript, Vitest.

---

## Plan Summary

### Objective

- 确认前端 review query 回退同步 TODO 已由现有 watcher 和测试覆盖。
- 将共享的 `EvidenceQuery` 迁移到 `internal/teaching/evidence`。
- 更新 teaching readmodel 与 assessment 的接口、实现和测试。
- 归档本次遗留项收尾 review 结果。

### Non-goals

- 不重写 evidence builder。
- 不新增新的事件类型或接口参数。
- 不调整教师复盘页面视觉或交互。

### Source docs

- `docs/reviews/general/2026-05-03-teaching-review-optimization-review-implementation.md`
- `docs/plan/impl-plan/2026-05-03-teaching-review-optimization-implementation-plan.md`

### Architecture Evaluation

- 这次改动是结构收敛，不是行为扩展。
- 共享查询契约属于教学 evidence 域，放在 `internal/teaching/evidence` 比放在 `teaching_readmodel/ports` 更稳定。
- evidence builder 仍未完全统一，但当前至少消除跨模块依赖方向错误；builder 全共享不在本次范围内。

## Task 1: Baseline and Frontend Residual Check

- [x] **Step 1: Confirm route-query sync is already implemented**
- [x] **Step 2: Confirm existing frontend test covers route query back-navigation**

## Task 2: Shared Evidence Query Contract Extraction

- [x] **Step 1: Add shared `Query` struct to `internal/teaching/evidence`**
- [x] **Step 2: Replace `readmodelports.EvidenceQuery` usages with `evidence.Query`**
- [x] **Step 3: Keep behavior unchanged while removing assessment -> teaching_readmodel ports coupling**

## Task 3: Verification and Review Archive

- [x] **Step 1: Run focused backend tests**
- [x] **Step 2: Run focused frontend test for route-query sync regression**
- [x] **Step 3: Archive review result for this cleanup**
- [x] **Step 4: Update this plan checklist and final status**

## Verification

- `cd code/backend && go test ./internal/module/assessment/... ./internal/module/teaching_readmodel/... -count=1`
- `cd code/frontend && npm run test:run -- src/views/teacher/__tests__/TeacherStudentAnalysis.test.ts`

## Risks

- 如果 `EvidenceQuery` 迁移时有遗漏，会在 assessment 或 readmodel 的编译/测试阶段直接暴露。
- 本次不收敛 builder 本体，后续若继续扩展事件源，仍需要关注双边同步。
