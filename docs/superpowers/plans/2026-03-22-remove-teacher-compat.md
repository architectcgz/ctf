# Remove Teacher Compatibility Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 物理删除 `internal/module/teacher` 兼容层，并让教师查询路由直接依赖 `teaching_readmodel`。

**Architecture:** 保留 `teaching_readmodel` 作为教师侧读模型模块，对外直接暴露 `Handler` 与 `TeachingQuery`。`app/composition` 与 `router` 直接装配 `teaching_readmodel`，同时删除架构规则中对 `teacher -> teaching_readmodel` 的放行特例，避免兼容壳回流。

**Tech Stack:** Go, Gin, GORM, go test

---

### Task 1: 固化架构目标与失败用例

**Files:**
- Modify: `code/backend/internal/app/architecture_rules_test.go`
- Modify: `code/backend/internal/app/router_test.go`
- Test: `code/backend/internal/app`

- [ ] **Step 1: 调整架构规则测试**

删除 `teacher -> teaching_readmodel` 的例外放行，确保跨模块具体实现依赖会被识别。

- [ ] **Step 2: 调整路由装配类型断言**

把 `TeacherModule` 相关断言改为 `TeachingReadmodelModule`，让编译先失败。

- [ ] **Step 3: 运行失败验证**

Run: `go test ./internal/app -run '^TestArchitectureRulesRejectConcreteCrossModuleImports$|^TestCompositionModulesExposeContracts$' -count=1`
Expected: FAIL，失败点指向 `teacher` 兼容层或旧类型名。

### Task 2: 物理迁移装配层

**Files:**
- Modify: `code/backend/internal/app/composition/teacher_module.go`
- Modify: `code/backend/internal/app/router.go`
- Modify: `code/backend/internal/app/router_routes.go`
- Modify: `code/backend/internal/app/full_router_integration_test.go`
- Modify: `code/backend/internal/app/full_router_state_matrix_integration_test.go`

- [ ] **Step 1: 重命名 composition 单元**

将 `TeacherModule` / `BuildTeacherModule` 改为直接表达 `teaching_readmodel` 的模块命名。

- [ ] **Step 2: 更新 router 装配与依赖注入**

同步更新 builder 变量、路由依赖结构体与测试钩子。

- [ ] **Step 3: 运行装配测试**

Run: `go test ./internal/app -run '^TestBuildRoot$|^TestCompositionModulesExposeContracts$|^TestNewRouter' -count=1`
Expected: PASS

### Task 3: 删除 teacher 兼容层

**Files:**
- Delete: `code/backend/internal/module/teacher/handler.go`
- Delete: `code/backend/internal/module/teacher/repository.go`
- Delete: `code/backend/internal/module/teacher/service.go`
- Delete: `code/backend/internal/module/teacher/service_test.go`

- [ ] **Step 1: 确认无生产代码依赖 `module/teacher`**

Run: `rg -n 'module/teacher|TeacherModule|buildTeacherModule' code/backend/internal -g '!**/*_test.go'`
Expected: no matches

- [ ] **Step 2: 删除兼容层文件**

仅删除 `teacher` 兼容壳，不连带做无关重构。

- [ ] **Step 3: 运行教学读模型与路由回归**

Run: `go test ./internal/module/teaching_readmodel ./internal/app -run 'TestArchitectureRulesRejectConcreteCrossModuleImports|TestBuildRoot|TestCompositionModulesExposeContracts|TestNewRouter' -count=1`
Expected: PASS

### Task 4: 最终最小验证

**Files:**
- Verify only

- [ ] **Step 1: 运行教师侧相关集成回归**

Run: `go test ./internal/app -run 'TestFullRouter|TestPracticeFlow' -count=1`
Expected: PASS

- [ ] **Step 2: 运行 teaching readmodel 定向测试**

Run: `go test ./internal/module/teaching_readmodel -count=1`
Expected: PASS

- [ ] **Step 3: 整理变更并准备提交**

记录删除的兼容层文件、重命名的 composition 类型，以及仍需后续迁移的边界。
