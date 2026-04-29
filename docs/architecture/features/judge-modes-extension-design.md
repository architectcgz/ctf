# 判题模式补强设计

## 目标

将当前平台的判题模型从 `static` / `dynamic` 扩展到能覆盖毕设与常见 CTF 训练场景。

## 必做范围

### 1. `regex` 判题

- 后台配置支持正则表达式
- 导入包 `challenge.yml` 支持 `regex`
- 查询、自检、提交校验逻辑一致

### 2. `manual_review` 判题

- 适用于非标准 Flag 文本题或需人工核验的题
- 学员提交答案后进入待审核状态
- 教师/管理员审核通过后计分

## 原始差距

- 契约与文档已出现 `regex`
- 实际导入和关键服务仍主要限制在 `static` / `dynamic`

## 设计原则

- 不搞插件化大重构
- 先在现有 `flag_type`/提交校验路径上做可控扩展
- 导入口径、后台配置口径、判题口径、OpenAPI 口径必须一致
- 自检逻辑需要明确区分：
  - 静态题自检
  - 动态题自检
  - 正则题配置校验
  - 人工题不参与自动校验

## 数据设计

- `regex` 独立存放在 `challenges.flag_regex`
- `manual_review` 继续复用 `submissions`，不新增平行提交表
- `submissions` 扩展：
  - `review_status`
  - `reviewed_by`
  - `reviewed_at`
  - `review_comment`
  - `updated_at`
- 自动判题仍以 `is_correct + score` 作为最终计分依据；人工题在审核通过前保持 `is_correct=false`

## 首版范围控制

- 练习侧 `manual_review` 闭环必须跑通
- 竞赛侧首版至少避免把 `manual_review` 误记为错误提交；完整比赛人工审核计分后续再扩
- 前端首版只补最短路径：
  - 后台题目详情页配置判题模式
  - 学员题目页显示 `pending_review`
  - 教师侧最小审核入口

## 最小交付

- 一个 `regex` 题可从导入到提交闭环跑通
- 一个 `manual_review` 题可配置、提交、审核、计分跑通

## 当前落地状态

- 后端已补齐：
  - `regex` 判题配置、导入、校验与提交链路
  - `manual_review` 练习提交流程、教师审核接口、审核后计分副作用
  - 竞赛侧对 `manual_review` 先做安全兜底，避免误记为错误提交
- 前端已补齐最短闭环：
  - 管理员题目详情页支持配置 `static` / `dynamic` / `regex` / `manual_review`
  - 学员题目页对提交结果区分 `correct` / `incorrect` / `pending_review`
  - 教师学员分析页提供人工审核列表、详情与通过/驳回操作
- 验证口径：
  - 后端以定向 `go test` 验证核心链路
  - 前端以 `vue-tsc` + 定向 `vitest` 验证 API 契约、学员页与教师页交互
