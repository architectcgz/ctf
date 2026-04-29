# 赛事资产导出与归档设计

## 目标

补齐赛事级运营能力，让平台能够导出关键结果与训练留档材料，为答辩展示、教学归档和后续备份能力打底。

## 本阶段聚焦

- 赛事结果导出
- 复盘证据摘要导出
- writeup 归档导出

## 设计原则

- 先做“结构稳定可落地”的导出，不做完整备份恢复系统
- 导出格式优先面向教学和评审使用
- 与现有报告导出能力保持风格一致

## 输出物建议

### 1. Contest Export

- 基本信息
- 榜单结果
- 题目解出情况
- 队伍/学生参与情况

### 2. Review Archive Export

- 学生基础信息
- 攻击会话摘要
- writeup 提交与评阅状态
- 教师评语

## API 建议

- `POST /api/v1/admin/contests/:id/export`
- `POST /api/v1/teacher/students/:id/review-archive/export`

## 格式建议

- 第一版优先 JSON
- 若当前报告导出链路成熟，可同时提供可下载文件地址

## 当前落地方案

- 直接复用现有 `reports` 导出任务链路与下载接口
- 新增：
  - `POST /api/v1/admin/contests/:id/export`
  - `POST /api/v1/teacher/students/:id/review-archive/export`
- 赛事导出当前包含：
  - 赛事基础信息
  - 榜单结果
  - 题目解出情况
  - 队伍与成员信息
- 学生复盘归档当前包含：
  - 学生基础信息
  - 训练摘要
  - 时间线与证据事件
  - 社区题解提交、公开状态与推荐状态
  - 人工审核记录

## 当前实现说明

- 两个新导出任务都会先写入 `reports` 表，再走异步生成，统一通过：
  - `GET /api/v1/reports/:id`
  - `GET /api/v1/reports/:id/download`
  查询状态与下载文件
- `reports.type` 扩展为：
  - `contest_export`
  - `review_archive`
- `reports.format` 扩展支持：
  - `json`
- 赛事导出 JSON 当前结构分为：
  - `contest`
  - `scoreboard`
  - `challenges`
  - `teams`
- 学生复盘归档 JSON 当前结构分为：
  - `student`
  - `summary`
  - `skill_profile`
  - `timeline`
  - `evidence`
  - `writeups`
  - `manual_reviews`
- 权限边界：
  - 赛事导出仅管理员可发起
  - 复盘归档由教师/管理员发起；教师仅能导出自己班级学生

## 最小交付

- 至少一个赛事导出接口可用
- 至少一个学生复盘归档导出接口可用
- 可被文档明确说明，结构可用于答辩展示
