# AWD Checker 校验状态设计

## 目标

在现有后台 `ContestManage -> AWD 运维视图 -> 题目配置` 这条链路上，把 `phase6` 已经具备的 checker 试跑能力升级成可持续使用的“配置可用性闭环”。

本轮要解决的问题：

- 试跑结果不再只存在于对话框临时状态里
- 管理员保存题目配置后，列表里能直接看到最近一次校验状态
- checker 配置变更后，旧校验结果会自动失效，而不是继续显示“已通过”
- 仍然保持 preview 不写正式轮次记录、不影响排行榜、不写 `awd_team_services`

## 非目标

- 不改正式 AWD 轮次巡检语义
- 不新增独立的 checker 调试页面
- 不引入真实队伍实例批量联调
- 不把 preview 结果写入 `awd_team_services` 或轮次历史

## 当前背景

phase4-6 已经完成这些能力：

- 管理端有 AWD 题目配置面板
- `legacy_probe / http_standard` 已有结构化编辑器
- 配置对话框内已经可以真实试跑 checker

但现在还有一个明显缺口：

- 试跑成功后，结果只保留在当前对话框实例中
- 关闭对话框、刷新页面或回到列表后，管理员无法判断某道题当前配置是否已经验证过
- 如果 checker 配置后来被改动，系统也不会显式提示“需要重新验证”

这会让 `phase6` 更像一次性调试工具，而不是可运营的 AWD 配置能力。

## 方案比较

### 方案 A：preview 接口直接写 `contest_challenges`

做法：

- 每次试跑成功后，preview 接口直接把最近结果写到 `contest_challenges`

优点：

- 实现路径最短

缺点：

- 创建新题目关联时，尚不存在可写入的 `contest_challenges` 行
- 编辑对话框里试跑的是“未保存草稿”，直接写现有行会造成“结果对应新配置，但配置本身还没保存”的错位
- 会破坏 `phase6` 中“preview 不写正式业务状态”的边界

不采用。

### 方案 B：preview 返回短时有效令牌，保存配置时再持久化校验状态

做法：

- preview 接口继续执行真实 checker，但额外返回一个短时有效的 `preview_token`
- token 内部绑定：
  - `contest_id`
  - `challenge_id`
  - `checker_type`
  - 归一化后的 `checker_config`
  - preview 结果快照
- 新增/更新题目配置时，前端把最近一次有效 token 一起提交
- 后端只在保存时消费 token，把校验状态和最近结果写入 `contest_challenges`

优点：

- 兼容新增与编辑两种场景
- 保持 preview 本身不落正式记录
- 后端仍然是校验状态的唯一可信来源，不需要相信前端直接上报“我已经验证通过”

缺点：

- 需要一层短期 token 存储
- 前端需要维护“当前草稿是否仍对应最近一次试跑”的本地状态

推荐采用。

### 方案 C：只在前端本地持久化

做法：

- 用 `sessionStorage` 或本地状态保存最近一次试跑结果

优点：

- 不改后端

缺点：

- 刷新或换设备后状态丢失
- 列表页无法共享结果
- 完全不能作为管理端可信状态

不采用。

## 决策

采用方案 B。

这轮新增一条“试跑令牌 -> 保存配置 -> 写入校验状态”的链路：

1. 管理员在对话框里试跑 checker
2. 后端返回 preview 结果和 `preview_token`
3. 只要当前草稿没再变更，前端保存配置时就把 token 一起提交
4. 后端消费 token，把校验状态和最近一次结果写入 `contest_challenges`
5. 题目配置列表与编辑对话框都能展示最近一次校验信息

## 数据设计

### `contest_challenges` 新增字段

- `awd_checker_validation_state`
  - 值域：
    - `pending`
    - `passed`
    - `failed`
    - `stale`
- `awd_checker_last_preview_at`
- `awd_checker_last_preview_result`

说明：

- `pending`：当前还没有可用校验结果
- `passed`：最近一次已保存的校验结果为 `service_status = up`
- `failed`：最近一次已保存的校验结果为 `down / compromised`
- `stale`：当前 checker 配置已经变更，最近一次校验结果不再对应当前配置

`awd_checker_last_preview_result` 保存完整 preview 响应快照，至少包含：

- `checker_type`
- `service_status`
- `check_result`
- `preview_context`

这样列表页和编辑对话框都可以直接复用现有结果展示 helper，不需要再拼另一套摘要结构。

## preview token 设计

### 生命周期

- 由 preview 接口生成
- 存在 Redis
- TTL 使用短期缓存策略，本轮先定为 30 分钟
- 保存配置时消费并删除

### token 内部绑定内容

- `contest_id`
- `challenge_id`
- `checker_type`
- 归一化后的 `checker_config`
- preview 响应快照
- 生成时间

### 绑定规则

保存配置时，只有下面条件同时满足，token 才会被接受：

- token 属于同一 `contest_id`
- token 对应同一 `challenge_id`
- token 内的 `checker_type` 与当前保存值一致
- token 内的归一化 `checker_config` 与当前保存值一致

不满足时：

- 配置保存仍然继续
- 但不写入新的已验证状态
- 当前题目的校验状态按普通配置变更规则处理为 `pending / stale`

这样可以避免“用旧试跑结果给新草稿背书”。

## 保存时的状态流转

### 新增题目关联

- 无 token
  - `awd_checker_validation_state = pending`
- 有有效 token
  - `passed` 或 `failed`
  - 同时写入 `last_preview_at / last_preview_result`

### 编辑已有题目

- 仅修改分值、顺序、可见性
  - 保留现有校验状态
- 修改 `checker_type` 或 `checker_config`
  - 有有效 token：写入新的 `passed / failed`
  - 无有效 token：
    - 若之前已有校验记录，置为 `stale`
    - 若之前没有校验记录，置为 `pending`

## 后端接口设计

### preview 接口返回补充

在现有：

- `checker_type`
- `service_status`
- `check_result`
- `preview_context`

之外，新增：

- `preview_token`

### 新增 / 更新题目配置请求补充

新增可选字段：

- `awd_checker_preview_token`

仅用于把当前草稿对应的最近一次试跑结果带入保存链路。

## 前端交互设计

### 列表页

在 `题目配置` 面板每行增加“校验状态”信息，至少展示：

- 状态标签：
  - `未验证`
  - `最近通过`
  - `最近失败`
  - `待重新验证`
- 最近时间
- 最近一次目标地址摘要（若存在）

视觉继续沿用当前 flat directory row，不额外叠厚卡片。

### 配置对话框

在现有试跑区基础上补两层状态：

1. 当前草稿试跑结果
   - 继续沿用 `phase6` 的即时结果区
2. 已保存校验状态
   - 编辑已有题目时，展示当前题目最近一次已保存校验结果

同时增加本地约束：

- 试跑成功后记录当前草稿签名和 `preview_token`
- 只要用户修改了 `checker_type / checker_config` 任一字段，就清空本地 token
- UI 提示当前草稿“需要重新试跑后，才能更新保存状态”

这样可以明确区分：

- “我刚刚试跑出来的临时结果”
- “当前题目已经保存下来的正式校验状态”

## UI 一致性约束

- 继续留在 `AWDChallengeConfigPanel / AWDChallengeConfigDialog` 内，不新增页面
- 列表区仍走 `workspace-directory-section` 与 flat row 结构
- 状态表达优先使用现有 badge / helper 文本语义，不做新视觉体系
- checker 结果摘要继续复用 `useAwdCheckResultPresentation`

## 测试策略

至少覆盖：

1. 后端命令与查询
   - preview 返回 `preview_token`
   - 新增题目时可消费 token 写入 `passed / failed`
   - 编辑题目修改 checker 配置但不带 token 时，状态会转 `stale`
   - 仅修改分值等非 checker 字段时，原状态保留
   - 管理端列表接口会返回校验状态与最近结果
2. 前端 API
   - preview 响应能归一化 `preview_token`
   - 新增/更新题目请求能携带 `awd_checker_preview_token`
   - 题目列表能归一化校验状态字段
3. 前端交互
   - 试跑成功后保存会带 token
   - 修改 checker 字段后 token 自动失效
   - 列表页可展示 `未验证 / 最近通过 / 最近失败 / 待重新验证`
   - 编辑已有题目时可看到最近一次已保存校验结果
