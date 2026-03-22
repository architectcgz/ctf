# CTF Backend 代码风格指南

## 1. 目标

这份文档参考 `zhi-file-service-go/docs/code-style-guide.md`，目标不是统一空格，而是统一：

- 包边界
- 文件职责
- 命名方式
- 错误处理
- 事务边界
- 日志与测试习惯

目标是让 CTF 后端在持续重构期间，新增代码仍然像同一套系统。

## 2. 总体原则

### 2.1 清晰优先于炫技

优先选择：

- 简单
- 可读
- 可调试
- 可测试

不要为了“更抽象”牺牲理解成本。

### 2.2 显式优先于隐式

优先选择：

- 明确依赖注入
- 明确参数与返回值
- 明确事务边界
- 明确 cache key 与配置来源

避免隐藏上下文和魔法行为。

### 2.3 约束优先于便利

如果实现破坏了：

- 模块边界
- 依赖方向
- 状态一致性

那它就不是合格实现。

## 3. 工具基线

后端最小工具基线：

- `gofmt`
- `goimports`
- `go test ./...`

本仓库默认不接受以下习惯：

- 格式没过先不管
- import 顺序手调但不统一
- 测试先跳过，后面再补

## 4. 命名规范

### 4.1 包名

规则：

- 全小写
- 简短
- 不带下划线
- 不带无意义后缀

正确示例：

- `contest`
- `practice`
- `teaching_readmodel`
- `infrastructure`

避免：

- `contestservice`
- `commonutil`
- `practice_api`

### 4.2 类型名

规则：

- 用清晰业务名词
- 不重复 package 语义

例如在 `practice_readmodel` 中：

- 用 `Module`
- 用 `QueryService`
- 用 `Repository`

不要写成：

- `PracticeReadmodelQueryServiceImpl`

### 4.3 方法名

规则：

- 用动词开头
- 体现业务行为

正确示例：

- `CreateContest`
- `SubmitFlag`
- `GetTimeline`
- `ListClassStudents`

避免：

- `DoContest`
- `HandleInfo`
- `ProcessData`

### 4.4 变量名

规则：

- 短作用域用短名
- 长作用域用全名

正确示例：

- `ctx`
- `tx`
- `userID`
- `challengeID`
- `practiceReadmodelService`

避免：

- `tmp1`
- `obj`
- `svc1`

## 5. 文件组织

### 5.1 一个文件一个主要关注点

不要把以下内容持续堆在同一个文件里：

- HTTP handler
- 查询拼装
- SQL 细节
- 事务控制
- 配置默认值

### 5.2 文件命名

优先按对象或用例命名：

- `handler.go`
- `query_service.go`
- `repository.go`
- `score_service.go`

避免继续新增：

- `misc.go`
- `util.go`
- `helper.go`

除非它真的是极小且边界明确的辅助文件。

## 6. 分层职责

### 6.1 `api/http`

负责：

- request 解析
- 参数校验
- 当前用户上下文提取
- response / error 映射

不负责：

- SQL 查询
- 事务编排
- 跨模块复杂业务规则

### 6.2 `application`

负责：

- use case 编排
- 事务边界
- 查询结果拼装
- cache 读写策略
- 跨 repository 调用顺序

不负责：

- HTTP 协议细节
- GORM SQL 细节
- 连接池、Redis client 初始化

### 6.3 `infrastructure`

负责：

- GORM 查询
- Redis / 外部系统适配
- SQL 投影与持久化细节

不负责：

- Gin request 解析
- 页面级或接口级错误文案
- 跨模块业务编排

## 7. 接口与契约

接口应服务于消费方，而不是为了“看起来解耦”。

优先在以下场景抽接口：

- 跨模块 contracts
- provider 替换点
- 测试隔离外部依赖

不满足这些条件时，不要为了抽象而抽象。

## 8. 错误处理与日志

### 8.1 错误处理

- `infrastructure` 返回原始错误或带上下文包装的错误
- `application` 负责收敛到 `errcode`
- 不要吞错
- 不要在多层重复包装成看不懂的错误链

### 8.2 日志

日志应打在边界和关键副作用处：

- 外部依赖失败
- 异步任务失败
- 关键状态推进
- 风险排查需要的业务节点

不要在每一层都对同一个错误重复打日志。

## 9. Context、事务与缓存

### 9.1 Context

所有可能阻塞的操作优先接收 `context.Context`：

- DB
- Redis
- 外部调用
- 异步任务入口

### 9.2 事务边界

事务边界应由 `application` 或模块 service 控制，不应散落在 handler 中。

### 9.3 缓存

- cache key 统一收敛到 `internal/constants` 或明确的 adapter 中
- TTL 必须从配置读取，不写死在业务逻辑里
- cache 失效策略要和写路径放在同一个业务决策面上

## 10. 测试规则

优先保证以下测试仍然齐全：

- 模块 service / application 的行为测试
- app/router 的装配测试
- architecture rule 测试
- 关键 HTTP / integration 测试

新增重构时，优先补：

- 契约边界测试
- 路由归属测试
- 读写职责分离后的回归测试

## 11. 增量迁移要求

这份规范不是要求一次性重写全仓，而是要求：

1. 新代码按规范写
2. 正在重构的模块优先按规范落地
3. 旧模块在被持续修改时逐步收敛
4. 不再继续制造新的大而平文件
