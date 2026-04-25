# CTF 平台代码 Review（team-management 第 1 轮）：组队管理功能实现

## Review 信息

| 字段 | 说明 |
|------|------|
| 变更主题 | team-management |
| 轮次 | 第 1 轮（首次审查） |
| 审查范围 | commit b1af6eb，5 个文件，488 行新增 |
| 变更概述 | 实现 CTF 竞赛的组队管理功能，包括队伍创建、加入、退出、解散 |
| 审查基准 | 项目 CLAUDE.md 中的后端开发规范 |
| 审查日期 | 2026-03-06 |
| 上轮问题数 | - |

## 问题清单

### 🔴 高优先级

#### [H1] 邀请码生成存在严重并发冲突风险
- **文件**：`code/backend/internal/module/contest/team_service.go:193-200`
- **问题描述**：
  - `generateInviteCode()` 使用 `math/rand` 生成 6 位邀请码，未初始化随机种子
  - 未使用 `crypto/rand`，随机性不足
  - 邀请码冲突时 `Create()` 会因唯一索引失败，但未重试
  - 高并发场景下可能生成重复邀请码导致创建失败
- **影响范围/风险**：
  - 多个用户同时创建队伍时可能生成相同邀请码
  - 用户创建队伍失败但不知道原因（数据库唯一约束错误）
  - 6 位字符空间仅 32^6 ≈ 10 亿，竞赛多时冲突概率上升
- **修正建议**：
```go
import (
    "crypto/rand"
    "encoding/base32"
)

func generateInviteCode() (string, error) {
    const maxRetries = 3
    for i := 0; i < maxRetries; i++ {
        bytes := make([]byte, 4)
        if _, err := rand.Read(bytes); err != nil {
            return "", err
        }
        code := base32.StdEncoding.EncodeToString(bytes)[:6]
        // 返回 code，由调用方检查唯一性并重试
        return code, nil
    }
    return "", errors.New("生成邀请码失败")
}
```
或在 Service 层添加重试逻辑：
```go
func (s *TeamService) CreateTeam(...) (*dto.TeamResp, error) {
    const maxRetries = 3
    for i := 0; i < maxRetries; i++ {
        team := &model.Team{
            InviteCode: generateInviteCode(),
            // ...
        }
        err := s.teamRepo.Create(team)
        if err == nil {
            break
        }
        // 检查是否为唯一索引冲突
        if !isUniqueViolation(err) {
            return nil, err
        }
        // 重试
    }
}
```

#### [H2] CreateTeam 存在事务一致性问题
- **文件**：`code/backend/internal/module/contest/team_service.go:50-62`
- **问题描述**：
  - 创建队伍（line 50）和添加队长成员（line 60）是两个独立操作
  - 如果 `AddMember` 失败，队伍已创建但队长未加入，数据不一致
  - 未使用数据库事务保证原子性
- **影响范围/风险**：
  - 队伍创建成功但队长不在成员列表中
  - 队伍人数统计错误（显示 0 人）
  - 队长无法管理自己创建的队伍
- **修正建议**：
```go
func (s *TeamService) CreateTeam(captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
    // 使用事务
    tx := s.teamRepo.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    team := &model.Team{...}
    if err := tx.Create(team).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    member := &model.TeamMember{...}
    if err := tx.Create(member).Error; err != nil {
        tx.Rollback()
        return nil, err
    }

    if err := tx.Commit().Error; err != nil {
        return nil, err
    }

    return s.toTeamResp(team, 1), nil
}
```
或在 Repository 层添加 `CreateWithMember(team, member)` 方法封装事务。

#### [H3] JoinTeam 存在并发竞态条件
- **文件**：`code/backend/internal/module/contest/team_service.go:82-96`
- **问题描述**：
  - 检查人数（line 83）和添加成员（line 96）之间无锁保护
  - 多个用户同时加入时，可能都通过人数检查，导致超员
  - 例如：队伍上限 4 人，当前 3 人，2 人同时加入都会成功
- **影响范围/风险**：
  - 队伍人数超过 `max_members` 限制
  - 业务规则失效
  - 影响竞赛公平性
- **修正建议**：
```go
func (s *TeamService) JoinTeam(userID int64, req *dto.JoinTeamReq) (*dto.TeamResp, error) {
    // 方案 1：使用分布式锁
    lockKey := fmt.Sprintf("team:join:%d", team.ID)
    lock := s.redisLock.Acquire(lockKey, 5*time.Second)
    defer lock.Release()

    // 方案 2：使用数据库行锁
    tx := s.teamRepo.db.Begin()
    var team model.Team
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&team, teamID)

    count := tx.Model(&model.TeamMember{}).Where("team_id = ?", team.ID).Count()
    if count >= team.MaxMembers {
        tx.Rollback()
        return nil, errcode.New(14004, "队伍人数已满", 403)
    }

    tx.Create(&model.TeamMember{...})
    tx.Commit()

    // 方案 3：使用数据库约束（推荐）
    // 在 team_members 表添加触发器或约束检查人数
}
```

#### [H4] DismissTeam 未级联删除成员记录
- **文件**：`code/backend/internal/module/contest/team_service.go:133`
- **问题描述**：
  - 解散队伍时只删除 `teams` 表记录（软删除）
  - `team_members` 表的成员记录未删除
  - 导致孤儿数据残留
- **影响范围/风险**：
  - 数据库垃圾数据累积
  - `FindUserTeamInContest` 可能查到已解散队伍的成员记录
  - 用户可能无法重新加入同一竞赛（因为旧成员记录仍存在）
- **修正建议**：
```go
func (s *TeamService) DismissTeam(captainID, teamID int64) error {
    // ...权限检查...

    // 使用事务
    tx := s.teamRepo.db.Begin()

    // 先删除所有成员
    if err := tx.Where("team_id = ?", teamID).Delete(&model.TeamMember{}).Error; err != nil {
        tx.Rollback()
        return err
    }

    // 再删除队伍
    if err := tx.Delete(&model.Team{}, teamID).Error; err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit().Error
}
```
或在数据库层添加外键级联删除：
```sql
ALTER TABLE team_members
ADD CONSTRAINT fk_team
FOREIGN KEY (team_id) REFERENCES teams(id)
ON DELETE CASCADE;
```

### 🟡 中优先级

#### [M1] FindUserTeamInContest 未过滤软删除记录
- **文件**：`code/backend/internal/module/contest/team_repository.go:57-63`
- **问题描述**：
  - JOIN 查询未添加 `teams.deleted_at IS NULL` 条件
  - 用户加入队伍后，队伍被解散（软删除），用户仍被认为在队伍中
  - 导致无法重新创建或加入新队伍
- **影响范围/风险**：
  - 用户被"幽灵队伍"锁定
  - 错误提示"您已加入该竞赛的队伍"但实际队伍已不存在
- **修正建议**：
```go
func (r *TeamRepository) FindUserTeamInContest(userID, contestID int64) (*model.Team, error) {
    var team model.Team
    err := r.db.Joins("JOIN team_members ON teams.id = team_members.team_id").
        Where("team_members.user_id = ? AND teams.contest_id = ? AND teams.deleted_at IS NULL", userID, contestID).
        First(&team).Error
    return &team, err
}
```
或使用 GORM 的 `Unscoped()` 明确处理软删除。

#### [M2] GetTeamInfo 存在 N+1 查询问题
- **文件**：`code/backend/internal/module/contest/team_service.go:150-161`
- **问题描述**：
  - 循环中逐个查询用户信息（line 152）
  - 队伍有 N 个成员就执行 N 次数据库查询
  - 性能随队伍人数线性下降
- **影响范围/风险**：
  - 大型队伍（10 人）查询慢
  - 数据库连接池压力大
  - 接口响应时间长
- **修正建议**：
```go
func (s *TeamService) GetTeamInfo(teamID int64) (*dto.TeamResp, []*dto.TeamMemberResp, error) {
    // ...

    members, err := s.teamRepo.GetMembers(teamID)
    if err != nil {
        return nil, nil, err
    }

    // 批量查询用户
    userIDs := make([]int64, len(members))
    for i, m := range members {
        userIDs[i] = m.UserID
    }
    users, err := s.userRepo.FindByIDs(userIDs) // 需要添加此方法
    if err != nil {
        return nil, nil, err
    }

    userMap := make(map[int64]*model.User)
    for _, u := range users {
        userMap[u.ID] = u
    }

    memberResps := make([]*dto.TeamMemberResp, 0, len(members))
    for _, m := range members {
        if user, ok := userMap[m.UserID]; ok {
            memberResps = append(memberResps, &dto.TeamMemberResp{
                UserID:   m.UserID,
                Username: user.Username,
                JoinedAt: m.JoinedAt,
            })
        }
    }

    return s.toTeamResp(team, len(members)), memberResps, nil
}
```

#### [M3] ListTeams 同样存在 N+1 查询
- **文件**：`code/backend/internal/module/contest/team_service.go:172-176`
- **问题描述**：
  - 循环中逐个查询队伍人数（line 174）
  - 竞赛有 N 个队伍就执行 N 次 COUNT 查询
- **影响范围/风险**：
  - 热门竞赛（100+ 队伍）列表接口极慢
  - 数据库负载高
- **修正建议**：
```go
func (s *TeamService) ListTeams(contestID int64) ([]*dto.TeamResp, error) {
    teams, err := s.teamRepo.ListByContest(contestID)
    if err != nil {
        return nil, err
    }

    // 批量查询所有队伍的人数
    teamIDs := make([]int64, len(teams))
    for i, t := range teams {
        teamIDs[i] = t.ID
    }

    countMap, err := s.teamRepo.GetMemberCountBatch(teamIDs) // 需要添加此方法
    if err != nil {
        return nil, err
    }

    result := make([]*dto.TeamResp, 0, len(teams))
    for _, team := range teams {
        count := countMap[team.ID]
        result = append(result, s.toTeamResp(team, count))
    }
    return result, nil
}
```
在 Repository 添加：
```go
func (r *TeamRepository) GetMemberCountBatch(teamIDs []int64) (map[int64]int, error) {
    type Result struct {
        TeamID int64
        Count  int
    }
    var results []Result
    err := r.db.Model(&model.TeamMember{}).
        Select("team_id, COUNT(*) as count").
        Where("team_id IN ?", teamIDs).
        Group("team_id").
        Scan(&results).Error

    countMap := make(map[int64]int)
    for _, r := range results {
        countMap[r.TeamID] = r.Count
    }
    return countMap, err
}
```

#### [M4] 错误码硬编码在 Service 层
- **文件**：`code/backend/internal/module/contest/team_service.go:34,71,79,88,107,114,124,130,140`
- **问题描述**：
  - 错误码（14001-14007）和错误消息直接写在代码中
  - 违反全局规范中的"禁止硬编码"原则
  - 错误消息重复定义，难以统一管理
- **影响范围/风险**：
  - 错误码冲突风险
  - 多语言支持困难
  - 错误消息不一致
- **修正建议**：
```go
// pkg/errcode/team.go
var (
    ErrAlreadyInTeam    = New(14001, "您已加入该竞赛的队伍", 409)
    ErrInvalidInviteCode = New(14002, "邀请码无效", 404)
    ErrTeamFull         = New(14004, "队伍人数已满", 403)
    ErrTeamNotFound     = New(14005, "队伍不存在", 404)
    ErrCaptainCannotLeave = New(14006, "队长不能退出队伍，请先解散队伍", 403)
    ErrNotCaptain       = New(14007, "只有队长可以解散队伍", 403)
)

// Service 中使用
return nil, errcode.ErrAlreadyInTeam
```

#### [M5] LeaveTeam 未检查用户是否在队伍中
- **文件**：`code/backend/internal/module/contest/team_service.go:117`
- **问题描述**：
  - 直接调用 `RemoveMember` 删除成员
  - 如果用户本不在队伍中，`DELETE` 操作影响 0 行但不报错
  - 用户无法区分"退出成功"和"本来就不在队伍"
- **影响范围/风险**：
  - 接口语义不清晰
  - 可能被恶意调用（尝试退出所有队伍）
- **修正建议**：
```go
func (s *TeamService) LeaveTeam(userID, teamID int64) error {
    team, err := s.teamRepo.FindByID(teamID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errcode.New(14005, "队伍不存在", 404)
        }
        return err
    }

    if team.CaptainID == userID {
        return errcode.New(14006, "队长不能退出队伍，请先解散队伍", 403)
    }

    // 检查用户是否在队伍中
    members, err := s.teamRepo.GetMembers(teamID)
    if err != nil {
        return err
    }

    found := false
    for _, m := range members {
        if m.UserID == userID {
            found = true
            break
        }
    }
    if !found {
        return errcode.New(14008, "您不在该队伍中", 400)
    }

    return s.teamRepo.RemoveMember(teamID, userID)
}
```

### 🟢 低优先级

#### [L1] Model 缺少数据库索引优化
- **文件**：`code/backend/internal/model/team.go:29-30`
- **问题描述**：
  - `TeamMember` 的复合索引 `idx_team_user` 定义在两个字段上
  - 但查询场景主要是 `WHERE team_id = ?` 或 `WHERE user_id = ?`
  - 复合索引 `(team_id, user_id)` 只能优化 `team_id` 开头的查询
- **影响范围/风险**：
  - `FindUserTeamInContest` 按 `user_id` 查询时索引命中率低
  - 大数据量时性能下降
- **修正建议**：
```go
type TeamMember struct {
    ID        int64     `gorm:"column:id;primaryKey"`
    TeamID    int64     `gorm:"column:team_id;index"`           // 单独索引
    UserID    int64     `gorm:"column:user_id;index"`           // 单独索引
    JoinedAt  time.Time `gorm:"column:joined_at"`
    CreatedAt time.Time `gorm:"column:created_at"`
}
```
并在 migration 中添加：
```sql
CREATE INDEX idx_team_members_team_id ON team_members(team_id);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE UNIQUE INDEX idx_team_members_team_user ON team_members(team_id, user_id);
```

#### [L2] CreateTeamReq 缺少队伍名称唯一性校验
- **文件**：`code/backend/internal/dto/team.go:8`
- **问题描述**：
  - 队伍名称只校验长度（2-50 字符）
  - 未校验同一竞赛内队伍名称是否重复
  - 可能出现多个同名队伍
- **影响范围/风险**：
  - 用户体验差（无法区分同名队伍）
  - 排行榜显示混乱
- **修正建议**：
在 `CreateTeam` 中添加检查：
```go
func (s *TeamService) CreateTeam(captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
    // 检查队伍名称是否重复
    existingTeam, err := s.teamRepo.FindByNameInContest(req.Name, req.ContestID)
    if err == nil && existingTeam.ID > 0 {
        return nil, errcode.New(14009, "队伍名称已存在", 409)
    }

    // ...
}
```
或在数据库层添加唯一约束：
```sql
CREATE UNIQUE INDEX idx_teams_contest_name ON teams(contest_id, name) WHERE deleted_at IS NULL;
```

#### [L3] Handler 层缺少用户身份验证
- **文件**：`code/backend/internal/module/contest/team_handler.go:26,43,60,76`
- **问题描述**：
  - 直接使用 `c.GetInt64("user_id")` 获取用户 ID
  - 未检查用户是否登录（`user_id` 可能为 0）
  - 未检查用户是否有权限参加该竞赛
- **影响范围/风险**：
  - 未登录用户可能绕过认证
  - 用户可能加入无权参加的竞赛队伍
- **修正建议**：
```go
func (h *TeamHandler) CreateTeam(c *gin.Context) {
    var req dto.CreateTeamReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.ValidationError(c, err)
        return
    }

    userID := c.GetInt64("user_id")
    if userID == 0 {
        response.Unauthorized(c, "请先登录")
        return
    }

    // 检查用户是否有权限参加该竞赛
    if err := h.contestService.CheckUserAccess(userID, req.ContestID); err != nil {
        response.FromError(c, err)
        return
    }

    teamResp, err := h.teamService.CreateTeam(userID, &req)
    // ...
}
```
或使用中间件统一处理认证。

#### [L4] Repository 返回值不一致
- **文件**：`code/backend/internal/module/contest/team_repository.go:21-25,27-31`
- **问题描述**：
  - `FindByID` 和 `FindByInviteCode` 即使查询失败也返回指针
  - 返回的 `team` 可能是零值结构体的指针
  - 调用方需要同时检查 `err` 和 `team.ID > 0`
- **影响范围/风险**：
  - 代码可读性差
  - 容易误用（只检查 err 不检查 ID）
- **修正建议**：
```go
func (r *TeamRepository) FindByID(id int64) (*model.Team, error) {
    var team model.Team
    err := r.db.Where("id = ?", id).First(&team).Error
    if err != nil {
        return nil, err  // 错误时返回 nil
    }
    return &team, nil
}
```

#### [L5] 缺少日志记录
- **文件**：所有 Service 方法
- **问题描述**：
  - 关键操作（创建队伍、加入队伍、解散队伍）无日志
  - 无法追踪用户行为
  - 问题排查困难
- **影响范围/风险**：
  - 运维困难
  - 安全审计缺失
- **修正建议**：
```go
func (s *TeamService) CreateTeam(captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
    log.Info("用户创建队伍",
        "user_id", captainID,
        "contest_id", req.ContestID,
        "team_name", req.Name)

    // ...

    log.Info("队伍创建成功",
        "team_id", team.ID,
        "invite_code", team.InviteCode)

    return s.toTeamResp(team, 1), nil
}
```

#### [L6] 数据库迁移缺少回滚测试
- **文件**：`code/backend/migrations/000010_create_teams_table.down.sql`
- **问题描述**：
  - down 迁移只有 `DROP TABLE`
  - 未考虑外键依赖（如果其他表引用 teams）
  - 未备份数据
- **影响范围/风险**：
  - 回滚时可能失败
  - 数据丢失风险
- **修正建议**：
```sql
-- 000010_create_teams_table.down.sql
-- 先删除依赖表
DROP TABLE IF EXISTS team_members;

-- 再删除主表
DROP TABLE IF EXISTS teams;

-- 生产环境建议先备份
-- CREATE TABLE teams_backup AS SELECT * FROM teams;
```

## 统计摘要

| 级别 | 数量 |
|------|------|
| 🔴 高 | 4 |
| 🟡 中 | 5 |
| 🟢 低 | 6 |
| 合计 | 15 |

## 总体评价

代码整体架构清晰，遵循了 Repository-Service-Handler 分层规范，Model 和 DTO 分离正确。但存在以下严重问题需要优先修复：

1. **并发安全**：邀请码生成冲突、加入队伍竞态条件是高风险问题，必须在上线前解决
2. **事务一致性**：创建队伍和解散队伍缺少事务保护，可能导致数据不一致
3. **性能问题**：N+1 查询在大规模使用时会成为瓶颈
4. **软删除处理**：多处查询未正确过滤软删除记录

建议修复所有高优先级和中优先级问题后再进行下一轮 review。低优先级问题可在后续迭代中优化。
