# 页面设计：AWD 竞赛页面 (AWD Contest)

> 继承：../design-system/MASTER.md | 角色：学生
> 当前覆盖：
> - 学生页 `/contests/:id?panel=challenges` 下的 AWD 战场
> - 不包含独立学生防守路由或浏览器文件工作台

---

## 页面定位

- AWD 模式继续复用竞赛详情页入口，通过 `ContestDetail.vue` 在 `challenges` panel 内挂载 `ContestAWDWorkspacePanel`。
- 页面不是多 tab 子系统，也不是独立“防守内容页”；学生在同一战场内完成防守态势查看、攻击提交和战报阅读。
- 当前页面事实明确排除两类旧方案：
  - 不新增 `ContestAWDDefenseWorkbench` 独立路由。
  - 不在学生页暴露 `AWDDefenseFileWorkbench`、目录树、文件读取或保存入口。

---

## 1. 页面骨架

- 顶部先显示 HUD 条，再进入三栏战场主体。
- HUD 固定展示五组信息：
  - `当前回合`
  - `我的战队`
  - `战队服务`
  - `最高分`
  - 右侧刷新按钮与最近同步时间
- 页面状态只有三种主分支：
  - 初次进入且 workspace 未返回时，显示 `正在建立战场连接...`
  - 未加入队伍时，显示 `先加入队伍`
  - 有队伍后进入三栏战场布局

---

## 2. 我的防守

- 左栏标题固定为 `我的防守`。
- 栏内顺序固定为：
  - 风险告警条
  - `AWDDefenseServiceList`
  - `AWDDefenseOperationsPanel`
- 风险告警只汇总当前稳定存在的服务态：
  - `已失陷`
  - `已离线`
  - `检测到 N 次攻击`
- 服务列表每项展示：
  - 题目名
  - `服务 #id`
  - 服务状态 badge
  - 实例状态文本
  - 风险原因 chips
  - 行内动作：打开服务、`SSH`、`重启`
- 选择服务只会刷新左栏下半部分，不会跳路由。

### Web 防守工作台

- 选中服务后，`AWDDefenseOperationsPanel` 作为当前唯一的 Web 防守承接区。
- 区块固定为：
  - `风险`：服务状态、实例状态、风险标签
  - `验证`：打开服务、重启服务、生成 SSH
  - `SSH 连接`：命令、密码、过期时间、复制动作
- 当前学生页不再展示：
  - 防守范围
  - 目录列表
  - 文件内容
  - 任意命令执行
  - `requestContestAWDDefenseDirectory`
  - `requestContestAWDDefenseFile`
  - `saveContestAWDDefenseFile`

---

## 3. 攻击向量

- 中栏标题固定为 `攻击向量`。
- 工具条只保留两个输入：
  - `目标题目`
  - `队伍筛选`
- 目标卡按当前题目展开，每张卡只表达：
  - 目标队伍名
  - 对应服务引用
  - 代理链路是否就绪
  - `打开`
  - Flag 输入框
  - `提交`
- 中栏空状态保持页内收口：
  - 没有题目时显示 `当前竞赛暂无可部署服务。`
  - 未选中题目时显示 `请选择目标题目后开始攻击。`
  - 当前筛选结果为空时显示 `当前题目下没有匹配的目标队伍。`
- 攻击反馈只在底部 `result-alert` 中回显本次提交结果，不额外弹出独立战术页。

---

## 4. 战场情报

- 右栏分成两段：
  - `战场情报`：排行榜前 10
  - `最近战报`：最近攻击 / 受击事件
- 排行榜行只保留：
  - 排名
  - 队伍名
  - 分数
- 最近战报项固定展示：
  - 方向：`对外攻击` / `受到攻击`
  - 时间
  - 对手队伍
  - 题目名
  - 服务引用
  - 成功 / 失败
  - 得分变化
- 右栏只承接当前快照和最近事件，不展开长时序复盘或独立情报工作台。

---

## 5. 共享交互规则

- 赛事处于 `running` 或 `frozen` 时，通过 `ScoreboardRealtimeBridge` 触发刷新。
- 防守服务选择、目标题目切换和队伍筛选都留在当前页面状态内，不通过 URL 派生子路由。
- `打开`、`SSH`、`重启`、`提交` 都以 service / target 维度的 pending key 禁止重复触发。
- SSH 复制失败时，只在当前页内提示 `复制失败，请手动选择文本`，不打开额外对话框。

---

## 代码落点

- 页面入口：
  - `code/frontend/src/views/contests/ContestDetail.vue`
- AWD 战场：
  - `code/frontend/src/components/contests/ContestAWDWorkspacePanel.vue`
  - `code/frontend/src/components/contests/awd/AWDDefenseServiceList.vue`
  - `code/frontend/src/components/contests/awd/AWDDefenseOperationsPanel.vue`
  - `code/frontend/src/components/contests/awd/AWDDefenseConnectionPanel.vue`
- 页面状态与派生：
  - `code/frontend/src/features/contest-awd-workspace/model/useContestAWDWorkspace.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/useAwdDefenseServiceSelection.ts`
  - `code/frontend/src/features/contest-awd-workspace/model/awdDefensePresentation.ts`
- 约束校验：
  - `code/frontend/src/views/contests/__tests__/contestAwdWorkspacePanelSource.test.ts`
  - `code/frontend/src/views/contests/__tests__/ContestDetail.test.ts`
