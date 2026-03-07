# CTF 前端任务拆分

## 文档信息

- 范围：`/home/azhi/workspace/projects/ctf/code/frontend`
- 目标：将当前前端工作拆分为可并行执行的独立任务，并明确每个任务的验收标准、依赖关系与阻塞项
- 流水线：`leader -> requirements-analyst（按需） -> frontend-engineer -> code-reviewer -> test-engineer`

## 当前判断

- 前端技术栈已经就绪：Vue 3、Vite、Pinia、Vue Router、Element Plus、Tailwind CSS。
- 学员端与管理员端已有较完整 API 可接入。
- 教师端前端已预定义部分 API，但当前后端代码中仅明确确认了：
  - 学员能力画像查看：`GET /teacher/students/:id/skill-profile`
  - 班级报告导出：`POST /reports/class`
- 教师端以下接口在当前后端代码中暂未确认存在，需先核实：
  - `GET /teacher/classes`
  - `GET /teacher/classes/:name/students`
  - `GET /teacher/students/:id/progress`
  - `GET /teacher/students/:id/recommendations`

## 任务拆分

### F0 基础对齐

- 目标：完成前后端契约核对，修正明显不一致项，恢复基础鉴权能力，输出阻塞清单。
- 范围：
  - `src/api/contracts.ts`
  - `src/router/guards.ts`
  - 相关 API 文件与类型
- 主要工作：
  - 校对管理员 dashboard、audit 等返回结构与后端 DTO 是否一致
  - 恢复当前被临时关闭的登录校验与角色校验
  - 形成教师端缺失接口清单，避免前端直接依赖不存在的 API
- 验收标准：
  - 路由守卫重新生效
  - 类型定义与已确认接口一致
  - 输出教师端阻塞接口列表
- 依赖：无
- 阻塞：无

### F1 学员侧页面完善

- 目标：把学员侧占位页替换为可用页面，优先覆盖真实可接入能力。
- 范围：
  - `src/views/dashboard/DashboardView.vue`
  - `src/views/profile/UserProfile.vue`
  - 如需要，补充相关 composable / 组件 / 测试
- 主要工作：
  - `Dashboard` 基于已有接口展示训练总览、进度、推荐入口
  - `Profile` 接入个人资料展示与密码修改
  - 接入个人报告导出入口（如放在 profile 或 dashboard 内）
- 验收标准：
  - 页面无占位文案
  - 数据来自真实 API
  - 错误态、空态、加载态完整
  - 关键交互有自动化测试
- 依赖：
  - 建议在 F0 完成后开始
- 阻塞：无

### F2 管理侧页面完善

- 目标：完成管理员核心页面，优先接通系统概览与审计日志。
- 范围：
  - `src/views/admin/AdminDashboard.vue`
  - `src/views/admin/AuditLog.vue`
  - `src/views/admin/CheatDetection.vue`
- 主要工作：
  - `AdminDashboard` 接入真实 dashboard API
  - `AuditLog` 实现筛选、分页、表格展示
  - `CheatDetection` 判断是否有后端接口；若没有，先给出降级实现方案或保持阻塞
- 验收标准：
  - dashboard 指标、容器/告警信息真实展示
  - audit log 可筛选、翻页、查看关键字段
  - 无 mock 数据残留
  - 关键页面有组件测试
- 依赖：
  - 建议在 F0 完成后开始
- 阻塞：
  - `CheatDetection` 是否存在专用后端接口待确认

### F3 教师侧报告导出

- 目标：优先落地教师端当前已确认有后端支撑的能力。
- 范围：
  - `src/views/teacher/ReportExport.vue`
- 主要工作：
  - 提供班级选择、导出格式选择、导出状态展示
  - 如接口返回下载地址，则提供下载入口
  - 如接口返回 processing，则给出状态说明与重试方式
- 验收标准：
  - 页面基于真实 `POST /reports/class`
  - 支持当前教师所属班级默认导出
  - 成功、失败、处理中状态可区分
  - 有基础测试覆盖
- 依赖：
  - 建议在 F0 完成后开始
- 阻塞：无

### F4 教师侧班级与教学概览

- 目标：实现教师视角的班级列表、学生详情、教学概览。
- 范围：
  - `src/views/teacher/TeacherDashboard.vue`
  - `src/views/teacher/ClassManagement.vue`
- 主要工作：
  - 展示班级、学生、进度、能力画像、推荐信息
  - 需要基于教师班级/学生相关 API 完成真实联动
- 验收标准：
  - 不使用伪造数据
  - 页面功能完整可交互
  - 支持加载态、空态、错误态
  - 有自动化测试覆盖主要交互
- 依赖：
  - F0 完成
  - 教师相关后端接口确认可用
- 阻塞：
  - 当前后端接口未全部确认，暂不能直接开工

### F5 统一质量闭环

- 目标：确保所有前端任务均按流水线完成 review 与测试。
- 主要工作：
  - 每个功能任务完成后调用 `code-reviewer`
  - review 问题修复后调用 `test-engineer`
  - 运行对应测试命令并记录结果
- 验收标准：
  - 所有任务均有 review 记录
  - 所有任务均有测试变更或明确说明无需新增测试的理由
  - 相关测试可在本地通过
- 依赖：
  - F1 / F2 / F3 / F4 各自实现完成后执行

## 并行执行建议

### Lane A

- 任务：F0 基础对齐
- 说明：必须最先开始，但周期应尽量短，为其他 lane 解锁

### Lane B

- 任务：F1 学员侧页面完善
- 说明：在 F0 核心契约确认后即可独立推进

### Lane C

- 任务：F2 管理侧页面完善
- 说明：与 F1 可并行，接口基础较明确

### Lane D

- 任务：F3 教师侧报告导出
- 说明：可与 F1/F2 并行，因为接口已确认存在

### 暂缓 Lane

- 任务：F4 教师侧班级与教学概览
- 说明：待后端教师接口确认后再启动

## 依赖关系

| 任务 | blockedBy |
|------|-----------|
| F0 | 无 |
| F1 | F0 |
| F2 | F0 |
| F3 | F0 |
| F4 | F0 + 教师端缺失接口补齐 |
| F5 | 对应实现任务完成 |

## 风险与阻塞项

1. 教师端接口未闭合，直接开发 `TeacherDashboard` / `ClassManagement` 会导致前端落到不存在的 API 上。
2. `CheatDetection` 页面当前未确认有专用后端接口，需先判断是新需求还是前端聚合视图。
3. 路由守卫当前存在临时关闭逻辑，若不先修复，会影响后续联调和权限测试可信度。

## 执行顺序建议

1. 先完成 F0，并输出接口核对结果。
2. 并发启动 F1、F2、F3，分别独立走完整流水线。
3. 等教师端接口确认后，再启动 F4。
4. 每个任务都必须经过 `code-reviewer` 与 `test-engineer`，不得跳过。

## 交付物

- 前端任务拆分文档：本文档
- 后续每个任务的：
  - engineer 完成报告
  - review 文档
  - 测试完成报告
