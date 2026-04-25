# AWD Checker 结构化编辑器设计

## 目标

在现有后台 `ContestManage -> AWD 运维视图 -> 题目配置` 的配置对话框内，把 phase4 里临时保留的 `Checker 配置 JSON` 文本框升级成可直接操作的结构化编辑器。

本轮要解决的问题：

- 管理员不再需要手写 `put_flag / get_flag / havoc` 的嵌套 JSON
- `legacy_probe` 与 `http_standard` 都有对应的可读表单
- 保留最终写入后端的 JSON 预览，便于核对契约
- UI 继续沿用当前 CTF 管理端工作台样式，不新增页面或路由

## 非目标

- 不修改后端 `contest_challenges` 接口契约
- 不新增新的 checker 类型
- 不把 `AWDChallengeConfigDialog` 拆成独立页面
- 不做完整的 checker 在线试跑或服务联调能力

## 当前背景

phase4 已经完成：

- AWD 运维页内有 `题目配置` 面板
- 可新增/编辑 contest challenge 的 AWD 配置
- 后端已经支持 `awd_checker_type / awd_checker_config / awd_sla_score / awd_defense_score`

但当前主要交互缺口仍然明显：

- `Checker 配置 JSON` 需要管理员自己拼嵌套对象
- `http_standard` 的字段多且层级深，手写成本高
- 输入错误只能在保存阶段靠 JSON 解析兜底，无法给出字段级反馈

这会直接削弱 AWD 对抗模式“可配置、可运营”的可用性，也不利于毕业设计里强调的“集中监控与管理”。

## 方案比较

### 方案 A：保留 JSON 文本框，只补格式示例

做法：

- 继续使用原始 JSON 文本框
- 增加示例和说明

优点：

- 改动最小

缺点：

- 依然要求管理员手写嵌套对象
- 字段级校验和默认值补齐都做不好
- phase4 已经证明这只是过渡方案

不推荐。

### 方案 B：结构化表单编辑 + JSON 只读预览

做法：

- `legacy_probe` 使用简单字段编辑
- `http_standard` 使用动作级表单编辑
- 最终 JSON 改为只读预览，不再作为主要输入入口

优点：

- 对管理员最友好
- 保持后端契约透明
- 改造范围集中在前端对话框，不影响现有数据流

缺点：

- 需要补配置解析、字段默认值和表单校验逻辑

推荐采用。

### 方案 C：新建独立 checker 配置向导

做法：

- 为 checker 配置单独做多步向导或独立页面

优点：

- 表单可以做得很完整

缺点：

- 超出当前阶段最小改动范围
- 会把 AWD 运维工作流割裂成两处入口

本轮不采用。

## 决策

采用方案 B。

这轮在 `AWDChallengeConfigDialog` 内做结构化编辑器，针对当前已支持的两个 checker 类型分别提供表单：

- `legacy_probe`
- `http_standard`

同时保留只读 JSON 预览，帮助管理员确认最终写入的 payload 结构。

## 交互设计

### 基础信息区

保留 phase4 已有字段：

- `challenge_id`
- `points`
- `order`
- `is_visible`
- `awd_checker_type`
- `awd_sla_score`
- `awd_defense_score`

这部分仍保持当前对话框顶部布局。

### `legacy_probe` 配置区

显示一个轻量配置区，仅暴露：

- `health_path`

文案明确说明：

- 留空时沿用当前全局健康检查路径
- 填写后用于 legacy 探活路径覆盖

这样可以覆盖当前 legacy 路径最常见的配置需求，不再让用户手写 `{"health_path":"/healthz"}`。

### `http_standard` 配置区

按动作拆成三个小节：

1. `PUT Flag`
2. `GET Flag`
3. `Havoc`

每个动作小节内按字段组织：

- `method`
- `path`
- `expected_status`
- `headers`（JSON 文本）

额外字段：

- `PUT Flag` 有 `body_template`
- `GET Flag` 有 `expected_substring`
- `Havoc` 为可选动作，路径留空即视为未启用

### 模板预置

对 `http_standard` 提供 3 组可直接套用的预置：

- `REST /api/flag`
- `Form /flag`
- `File /flag.txt`

点击后填充表单默认值，管理员再按题目实际路径微调。

这组预置既对应“格式草案 + 样例”的文档要求，也能降低首轮配置门槛。

### JSON 预览

在配置区底部增加只读 JSON 预览：

- 动态展示当前表单将提交的 `awd_checker_config`
- 不允许直接编辑

这样可以保留和后端契约的可见对应关系，同时避免双向数据源导致的状态混乱。

## 数据流设计

本轮不改 API，只补前端表单层的解析与构建逻辑。

新增一个前端配置支持文件，负责：

1. 从已有 `awd_checker_config` 解析出表单草稿
2. 根据当前表单状态构建最终提交的 `awd_checker_config`
3. 输出字段级校验错误
4. 提供 `http_standard` 预置模板

`AWDChallengeConfigDialog` 只负责：

- 选择 checker 类型
- 渲染对应表单
- 调用支持函数构建 payload

## 校验规则

### 通用

- `points > 0`
- `order >= 0`
- `awd_sla_score >= 0`
- `awd_defense_score >= 0`

### `legacy_probe`

- `health_path` 可留空
- 不做更深层 JSON 校验

### `http_standard`

- `put_flag.path` 必填
- `get_flag.path` 必填
- `expected_status` 必须大于 0
- `headers` 若填写，必须是 JSON 对象
- `havoc` 路径为空时不写入该动作

本轮不做：

- 模板变量合法性深度校验
- URL 连通性校验
- 与真实服务实例的试跑验证

## UI 一致性约束

这轮继续遵守现有 CTF 管理端工作台语言：

- 使用现有暗色 surface 和输入控件风格
- 配置区优先使用扁平分区和分隔，不堆叠厚重卡片
- 主动作保持高亮，辅助动作保持中性
- 表单说明文案短而直接，不渲染实现说明

## 测试策略

至少覆盖：

1. 结构化配置支持逻辑
   - 能解析既有 `http_standard` 配置为表单草稿
   - 能从表单构建出后端需要的 JSON 结构
   - 预置模板输出稳定
2. 配置对话框行为
   - `http_standard` 会显示结构化字段和 JSON 预览
   - `legacy_probe` 会显示 `health_path`
   - 保存时发出正确 payload
3. 管理页集成
   - 在 AWD 题目配置对话框内可通过结构化字段新增 `http_standard`
   - 可通过结构化字段或简化字段编辑已有 checker 配置

## 验收标准

完成后应满足：

1. 管理员新增或编辑 AWD 题目时，不再需要手写原始 JSON
2. `http_standard` 的主要动作字段都能通过结构化表单配置
3. `legacy_probe` 的健康检查路径可通过简化字段编辑
4. 对话框内能看到最终将提交的 JSON 预览
5. 现有 AWD 运维页和后端接口保持兼容，视觉风格与当前 CTF 管理端一致
