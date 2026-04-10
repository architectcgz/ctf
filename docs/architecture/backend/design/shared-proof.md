# 共享实例 Proof 提交流程设计

> 状态：已采用

## 1. 设计概览

共享实例模式解决了“多人启动同一道题导致容器数量线性增长”的问题，但共享运行环境本身不能解决“提交凭证可转发”的问题。

因此，平台对共享题采用独立的提交凭证类型 `shared_proof`：

- 实例仍然共享
- 提交结果仍然按用户归属
- proof 由平台签发、校验和消费
- proof 短时有效且一次性使用

这套设计把“共享环境”与“用户提交凭证”拆开，避免一个用户拿到答案后其他用户直接复用同一份提交内容通关。

## 2. 适用范围

`shared_proof` 只用于共享实例题目，并且是共享题的正式提交模式之一。

配置约束固定如下：

- `flag_type = shared_proof` 时，强制要求 `instance_sharing = shared`
- `instance_sharing = shared` 时，允许 `flag_type = static / regex / manual_review / shared_proof`
- `instance_sharing = shared` 时，禁止 `flag_type = dynamic`

## 3. 配置与数据模型

### 3.1 题目配置

`challenges.flag_type` 新增枚举值：

- `shared_proof`

它表示该题的有效提交内容不是静态 flag，也不是实例内动态 flag，而是由平台签发的用户绑定 proof。

### 3.2 `shared_proofs` 表

平台使用 `shared_proofs` 表保存 proof 元数据，明文 proof 不落库，只保存 hash。

核心字段包括：

- `token_hash`
- `user_id`
- `challenge_id`
- `contest_id`
- `instance_id`
- `issued_by_ticket`
- `status`
- `expires_at`
- `consumed_at`

状态值固定为：

- `issued`
- `consumed`
- `expired`

核心索引包括：

- `token_hash` 唯一索引
- `(user_id, challenge_id, contest_id, status)` 查询索引
- `expires_at` 清理索引

## 4. 代理票据上下文

proof 签发不依赖题目服务自报用户上下文，而是复用现有 runtime proxy ticket。

当前 ticket claims 包含：

- `user_id`
- `username`
- `role`
- `instance_id`
- `challenge_id`
- `contest_id`
- `share_scope`
- `issued_at`

这组 claims 由平台在实例访问阶段的代理链路中签发，题目服务只能携带 ticket 调平台内部接口，不能自行声明用户、题目或竞赛身份。

## 5. Proof 签发接口

平台提供内部接口：

- `POST /api/v1/internal/runtime/shared-proofs/issue`

请求方式：

- 请求头：`X-CTF-Proxy-Ticket`
- 请求体：可为空

响应字段：

- `proof`
- `expires_at`

proof 的有效时间当前复用 `container.proxy_ticket_ttl`，不单独增加新的 TTL 配置项。

## 6. Proof 签发流程

共享题服务在用户完成题目要求后，使用当前访问票据向平台换取 proof。平台按以下顺序处理：

1. 解析并校验 `X-CTF-Proxy-Ticket`
2. 校验实例存在且当前仍可访问
3. 校验 `share_scope = shared`
4. 校验题目 `flag_type = shared_proof`
5. 生成 proof 明文并计算 hash
6. 将 hash 和上下文写入 `shared_proofs`
7. 返回 proof 与过期时间

题目服务本身不持有平台签名密钥，也不直接写入提交结果。

## 7. 提交流程

### 7.1 练习模式

练习模式仍使用原有提交入口：

- `POST /api/v1/challenges/{id}/submit`

当题目 `flag_type = shared_proof` 时，后端把提交值视为 proof，并执行以下校验：

1. proof hash 命中
2. `user_id` 与当前用户一致
3. `challenge_id` 与当前题目一致
4. `contest_id IS NULL`
5. `instance_id` 与 proof 所属实例一致
6. proof 未过期
7. proof 未消费

校验通过后，proof 被标记为已消费，再沿用现有练习正确提交流程记分和落库。

### 7.2 竞赛模式

竞赛仍走现有提交入口，区别只在于校验内容改为 proof：

1. proof hash 命中
2. `user_id` 与当前用户一致
3. `challenge_id` 与当前题目一致
4. `contest_id` 与当前竞赛一致
5. `instance_id` 与 proof 所属实例一致
6. proof 未过期
7. proof 未消费

校验通过后，proof 被消费，再进入现有竞赛计分流程。

这样保证共享实例不影响提交归属，竞赛计分仍保持用户或队伍侧的既有规则。

## 8. 安全边界

`shared_proof` 的安全边界固定如下：

- 平台签发能力不能进入共享容器
- proof 必须绑定 `user_id + challenge_id + contest_id + instance_id`
- proof 必须一次性消费
- proof 必须短时有效
- proof 校验只信任平台代理票据带来的上下文
- 共享题服务不能通过前端参数冒充其他用户签发 proof

这意味着即使多个用户共享同一运行环境，提交成功仍然需要当前用户自己完成一次独立的 proof 签发。

## 9. 前端与接口表现

前端只做与提交流程直接相关的同步：

- 题目详情识别 `flag_type = shared_proof`
- 提交区文案显示为“提交 proof”
- 仍使用现有提交入口，不引入新的学员端提交流程

管理端配置层面，`shared_proof` 作为正式 flag 类型出现在题目配置中。

## 10. 运行边界

这套设计负责的是共享题的提交凭证绑定，不负责共享应用内部如何判定“用户已完成挑战”。

当前边界明确如下：

- 平台负责 proof 的签发、存储、校验和消费
- 题目服务负责在合适的业务节点向平台申请 proof
- 不引入题目服务直接回调“判题成功”的旁路接口
- 不引入平台与题目服务之间的第二套会话系统

## 11. 当前落地结果

当前后端设计已经固定为：

- `shared_proof` 是共享题的正式 flag 类型
- proof 通过 runtime proxy ticket 上下文签发
- 练习和竞赛共用现有提交流程，只替换校验逻辑
- proof 绑定用户、题目、竞赛和实例，并且一次性消费
