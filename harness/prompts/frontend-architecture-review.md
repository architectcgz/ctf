# Frontend Architecture Review Prompt（CTF 入口）

> 状态：Legacy local entry
> 说明：保留给历史 review 引用；已移除的共享 code-reviewer prompt 不再作为本文件前置依赖。

本文件仅作为 CTF 前端架构审计的项目化提示词入口。

## CTF 补充

- 使用前先读目标仓库的 `AGENTS.md`、前端架构测试、路由入口和主要 feature owner。
- 如果 review 结果会进入仓库，优先写入 `docs/reviews/{frontend|architecture|general}/`，不要只停留在聊天结论。
- 如果这轮 review 产出了可复用方法，而不是一次性结论，再回写 `feedback/`；跨项目通用内容再提升到共享 skill 或共享 harness prompt。

## CTF 前端架构审计提示词

```text
你是资深 Vue 3 / Vite / TypeScript 前端架构审查者。请分析本地仓库 `/home/azhi/workspace/projects/ctf/code/frontend` 的前端架构，目标不是重写架构，而是基于真实代码给出现状判断、边界问题、风险和最小可行改进建议。

分析前必须先读取这些事实源：
1. `/home/azhi/workspace/projects/ctf/AGENTS.md`
2. `/home/azhi/workspace/projects/ctf/code/frontend/package.json`
3. `/home/azhi/workspace/projects/ctf/docs/architecture/frontend/README.md`
4. `/home/azhi/workspace/projects/ctf/docs/architecture/frontend/*.md`
5. `/home/azhi/workspace/projects/ctf/code/frontend/src` 下的主要目录：
   - `pages`
   - `widgets`
   - `features`
   - `entities`
   - `shared`
   - `api`
   - `router`
   - `stores`
   - `config`
   - `runtime`
   - `__tests__`

请重点检查：

1. 技术栈与工程入口
   - Vue 3、Vite、TypeScript、Pinia、Vue Router、Tailwind、Vitest 的实际使用方式。
   - `vite.config.ts`、`tsconfig.json`、测试配置、路径别名和自动导入配置是否影响架构边界。

2. 分层结构
   - 当前是否符合类似 Feature-Sliced Design 的分层：`pages -> widgets -> features -> entities -> shared`。
   - 是否存在下层反向依赖上层、同层 slice 互相深度依赖、绕过 public API 的 deep import。
   - `shared` 是否混入了明显的业务语义。
   - `entities` 是否只表达稳定业务对象，还是承载了用户流程。
   - `features` 是否承载用户动作、异步流程和页面工作流。
   - `widgets` 是否只是组合层，还是变成了隐形应用层。
   - `pages` 是否只是路由组合面，还是混入过多 API 请求、watcher、状态机、权限判断和业务流程。

3. 角色与路由边界
   - 学生端、教师端、平台管理员端的页面和 feature 是否清晰分离。
   - 教师端是否只使用 `/academy/*`，平台端是否只使用 `/platform/*`。
   - `features/platform/*`、`features/teacher/*` 和顶层 `features/*` 的归属是否一致。
   - 是否存在语义属于平台/教师后台但仍放在顶层 feature 的情况。

4. 状态与数据流
   - Pinia store、composable、本地组件状态、route query、API DTO/view model 的 owner 是否清晰。
   - API 响应是否直接泄漏到大型 template 或跨层传递。
   - loading / empty / error / loaded 状态是否由合适层级负责。
   - WebSocket、轮询、实时榜单、通知等运行时状态是否有清晰生命周期和清理逻辑。

5. API 层
   - `src/api` 是否只是请求适配层，还是混入页面流程。
   - teacher/platform/admin API 命名和调用边界是否稳定。
   - 是否存在重复 mapper、重复 normalize/default/validate 逻辑。
   - 错误处理、鉴权、重试、取消请求、空状态处理是否分散。

6. 测试与架构护栏
   - Vitest / Vue Test Utils 测试是否覆盖用户行为、状态 owner 和边界。
   - 是否存在只断言 class、DOM 细节、源码字符串的脆弱测试。
   - `scripts/check-frontend-growth-guard.mjs`、`check-theme-tail.mjs`、`check-vue-deep-guard.mjs` 等 guard 保护了哪些边界，是否有缺口。

7. 可维护性风险
   - 找出“大页面组件”“万能 feature”“过宽 shared”“跨角色复用不清”“API DTO 直通 UI”“route page 变应用层”“store 变服务层”等问题。
   - 每个问题都必须给出文件路径、证据、为什么是风险、建议的最小修正方向。
   - 不要凭目录名直接下结论，必须读具体代码和 import/call path。

建议使用这些命令辅助扫描，但结论必须来自实际阅读：
- `find src -maxdepth 2 -type d | sort`
- `rg "^import .* from ['\"]@/(pages|widgets|features|entities|shared|api|stores|router)" src`
- `rg "useRoute|useRouter|defineStore|watch\\(|watchEffect\\(|onMounted\\(|axios|api" src/pages src/widgets src/features`
- `rg "from ['\"]@/features|from ['\"]@/widgets|from ['\"]@/pages" src/entities src/shared`
- `rg "platform|teacher|academy|admin" src/features src/pages src/router`
- `rg "\\?raw|class|padding|gap" src/**/*.test.* src/**/__tests__`

输出格式：

## 结论
用 5-8 条说明当前前端架构整体健康度、主要边界、最大风险。

## 当前架构地图
列出实际目录职责、主要技术栈、路由/状态/API/组件组合方式。

## 关键发现
按严重程度列出，每条包含：
- 文件路径
- 证据
- 问题说明
- 影响范围
- 建议处理方式

## 分层与依赖检查
明确说明哪些依赖方向是健康的，哪些违反边界，哪些只是命名或归属可疑但需要进一步确认。

## 数据流与状态 owner
说明 route、store、composable、API、view model、组件本地状态之间的实际分工。

## 测试与护栏
说明现有测试和 guard 能防住什么，防不住什么。

## 最小改进路线
给出 3 个阶段：
1. 低风险清理
2. owner 边界收口
3. 需要计划和 review 的结构性迁移

要求：
- 不要提出大而空的“重构为 Clean Architecture”。
- 不要建议为了模式而移动文件。
- 优先保留当前 Feature-Sliced / vertical slice 风格。
- 所有建议必须能落到具体文件或具体边界。
- 如果信息不足，明确列出 unknowns，而不是猜测。
```
