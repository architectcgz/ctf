# CTF Frontend

当前前端实现目录，保留学生、教师、管理员三侧页面、共享组件原语与页面级能力封装。

## 技术栈

- Vue 3 + `<script setup lang="ts">`
- Vite 7
- Vue Router 4
- Pinia 3
- Axios
- Tailwind CSS 4
- Lucide Vue Next

## 已就绪能力

- 路由骨架与角色守卫
- Pinia 全局状态
- Axios 请求封装与 401 刷新队列
- 登录/注册页与主布局壳
- 学员、教师、管理员页面占位视图

## 常用命令

```bash
npm run dev
npm run build
npm run typecheck
```

## 测试文件放置约定

前端测试默认贴近行为 owner，不按“页面 / 组件 / 工具”之外再建一个大测试目录。新增测试前先判断这条测试证明的是谁的契约。

- `src/shared/**/__tests__`：放 shared model、shared lib、shared UI 原语自己的行为测试；不得在这里测试具体 feature 或页面流程。
- `src/features/**/__tests__` 或 feature 内邻近 `__tests__`：放 feature owner 的状态、校验、权限、异步流程、组合逻辑和 feature 私有 UI 行为。
- `src/pages/**/__tests__`：放 route / page 入口集成测试，只覆盖页面装配、路由参数、入口状态和关键用户流程；不重复锁定 shared 组件或 feature 内部结构。
- `src/api/__tests__`：放 API adapter、请求参数、响应映射和后端 contract 对齐测试。
- `src/stores/__tests__`：放 Pinia store 行为、持久化、权限状态和跨页面共享状态。
- `src/router/__tests__`、`src/runtime/__tests__`、`src/config/__tests__`、`src/utils/__tests__`：分别放对应基础设施 owner 的测试，不承接业务场景。
- `src/__tests__`：只放架构边界、设计系统 guard、跨切面回归防线和无法归属到单一 feature / page / shared owner 的测试；新增前必须能说明它为什么不能贴近具体 owner。
- `src/test`：放 Vitest setup、测试环境适配和稳定复用的测试工具；只服务单个测试文件的 helper 先留在测试文件本地。

测试文件命名优先使用被测 owner 或业务行为，例如 `ContestDetail.test.ts`、`usePagination.test.ts`、`featureBoundaries.test.ts`。同一行为信号不要在 page、feature、shared 多层重复断言；需要多层测试时，每一层必须证明不同契约。

UI 测试规则以项目根 `AGENTS.md` 为准：优先覆盖用户行为、状态 owner 和架构边界；不要新增只断言 class、CSS 文本块、utility class 顺序或组件内部 markup 细节的过细测试。写完前端测试后，在项目根运行：

```bash
bash scripts/check-frontend-test-guard.sh
```

## 页面风格约定

当前学生端与教师端页面改造，统一按下面这套规则执行。

### 结构

- 页面以一张主卡片为核心，标题、摘要、筛选、列表、说明尽量收进同一个主容器。
- 避免“主卡片外再挂一张完整次卡片”或“卡片里继续套卡片”的结构。
- 同一主卡片里的不同内容区，优先用留白和分隔线切开，而不是继续堆很多独立小卡。
- 列表页的筛选条、工具条和列表主体优先放进同一主卡片。
- 教师端保留管理后台的信息架构，但卡片语言和学生端保持同一套轻量风格。

### 分隔

- 纵向相邻内容使用横向虚线分隔。
- 横向并列内容如果视觉上需要切开，使用竖向虚线分隔。
- 虚线不能太虚，颜色控制在浅灰蓝范围内，避免发黑或过轻。

### 视觉

- 主容器圆角保持克制，通常为 `16px` 左右。
- 阴影保持轻量，只用于提供层次，不做厚重悬浮感。
- **禁止硬编码色彩**：必须使用语义化变量（如 `var(--color-bg-base)`），严禁在 Light Mode 下使用 Dark UI 硬编码值。
- **禁止 !important**：通过选择器优先级或变量继承解决样式覆盖，禁止暴力使用 `!important`。
- **弹窗安全**：所有弹窗/抽屉必须具备 `max-height: calc(100vh - 4rem)` 和内容区滚动保护。
- **抽屉契约**：复杂配置统一使用右侧抽屉布局，主体实色背景，遮罩强模糊。
- 背景以浅白、浅灰蓝渐变和弱对比面为主，优先按白天模式校准。
...
### 文案

- 文案保持短、直、可操作。
- 去掉机械化、说明书式、装饰性强的描述。
- 不保留与当前界面无关的“设计说明”或“状态同步说明”文字。
