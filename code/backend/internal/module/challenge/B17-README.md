# B17 - 靶场列表查询（学员视图）实现说明

## 实现内容

### 1. 扩展 DTO（internal/dto/challenge.go）
- 扩展 `ChallengeQuery` 添加 `Keyword` 和 `SortBy` 字段
- 新增 `ChallengeListItem` - 学员视图列表项
- 新增 `ChallengeDetailResp` - 学员视图详情响应

### 2. 新增 Model（internal/model/submission.go）
- `Submission` - 提交记录模型

### 3. 扩展 Repository（internal/module/challenge/repository.go）
- `ListPublished()` - 查询已发布靶场列表
- `GetSolvedStatus()` - 获取用户完成状态
- `GetSolvedCount()` - 获取完成人数
- `GetTotalAttempts()` - 获取总尝试次数

### 4. 扩展 Service（internal/module/challenge/service.go）
- 添加 Redis 依赖注入
- `ListPublishedChallenges()` - 学员视图靶场列表
- `GetPublishedChallenge()` - 学员视图靶场详情
- `getSolvedCountCached()` - 带缓存的完成人数查询（TTL 5分钟）

### 5. 扩展 Handler（internal/module/challenge/handler.go）
- `ListPublishedChallenges()` - GET /api/v1/challenges
- `GetPublishedChallenge()` - GET /api/v1/challenges/:id

### 6. 数据库迁移
- `migrations/000008_create_submissions_table.up.sql`
- `migrations/000008_create_submissions_table.down.sql`

## 功能特性

✅ 分页支持（page, size）
✅ 按分类筛选（category）
✅ 按难度筛选（difficulty）
✅ 关键词搜索（keyword）
✅ 排序支持（sort_by: created_at, difficulty）
✅ 返回用户完成状态（is_solved）
✅ 返回完成人数（solved_count，带缓存）
✅ 返回总尝试次数（total_attempts）
✅ 权限控制（仅返回已发布靶场）

## 验收标准

✅ 学员只能看到已发布靶场
✅ 筛选和排序功能正常
✅ 已完成的靶场有标识
✅ 缓存机制生效（完成人数缓存 5 分钟）
✅ 遵循分层架构
✅ Model 和 DTO 分离
✅ 代码编译通过
