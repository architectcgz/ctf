# CTF 平台前端任务拆解

## 基本信息

| 字段 | 说明 |
|------|------|
| 来源 | ctf-platform-task-breakdown.md（T4, T11, T15, T18, T33） |
| 拆分日期 | 2026-03-03 |
| 技术栈 | Vue 3 + Vite + TypeScript + Tailwind CSS + Element Plus |
| 工作模式 | 每个任务在独立 worktree 中开发 → code-reviewer 审查 → 合并主分支 |

## 任务列表

### FE-T1：前端项目骨架初始化（对应 T4）

**依赖**：后端 T3（认证接口）

**交付物**：
- Vite + Vue 3 + TypeScript 项目初始化
- Tailwind CSS 4 + Element Plus 集成
- 目录结构搭建（api/components/views/stores/router/utils）
- Axios 封装（拦截器：Token 注入、401 处理、错误映射）
- Pinia 状态管理初始化
- Vue Router 路由框架 + 权限守卫
- 登录/注册页面 + 主布局框架（TopNav + Sidebar + Content）

**验收标准**：
- `npm run dev` 启动成功
- 登录后跳转主页，Token 存储正确
- 未登录访问受保护路由自动跳转登录页
- 不同角色看到不同侧边栏菜单

**Worktree 分支**：`feature/frontend-skeleton`

---

### FE-T2：靶场管理前端页面（对应 T11）

**依赖**：FE-T1、后端 T8/T9/T10

**交付物**：
- 管理员端：镜像管理页面（上传、列表、构建状态）
- 管理员端：靶机管理页面（CRUD、难度配置、Flag 配置）
- 管理员端：靶机详情页（镜像关联、标签、资源限制）
- 表单校验与错误提示
- 镜像构建状态轮询刷新

**验收标准**：
- 管理员可完成靶机从创建到发布全流程
- 镜像构建状态实时更新
- 表单校验拦截不合法输入
- 发布前校验（镜像、Flag、标签）生效

**Worktree 分支**：`feature/admin-challenge-manage`

---

### FE-T3：攻防演练前端页面（对应 T15）

**依赖**：FE-T1、后端 T12/T13/T14

**交付物**：
- 学员端：靶场列表页（筛选、搜索、难度标签、完成状态）
- 学员端：靶场详情页（描述、开始挑战按钮、Flag 提交框）
- 实例状态面板（运行中实例列表、倒计时、销毁/延时按钮）
- 即将超时弹窗提醒（< 5 分钟）
- 个人得分卡片（总分、解题数、排名）
- Flag 提交结果即时反馈

**验收标准**：
- 学员可完成浏览 → 启动 → 提交 Flag → 查看得分全流程
- 倒计时准确，超时提醒正常弹出
- Flag 提交后即时反馈（成功/失败/已完成/被锁定）
- 筛选和搜索功能正确

**Worktree 分支**：`feature/student-challenge`

---

### FE-T4：竞赛管理前端页面（对应 T18）

**依赖**：FE-T1、后端 T16/T17

**交付物**：
- 管理员端：竞赛创建/编辑页面（时间、题目、计分规则、冻结时间）
- 管理员端：竞赛列表与状态监控
- 学员端：竞赛列表页（可报名竞赛、进行中竞赛入口）
- 学员端：竞赛详情页（题目列表、Flag 提交、倒计时）
- 学员端：组队页面（创建队伍、邀请码、成员管理）
- 实时排行榜页面（WebSocket 驱动，支持冻结状态）

**验收标准**：
- 管理员可完成竞赛创建到发布全流程
- 学员可组队、参赛、提交 Flag、查看排行榜
- 排行榜实时刷新，冻结后显示"榜单已冻结"提示
- WebSocket 连接断开自动重连

**Worktree 分支**：`feature/contest-system`

---

### FE-T5：能力画像与推荐前端页面（对应 T33）

**依赖**：FE-T1、后端 T31/T32

**交付物**：
- 学员端：能力画像页面（ECharts 雷达图，薄弱项高亮）
- 学员端：推荐靶场列表（推荐理由、难度标签、一键跳转）
- 教师端：查看指定学员能力画像
- 空状态处理（新学员引导）

**验收标准**：
- 雷达图渲染正确，交互流畅（hover 展示详细数据）
- 推荐列表点击可跳转靶场详情页
- 教师可查看任意学员画像数据
- 新学员显示空状态引导

**Worktree 分支**：`feature/skill-assessment`

---

## 工作流程

### 1. 创建 Worktree

```bash
cd /home/azhi/workspace/projects/ctf
git worktree add .claude/worktrees/<分支名> -b <分支名>
```

### 2. 开发与自检

在 worktree 中完成开发，自检清单：
- TypeScript 类型检查通过
- ESLint 无错误
- 组件可正常渲染
- API 对接正确
- 路由跳转正常
- 权限控制生效

### 3. 提交 Commit

```bash
git commit -m "feat(前端): <功能描述>"
```

### 4. 调用 code-reviewer

完成开发后，调用 code-reviewer agent 进行审查：
- 检查代码质量
- 检查架构一致性
- 检查安全问题
- 生成 review 报告到 `docs/reviews/`

### 5. 修复 Review 问题

根据 review 报告修复所有问题（高/中/低优先级全部修复）

### 6. 复审通过后合并

```bash
cd /home/azhi/workspace/projects/ctf
git merge <分支名>
git push origin main
git worktree remove .claude/worktrees/<分支名>
```

---

## 任务依赖关系

```
FE-T1（前端骨架）
├── FE-T2（靶场管理）
├── FE-T3（攻防演练）
├── FE-T4（竞赛系统）
└── FE-T5（能力画像）
```

所有前端任务都依赖 FE-T1，但 FE-T2 ~ FE-T5 之间无依赖，可并行开发。

---

## Review 报告命名规范

格式：`ctf-frontend-code-review-<任务主题>-round<轮次>-<commitHash>.md`

示例：
- `ctf-frontend-code-review-skeleton-round1-abc1234.md`
- `ctf-frontend-code-review-challenge-manage-round1-def5678.md`
- `ctf-frontend-code-review-challenge-manage-round2-ghi9012.md`（修复后复审）
