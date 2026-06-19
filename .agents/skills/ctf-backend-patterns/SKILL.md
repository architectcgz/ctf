---
name: ctf-backend-patterns
description: CTF 项目后端重构和架构 pattern 索引，引用 feedback 中已验证的最佳实践
---

# CTF Backend Patterns

项目后端开发中已验证的 pattern 和最佳实践索引。

## Refactoring Patterns

### Package Split By Responsibility

单个文件超过 800-1000 行且职责混合时的拆分方法：
- 同包内拆分，不创建新 package（除非有明确封装需求）
- 按职责拆分：types / load / validate / defaults / domain-logic
- 先写结构护栏测试，再执行拆分
- 保持所有导出 API 不变
- 详细模板：`feedback/2026-06-12-package-split-by-responsibility-template.md`

适用场景：
- 单文件行数超标且职责混合
- 多人频繁触碰导致 merge conflict
- 可以按明确维度拆分

不适用：
- 拆分后会引入循环依赖 → 应考虑 package 级重构

## 使用方式

遇到类似问题时：
1. 在本 skill 找相关 pattern
2. 读对应 feedback 获取详细模板和案例
3. 按 checklist 执行

## 添加新 Pattern

当 feedback 中出现高质量 pattern 时：
1. 在对应章节增加简要索引（5-10 行）
2. 指向详细 feedback 路径
3. 更新 feedback 的沉淀状态
