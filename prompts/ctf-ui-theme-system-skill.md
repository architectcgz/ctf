# CTF UI Theme System Skill

## Migrated From

历史 CTF UI theme skill 原文

## 用途

在构建或重构 CTF 前端页面时，保持布局、排版、色彩、文案和交互模式符合已验证的 academy / challenges / platform 工作台风格。

## 核心提示词

```text
请按 CTF UI Theme System 处理当前页面：学生刷题优先，教师/管理员界面保持分析和治理效率；整体视觉专业、克制、技术感来自秩序而不是霓虹或装饰；页面使用一个主 workspace surface，内部优先用分隔线、层级和留白组织信息；列表使用 flat directory/list pattern；可见 UI 只保留任务相关文案，不展示设计说明。
```

## 关键规则摘要

- 学生流程优先：浏览题目 -> 启动环境 -> 提交 Flag -> 查看题解/复盘。
- 教师/管理员页面偏分析和运维，不做 OA 后台模板感。
- 默认 light-first，同时支持 dark。
- 避免二次元、游戏商城、霓虹黑客风和企业 OA 感。
- 页面主标题用明确层级，避免重复声明标题字体指标。
- 列表优先扁平 row + header，不堆叠小卡片。
- 搜索、筛选、分页要明确状态保留和 stale response 处理。
- 行操作保持主次层级；超过 2 个主要可见操作时使用 overflow。
- CSS 颜色使用语义变量，不直接写硬编码 hex/rgb。
- 不在可见 UI 中写设计意图、布局介绍、实现说明或 todo。

## 迁移说明

原文件是完整 skill 定义，仍可作为细则事实源；本文件迁入 `prompts/` 后作为严格 harness 下的可复用提示词入口。
