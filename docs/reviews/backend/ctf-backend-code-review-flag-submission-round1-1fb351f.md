# CTF Backend 代码 Review（flag-submission 第 1 轮）：Flag 提交与验证功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | flag-submission |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 1fb351f，10 个文件，281 行新增 |
| 变更概述 | 实现 Flag 提交与验证功能（B20），包括静态/动态 Flag 验证、防暴力破解、重复提交检测 |
| 审查基准 | `docs/architecture/backend/05-key-flows.md`（Flag 提交流程）<br>`docs/tasks/backend-task-breakdown.md`（B20 任务定义）<br>`CLAUDE.md`（代码规范） |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 时序攻击风险：Flag 验证未使用 constant-time 比较
- **文件**：`code/backend/pkg/crypto/flag.go:30`
- **问题描述**：`ValidateFlag` 函数使用 `==` 进行字符串比较，存在时序攻击风险。攻击者可以通过测量响应时间推断 Flag 的部分内容。
- **影响范围/风险**：安全漏洞，攻击者可通过时序分析逐字符爆破 Flag
- **修正建议**：使用 `crypto/subtle.ConstantTimeCompare` 进行恒定时间比较
```go
// 修改 pkg/crypto/flag.go
import "crypto/subtle"

func ValidateFlag(input, expected string) bool {
    return subtle.ConstantTimeCompare([]byte(input), []byte(expected)) == 1
}
```

#### [H2] 配置硬编码：flag_global_secret 放在 config.yaml 中
- **文件**：`code/backend/configs/config.yaml:109`
- **问题描述**：`flag_global_secret` 直接写在配置文件中，且使用明文占位符 `"change-this-in-production-use-env-var"`。这是敏感密钥，不应提交到版本控制。
- **影响范围/风险**：生产环境密钥泄露风险，动态 Flag 可被伪造
- **修正建议**：
  1. 从 `config.yaml` 中删除该字段，改为仅通过环境变量注入
  2. 在 `config.go` 的 `setDefaults` 中设置空字符串默认值
  3. 启动时检查该值是否为空，为空则报错退出
  4. 在 `.env.example` 或文档中说明需要设置 `CTF_CONTAINER_FLAG_GLOBAL_SECRET` 环境变量

#### [H3] 配置结构错误：flag_global_secret 放在 pagination 下
- **文件**：`code/backend/configs/config.yaml:109`，`code/backend/internal/config/config.go:116`
- **问题描述**：`flag_global_secret` 配置项错误地放在 `pagination` 配置块下，但在 `config.go` 中定义在 `ContainerConfig` 结构体中。这会导致配置无法正确加载。
- **影响范围/风险**：配置加载失败，动态 Flag 验证功能不可用
- **修正建议**：将 `config.yaml:109` 的 `flag_global_secret` 移动到 `container` 配置块内（第 92-104 行之间）

#### [H4] 防暴力破解机制失效：限流逻辑重复且配置不一致
- **文件**：`code/backend/internal/app/router.go:126-128`，`code/backend/internal/module/practice/service.go:66-74`
- **问题描述**：
  1. 路由层已经应用了 `RateLimitByIP` 中间件（限制 5 次/分钟）
  2. Service 层又实现了一次频率检查（使用 `CountRecentSubmissions` 查询数据库）
  3. 两层限流逻辑不一致：中间件按 IP 限流，Service 按用户+题目限流
  4. Service 层的限流依赖数据库查询，性能差且无法防止高频请求打爆数据库
- **影响范围/风险**：防暴力破解机制不完整，攻击者可通过多 IP 绕过限流
- **修正建议**：
  1. 移除路由层的通用 IP 限流（或保留作为全局保护）
  2. Service 层改用 Redis 实现 `user:{userID}:challenge:{challengeID}:submit` 计数器（`INCR` + `EXPIRE`）
  3. 限流粒度：每用户每题 5 次/分钟（与架构文档一致）

#### [H5] 重复提交检测不完整：未处理并发场景
- **文件**：`code/backend/internal/module/practice/service.go:56-64`
- **问题描述**：
  1. 先查询是否已完成（第 57 行），再创建提交记录（第 99 行），存在 TOCTOU（Time-of-Check-Time-of-Use）竞态
  2. 并发请求可能同时通过"已完成"检查，导致重复计分
  3. 架构文档要求使用数据库唯一索引兜底，但代码中未处理唯一约束冲突异常
- **影响范围/风险**：并发提交可能导致重复计分，破坏计分公平性
- **修正建议**：
  1. 在 `submissions` 表添加部分唯一索引：`UNIQUE(user_id, challenge_id) WHERE is_correct = true`
  2. 在 `CreateSubmission` 后捕获唯一约束冲突错误（GORM 错误码 23505）
  3. 冲突时返回"已完成"响应，而非报错

### 🟡 中优先级

#### [M1] 错误处理不一致：动态 Flag 验证失败时静默返回 false
- **文件**：`code/backend/internal/module/practice/service.go:84-88`
- **问题描述**：
  1. 查询实例失败时（第 84 行 `err`），代码静默忽略错误，直接判定 Flag 错误
  2. 无法区分"实例不存在"和"数据库查询失败"两种情况
  3. 缺少日志记录，排查问题困难
- **影响范围/风险**：用户体验差（数据库故障时提示"Flag 错误"而非"系统错误"），可观测性不足
- **修正建议**：
```go
instance, err := s.instanceRepo.FindByUserAndChallenge(userID, challengeID)
if err != nil {
    if err == gorm.ErrRecordNotFound {
        // 动态题目但实例不存在，判定为 Flag 错误（用户可能未启动实例）
        isCorrect = false
    } else {
        // 数据库查询失败，返回系统错误
        return nil, errcode.ErrInternal.WithCause(err)
    }
} else if instance.Nonce != "" {
    expectedFlag := crypto.GenerateDynamicFlag(userID, challengeID, s.globalSecret, instance.Nonce)
    isCorrect = crypto.ValidateFlag(flag, expectedFlag)
}
```

#### [M2] 敏感信息泄露：提交记录保存明文 Flag
- **文件**：`code/backend/internal/module/practice/service.go:95`
- **问题描述**：`submission.Flag = flag` 将用户提交的明文 Flag 存入数据库。架构文档要求"Flag 不存明文"。
- **影响范围/风险**：数据库泄露后 Flag 暴露，且错误的 Flag 也被记录（可能包含用户输入的敏感信息）
- **修正建议**：
  1. 正确的 Flag 不存储明文，仅存储 `is_correct = true` 标记
  2. 错误的 Flag 可选择性存储哈希值（用于作弊检测分析），或完全不存储
```go
submission := &model.Submission{
    UserID:      userID,
    ChallengeID: challengeID,
    Flag:        "", // 不存储明文
    FlagHash:    crypto.HashSubmittedFlag(flag), // 可选：存储哈希用于分析
    IsCorrect:   isCorrect,
    SubmittedAt: time.Now(),
}
```

#### [M3] 配置注入缺失：submitLimit 和 submitWindow 未通过配置注入
- **文件**：`code/backend/internal/app/router.go:119-121`
- **问题描述**：Service 构造函数直接使用 `cfg.RateLimit.FlagSubmit.Limit` 和 `Window`，但这两个值是全局 IP 限流配置，不是"每用户每题"的限流配置。
- **影响范围/风险**：限流配置语义混乱，无法独立调整 Flag 提交的限流策略
- **修正建议**：
  1. 在 `config.yaml` 的 `rate_limit.flag_submit` 下新增 `per_user_per_challenge_limit` 和 `per_user_per_challenge_window`
  2. 或在 `container` 配置块下新增 `flag_submit_limit` 和 `flag_submit_window`
  3. 通过配置结构体注入到 Service

#### [M4] 日志缺失：关键操作无日志记录
- **文件**：`code/backend/internal/module/practice/service.go`（整个文件）
- **问题描述**：
  1. Flag 验证成功/失败无日志
  2. 防暴力破解触发无日志
  3. 数据库操作失败无详细日志
  4. 无法追踪用户提交行为和系统异常
- **影响范围/风险**：可观测性不足，无法排查问题和检测作弊行为
- **修正建议**：在 Service 构造函数中注入 `*zap.Logger`，在关键节点记录日志：
  - Flag 提交（INFO）：`userID`, `challengeID`, `isCorrect`, `submittedAt`
  - 频率超限（WARN）：`userID`, `challengeID`, `currentCount`, `limit`
  - 验证失败（DEBUG）：`userID`, `challengeID`, `flagPrefix`（仅记录前缀，不记录完整 Flag）
  - 数据库错误（ERROR）：完整错误堆栈

#### [M5] 响应信息泄露：错误提示过于详细
- **文件**：`code/backend/internal/module/practice/service.go:59-63`
- **问题描述**：已完成题目时返回 `SubmittedAt: time.Now()`，这是当前时间而非首次提交时间，信息不准确且无意义。
- **影响范围/风险**：用户体验差，返回信息不准确
- **修正建议**：
```go
existingSubmission, err := s.repo.FindCorrectSubmission(userID, challengeID)
if err == nil {
    return &dto.SubmissionResp{
        IsCorrect:   true,
        Message:     "该题目已完成",
        SubmittedAt: existingSubmission.SubmittedAt, // 返回首次提交时间
    }, nil
}
```

#### [M6] 错误码使用不当：已完成题目应返回 409 Conflict
- **文件**：`code/backend/internal/module/practice/service.go:59-63`
- **问题描述**：已完成题目时返回 200 成功，但架构文档中定义了 `ErrAlreadySolved` 错误码（13002, 409 Conflict）。
- **影响范围/风险**：API 语义不清晰，前端无法通过 HTTP 状态码区分"首次成功"和"重复提交"
- **修正建议**：
```go
if err == nil {
    return nil, errcode.ErrAlreadySolved
}
```
前端根据错误码 13002 显示"该题目已完成"提示。

### 🟢 低优先级

#### [L1] 命名不规范：Repository 方法命名不一致
- **文件**：`code/backend/internal/module/practice/repository.go:24`
- **问题描述**：`FindCorrectSubmission` 方法名暗示返回单条记录，但实际上应该是"检查是否存在"的语义。
- **影响范围/风险**：代码可读性略差
- **修正建议**：重命名为 `ExistsCorrectSubmission(userID, challengeID int64) (bool, error)` 或保持现有命名但在注释中明确说明用途

#### [L2] 性能优化：已完成检查可使用 EXISTS 查询
- **文件**：`code/backend/internal/module/practice/repository.go:24-28`
- **问题描述**：`FindCorrectSubmission` 使用 `First()` 查询完整记录，但只需要判断是否存在。
- **影响范围/风险**：轻微性能浪费（单条记录影响不大）
- **修正建议**：
```go
func (r *Repository) ExistsCorrectSubmission(userID, challengeID int64) (bool, error) {
    var exists bool
    err := r.db.Model(&model.Submission{}).
        Select("1").
        Where("user_id = ? AND challenge_id = ? AND is_correct = ?", userID, challengeID, true).
        Limit(1).
        Find(&exists).Error
    return exists, err
}
```

#### [L3] 代码注释不足：crypto 包缺少安全说明
- **文件**：`code/backend/pkg/crypto/flag.go`
- **问题描述**：
  1. `HashStaticFlag` 使用 SHA-256 但未说明为何不用 bcrypt/argon2（因为 Flag 不是密码，不需要慢哈希）
  2. `GenerateDynamicFlag` 的 HMAC 算法未说明安全假设（依赖 globalSecret 保密性）
- **影响范围/风险**：代码可维护性略差，后续维护者可能误改算法
- **修正建议**：补充注释说明设计决策和安全假设

#### [L4] 测试覆盖缺失：核心验证逻辑无单元测试
- **文件**：整个 `practice` 模块
- **问题描述**：Flag 验证、防暴力破解、重复提交检测等核心逻辑未编写单元测试。
- **影响范围/风险**：代码质量无保障，重构时容易引入 bug
- **修正建议**：补充单元测试覆盖：
  - 静态 Flag 验证（正确/错误/格式非法）
  - 动态 Flag 验证（正确/错误/实例不存在）
  - 防暴力破解（未超限/超限）
  - 重复提交（首次/重复）
  - 并发提交（使用 goroutine 模拟）

#### [L5] 中间件职责不清：ParseChallengeID 应该是通用中间件
- **文件**：`code/backend/internal/middleware/parse_id.go`
- **问题描述**：`ParseChallengeID` 中间件只在一个路由中使用，但实现为独立中间件文件。如果后续有更多 ID 解析需求，会产生大量类似文件。
- **影响范围/风险**：代码组织略显冗余
- **修正建议**：
  1. 重命名为通用的 `ParseInt64Param(paramName, contextKey string)` 工厂函数
  2. 或保持现状，后续按需添加 `ParseImageID`、`ParseInstanceID` 等

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 5 |
| 🟡 中 | 6 |
| 🟢 低 | 5 |
| 合计 | 16 |

## 总体评价

本次实现完成了 Flag 提交与验证的基本功能，代码结构清晰，遵循了分层架构规范。但存在以下关键问题需要修复：

**必须修复（高优先级）**：
1. 时序攻击防护缺失（H1）
2. 敏感配置管理不当（H2, H3）
3. 防暴力破解机制不完整（H4）
4. 并发安全问题（H5）

**建议修复（中优先级）**：
1. 错误处理和日志记录不完善（M1, M4）
2. 敏感信息存储不当（M2）
3. 配置注入和错误码使用需优化（M3, M6）

**可选优化（低优先级）**：
1. 代码可读性和性能优化（L1, L2, L5）
2. 测试覆盖和文档补充（L3, L4）

修复高优先级问题后，该功能可进入测试阶段。中优先级问题建议在第 2 轮修复后再合并到主分支。
