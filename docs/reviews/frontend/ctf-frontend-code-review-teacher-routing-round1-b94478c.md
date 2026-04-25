# ctf-frontend 代码 Review（teacher-routing 第 1 轮）：教师端路由与页面拆分调整

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | teacher-routing |
| 轮次 | 第 1 轮 |
| 审查范围 | HEAD `b94478c` 上的当前 worktree 变更；前后端共 39 个文件（含未跟踪文件），其中前端重点审查教师端路由、导航和页面拆分相关文件 |
| 变更概述 | 将教师工作台拆分为班级列表、学生列表、学员分析等独立页面，并补充用户姓名字段展示 |
| 审查基准 | `docs/architecture/frontend/01-architecture-overview.md`、`docs/architecture/frontend/02-routing.md`、`docs/architecture/frontend/07-pages-dataflow.md` |
| 审查日期 | 2026-03-09 |

## 问题清单

### 🟡 中优先级

#### [M1] 侧边栏把“教学概览”入口从教师导航中排除了
- **文件**：[code/frontend/src/components/layout/Sidebar.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/components/layout/Sidebar.vue#L362)
- **问题描述**：当前教师导航在构建分组时显式排除了 `TeacherDashboard`。但路由仍保留该页面的 `title/icon` 元信息，且多个教师页面继续把“教学概览”当作一级页面入口使用。
- **影响范围/风险**：教师登录后虽然仍会被重定向到 `/teacher/dashboard`，但离开后无法再通过全局侧边栏返回该页，形成可访问性回退，也与路由文档中的教师一级导航不一致。
- **修正建议**：保留 `TeacherDashboard` 在 teacher 分组中的展示；如果想弱化入口，也应提供稳定的替代导航，而不是直接从全局菜单移除。

#### [M2] 对已解码的路由参数再次 `decodeURIComponent`，班级名包含 `%` 时会直接抛错
- **文件**：[code/frontend/src/views/teacher/TeacherClassStudents.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherClassStudents.vue#L19)
- **文件**：[code/frontend/src/views/teacher/TeacherStudentAnalysis.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherStudentAnalysis.vue#L46)
- **问题描述**：`route.params.className` 在 Vue Router 中已经是解码后的参数值，这里再次执行 `decodeURIComponent()` 会把普通 `%` 当成转义前缀处理。班级名称只要出现 `100%`、`A%B` 这类合法文本，就会触发 `URI malformed`。
- **影响范围/风险**：相关教师页面会在初始化阶段直接报错，导致该班级无法进入学生列表或学员分析页。
- **修正建议**：直接返回 `String(route.params.className || '')`；如果确实需要手动解码，应只对原始 URL 片段做一次，并对异常输入做保护。

#### [M3] 实时搜索改成远程请求后没有做请求序列保护，旧响应会覆盖新结果
- **文件**：[code/frontend/src/views/teacher/TeacherStudentManagement.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherStudentManagement.vue#L31)
- **文件**：[code/frontend/src/views/teacher/TeacherStudentManagement.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherStudentManagement.vue#L87)
- **文件**：[code/frontend/src/views/teacher/TeacherClassStudents.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherClassStudents.vue#L31)
- **文件**：[code/frontend/src/views/teacher/TeacherClassStudents.vue](/home/azhi/workspace/projects/ctf/code/frontend/src/views/teacher/TeacherClassStudents.vue#L81)
- **问题描述**：输入框每次变更都会立即触发新的 `getClassStudents()` 请求，但组件没有使用请求版本号、AbortController 或“仅接受最后一次响应”的保护。用户连续输入时，较慢的旧请求返回后会直接覆盖较新的查询结果。
- **影响范围/风险**：页面会间歇性显示“上一关键字”的学生列表，造成搜索结果闪回、误点错误学生，属于典型的异步竞态回归。
- **修正建议**：为列表请求增加 request id/sequence 校验，或在发起新请求时取消上一个请求；如果交互允许，也建议加 200~300ms debounce，减少不必要的网络抖动。

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 0 |
| 🟡 中 | 3 |
| 🟢 低 | 0 |
| 合计 | 3 |

## 总体评价

这批拆分把教师端页面边界理顺了，测试也覆盖了新增路由的基本跳转。但当前 worktree 仍存在 3 个实际回退点：全局导航入口缺失、路由参数重复解码、实时搜索的异步竞态。它们都会落到真实用户路径上，建议在继续扩展教师端功能前先修掉。
