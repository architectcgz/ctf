# CTF Backend 代码 Review（instance-destroy 第 1 轮）：实例销毁与延时功能（B19）

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | instance-destroy |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 1f70503，4 个文件，10 增 / 190 删 |
| 变更概述 | 声称实现 B19 实例销毁与延时功能，但实际只注册了路由 |
| 审查基准 | `docs/architecture/backend/`、`docs/tasks/backend-task-breakdown.md` B19 部分 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 功能完全未实现，commit message 严重误导
- **文件**：`code/backend/internal/app/router.go:111-117`
- **问题描述**：
  - commit message 声称"实现实例销毁与延时功能 (B19)"
  - 实际只在路由中注册了 4 个接口：
    ```go
    protected.POST("/challenges/:id/instances", containerHandler.CreateInstance)
    protected.GET("/instances", containerHandler.ListInstances)
    protected.DELETE("/instances/:id", containerHandler.DestroyInstance)
    protected.POST("/instances/:id/extend", containerHandler.ExtendInstance)
    ```
  - 但 `internal/module/container/` 目录下根本不存在以下文件：
    - `service.go`（业务逻辑层）
    - `handler.go`（HTTP 处理层）
    - `repository.go`（数据访问层）
  - 路由中引用的 `containerModule.NewRepository`、`containerModule.NewService`、`containerModule.NewHandler` 全部不存在
  - 这会导致编译失败，无法启动服务
- **影响范围/风险**：
  - 代码无法编译通过
  - 功能完全不可用
  - commit message 与实际内容严重不符，误导后续开发者
  - 违反任务拆分原则："每个任务可独立提交、独立验收、可回滚"
- **修正建议**：
  1. 立即回滚此 commit，或修改 commit message 为"chore: 预留实例管理路由接口"
  2. 重新实现 B19 任务，必须包含：
     - `internal/module/container/repository.go`：实例数据访问层
     - `internal/module/container/service.go`：实例业务逻辑层（销毁、延时）
     - `internal/module/container/handler.go`：HTTP 处理层
     - 数据库迁移：添加 `extend_count`、`max_extends` 字段
  3. 确保代码可编译、可运行、可测试后再提交

#### [H2] 删除了正在使用的 Flag 功能代码
- **文件**：
  - `code/backend/internal/module/challenge/flag_handler.go`（已删除）
  - `code/backend/internal/module/challenge/flag_service.go`（已删除）
- **问题描述**：
  - commit message 说"删除已废弃的 flag_handler 和 flag_service 文件"
  - 但这些文件实现了 B15 任务（Flag 管理功能），包括：
    - 静态 Flag 配置
    - 动态 Flag 生成
    - Flag 验证逻辑
  - 删除后会导致 Flag 相关功能完全失效
  - 没有提供替代实现或迁移说明
- **影响范围/风险**：
  - B15 功能被破坏
  - B20（Flag 提交与验证）依赖 B15，无法继续开发
  - 违反"不要无必要地大范围重构"原则
- **修正建议**：
  1. 恢复被删除的文件：
     ```bash
     git checkout 1f70503^ -- code/backend/internal/module/challenge/flag_handler.go
     git checkout 1f70503^ -- code/backend/internal/module/challenge/flag_service.go
     ```
  2. 如果确实需要重构 Flag 功能，必须：
     - 先实现新的替代方案
     - 确保功能完整迁移
     - 单独提交，commit message 明确说明重构原因和影响范围

#### [H3] 缺少数据库迁移文件
- **文件**：无
- **问题描述**：
  - B19 任务要求扩展 Instance 表，添加：
    - `extend_count`：延时次数
    - `max_extends`：最大延时次数
  - 但 commit 中没有包含数据库迁移文件
  - 即使后续补充实现代码，也无法正常运行
- **影响范围/风险**：
  - 延时功能无法记录延时次数
  - 无法限制延时次数上限
  - 数据库结构与代码不一致
- **修正建议**：
  创建迁移文件 `migrations/000019_add_instance_extend_fields.up.sql`：
  ```sql
  ALTER TABLE instances
  ADD COLUMN extend_count INT NOT NULL DEFAULT 0,
  ADD COLUMN max_extends INT NOT NULL DEFAULT 2;
  ```

### 🟡 中优先级

#### [M1] 路由注册位置不当
- **文件**：`code/backend/internal/app/router.go:111-117`
- **问题描述**：
  - 实例管理接口注册在 `protected` 路由组（学员权限）
  - 但创建实例的接口 `POST /challenges/:id/instances` 应该在靶场详情页触发
  - 列表接口 `GET /instances` 应该返回当前用户的所有实例
  - 这些接口的权限控制逻辑不清晰
- **影响范围/风险**：
  - 权限控制可能不符合业务需求
  - 接口设计与前端交互流程不匹配
- **修正建议**：
  1. 明确接口权限：
     - 创建实例：学员可创建自己的实例
     - 列表查询：只能查看自己的实例
     - 销毁实例：只能销毁自己的实例
     - 延时实例：只能延时自己的实例
  2. 在 Service 层添加权限校验逻辑：
     ```go
     func (s *Service) DestroyInstance(userID, instanceID int64) error {
         instance, err := s.repo.FindByID(instanceID)
         if err != nil {
             return err
         }
         if instance.UserID != userID {
             return errcode.ErrForbidden()
         }
         // 执行销毁逻辑
     }
     ```

#### [M2] 缺少定时清理任务的启动代码
- **文件**：无
- **问题描述**：
  - commit message 说"定时清理任务已在 HTTPServer 启动时自动启动"
  - 但没有看到相关代码变更
  - B18 任务要求实现"过期实例自动清理（定时任务，每 5 分钟）"
  - 如果定时任务未启动，过期实例会一直占用资源
- **影响范围/风险**：
  - 过期实例无法自动清理
  - 资源泄漏（容器、网络、端口）
  - 系统资源耗尽
- **修正建议**：
  在 `cmd/server/main.go` 或 `internal/app/server.go` 中启动定时任务：
  ```go
  // 启动定时清理任务
  go func() {
      ticker := time.NewTicker(5 * time.Minute)
      defer ticker.Stop()
      for range ticker.C {
          if err := containerService.CleanupExpiredInstances(); err != nil {
              log.Error("清理过期实例失败", zap.Error(err))
          }
      }
  }()
  ```

### 🟢 低优先级

#### [L1] 删除了测试辅助函数但未说明原因
- **文件**：`code/backend/internal/module/challenge/test_helper.go:19-28`
- **问题描述**：
  - 删除了 `setupTagTestDB` 函数
  - commit message 说"修复测试辅助函数中的 Tag 模型引用问题"
  - 但删除函数不是"修复"，而是"移除"
  - 如果后续需要 Tag 相关测试，需要重新实现
- **影响范围/风险**：
  - Tag 相关测试可能无法运行
  - 测试覆盖率下降
- **修正建议**：
  1. 如果 Tag 功能已废弃，在 commit message 中明确说明
  2. 如果 Tag 功能仍在使用，保留测试辅助函数并修复引用问题

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 3 |
| 🟡 中 | 2 |
| 🟢 低 | 1 |
| 合计 | 6 |

## 总体评价

**严重问题：此 commit 不可接受，必须立即回滚或重做。**

核心问题：
1. **功能完全未实现**：只注册了路由，没有实现任何业务逻辑，代码无法编译
2. **误导性 commit message**：声称"实现功能"，实际只是"预留接口"
3. **破坏现有功能**：删除了正在使用的 Flag 功能代码
4. **缺少数据库迁移**：表结构未更新，即使补充代码也无法运行

建议处理方式：
1. 回滚此 commit
2. 恢复被删除的 Flag 功能代码
3. 按照 B19 任务要求完整实现：
   - Repository 层（数据访问）
   - Service 层（业务逻辑：销毁、延时、权限校验）
   - Handler 层（HTTP 处理）
   - 数据库迁移（extend_count、max_extends）
   - 定时清理任务启动
4. 编写单元测试验证功能
5. 确保代码可编译、可运行后再提交

**此 commit 违反了项目开发规范中的多项原则，不符合"每个任务可独立提交、独立验收、可回滚"的要求。**
