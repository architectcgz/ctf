# 实现后缺少强制独立 review 闭环

## Migrated From

历史 improvement 条目 `2026-05-03-实现后缺少强制独立-review-闭环`

## Status

not-impl

## 问题描述

当前工作流对 verification-before-completion 有硬约束，但 backend-engineer/frontend-engineer 的默认收尾没有强制进入 code-reviewer 或 requesting-code-review 流程，导致实现者容易在定向验证通过后直接结束。

## 原因分析

实现完成且测试通过之后，如果没有独立 review，容易遗漏代码质量、结构债、回归风险和契约一致性问题。历史任务中曾出现后续补 review 才发现前端大文件继续膨胀、AWD 个人归因错误、Jeopardy 上下文丢失等问题。

## 解决方案

中高复杂度实现任务的默认闭环应升级为：

```text
实现 -> 最小充分验证 -> 独立 review -> 修复 findings -> 再验证 -> 完成
```

## 收获

这条反馈应继续推动实现型 skill 和 pipeline 的收尾规则，避免“验证通过”被误当成“review gate 已满足”。
