# 共享实例 Proof 提交流程设计

## 背景

上一轮已经把实例共享策略补到了平台里，`shared` 题可以在练习模式和竞赛模式下复用同一份实例。

但只要题目仍使用可转发的提交凭证，`shared` 就会天然带来一类风险：

- `static` 题：一个人拿到 flag，其他人直接抄提交通关
- `regex` 题：如果答案格式固定，其他人同样可以复用
- `manual_review` 题：平台不会自动阻止“抄答案提交”，只能依赖人工审核识别

当前平台已经禁止了 `shared + dynamic`，因为动态 flag 依赖实例启动期注入环境变量，而共享实例没有办法同时承载多个用户的私有 flag。

这说明还缺最后一层能力：共享环境与用户提交凭证解耦。

## 目标

本轮要补的不是“共享实例里的动态环境变量”，而是“共享实例里的用户绑定提交凭证”：

1. 新增 `flag_type = shared_proof`
2. 仅允许 `instance_sharing = shared` 的题目使用 `shared_proof`
3. 共享题完成后，由平台为当前访问用户签发一次性 proof
4. 练习和竞赛都走现有提交入口，但提交内容改为 proof
5. proof 必须绑定 `user_id + challenge_id + contest_id + instance_id`
6. proof 必须短时有效，且一次性消费

## 非目标

本轮不做下面这些内容：

- 题目服务直接回调平台“强制判题成功”
- 在共享容器内存放平台签名密钥
- 独立于实例代理之外再做一套新的会话系统
- 前端完整的题目管理 UI 重构
- 对已有 `static / regex` 共享题自动迁移

## 可选方案

### 方案 A：共享应用自行生成 proof，平台只做验签

优点：

- 共享题服务可以完全自主生成结果
- 平台提交流程改动较小

缺点：

- 平台签名能力必须进入共享容器
- 共享应用一旦被打穿，proof 体系会一起失守
- 不符合当前平台的密钥边界

不推荐。

### 方案 B：共享应用向平台换 proof，平台负责签发和消费

优点：

- 不需要把平台密钥放进共享容器
- 可以复用现有实例代理和 proxy ticket
- proof 可以做成短时、一次性、强绑定上下文

缺点：

- 需要新增 proof 存储和内部签发接口
- 需要同时改练习和竞赛提交校验

推荐采用。

### 方案 C：共享应用直接调用平台标记 solve

优点：

- 用户不需要再提交 proof
- 流程最短

缺点：

- 绕开现有提交链路、限流、记录与反馈语义
- 会把“提交流程”和“题目服务回调”变成两套判题模型

不推荐作为第一版。

## 推荐方案

采用方案 B，并尽量复用现有运行时代理链路：

1. 用户访问共享实例时，平台继续签发 `proxy ticket`
2. `proxy ticket` claims 补充当前题目和上下文信息
3. 共享题服务在完成关键步骤后，使用平台代理注入的 ticket 调平台内部接口换取一次性 `shared_proof`
4. 用户仍通过现有 `/submit` 接口提交 proof
5. 平台在提交时校验 proof 是否属于当前用户、当前题目、当前 contest 和当前实例
6. 校验成功后消费 proof，并写入现有 submission 记录

## 数据设计

### `challenges.flag_type`

新增枚举值：

- `shared_proof`

语义：

- 共享实例题目的“用户绑定提交凭证”
- 不是静态 flag，也不是实例环境变量
- 必须通过平台内部签发流程获得

约束：

- `shared_proof` 仅允许和 `instance_sharing = shared` 搭配
- `per_user / per_team` 题不允许配置为 `shared_proof`

### `shared_proofs`

新增表，用于保存已签发的 proof 元数据。

建议字段：

- `id`
- `token_hash`
- `user_id`
- `challenge_id`
- `contest_id`
- `instance_id`
- `issued_by_ticket`
- `status`
  - `issued`
  - `consumed`
  - `expired`
- `expires_at`
- `consumed_at`
- `created_at`
- `updated_at`

说明：

- 平台不直接存明文 proof，只存 hash
- contest 为空表示练习模式
- `issued_by_ticket` 用于审计和问题排查

### 索引建议

- `token_hash` 唯一索引
- `(user_id, challenge_id, contest_id, status)` 普通索引
- `expires_at` 索引，便于后续清理

## 代理票据设计

当前平台已经有 `proxy ticket`，但 claims 里只有：

- `user_id`
- `username`
- `role`
- `instance_id`

本轮扩展为：

- `user_id`
- `username`
- `role`
- `instance_id`
- `challenge_id`
- `contest_id`
- `share_scope`
- `issued_at`

原因：

- shared proof 签发时需要明确上下文
- 不应要求共享题服务自己再上传 `user_id / challenge_id / contest_id`
- proof 服务应该优先信任平台代理票据，而不是题目服务自报身份

## 内部接口设计

新增一个仅供共享题服务调用的内部接口，例如：

- `POST /api/v1/internal/runtime/shared-proofs/issue`

请求头：

- `X-CTF-Proxy-Ticket`

请求体最小化：

- 可为空，或只带调试字段

响应：

- `proof`
- `expires_at`

当前实现里，`shared_proof` 的过期时间先复用 `container.proxy_ticket_ttl`，不额外新增单独 TTL 配置项。

签发流程：

1. 平台解析并校验 proxy ticket
2. 校验对应实例仍可访问，且 `share_scope = shared`
3. 校验 challenge 的 `flag_type = shared_proof`
4. 签发 proof，保存 hash 和上下文
5. 返回 proof 给共享题服务

## 提交流程设计

### 练习模式

现有入口不变：

- `POST /api/v1/challenges/{id}/submit`

提交值仍放在：

- `flag`

但当题目 `flag_type = shared_proof` 时，平台改为：

1. 把输入值当作 proof
2. 查 proof hash
3. 校验：
   - `user_id` 匹配当前用户
   - `challenge_id` 匹配当前题目
   - `contest_id IS NULL`
   - proof 未过期
   - proof 未消费
4. 成功后消费 proof，并继续沿用现有正确提交流程

### 竞赛模式

现有竞赛提交入口不变。

当题目 `flag_type = shared_proof` 时，校验改为：

1. proof hash 命中
2. `user_id` 匹配当前用户
3. `challenge_id` 匹配当前题目
4. `contest_id` 匹配当前竞赛
5. proof 未过期、未消费
6. 成功后消费 proof，再进入现有记分逻辑

这样可以保证：

- 实例可以共享
- 提交结果仍然是用户级
- 队伍赛的计分仍由现有 contest scoring 决定

## 安全边界

本轮必须守住下面几条：

1. 平台签名能力不能进入共享容器
2. proof 必须一次性消费
3. proof 必须短期有效
4. proof 必须绑定用户、题目、contest、实例
5. 题目服务不能靠前端自报 `user_id`
6. internal issue 接口必须只信平台代理注入的 ticket

## 兼容性规则

### 题目配置

- `instance_sharing = shared`
  - 允许：`static / regex / manual_review / shared_proof`
  - 禁止：`dynamic`
- `flag_type = shared_proof`
  - 强制要求：`instance_sharing = shared`

### 前端

本轮前端只做最小同步：

- 题目详情页识别 `shared_proof`
- 共享题的提交文案可改为“提交 proof”
- 不要求新增出题端高级配置面板

## 测试策略

至少覆盖下面场景：

1. 题目配置
   - `shared_proof + per_user` 被拒绝
   - `shared_proof + shared` 允许保存
2. 签发 proof
   - 共享实例访问者可通过 ticket 换取 proof
   - 非共享实例或非共享题不能签发
3. 练习提交
   - 正确 proof 允许当前用户通过
   - 其他用户复用 proof 被拒绝
   - proof 第二次提交被拒绝
4. 竞赛提交
   - 同题同赛 proof 可用
   - 不同 contest 复用 proof 被拒绝
5. 安全边界
   - 过期 proof 被拒绝
   - 实例不匹配的 proof 被拒绝

## 分阶段后续工作

### 下一阶段

- shared proof 的后台清理任务
- internal issue 接口的速率限制和审计统计
- 前端题目详情页更明确地区分 flag / proof 文案

### 更长期

- 共享实例中的多租户 checker SDK
- 题目服务与平台间的双向鉴权
- 共享题 proof 模板和出题脚手架
