# 选手 Writeup 提交与评阅设计

## 目标

把当前“管理员维护题解”的 writeup 能力扩展为“选手提交解题过程、教师评阅、结果归档”的教学工作流。

## 现状

- 当前 writeup 主要服务于题目官方题解。
- 学生侧没有提交 writeup 的入口。
- 教师侧没有按学生/题目查看 writeup、给评语或标记质量的流程。

## 设计范围

### 包含

- 选手按题目提交 writeup
- 教师查看、筛选、评阅 writeup
- writeup 与学生、题目、赛事建立关联
- 为报告导出和后续证据归档预留引用字段

### 不包含

- 富文本协同编辑
- 多人协作审稿
- 外部文档平台集成

## 核心设计

### 1. 区分两类 writeup

- `challenge_writeup`：官方题解，保留现有能力
- `submission_writeup`：选手提交记录，新增能力

避免把教师题解编辑和学生提交混在一个模型里。

### 2. 学生提交流程

- 学生在题目详情或个人训练记录中提交 writeup
- 支持草稿与正式提交
- 内容至少包含：
  - 标题
  - 解题思路
  - 关键利用步骤
  - 遇到的问题

### 3. 教师评阅流程

- 教师按学生、班级、题目筛选 writeup
- 可写评语
- 可给出结果状态：
  - `pending`
  - `reviewed`
  - `excellent`
  - `needs_revision`

### 4. 与画像/报告的关系

- 第一版只在教师报告和学生详情中展示“writeup 已提交/已评阅”状态
- 第二版再考虑把评阅结果纳入画像权重

## 数据模型建议

- `submission_writeups`
  - `id`
  - `user_id`
  - `challenge_id`
  - `contest_id`
  - `status`
  - `title`
  - `content`
  - `submitted_at`
  - `reviewed_by`
  - `reviewed_at`
  - `review_comment`

## API 建议

- 学生侧
  - `POST /api/v1/challenges/:id/writeup-submissions`
  - `GET /api/v1/challenges/:id/writeup-submissions/me`
- 教师侧
  - `GET /api/v1/teacher/writeup-submissions`
  - `GET /api/v1/teacher/writeup-submissions/:id`
  - `PUT /api/v1/teacher/writeup-submissions/:id/review`

## 验收标准

- 学生可提交与更新自己的 writeup
- 教师可查看并评阅
- 记录可按学生/题目检索
- 文档、接口、测试齐全
