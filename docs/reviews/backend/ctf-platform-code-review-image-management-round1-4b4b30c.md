# ctf-platform 代码 Review（image-management 第 1 轮）：实现 B13 镜像管理功能

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | image-management |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit 4b4b30c，10 个文件，449 行新增 |
| 变更概述 | 实现镜像 CRUD 接口、Docker 镜像验证、分页搜索功能 |
| 审查基准 | `docs/tasks/backend-task-breakdown.md` B13 任务定义 |
| 审查日期 | 2026-03-05 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 错误码硬编码，违反配置外部化原则
- **文件**：`code/backend/internal/module/challenge/image_service.go:35, 47`
- **问题描述**：错误码 `12101`、`12102`、`12103` 直接硬编码在业务逻辑中，未提取为常量或配置
- **影响范围/风险**：
  - 错误码分散在多处，维护困难
  - 错误码冲突风险高
  - 违反全局 CLAUDE.md 中的"禁止硬编码"规范
- **修正建议**：
  ```go
  // pkg/errcode/image.go
  var (
      ErrImageAlreadyExists = New(12101, "镜像已存在", 409)
      ErrImageNotAccessible = New(12102, "Docker 镜像不存在或无法访问", 400)
      ErrImageNotFound      = New(12103, "镜像不存在", 404)
  )

  // 使用时
  return nil, errcode.ErrImageAlreadyExists
  ```

#### [H2] 分页默认值硬编码
- **文件**：`code/backend/internal/module/challenge/image_service.go:85`
- **问题描述**：分页默认值 `size = 20` 硬编码在代码中
- **影响范围/风险**：无法通过配置调整默认分页大小，不同环境可能需要不同值
- **修正建议**：
  ```go
  // internal/config/config.go
  type PaginationConfig struct {
      DefaultPageSize int `mapstructure:"default_page_size"`
      MaxPageSize     int `mapstructure:"max_page_size"`
  }

  // 注入到 Service
  func (s *ImageService) ListImages(query *dto.ImageQuery) (*dto.PageResult, error) {
      size := query.Size
      if size < 1 {
          size = s.config.Pagination.DefaultPageSize // 从配置读取
      }
      // ...
  }
  ```

#### [H3] Docker 客户端未注入，路由层传 nil
- **文件**：`code/backend/internal/app/router.go:91`
- **问题描述**：`NewImageService(imageRepo, nil, log)` 传入 `nil` 作为 Docker 客户端，导致镜像验证功能失效
- **影响范围/风险**：
  - 创建镜像时无法验证 Docker 镜像是否存在
  - 删除镜像时无法同步删除 Docker 镜像
  - 核心功能缺失
- **修正建议**：
  ```go
  // router.go
  dockerClient, err := docker.NewClient() // 从配置初始化
  if err != nil {
      log.Warn("Docker 客户端初始化失败，镜像验证功能将不可用", zap.Error(err))
  }
  imageService := challengeModule.NewImageService(imageRepo, dockerClient, log.Named("image_service"))
  ```

#### [H4] 删除镜像前未检查关联关系
- **文件**：`code/backend/internal/module/challenge/image_service.go:131-158`
- **问题描述**：删除镜像时未检查是否有靶场（Challenge）正在使用该镜像
- **影响范围/风险**：
  - 删除正在使用的镜像会导致靶场无法启动
  - 数据一致性问题
  - 用户体验差
- **修正建议**：
  ```go
  func (s *ImageService) DeleteImage(id int64) error {
      // 检查是否有靶场使用该镜像
      count, err := s.challengeRepo.CountByImageID(id)
      if err != nil {
          return errcode.ErrInternal.WithCause(err)
      }
      if count > 0 {
          return errcode.New(12104, fmt.Sprintf("镜像正在被 %d 个靶场使用，无法删除", count), 400)
      }

      // 继续删除逻辑...
  }
  ```

### 🟡 中优先级

#### [M1] 更新镜像时允许清空 Description
- **文件**：`code/backend/internal/module/challenge/image_service.go:116-117`
- **问题描述**：`if req.Description != ""` 逻辑导致无法清空描述字段
- **影响范围/风险**：用户无法删除已有的描述信息
- **修正建议**：
  ```go
  type UpdateImageReq struct {
      Description *string `json:"description"` // 使用指针区分"未传"和"传空值"
      Status      string  `json:"status" binding:"omitempty,oneof=pending building available failed"`
  }

  // Service 层
  if req.Description != nil {
      img.Description = *req.Description // 允许设置为空字符串
  }
  ```

#### [M2] 镜像名称未做格式校验
- **文件**：`code/backend/internal/dto/image.go:6`
- **问题描述**：镜像名称和标签未校验格式，可能包含非法字符
- **影响范围/风险**：
  - 可能导致 Docker 镜像引用失败
  - 安全风险（注入攻击）
- **修正建议**：
  ```go
  type CreateImageReq struct {
      Name string `json:"name" binding:"required,image_name"` // 自定义校验器
      Tag  string `json:"tag" binding:"required,image_tag"`
      Description string `json:"description"`
  }

  // validation/custom.go
  func ValidateImageName(fl validator.FieldLevel) bool {
      name := fl.Field().String()
      // Docker 镜像名称规则：[a-z0-9]+(?:[._-][a-z0-9]+)*
      matched, _ := regexp.MatchString(`^[a-z0-9]+(?:[._/-][a-z0-9]+)*$`, name)
      return matched
  }
  ```

#### [M3] 删除镜像的 goroutine 缺少 context 超时控制
- **文件**：`code/backend/internal/module/challenge/image_service.go:148-153`
- **问题描述**：异步删除 Docker 镜像使用 `context.Background()`，无超时控制
- **影响范围/风险**：
  - 删除操作可能永久阻塞
  - goroutine 泄漏风险
- **修正建议**：
  ```go
  go func() {
      ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
      defer cancel()

      if _, err := s.dockerClient.ImageRemove(ctx, imageRef, image.RemoveOptions{}); err != nil {
          s.logger.Warn("删除 Docker 镜像失败", zap.String("image", imageRef), zap.Error(err))
      }
  }()
  ```

#### [M4] 镜像验证超时未配置
- **文件**：`code/backend/internal/module/challenge/image_service.go:160-167`
- **问题描述**：`verifyDockerImage` 使用 `context.Background()`，无超时控制
- **影响范围/风险**：镜像验证可能长时间阻塞，影响用户体验
- **修正建议**：
  ```go
  func (s *ImageService) verifyDockerImage(imageRef string) (int64, error) {
      ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()

      inspect, _, err := s.dockerClient.ImageInspectWithRaw(ctx, imageRef)
      if err != nil {
          return 0, err
      }
      return inspect.Size, nil
  }
  ```

#### [M5] 日志记录不完整
- **文件**：`code/backend/internal/module/challenge/image_service.go:127, 156`
- **问题描述**：更新和删除操作的日志未记录变更前的状态
- **影响范围/风险**：审计追溯困难，无法回溯历史状态
- **修正建议**：
  ```go
  // 更新时
  s.logger.Info("更新镜像",
      zap.Int64("id", id),
      zap.String("old_status", oldStatus),
      zap.String("new_status", img.Status),
      zap.String("old_desc", oldDesc),
      zap.String("new_desc", img.Description))

  // 删除时
  s.logger.Info("删除镜像",
      zap.Int64("id", id),
      zap.String("name", img.Name),
      zap.String("tag", img.Tag),
      zap.Int64("size", img.Size))
  ```

#### [M6] 数据库索引设计不完整
- **文件**：`code/backend/migrations/000003_create_images_table.up.sql:14-15`
- **问题描述**：缺少 `deleted_at` 索引，软删除查询性能差
- **影响范围/风险**：列表查询时需要全表扫描已删除记录
- **修正建议**：
  ```sql
  CREATE INDEX idx_images_deleted_at ON images(deleted_at);
  -- 或使用部分索引（仅索引未删除记录）
  CREATE INDEX idx_images_active ON images(id) WHERE deleted_at IS NULL;
  ```

### 🟢 低优先级

#### [L1] PageResult 使用 interface{} 类型不安全
- **文件**：`code/backend/internal/dto/common.go:4`
- **问题描述**：`List interface{}` 缺少类型约束，运行时类型不安全
- **影响范围/风险**：前端可能收到意外类型，调试困难
- **修正建议**：
  ```go
  // 使用泛型（Go 1.18+）
  type PageResult[T any] struct {
      List  []T   `json:"list"`
      Total int64 `json:"total"`
      Page  int   `json:"page"`
      Size  int   `json:"size"`
  }

  // 或明确类型
  type ImagePageResult struct {
      List  []*dto.ImageResp `json:"list"`
      Total int64            `json:"total"`
      Page  int              `json:"page"`
      Size  int              `json:"size"`
  }
  ```

#### [L2] 错误消息未提取为常量
- **文件**：`code/backend/internal/module/challenge/image_handler.go:39, 71, 92`
- **问题描述**：错误消息 `"无效的镜像 ID"` 重复出现 3 次
- **影响范围/风险**：维护困难，国际化支持差
- **修正建议**：
  ```go
  // internal/module/challenge/errors.go
  const (
      ErrMsgInvalidImageID = "无效的镜像 ID"
  )

  // 使用时
  response.InvalidParams(c, ErrMsgInvalidImageID)
  ```

#### [L3] 镜像列表查询未考虑软删除
- **文件**：`code/backend/internal/module/challenge/image_repository.go:43`
- **问题描述**：`r.db.Model(&model.Image)` 未显式排除软删除记录（虽然 GORM 默认会处理）
- **影响范围/风险**：依赖 GORM 隐式行为，代码可读性差
- **修正建议**：
  ```go
  query := r.db.Model(&model.Image{}).Where("deleted_at IS NULL")
  // 或显式使用 Unscoped() 说明意图
  ```

#### [L4] 缺少镜像大小的人类可读格式
- **文件**：`code/backend/internal/dto/image.go:21`
- **问题描述**：镜像大小只返回字节数，前端需要自行转换
- **影响范围/风险**：前端重复实现格式化逻辑
- **修正建议**：
  ```go
  type ImageResp struct {
      // ...
      Size           int64  `json:"size"`
      SizeFormatted  string `json:"size_formatted"` // "256 MB"
  }

  func toImageResp(img *model.Image) *dto.ImageResp {
      return &dto.ImageResp{
          // ...
          Size:          img.Size,
          SizeFormatted: formatBytes(img.Size),
      }
  }
  ```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 6 |
| 🟢 低 | 4 |
| 合计 | 14 |

## 总体评价

**功能完整性**：✅ 基本功能已实现，CRUD 接口完整，分页和搜索功能正常。

**架构一致性**：✅ 遵循三层架构（Handler → Service → Repository），Model 和 DTO 正确分离。

**主要问题**：
1. **配置外部化不足**：错误码、分页默认值等硬编码
2. **Docker 集成缺失**：路由层未注入 Docker 客户端，核心功能失效
3. **数据一致性风险**：删除镜像前未检查关联关系
4. **超时控制缺失**：Docker 操作无超时保护

**改进方向**：
1. 修复 H3（Docker 客户端注入）和 H4（关联检查）后，核心功能可用
2. 补充配置外部化（H1、H2）提升可维护性
3. 完善超时控制（M3、M4）提升稳定性
4. 优化日志和索引（M5、M6）提升可观测性和性能

**建议**：优先修复 4 个高优先级问题后再合并，中低优先级问题可在后续迭代中优化。
