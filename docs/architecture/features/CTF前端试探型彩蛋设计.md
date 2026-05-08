# CTF 前端试探型彩蛋架构

## 文档元信息

- 状态：`implemented`
- 事实源级别：`final`
- 适用范围：`frontend`
- 关联模块：
  - `code/frontend/src/composables`
  - `code/frontend/src/features/auth/model`
  - `code/frontend/src/components/errors`
  - `code/frontend/src/features/notifications/model`
  - `code/frontend/src/features/challenge-detail/model`
- 过程追溯：旧稿 `CTF前端试探型彩蛋设计.md`
- 最后更新：`2026-05-07`

## 1. 背景与问题

这组能力已经不是“要不要做彩蛋系统”的方案讨论，而是已经落成的前端局部行为。这里需要明确的是当前实现边界：

- 彩蛋完全是前端本地态
- 触发点分散在若干已有页面与 composable 中
- 不影响真实业务返回、权限和接口结果

## 2. 架构结论

- 当前彩蛋统一由 `useProbeEasterEggs()` 管理。
- 所有状态都保存在浏览器 `sessionStorage`，键名固定为 `ctf.probe-easter-eggs`。
- 存储模型当前只有两类数据：
  - `counts`
  - `activated`
- 当前已接入的五个真实探测点是：
  - `login-brand`，阈值 `4`
  - `error-status`，阈值 `4`
  - `notification-refresh`，阈值 `3`
  - `notification-id`，阈值 `4`
  - `challenge-side-rail`，阈值 `4`
- 彩蛋反馈当前只表现为：
  - 控制台提示
  - 短时局部文案
  - 轻量 toast / note
- 当前没有后端接口、数据库字段、全局 feature flag 或统一 AppLayout 彩蛋总线。

## 3. 模块边界与职责

### 3.1 模块清单

- `useProbeEasterEggs`
  - 负责：维护会话内计数和激活状态，提供 `track` / `isActivated`
  - 不负责：渲染任何页面反馈

- `useLoginPage`
  - 负责：登录页控制台提示与品牌区 probe 文案
  - 不负责：真实登录校验逻辑

- `ErrorStatusShell`
  - 负责：错误页状态码区域 probe 提示
  - 不负责：路由重试或错误恢复逻辑本身

- `useNotificationListPage`
  - 负责：刷新按钮高频点击 probe
  - 不负责：修改真实通知刷新接口

- `useNotificationDetailPage`
  - 负责：通知详情 ID 区块 probe
  - 不负责：修改通知正文或链接跳转

- `useChallengeDetailPage`
  - 负责：题目详情侧栏 probe
  - 不负责：题目判题和题解读取

### 3.2 事实源与所有权

- 彩蛋状态事实源：浏览器 `sessionStorage`
- 彩蛋触发逻辑事实源：各页面调用 `useProbeEasterEggs.track`
- 真实业务状态事实源：原有登录、通知、题目详情链路，彩蛋不覆盖它们

## 4. 关键模型与不变量

### 4.1 核心实体

- `ProbeStorageState`
  - `version`
  - `counts`
  - `activated`

- `ProbeTrackResult`
  - `count`
  - `unlocked`
  - `activated`

### 4.2 不变量

- 当前彩蛋状态只在单个浏览器会话内持久化，不回写后端。
- 每个 probe key 第一次达到阈值时，`unlocked = true`；之后只保留 `activated = true`，不会反复触发首次解锁效果。
- `sessionStorage` 不可用时，逻辑退回当前实例内存态，不影响页面主功能。
- 彩蛋不会改变真实登录结果、通知加载结果、题目详情结果和错误页状态。
- 当前 AWD 战场页面没有接入这套 probe，也没有与浏览器文件工作台耦合。

## 5. 关键链路

### 5.1 通用触发链路

1. 页面在局部交互点调用 `track(key, threshold)`。
2. composable 为对应 key 累加计数。
3. 若该 key 首次达到阈值，则标记 `activated[key] = true`。
4. 页面根据 `unlocked` 结果展示一次性短时反馈。
5. 状态写回 `sessionStorage`。

### 5.2 当前已落地触发点

1. 登录页：
   - 页面初始化输出控制台提示
   - 品牌区连续点击 4 次后显示“隐藏入口排查完毕，结果让你失望了。”
2. 错误页：
   - 状态区连续点击 4 次后显示“路径枚举记录已写入……”
3. 通知列表：
   - 刷新按钮连续触发 3 次后显示“新消息不会因为执念刷新得更快。”
4. 通知详情：
   - ID 区块连续点击 4 次后显示“值守备注：有人开始认真看编号了。”
5. 题目详情：
   - 侧栏区域连续点击 4 次后显示“这块区域的情报价值，低于你现在的期待。”

## 6. 接口与契约

### 6.1 本地存储契约

- storage key：`ctf.probe-easter-eggs`
- payload 当前结构：
  - `version: 1`
  - `counts: Record<string, number>`
  - `activated: Record<string, boolean>`

### 6.2 非契约边界

当前没有以下任何后端或跨页面契约：

- 无 HTTP API
- 无数据库表
- 无服务端计数
- 无统一彩蛋配置下发

## 7. 兼容与迁移

- 这组能力当前完全是前端局部增强，不属于产品主流程。
- 旧稿里的 ARG 式串线玩法、全局系统回应或隐藏入口，不是当前实现事实。
- 若未来要引入服务端开关或更复杂串联机制，应新开专题；不能把当前本地态实现写成更大系统。

## 8. 代码落点

- `code/frontend/src/composables/useProbeEasterEggs.ts`
- `code/frontend/src/features/auth/model/useLoginPage.ts`
- `code/frontend/src/components/errors/ErrorStatusShell.vue`
- `code/frontend/src/features/notifications/model/useNotificationListPage.ts`
- `code/frontend/src/features/notifications/model/useNotificationDetailPage.ts`
- `code/frontend/src/features/challenge-detail/model/useChallengeDetailPage.ts`

## 9. 验证标准

- `sessionStorage` 中存在 `ctf.probe-easter-eggs`，并记录 `counts / activated`。
- 五个已接入页面都只在达到指定阈值时显示一次首次解锁反馈。
- 关闭或刷新页面后，只保留会话级状态，不会写入后端。
- 登录、通知、题目详情和错误页的真实业务行为不因彩蛋而改变。
- 当前 AWD 学生战场不包含额外彩蛋入口。
