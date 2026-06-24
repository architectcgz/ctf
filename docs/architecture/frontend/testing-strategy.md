# 前端测试策略与覆盖

> 状态：Current
> 事实源：`code/frontend/src/__tests__/`、`code/frontend/vitest.config.ts`、`code/frontend/package.json`
> 替代：无

## 本文档范围

| 覆盖 | 不覆盖 |
|------|--------|
| 前端测试工具栈与配置 | 后端测试策略（见后端文档） |
| E2E 测试覆盖范围与关键路径 | CI/CD 配置细节 |
| 组件测试、状态测试、架构测试的职责边界 | 具体页面的测试用例细节 |
| 测试质量守卫与过细测试拦截 | 生产环境监控策略 |

## 定位

本文档只说明前端测试工具选型、测试类型划分、E2E 覆盖范围、关键路径测试清单和测试质量守卫机制。

## 当前设计

- `code/frontend/vitest.config.ts`
  - 负责：Vitest 配置、测试环境、覆盖率配置
  - 不负责：具体测试用例编写

- `code/frontend/src/__tests__/`
  - 负责：全局前端架构守卫与跨页面稳定策略测试
  - 不负责：替代具体 page / feature / widget 的行为测试

- `code/frontend/src/features/**/model/*.test.ts`
  - 负责：页面行为、状态机、请求编排的测试
  - 不负责：UI 渲染细节测试

## 1. 测试工具栈

### 1.1 核心工具

| 工具 | 用途 | 版本 |
|------|------|------|
| Vitest | 测试运行器 | 最新稳定版 |
| Vue Test Utils | Vue 组件测试 | 最新稳定版 |
| jsdom | DOM 环境模拟 | 通过 Vitest 配置 |
| MSW（可选） | API Mock | 未使用（当前使用手写 Mock） |

### 1.2 测试命令

```bash
# 运行所有测试
npm run test

# 运行测试并生成覆盖率报告
npm run test:coverage

# 监听模式
npm run test:watch

# 运行单个测试文件
npm run test src/features/auth/model/useLogin.test.ts
```

### 1.3 测试配置

**测试环境**：`jsdom`

**覆盖率配置**（当前未强制）：

- 目标：60%+ 语句覆盖率
- 优先覆盖：`features/**/model`、`shared/model`、关键 composables

## 2. 测试类型划分

### 2.1 单元测试（Unit Tests）

**位置**：与源码同目录，`*.test.ts` 或 `*.spec.ts`

**职责**：

- 测试单个函数、composable、工具函数
- 验证状态机转换、输入验证、边界条件

**示例**：

```typescript
// features/auth/model/useLogin.test.ts
describe('useLogin', () => {
  it('should set loading state during login', async () => {
    const { login, isLoading } = useLogin()
    expect(isLoading.value).toBe(false)
    
    login('user', 'pass')
    expect(isLoading.value).toBe(true)
  })
})
```

**特征**：

- 不需要渲染完整组件
- 执行时间 < 100ms
- 验证逻辑正确性

### 2.2 组件测试（Component Tests）

**位置**：与组件同目录，`*.test.ts`

**职责**：

- 测试组件渲染、事件响应、props 传递
- 验证用户交互（点击、输入）

**示例**：

```typescript
// shared/ui/common/AppEmpty.test.ts
describe('AppEmpty', () => {
  it('should render empty state with custom message', () => {
    const wrapper = mount(AppEmpty, {
      props: { message: 'No data' }
    })
    expect(wrapper.text()).toContain('No data')
  })
})
```

**特征**：

- 使用 `mount()` 或 `shallowMount()`
- 验证 DOM 输出和事件触发
- 执行时间 100ms - 500ms

### 2.3 架构测试（Architecture Tests）

**位置**：`code/frontend/src/__tests__/`

**职责**：

- 验证架构约束（依赖方向、层级边界）
- 检查命名规范、文件结构
- 验证 route page 不直接持有复杂业务 owner

**示例**：

```typescript
// __tests__/routePageArchitectureBoundary.test.ts
describe('Route Page Architecture Boundary', () => {
  it('should not use route hooks directly in route pages', () => {
    const routePages = glob.sync('src/pages/**/*.vue')
    routePages.forEach(file => {
      const content = fs.readFileSync(file, 'utf-8')
      expect(content).not.toMatch(/useRouteQueryTabs/)
    })
  })
})
```

**特征**：

- 通过静态分析或源码扫描
- 不运行业务逻辑
- 执行时间 < 1s

### 2.4 E2E 测试（End-to-End Tests）

**位置**：`code/frontend/e2e/`（当前未系统化）

**职责**：

- 端到端用户流程测试
- 真实浏览器环境验证
- 验证关键业务路径

**示例场景**：

- 用户登录 → 进入比赛 → 提交 flag → 查看排行榜
- 教师创建题目 → 发布 → 学生查看

**特征**：

- 需要真实后端或 Mock 服务
- 执行时间 10s - 60s
- 优先覆盖关键路径，不追求全覆盖

## 3. E2E 测试覆盖范围

当前 E2E 测试处于早期阶段，优先覆盖以下关键路径：

### 3.1 学生侧关键路径

| 路径 | 覆盖状态 | 优先级 |
|------|---------|--------|
| 登录 → 查看比赛列表 → 进入比赛 | ⚠️ 未覆盖 | P0 |
| 查看题目 → 提交 flag → 验证得分 | ⚠️ 未覆盖 | P0 |
| 查看排行榜 → 实时更新 | ⚠️ 未覆盖 | P1 |
| 查看公告 → 实时推送 | ⚠️ 未覆盖 | P1 |
| 查看通知 → 标记已读 | ⚠️ 未覆盖 | P2 |

### 3.2 教师侧关键路径

| 路径 | 覆盖状态 | 优先级 |
|------|---------|--------|
| 创建题目 → 上传题包 → 预览 | ⚠️ 未覆盖 | P0 |
| 创建比赛 → 添加题目 → 发布 | ⚠️ 未覆盖 | P0 |
| 查看提交记录 → 导出报表 | ⚠️ 未覆盖 | P1 |

### 3.3 管理员侧关键路径

| 路径 | 覆盖状态 | 优先级 |
|------|---------|--------|
| 用户管理 → 批量导入 → 分配权限 | ⚠️ 未覆盖 | P1 |
| 班级管理 → 创建班级 → 添加学生 | ⚠️ 未覆盖 | P1 |
| 系统监控 → 查看实例状态 | ⚠️ 未覆盖 | P2 |

**说明**：当前 E2E 测试未系统化，优先通过手动测试验证关键路径。

## 4. 关键路径测试清单

以下场景必须有对应测试（单元测试或组件测试）：

### 4.1 状态管理

- [ ] 登录状态恢复
- [ ] 用户信息存储与读取
- [ ] WebSocket 连接状态转换
- [ ] 表单草稿保存与恢复

### 4.2 权限与路由

- [ ] 未登录用户访问受保护页面 → 跳转登录
- [ ] 学生访问教师页面 → 403
- [ ] 路由守卫正确拦截

### 4.3 数据加载与错误处理

- [ ] 分页加载正确处理
- [ ] 空状态正确显示
- [ ] 网络错误降级提示
- [ ] 401 错误自动登出

### 4.4 实时功能

- [ ] WebSocket 重连机制
- [ ] 心跳超时处理
- [ ] 断线后降级为手动刷新

### 4.5 表单与验证

- [ ] 必填字段验证
- [ ] 格式验证（邮箱、手机号）
- [ ] 提交失败显示错误

## 5. 测试质量守卫

### 5.1 过细测试拦截

前端测试不应锁定实现细节，以下测试类型被禁止或需要 owner 说明：

**禁止的测试模式**：

- 只断言 class 名的测试
- 只断言 CSS 文本块的测试
- 只断言 utility class 顺序的测试
- 只断言组件内部 markup 细节的测试

**检查机制**：

```bash
# 检查当前工作区改动
bash scripts/check-frontend-test-guard.sh

# 只检查暂存文件
bash scripts/check-frontend-test-guard.sh --staged

# 检查指定文件
bash scripts/check-frontend-test-guard.sh --files src/features/auth/ui/LoginForm.test.ts
```

**Pre-commit Hook**：

自动运行 `check-frontend-test-guard.sh --staged`，阻止新增过细测试。

### 5.2 允许的细粒度测试

以下场景可以包含 `?raw` 源码字符串测试：

- **架构边界守卫**：验证 route page 不直接持有业务 owner
- **依赖方向守卫**：验证 `entities/*` 不反向依赖 `features/*`
- **危险 surface 回归防线**：防止误用全局状态或直接访问浏览器 API
- **迁移 guard**：明确标注 owner 和移除条件的临时守卫

**要求**：测试必须在注释中说明 owner 和边界理由。

### 5.3 状态测试优先

对 `loading / empty / error / loaded` 这类状态，优先用 Vue Test Utils 做组件渲染测试：

```typescript
// ✅ 正确：测试状态渲染
describe('ChallengeList', () => {
  it('should show empty state when no challenges', () => {
    const wrapper = mount(ChallengeList, {
      props: { challenges: [] }
    })
    expect(wrapper.find('[data-testid="empty-state"]').exists()).toBe(true)
  })
})

// ❌ 错误：只锁定 class
it('should have empty-state class', () => {
  expect(wrapper.html()).toContain('class="empty-state"')
})
```

## 6. 测试文件放置规则

### 6.1 单元测试与组件测试

**规则**：与源码同目录

```
src/features/auth/model/
  ├── useLogin.ts
  └── useLogin.test.ts

src/shared/ui/common/
  ├── AppEmpty.vue
  └── AppEmpty.test.ts
```

### 6.2 架构测试

**规则**：放在 `src/__tests__/`

```
src/__tests__/
  ├── architectureBoundaries.test.ts
  ├── routePageArchitectureBoundary.test.ts
  └── appRouteTransition.test.ts
```

### 6.3 E2E 测试

**规则**：放在 `e2e/`（当前未系统化）

```
e2e/
  ├── student/
  │   ├── login.spec.ts
  │   └── submit-flag.spec.ts
  └── teacher/
      └── create-challenge.spec.ts
```

## 7. 边界

### 7.1 本文档不覆盖

- **具体页面的测试用例**：见各 feature 目录
- **CI/CD 配置**：见 `.github/workflows/`
- **性能测试**：暂未系统化
- **后端测试**：见 `docs/architecture/backend/design/testing-strategy.md`

### 7.2 与其他文档的关系

- 前端架构：见 `docs/architecture/frontend/`
- 组件体系：见 `docs/architecture/frontend/06-components.md`
- 页面数据流：见 `docs/architecture/frontend/07-pages-dataflow.md`
- 后端测试：见 `docs/architecture/backend/design/testing-strategy.md`

## 8. Guardrail

- 新增 feature 必须有对应测试
- 修复 bug 必须先写回归测试
- 测试不得锁定实现细节（class 名、markup 结构）
- Pre-commit hook 自动拦截过细测试
- 架构测试失败必须修复，不得修改测试让它通过

## 9. 已知限制

- E2E 测试未系统化，优先依赖手动测试
- 覆盖率报告未接入 CI/CD
- 部分旧测试仍然锁定了实现细节（正在逐步迁移）
- 缺少性能测试和可访问性测试自动化
