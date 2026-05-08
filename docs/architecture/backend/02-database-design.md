
# CTF 网络攻防靶场平台 — 数据库设计文档

> 版本：v1.0 | 日期：2026-03-01 | 状态：Current

---

## 目录

1. [设计原则与约定](#1-设计原则与约定)
2. [ER 关系概览](#2-er-关系概览)
3. [枚举值说明](#3-枚举值说明)
4. [基础模块](#4-基础模块)
5. [靶场管理模块](#5-靶场管理模块)
6. [攻防演练模块](#6-攻防演练模块)
7. [竞赛管理模块](#7-竞赛管理模块)
8. [AWD 模块](#8-awd-模块)
9. [技能评估模块](#9-技能评估模块)
10. [系统模块](#10-系统模块)
11. [JSONB 字段结构说明](#11-jsonb-字段结构说明)
12. [Redis 缓存键设计](#12-redis-缓存键设计)
13. [数据归档策略](#13-数据归档策略)

---

## 1. 设计原则与约定

| 约定项 | 规则 |
|--------|------|
| 命名风格 | 表名、字段名统一 `snake_case` |
| 主键 | 所有表使用 `BIGSERIAL` 自增主键 `id` |
| 时间戳 | 每张表包含 `created_at`、`updated_at`（`TIMESTAMPTZ`，默认 `NOW()`） |
| 软删除 | 需要软删除的表增加 `deleted_at TIMESTAMPTZ DEFAULT NULL` |
| 外键 | 定义逻辑外键关系，DDL 中显式声明 `REFERENCES`，级联策略以 `RESTRICT` 为主 |
| JSONB | 用于存储结构灵活但查询频率较低的配置数据，核心查询字段不放 JSONB |
| 索引 | 每个索引附带用途说明；复合索引遵循最左前缀原则 |
| 字符串长度 | 用户可见文本用 `VARCHAR(n)` 限制上限；大文本用 `TEXT` |
| 布尔字段 | 统一用 `BOOLEAN DEFAULT FALSE` |

---

## 2. ER 关系概览

### 核心实体关系（文字描述）

**用户与角色（RBAC）：**
- `users` 1:N `user_roles` N:1 `roles` — 用户通过关联表拥有多个角色

**靶场管理：**
- `images` 1:N `challenges` — 一个镜像可被多个靶机引用
- `challenges` 1:N `hints` — 一个靶机可配置多级提示
- `challenges` 1:N `writeups` — 一个靶机可有多篇 Writeup
- `challenges` M:N `tags`（通过 `challenge_tags`）— 靶机多维度标签
- `topologies` 1:N `topology_nodes` — 拓扑包含多个节点
- `topology_nodes` N:1 `challenges` — 节点关联靶机

**攻防演练：**
- `users` 1:N `instances` N:1 `challenges` — 用户为靶机启动容器实例
- `users` 1:N `submissions` N:1 `challenges` — 用户提交 Flag
- `submissions` N:1 `contests`（可空）— 提交可关联竞赛
- `users` 1:1 `user_scores` — 用户得分汇总
- `users` 1:N `hint_unlocks` N:1 `hints` — 用户解锁提示记录

**竞赛管理：**
- `contests` 1:N `contest_challenges` N:1 `challenges` — 竞赛包含多道题
- `contests` 1:N `teams` — 队伍归属竞赛
- `teams` 1:N `team_members` N:1 `users` — 队伍成员
- `contests` 1:N `contest_registrations` N:1 `users` — 参赛报名
- `contests` 1:N `contest_announcements` — 竞赛公告
- `contests` 1:N `cheat_reports` — 作弊检测记录

**AWD：**
- `contests` 1:N `awd_rounds` — AWD 竞赛的轮次
- `awd_rounds` + `teams` → `awd_team_services` — 每轮每队的服务状态
- `awd_rounds` 1:N `awd_attack_logs` — 攻击日志
- `awd_rounds` 1:N `awd_traffic_events` — 轮次代理流量事实表

**技能评估：**
- `skill_dimensions` 1:N `user_skill_profiles` N:1 `users` — 用户多维能力画像
- `learning_paths` 独立实体，可关联 `challenges`

**系统：**
- `audit_logs` — 独立审计流水，记录关键操作
- `notifications` N:1 `users` — 用户通知

---

## 3. 枚举值说明

> PostgreSQL 中使用 `VARCHAR` 存储枚举值（而非 `ENUM` 类型），便于后续扩展，Go 代码中用常量约束。

| 字段 | 枚举值 | 说明 |
|------|--------|------|
| `users.status` | `active`, `inactive`, `locked`, `banned` | 正常 / 未激活 / 锁定（登录失败过多）/ 封禁 |
| `roles.code` | `student`, `teacher`, `admin` | 学员 / 教师 / 管理员 |
| `images.source_type` | `registry`, `dockerfile`, `upload` | 镜像仓库拉取 / Dockerfile 构建 / 手动上传 |
| `images.status` | `pending`, `building`, `ready`, `failed`, `deprecated` | 待构建 / 构建中 / 可用 / 失败 / 已废弃 |
| `challenges.difficulty` | `beginner`, `easy`, `medium`, `hard`, `hell` | 入门 / 简单 / 中等 / 困难 / 地狱 |
| `challenges.status` | `draft`, `review`, `active`, `archived` | 草稿 / 审核中 / 上线 / 归档 |
| `challenges.flag_type` | `static`, `dynamic`, `regex` | 静态 Flag / 动态生成 / 正则匹配 |
| `tags.dimension` | `category`, `technique`, `tool`, `platform` | 题目分类 / 技术点 / 工具 / 平台 |
| `writeups.visibility` | `private`, `unlocked`, `public` | 仅作者 / 解题后可见 / 公开 |
| `instances.status` | `pending`, `creating`, `running`, `expired`, `destroying`, `destroyed`, `failed`, `crashed` | 排队中 / 创建中 / 运行中 / 已过期 / 销毁中 / 已销毁 / 创建失败 / 运行崩溃 |
| `submissions.result` | `correct`, `incorrect`, `duplicate`, `rate_limited` | 正确 / 错误 / 重复提交 / 频率限制 |
| `contests.mode` | `jeopardy`, `awd`, `awd_plus`, `king_of_hill` | 解题赛 / AWD / AWD+ / 抢占赛 |
| `contests.status` | `draft`, `published`, `registering`, `running`, `frozen`, `ended`, `cancelled`, `archived` | 草稿 / 已发布 / 报名中 / 进行中 / 已冻结 / 已结束 / 已取消 / 归档 |
| `contest_registrations.status` | `pending`, `approved`, `rejected` | 待审核 / 通过 / 拒绝 |
| `cheat_reports.type` | `flag_sharing`, `ip_collision`, `time_anomaly`, `similarity` | Flag 共享 / IP 碰撞 / 时间异常 / 答案相似 |
| `cheat_reports.status` | `pending`, `confirmed`, `dismissed` | 待处理 / 已确认 / 已驳回 |
| `awd_team_services.service_status` | `up`, `down`, `compromised` | 正常 / 宕机 / 被攻破 |
| `awd_traffic_events.source` | `runtime_proxy` | 当前平台代理访问链路 |
| `notifications.type` | `system`, `contest`, `challenge`, `team` | 系统通知 / 竞赛通知 / 靶机通知 / 队伍通知 |
| `audit_logs.action` | `login`, `logout`, `create`, `update`, `delete`, `submit`, `admin_op` | 登录 / 登出 / 创建 / 更新 / 删除 / 提交 / 管理操作 |

---

## 4. 基础模块

### 4.1 users — 用户表

```sql
CREATE TABLE users (
    id            BIGSERIAL       PRIMARY KEY,
    student_id    VARCHAR(32)     NOT NULL,                -- 学号/工号，校园唯一标识
    username      VARCHAR(64)     NOT NULL,                -- 登录用户名
    realname      VARCHAR(64)     NOT NULL,                -- 真实姓名
    class_name    VARCHAR(128)    DEFAULT NULL,            -- 班级（教师/管理员可为空）
    email         VARCHAR(255)    DEFAULT NULL,            -- 邮箱
    password_hash VARCHAR(255)    NOT NULL,                -- bcrypt 哈希
    avatar_url    VARCHAR(512)    DEFAULT NULL,            -- 头像地址
    status        VARCHAR(16)     NOT NULL DEFAULT 'active', -- active/inactive/locked/banned
    login_fail_count INT          NOT NULL DEFAULT 0,      -- 连续登录失败次数
    locked_until  TIMESTAMPTZ     DEFAULT NULL,            -- 锁定截止时间，NULL 表示未锁定
    last_login_at TIMESTAMPTZ     DEFAULT NULL,            -- 最近登录时间
    last_login_ip VARCHAR(45)     DEFAULT NULL,            -- 最近登录 IP（兼容 IPv6）
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE UNIQUE INDEX uk_users_student_id ON users(student_id) WHERE deleted_at IS NULL;
  -- 用途：学号唯一约束（软删除感知），登录/查询按学号定位
CREATE UNIQUE INDEX uk_users_username ON users(username) WHERE deleted_at IS NULL;
  -- 用途：用户名唯一约束，登录时按用户名查找
CREATE INDEX idx_users_status ON users(status);
  -- 用途：管理后台按状态筛选用户列表
CREATE INDEX idx_users_class_name ON users(class_name);
  -- 用途：按班级筛选学员
```

### 4.2 roles — 角色表

```sql
CREATE TABLE roles (
    id          BIGSERIAL       PRIMARY KEY,
    code        VARCHAR(32)     NOT NULL,                -- 角色编码：student/teacher/admin
    name        VARCHAR(64)     NOT NULL,                -- 角色显示名称
    description VARCHAR(256)    DEFAULT NULL,            -- 角色描述
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_roles_code ON roles(code);
  -- 用途：角色编码全局唯一，程序中按 code 查找角色
```

### 4.3 user_roles — 用户角色关联表

```sql
CREATE TABLE user_roles (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    BIGINT       NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_user_roles ON user_roles(user_id, role_id);
  -- 用途：防止重复授权，同时支持按 user_id 查询用户所有角色
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
  -- 用途：按角色反查拥有该角色的所有用户
```

---

## 5. 靶场管理模块

### 5.1 images — 镜像表

```sql
CREATE TABLE images (
    id           BIGSERIAL       PRIMARY KEY,
    name         VARCHAR(128)    NOT NULL,                -- 镜像名称（如 ctf/web-sqli）
    tag          VARCHAR(64)     NOT NULL DEFAULT 'latest', -- 镜像标签
    source_type  VARCHAR(16)     NOT NULL,                -- registry/dockerfile/upload
    registry_url VARCHAR(512)    DEFAULT NULL,            -- 镜像仓库地址（source_type=registry 时）
    dockerfile   TEXT            DEFAULT NULL,            -- Dockerfile 内容（source_type=dockerfile 时）
    status       VARCHAR(16)     NOT NULL DEFAULT 'pending', -- pending/building/ready/failed/deprecated
    size_bytes   BIGINT          DEFAULT NULL,            -- 镜像大小（字节）
    build_log    TEXT            DEFAULT NULL,            -- 构建日志
    checksum     VARCHAR(128)    DEFAULT NULL,            -- 镜像摘要（sha256）
    created_by   BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE UNIQUE INDEX uk_images_name_tag ON images(name, tag) WHERE deleted_at IS NULL;
  -- 用途：同名同标签镜像唯一，防止重复注册
CREATE INDEX idx_images_status ON images(status);
  -- 用途：按状态筛选镜像列表（如只看 ready 状态）
CREATE INDEX idx_images_deleted_at ON images(deleted_at);
  -- 用途：加速软删除过滤与镜像目录查询
```

### 5.2 challenges — 靶机表

```sql
CREATE TABLE challenges (
    id               BIGSERIAL       PRIMARY KEY,
    title            VARCHAR(128)    NOT NULL,                -- 靶机名称
    description      TEXT            NOT NULL,                -- 题目描述（支持 Markdown）
    category         VARCHAR(32)     NOT NULL,                -- 大类：web/pwn/reverse/crypto/misc/forensics
    difficulty       VARCHAR(16)     NOT NULL,                -- beginner/easy/medium/hard/hell
    base_score       INT             NOT NULL DEFAULT 500,    -- 基础分值
    status           VARCHAR(16)     NOT NULL DEFAULT 'draft', -- draft/review/active/archived
    image_id         BIGINT          DEFAULT NULL REFERENCES images(id) ON DELETE SET NULL,
    resource_limits  JSONB           NOT NULL DEFAULT '{}',   -- 容器资源限制（CPU/内存/网络），见 JSONB 说明
    flag_type        VARCHAR(16)     NOT NULL DEFAULT 'static', -- static/dynamic/regex
    flag_hash        VARCHAR(128)    DEFAULT NULL,            -- 静态 Flag 的 SHA-256(flag + salt) 哈希值（禁止存明文，防拖库泄露）
    flag_salt        VARCHAR(64)     DEFAULT NULL,            -- Flag 哈希盐值（每题独立随机 32 字节）
    flag_rule        VARCHAR(512)    DEFAULT NULL,            -- 动态 Flag 生成规则或正则表达式
    solve_count      INT             NOT NULL DEFAULT 0,      -- 解出人数（冗余计数，定期校准）
    attachment_url   VARCHAR(512)    DEFAULT NULL,            -- 附件下载地址
    created_by       BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE INDEX idx_challenges_status ON challenges(status) WHERE deleted_at IS NULL;
  -- 用途：前台只展示 active 状态靶机，后台按状态筛选
CREATE INDEX idx_challenges_category ON challenges(category) WHERE deleted_at IS NULL;
  -- 用途：按分类浏览靶机列表
CREATE INDEX idx_challenges_difficulty ON challenges(difficulty) WHERE deleted_at IS NULL;
  -- 用途：按难度筛选
CREATE INDEX idx_challenges_image_id ON challenges(image_id);
  -- 用途：查询某镜像被哪些靶机引用（镜像删除前检查）
```

### 5.3 tags — 标签表

```sql
CREATE TABLE tags (
    id         BIGSERIAL       PRIMARY KEY,
    name       VARCHAR(64)     NOT NULL,                -- 标签名称（如 SQL注入、栈溢出）
    dimension  VARCHAR(32)     NOT NULL DEFAULT 'category', -- 标签维度：category/technique/tool/platform
    created_at TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_tags_name_dimension ON tags(name, dimension);
  -- 用途：同维度下标签名唯一
CREATE INDEX idx_tags_dimension ON tags(dimension);
  -- 用途：按维度列出标签（如列出所有技术点标签）
```

### 5.4 challenge_tags — 靶机标签关联表

```sql
CREATE TABLE challenge_tags (
    id            BIGSERIAL    PRIMARY KEY,
    challenge_id  BIGINT       NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    tag_id        BIGINT       NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_challenge_tags ON challenge_tags(challenge_id, tag_id);
  -- 用途：防止重复打标，同时支持按 challenge_id 查询所有标签
CREATE INDEX idx_challenge_tags_tag_id ON challenge_tags(tag_id);
  -- 用途：按标签反查关联的靶机列表
```

### 5.5 topologies — 网络拓扑表

```sql
CREATE TABLE topologies (
    id          BIGSERIAL       PRIMARY KEY,
    name        VARCHAR(128)    NOT NULL,                -- 拓扑名称
    description TEXT            DEFAULT NULL,            -- 拓扑描述
    structure   JSONB           NOT NULL DEFAULT '{}',   -- 拓扑结构定义（网络段、连接关系），见 JSONB 说明
    created_by  BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE INDEX idx_topologies_created_by ON topologies(created_by);
  -- 用途：教师查看自己创建的拓扑
```

### 5.6 topology_nodes — 拓扑节点表

```sql
CREATE TABLE topology_nodes (
    id            BIGSERIAL       PRIMARY KEY,
    topology_id   BIGINT          NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    challenge_id  BIGINT          DEFAULT NULL REFERENCES challenges(id) ON DELETE SET NULL,
    node_name     VARCHAR(64)     NOT NULL,              -- 节点显示名称
    node_type     VARCHAR(32)     NOT NULL,              -- 节点类型：challenge/gateway/dns/custom
    network_segment VARCHAR(64)   DEFAULT NULL,          -- 所属网段（如 10.0.1.0/24）
    ip_address    VARCHAR(45)     DEFAULT NULL,          -- 分配的 IP 地址
    sort_order    INT             NOT NULL DEFAULT 0,    -- 排序权重
    config        JSONB           NOT NULL DEFAULT '{}', -- 节点额外配置
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_topology_nodes_topology_id ON topology_nodes(topology_id);
  -- 用途：按拓扑 ID 查询所有节点
CREATE INDEX idx_topology_nodes_challenge_id ON topology_nodes(challenge_id);
  -- 用途：查询某靶机在哪些拓扑中被引用
```

### 5.7 writeups — Writeup 表

```sql
CREATE TABLE writeups (
    id            BIGSERIAL       PRIMARY KEY,
    challenge_id  BIGINT          NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    author_id     BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title         VARCHAR(256)    NOT NULL,                -- Writeup 标题
    content       TEXT            NOT NULL,                -- 正文内容（Markdown）
    visibility    VARCHAR(16)     NOT NULL DEFAULT 'private', -- private/unlocked/public
    publish_at    TIMESTAMPTZ     DEFAULT NULL,            -- 定时公开时间（visibility=public 时生效）
    view_count    INT             NOT NULL DEFAULT 0,      -- 浏览次数
    like_count    INT             NOT NULL DEFAULT 0,      -- 点赞数
    created_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE INDEX idx_writeups_challenge_id ON writeups(challenge_id) WHERE deleted_at IS NULL;
  -- 用途：按靶机查询关联的 Writeup 列表
CREATE INDEX idx_writeups_author_id ON writeups(author_id) WHERE deleted_at IS NULL;
  -- 用途：用户查看自己撰写的 Writeup
CREATE INDEX idx_writeups_visibility ON writeups(visibility) WHERE deleted_at IS NULL;
  -- 用途：前台按可见性筛选（如只展示 public 的 Writeup）
```

### 5.8 hints — 提示表

```sql
CREATE TABLE hints (
    id              BIGSERIAL       PRIMARY KEY,
    challenge_id    BIGINT          NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    level           INT             NOT NULL,                -- 提示级别（1=最轻微，数字越大提示越明显）
    content         TEXT            NOT NULL,                -- 提示内容
    cost_percent    INT             NOT NULL DEFAULT 10,     -- 扣分比例（百分比，如 10 表示扣除该题基础分的 10%）
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_hints_challenge_level ON hints(challenge_id, level);
  -- 用途：同一靶机的提示级别唯一，同时支持按 challenge_id 查询所有提示并按 level 排序
```

---

## 6. 攻防演练模块

### 6.1 instances — 容器实例表

```sql
CREATE TABLE instances (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    challenge_id    BIGINT          NOT NULL REFERENCES challenges(id) ON DELETE RESTRICT,
    contest_id      BIGINT          DEFAULT NULL REFERENCES contests(id) ON DELETE SET NULL, -- 关联竞赛（可空，练习模式为 NULL）
    container_id    VARCHAR(128)    DEFAULT NULL,            -- Docker 容器 ID（创建成功后回填）
    network_id      VARCHAR(128)    DEFAULT NULL,            -- Docker 网络 ID
    access_url      VARCHAR(512)    DEFAULT NULL,            -- 用户访问地址（如 http://ip:port）
    instance_nonce  VARCHAR(64)     DEFAULT NULL,            -- 动态 Flag 的实例随机值（建议 base64url(32bytes)）；用于服务端重算校验
    status          VARCHAR(16)     NOT NULL DEFAULT 'pending', -- pending/creating/running/expired/destroying/destroyed/failed/crashed
    started_at      TIMESTAMPTZ     DEFAULT NULL,            -- 容器实际启动时间
    expires_at      TIMESTAMPTZ     NOT NULL,                -- 过期时间
    extend_count    INT             NOT NULL DEFAULT 0,      -- 已延时次数
    max_extend      INT             NOT NULL DEFAULT 3,      -- 最大允许延时次数
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_instances_user_id_status ON instances(user_id, status);
  -- 用途：查询用户当前运行中的实例（限制并发实例数）
CREATE INDEX idx_instances_challenge_id ON instances(challenge_id);
  -- 用途：查询某靶机的所有实例（资源统计）
CREATE INDEX idx_instances_status_expires ON instances(status, expires_at);
  -- 用途：定时任务扫描过期实例进行清理
CREATE INDEX idx_instances_container_id ON instances(container_id);
  -- 用途：容器事件回调时按 container_id 定位记录
```

### 6.1.1 instance_ports — 实例端口映射表

> 用于记录每个实例实际分配到的宿主机端口，支持多端口题目与服务重启后的端口占用恢复。

```sql
CREATE TABLE instance_ports (
    id              BIGSERIAL       PRIMARY KEY,
    instance_id     BIGINT          NOT NULL REFERENCES instances(id) ON DELETE CASCADE,
    host_port       INT             NOT NULL,                -- 宿主机端口
    container_port  INT             NOT NULL,                -- 容器内端口
    protocol        VARCHAR(8)      NOT NULL DEFAULT 'tcp',  -- tcp/udp
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_instance_ports_host ON instance_ports(host_port, protocol);
  -- 用途：保证宿主机端口不被重复分配
CREATE INDEX idx_instance_ports_instance ON instance_ports(instance_id);
  -- 用途：查询实例映射的所有端口
```

### 6.2 submissions — Flag 提交记录表

```sql
CREATE TABLE submissions (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    challenge_id    BIGINT          NOT NULL REFERENCES challenges(id) ON DELETE RESTRICT,
    contest_id      BIGINT          DEFAULT NULL REFERENCES contests(id) ON DELETE SET NULL, -- 关联竞赛（练习模式为 NULL）
    team_id         BIGINT          DEFAULT NULL,            -- 关联队伍（个人赛为 NULL）
    submitted_flag  VARCHAR(512)    NOT NULL,                -- 用户提交的 Flag 内容
    result          VARCHAR(16)     NOT NULL,                -- correct/incorrect/duplicate/rate_limited
    score           INT             NOT NULL DEFAULT 0,      -- 本次获得分数（仅 correct 时 > 0）
    submit_ip       VARCHAR(45)     NOT NULL,                -- 提交者 IP（用于作弊检测）
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_submissions_user_challenge ON submissions(user_id, challenge_id);
  -- 用途：查询用户对某靶机的提交历史，判断是否已解出
CREATE INDEX idx_submissions_contest_id ON submissions(contest_id) WHERE contest_id IS NOT NULL;
  -- 用途：按竞赛查询所有提交记录（排行榜计算）
CREATE INDEX idx_submissions_result ON submissions(result, created_at);
  -- 用途：统计正确/错误提交趋势
CREATE INDEX idx_submissions_ip ON submissions(submit_ip, created_at);
  -- 用途：作弊检测 — 同 IP 多用户提交分析
CREATE UNIQUE INDEX uk_submissions_contest_solve ON submissions(user_id, challenge_id, contest_id) WHERE result = 'correct' AND contest_id IS NOT NULL;
  -- 用途：竞赛维度「是否已解出」快速判断，200 并发下避免全表扫描；同时保证同一用户同一竞赛同一题只有一条正确记录
CREATE UNIQUE INDEX uk_submissions_practice_solve ON submissions(user_id, challenge_id) WHERE result = 'correct' AND contest_id IS NULL;
  -- 用途：练习模式「是否已解出」快速判断；同时保证同一用户同一题在练习模式下只有一条正确记录（防并发重复计分）
```

### 6.3 user_scores — 用户得分汇总表

```sql
CREATE TABLE user_scores (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_score     INT             NOT NULL DEFAULT 0,      -- 总得分（练习模式累计）
    solve_count     INT             NOT NULL DEFAULT 0,      -- 总解题数
    last_solve_at   TIMESTAMPTZ     DEFAULT NULL,            -- 最近一次解题时间（排行榜同分排序依据）
    version         INT             NOT NULL DEFAULT 0,      -- 乐观锁版本号，并发更新时 CAS 校验
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_user_scores_user_id ON user_scores(user_id);
  -- 用途：每个用户一条汇总记录，按 user_id 快速定位
CREATE INDEX idx_user_scores_ranking ON user_scores(total_score DESC, last_solve_at ASC);
  -- 用途：全站排行榜查询（分数降序，同分按最早解出排序）
```

### 6.4 hint_unlocks — 提示解锁记录表

```sql
CREATE TABLE hint_unlocks (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    hint_id         BIGINT          NOT NULL REFERENCES hints(id) ON DELETE CASCADE,
    challenge_id    BIGINT          NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    cost_score      INT             NOT NULL DEFAULT 0,      -- 实际扣除的分数
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_hint_unlocks_user_hint ON hint_unlocks(user_id, hint_id);
  -- 用途：防止重复解锁同一提示
CREATE INDEX idx_hint_unlocks_user_challenge ON hint_unlocks(user_id, challenge_id);
  -- 用途：查询用户在某靶机已解锁的所有提示
```

---

## 7. 竞赛管理模块

### 7.1 contests — 竞赛表

```sql
CREATE TABLE contests (
    id              BIGSERIAL       PRIMARY KEY,
    title           VARCHAR(128)    NOT NULL,                -- 竞赛名称
    description     TEXT            DEFAULT NULL,            -- 竞赛描述（Markdown）
    mode            VARCHAR(16)     NOT NULL,                -- jeopardy/awd/awd_plus/king_of_hill
    status          VARCHAR(16)     NOT NULL DEFAULT 'draft', -- draft/published/registering/running/frozen/ended/cancelled/archived
    is_public       BOOLEAN         NOT NULL DEFAULT TRUE,   -- 是否公开（FALSE 则仅邀请制）
    password        VARCHAR(128)    DEFAULT NULL,            -- 加密竞赛的访问密码（哈希存储）
    contest_salt_enc TEXT           DEFAULT NULL,            -- 动态 Flag 的 contest_salt（AES-GCM 加密后 base64 存储）
    team_mode       BOOLEAN         NOT NULL DEFAULT FALSE,  -- 是否为团队赛
    max_team_size   INT             DEFAULT NULL,            -- 团队赛最大队伍人数
    registration_start TIMESTAMPTZ  DEFAULT NULL,            -- 报名开始时间
    registration_end   TIMESTAMPTZ  DEFAULT NULL,            -- 报名截止时间
    start_at        TIMESTAMPTZ     NOT NULL,                -- 竞赛开始时间
    end_at          TIMESTAMPTZ     NOT NULL,                -- 竞赛结束时间
    config          JSONB           NOT NULL DEFAULT '{}',   -- 竞赛高级配置，见 JSONB 说明
    created_by      BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE INDEX idx_contests_status ON contests(status) WHERE deleted_at IS NULL;
  -- 用途：前台按状态筛选竞赛列表（进行中、即将开始等）
CREATE INDEX idx_contests_start_at ON contests(start_at) WHERE deleted_at IS NULL;
  -- 用途：按开始时间排序展示竞赛列表
CREATE INDEX idx_contests_created_by ON contests(created_by);
  -- 用途：教师查看自己创建的竞赛
```

### 7.2 contest_challenges — 竞赛题目关联表

```sql
CREATE TABLE contest_challenges (
    id              BIGSERIAL       PRIMARY KEY,
    contest_id      BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    challenge_id    BIGINT          NOT NULL REFERENCES challenges(id) ON DELETE RESTRICT,
    contest_score   INT             DEFAULT NULL,            -- 竞赛专属分值（NULL 则使用靶机基础分值）
    solve_count     INT             NOT NULL DEFAULT 0,      -- 本竞赛内解出人数
    sort_order      INT             NOT NULL DEFAULT 0,      -- 题目排序权重
    is_hidden       BOOLEAN         NOT NULL DEFAULT FALSE,  -- 是否隐藏（定时放题场景）
    unlock_at       TIMESTAMPTZ     DEFAULT NULL,            -- 定时解锁时间
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_contest_challenges ON contest_challenges(contest_id, challenge_id);
  -- 用途：同一竞赛不重复添加同一靶机
CREATE INDEX idx_contest_challenges_challenge_id ON contest_challenges(challenge_id);
  -- 用途：查询某靶机被哪些竞赛引用
```

### 7.3 teams — 队伍表

```sql
CREATE TABLE teams (
    id              BIGSERIAL       PRIMARY KEY,
    name            VARCHAR(64)     NOT NULL,                -- 队伍名称
    contest_id      BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    captain_id      BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    invite_code     VARCHAR(32)     NOT NULL,                -- 邀请码（加入队伍用）
    max_members     INT             NOT NULL DEFAULT 4,      -- 人数上限
    member_count    INT             NOT NULL DEFAULT 1,      -- 当前人数（冗余计数）
    total_score     INT             NOT NULL DEFAULT 0,      -- 队伍总分（冗余，排行榜用）
    last_solve_at   TIMESTAMPTZ     DEFAULT NULL,            -- 最近解题时间（同分排序）
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_teams_contest_name ON teams(contest_id, name);
  -- 用途：同一竞赛内队伍名唯一
CREATE UNIQUE INDEX uk_teams_invite_code ON teams(invite_code);
  -- 用途：通过邀请码查找队伍（加入队伍时）
CREATE INDEX idx_teams_captain_id ON teams(captain_id);
  -- 用途：查询用户作为队长的队伍
CREATE INDEX idx_teams_contest_score ON teams(contest_id, total_score DESC, last_solve_at ASC);
  -- 用途：竞赛队伍排行榜
```

### 7.4 team_members — 队伍成员表

```sql
CREATE TABLE team_members (
    id          BIGSERIAL       PRIMARY KEY,
    team_id     BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id     BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(16)     NOT NULL DEFAULT 'member',   -- captain/member
    joined_at   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_team_members ON team_members(team_id, user_id);
  -- 用途：防止重复加入同一队伍
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
  -- 用途：查询用户加入的所有队伍
```

### 7.5 contest_registrations — 参赛报名表

```sql
CREATE TABLE contest_registrations (
    id          BIGSERIAL       PRIMARY KEY,
    contest_id  BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    user_id     BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id     BIGINT          DEFAULT NULL REFERENCES teams(id) ON DELETE SET NULL,
    status      VARCHAR(16)     NOT NULL DEFAULT 'pending', -- pending/approved/rejected
    reviewed_by BIGINT          DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ     DEFAULT NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_contest_reg_user ON contest_registrations(contest_id, user_id);
  -- 用途：同一用户不重复报名同一竞赛
CREATE INDEX idx_contest_reg_status ON contest_registrations(contest_id, status);
  -- 用途：管理员按状态审核报名列表
```

### 7.6 contest_announcements — 竞赛公告表

```sql
CREATE TABLE contest_announcements (
    id          BIGSERIAL       PRIMARY KEY,
    contest_id  BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    title       VARCHAR(256)    NOT NULL,                -- 公告标题
    content     TEXT            NOT NULL,                -- 公告内容（Markdown）
    is_pinned   BOOLEAN         NOT NULL DEFAULT FALSE,  -- 是否置顶
    created_by  BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_contest_ann_contest ON contest_announcements(contest_id, is_pinned DESC, created_at DESC);
  -- 用途：按竞赛查询公告列表，置顶优先、时间倒序
```

### 7.7 cheat_reports — 作弊检测记录表

```sql
CREATE TABLE cheat_reports (
    id              BIGSERIAL       PRIMARY KEY,
    contest_id      BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    type            VARCHAR(32)     NOT NULL,                -- flag_sharing/ip_collision/time_anomaly/similarity
    status          VARCHAR(16)     NOT NULL DEFAULT 'pending', -- pending/confirmed/dismissed
    suspect_user_ids BIGINT[]       NOT NULL,                -- 嫌疑用户 ID 列表（PostgreSQL 数组类型）
    evidence        JSONB           NOT NULL DEFAULT '{}',   -- 证据详情，见 JSONB 说明
    description     TEXT            DEFAULT NULL,            -- 人工备注
    reviewed_by     BIGINT          DEFAULT NULL REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at     TIMESTAMPTZ     DEFAULT NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_cheat_reports_contest ON cheat_reports(contest_id, status);
  -- 用途：按竞赛查看待处理的作弊报告
CREATE INDEX idx_cheat_reports_type ON cheat_reports(type, created_at DESC);
  -- 用途：按作弊类型统计分析
```

---

## 8. AWD 模块

当前实现补充：

- AWD 赛事服务定义由 `contest_awd_services` 承接，`runtime_config / score_config / awd_checker_validation_state` 是运行态配置与 readiness 的主事实源。
- 下列运行态事实表统一以 `service_id = contest_awd_services.id` 作为主身份，并保留 `awd_challenge_id` 作为题目资产引用字段。

### 8.1 awd_rounds — AWD 轮次表

```sql
CREATE TABLE awd_rounds (
    id              BIGSERIAL       PRIMARY KEY,
    contest_id      BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    round_number    INT             NOT NULL,                -- 轮次编号（从 1 开始）
    status          VARCHAR(16)     NOT NULL DEFAULT 'pending', -- pending/running/finished
    started_at      TIMESTAMPTZ     DEFAULT NULL,            -- 本轮实际开始时间
    ended_at        TIMESTAMPTZ     DEFAULT NULL,            -- 本轮实际结束时间
    attack_score    INT             NOT NULL DEFAULT 50,     -- 本轮攻击得分
    defense_score   INT             NOT NULL DEFAULT 50,     -- 本轮防御得分
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_awd_rounds ON awd_rounds(contest_id, round_number);
  -- 用途：同一竞赛轮次编号唯一
CREATE INDEX idx_awd_rounds_status ON awd_rounds(contest_id, status);
  -- 用途：查询当前进行中的轮次
```

### 8.2 awd_team_services — 各队服务状态表

```sql
CREATE TABLE awd_team_services (
    id              BIGSERIAL       PRIMARY KEY,
    round_id        BIGINT          NOT NULL REFERENCES awd_rounds(id) ON DELETE CASCADE,
    team_id         BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    service_id      BIGINT          NOT NULL,
    awd_challenge_id BIGINT         NOT NULL REFERENCES awd_challenges(id) ON DELETE RESTRICT,
    service_status  VARCHAR(16)     NOT NULL DEFAULT 'up',   -- up/down/compromised
    check_result    TEXT            NOT NULL DEFAULT '{}',   -- 健康检查详情 JSON
    checker_type    VARCHAR(32)     NOT NULL DEFAULT '',
    attack_received INT             NOT NULL DEFAULT 0,      -- 本轮被攻击次数
    sla_score       INT             NOT NULL DEFAULT 0,      -- 本轮 SLA 得分
    defense_score   INT             NOT NULL DEFAULT 0,      -- 本轮防御得分
    attack_score    INT             NOT NULL DEFAULT 0,      -- 本轮攻击得分
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_awd_team_services ON awd_team_services(round_id, team_id, service_id);
  -- 用途：每轮每队每个服务一条记录，防止重复
CREATE INDEX idx_awd_ts_team ON awd_team_services(team_id, round_id);
  -- 用途：查询某队伍各轮次的服务状态
CREATE INDEX idx_awd_ts_round_team_service ON awd_team_services(round_id, team_id, service_id);
  -- 用途：按轮次、队伍、服务快速定位官方状态
```

### 8.3 awd_attack_logs — 攻击日志表

```sql
CREATE TABLE awd_attack_logs (
    id              BIGSERIAL       PRIMARY KEY,
    round_id        BIGINT          NOT NULL REFERENCES awd_rounds(id) ON DELETE CASCADE,
    attacker_team_id BIGINT         NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    victim_team_id  BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    service_id      BIGINT          NOT NULL,
    awd_challenge_id BIGINT         NOT NULL REFERENCES awd_challenges(id) ON DELETE RESTRICT,
    attack_type     VARCHAR(32)     NOT NULL,                -- flag_capture/service_exploit
    source          VARCHAR(32)     NOT NULL DEFAULT 'legacy',
    submitted_flag  VARCHAR(512)    DEFAULT NULL,            -- 提交的 Flag
    submitted_by_user_id BIGINT     DEFAULT NULL,
    is_success      BOOLEAN         NOT NULL DEFAULT FALSE,  -- 是否攻击成功
    score_gained    INT             NOT NULL DEFAULT 0,      -- 获得分数
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_awd_attack_round ON awd_attack_logs(round_id, attacker_team_id);
  -- 用途：按轮次查询某队的攻击记录
CREATE INDEX idx_awd_attack_victim ON awd_attack_logs(round_id, victim_team_id);
  -- 用途：按轮次查询某队被攻击记录
CREATE INDEX idx_awd_attack_round_service_success ON awd_attack_logs(round_id, attacker_team_id, victim_team_id, service_id, is_success);
  -- 用途：按轮次和服务去重成功攻击记录
CREATE INDEX idx_awd_attack_success ON awd_attack_logs(round_id, is_success) WHERE is_success = TRUE;
  -- 用途：统计每轮成功攻击次数（部分索引，只索引成功记录）
```

### 8.4 awd_traffic_events — AWD 代理流量事实表

```sql
CREATE TABLE awd_traffic_events (
    id               BIGSERIAL       PRIMARY KEY,
    contest_id       BIGINT          NOT NULL REFERENCES contests(id) ON DELETE CASCADE,
    round_id         BIGINT          NOT NULL REFERENCES awd_rounds(id) ON DELETE CASCADE,
    attacker_team_id BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    victim_team_id   BIGINT          NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    service_id       BIGINT          NOT NULL,
    awd_challenge_id BIGINT          NOT NULL REFERENCES awd_challenges(id) ON DELETE RESTRICT,
    method           VARCHAR(16)     NOT NULL,
    path             VARCHAR(1024)   NOT NULL,
    status_code      INT             NOT NULL,
    source           VARCHAR(32)     NOT NULL DEFAULT 'runtime_proxy',
    created_at       TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_awd_traffic_round_created ON awd_traffic_events(round_id, created_at DESC, id DESC);
  -- 用途：按轮次倒序拉取最新流量明细
CREATE INDEX idx_awd_traffic_round_summary ON awd_traffic_events(round_id, method, path, status_code);
  -- 用途：按轮次聚合热点路径、状态码和趋势
CREATE INDEX idx_awd_traffic_attacker ON awd_traffic_events(round_id, attacker_team_id);
  -- 用途：统计攻击方活跃度和队伍筛选
CREATE INDEX idx_awd_traffic_victim ON awd_traffic_events(round_id, victim_team_id);
  -- 用途：统计受害方热点和队伍筛选
CREATE INDEX idx_awd_traffic_service ON awd_traffic_events(service_id);
  -- 用途：按显式 AWD service 归因流量事件，避免再依赖 challenge_id 推断运行态服务
```

说明：

- 该表是 AWD 管理后台“攻击流量态势”能力的轻量事实表，只保存轮次、队伍、服务、题目和路径级摘要。
- 第一版不持久化请求体；若需要请求体预览，仍从通用审计链路读取。
- `source` 当前固定为 `runtime_proxy`，为后续接入其他流量来源预留扩展位。

---

## 9. 技能评估模块

### 9.1 skill_dimensions — 能力维度表

```sql
CREATE TABLE skill_dimensions (
    id          BIGSERIAL       PRIMARY KEY,
    name        VARCHAR(64)     NOT NULL,                -- 维度名称（如 Web安全、逆向工程、密码学）
    code        VARCHAR(32)     NOT NULL,                -- 维度编码（如 web/pwn/reverse/crypto/misc/forensics）
    description VARCHAR(256)    DEFAULT NULL,            -- 维度描述
    icon        VARCHAR(128)    DEFAULT NULL,            -- 图标标识（前端雷达图用）
    max_level   INT             NOT NULL DEFAULT 5,      -- 最高等级
    sort_order  INT             NOT NULL DEFAULT 0,      -- 显示排序
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_skill_dimensions_code ON skill_dimensions(code);
  -- 用途：按编码唯一定位维度
```

### 9.2 user_skill_profiles — 用户能力画像表

```sql
CREATE TABLE user_skill_profiles (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    dimension_id    BIGINT          NOT NULL REFERENCES skill_dimensions(id) ON DELETE CASCADE,
    score_rate      NUMERIC(5,4)    NOT NULL DEFAULT 0.0000, -- 得分率（0.0000 ~ 1.0000）
    solve_count     INT             NOT NULL DEFAULT 0,      -- 该维度解题数
    total_attempts  INT             NOT NULL DEFAULT 0,      -- 该维度总尝试数
    current_level   INT             NOT NULL DEFAULT 0,      -- 当前等级（根据 score_rate 计算）
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE UNIQUE INDEX uk_user_skill ON user_skill_profiles(user_id, dimension_id);
  -- 用途：每个用户每个维度一条记录
CREATE INDEX idx_user_skill_dimension ON user_skill_profiles(dimension_id, score_rate DESC);
  -- 用途：按维度查看排名（如 Web 安全 Top 10）
```

### 9.3 learning_paths — 学习路径表

```sql
CREATE TABLE learning_paths (
    id              BIGSERIAL       PRIMARY KEY,
    title           VARCHAR(128)    NOT NULL,                -- 路径名称（如"Web 安全入门"）
    description     TEXT            DEFAULT NULL,            -- 路径描述
    dimension_id    BIGINT          DEFAULT NULL REFERENCES skill_dimensions(id) ON DELETE SET NULL,
    difficulty      VARCHAR(16)     NOT NULL DEFAULT 'easy', -- beginner/easy/medium/hard/hell
    challenge_ids   BIGINT[]        NOT NULL DEFAULT '{}',   -- 有序靶机 ID 列表（PostgreSQL 数组）
    estimated_hours NUMERIC(5,1)    DEFAULT NULL,            -- 预计学习时长（小时）
    is_published    BOOLEAN         NOT NULL DEFAULT FALSE,  -- 是否已发布
    created_by      BIGINT          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ     DEFAULT NULL
);

-- 索引
CREATE INDEX idx_learning_paths_dimension ON learning_paths(dimension_id) WHERE deleted_at IS NULL;
  -- 用途：按能力维度筛选学习路径
CREATE INDEX idx_learning_paths_published ON learning_paths(is_published, difficulty) WHERE deleted_at IS NULL;
  -- 用途：前台展示已发布的学习路径，按难度筛选
```

---

## 10. 系统模块

### 10.1 audit_logs — 审计日志表

> 审计日志为追加写入（append-only），不做软删除，不做 UPDATE。

```sql
CREATE TABLE audit_logs (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          DEFAULT NULL,            -- 操作用户（系统自动操作时为 NULL）
    action          VARCHAR(32)     NOT NULL,                -- login/logout/create/update/delete/submit/admin_op
    resource_type   VARCHAR(64)     NOT NULL,                -- 操作对象类型（如 challenge/contest/user）
    resource_id     BIGINT          DEFAULT NULL,            -- 操作对象 ID
    detail          JSONB           NOT NULL DEFAULT '{}',   -- 操作详情（变更前后值等）
    ip_address      VARCHAR(45)     NOT NULL,                -- 操作者 IP
    user_agent      VARCHAR(512)    DEFAULT NULL,            -- 浏览器 User-Agent
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
  -- 用途：查询某用户的操作历史
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, created_at DESC);
  -- 用途：查询某资源的变更历史（如某靶机的所有操作记录）
CREATE INDEX idx_audit_logs_action ON audit_logs(action, created_at DESC);
  -- 用途：按操作类型筛选（如查看所有登录记录）
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at);
  -- 用途：按时间范围查询 + 归档任务扫描
```

### 10.2 notifications — 通知表

```sql
CREATE TABLE notifications (
    id              BIGSERIAL       PRIMARY KEY,
    user_id         BIGINT          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            VARCHAR(16)     NOT NULL,                -- system/contest/challenge/team
    title           VARCHAR(256)    NOT NULL,                -- 通知标题
    content         TEXT            NOT NULL,                -- 通知内容
    is_read         BOOLEAN         NOT NULL DEFAULT FALSE,  -- 是否已读
    link            VARCHAR(512)    DEFAULT NULL,            -- 点击跳转链接
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    read_at         TIMESTAMPTZ     DEFAULT NULL             -- 阅读时间
);

-- 索引
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read, created_at DESC);
  -- 用途：查询用户未读通知列表（最常用查询路径）
CREATE INDEX idx_notifications_user_type ON notifications(user_id, type, created_at DESC);
  -- 用途：按通知类型筛选
```

---

## 11. JSONB 字段结构说明

### 11.1 challenges.resource_limits — 容器资源限制

```json
{
  "cpu_limit": "1.0",
  "memory_limit": "512Mi",
  "disk_limit": "1Gi",
  "network_limit": {
    "bandwidth_in": "10Mbps",
    "bandwidth_out": "10Mbps",
    "allow_internet": false
  },
  "port_mappings": [
    { "container_port": 80, "protocol": "tcp" },
    { "container_port": 22, "protocol": "tcp" }
  ],
  "env_vars": {
    "FLAG": "{{FLAG}}",
    "DIFFICULTY": "medium"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `cpu_limit` | string | 否 | CPU 核数限制，默认 "1.0" |
| `memory_limit` | string | 否 | 内存限制，K8s 格式，默认 "512Mi" |
| `disk_limit` | string | 否 | 磁盘限制，默认 "1Gi" |
| `network_limit.bandwidth_in` | string | 否 | 入站带宽限制 |
| `network_limit.bandwidth_out` | string | 否 | 出站带宽限制 |
| `network_limit.allow_internet` | bool | 否 | 是否允许访问外网，默认 false |
| `port_mappings` | array | 否 | 端口映射列表 |
| `env_vars` | object | 否 | 环境变量，`{{FLAG}}` 为动态 Flag 占位符 |

### 11.2 topologies.structure — 拓扑结构定义

```json
{
  "networks": [
    {
      "name": "dmz",
      "subnet": "10.0.1.0/24",
      "gateway": "10.0.1.1"
    },
    {
      "name": "internal",
      "subnet": "10.0.2.0/24",
      "gateway": "10.0.2.1"
    }
  ],
  "connections": [
    {
      "from_node": "gateway-1",
      "to_node": "web-server",
      "network": "dmz",
      "bidirectional": true
    }
  ],
  "firewall_rules": [
    {
      "network": "internal",
      "direction": "inbound",
      "allow_from": ["dmz"],
      "ports": [22, 3306]
    }
  ]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `networks` | array | 是 | 网络段定义列表 |
| `networks[].name` | string | 是 | 网段名称 |
| `networks[].subnet` | string | 是 | CIDR 格式子网 |
| `networks[].gateway` | string | 否 | 网关地址 |
| `connections` | array | 是 | 节点间连接关系 |
| `connections[].from_node` | string | 是 | 源节点名称（对应 topology_nodes.node_name） |
| `connections[].to_node` | string | 是 | 目标节点名称 |
| `connections[].bidirectional` | bool | 否 | 是否双向连接，默认 true |
| `firewall_rules` | array | 否 | 防火墙规则 |

### 11.3 contests.config — 竞赛高级配置

```json
{
  "scoring": {
    "type": "dynamic",
    "min_score": 100,
    "max_score": 1000,
    "decay_factor": 20
  },
  "freeze_scoreboard_at": "2026-03-15T17:00:00Z",
  "allow_hint": true,
  "max_instances_per_user": 3,
  "submission_rate_limit": {
    "max_attempts": 10,
    "window_seconds": 60
  },
  "awd_config": {
    "round_duration_seconds": 300,
    "total_rounds": 20,
    "check_interval_seconds": 30,
    "flag_rotate_per_round": true
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `scoring.type` | string | 否 | 计分方式：`static`（固定分）/ `dynamic`（动态递减） |
| `scoring.min_score` | int | 否 | 动态计分最低分 |
| `scoring.max_score` | int | 否 | 动态计分最高分（初始分） |
| `scoring.decay_factor` | int | 否 | 衰减因子（解出人数达到此值时分数降至最低） |
| `freeze_scoreboard_at` | string | 否 | 封榜时间（ISO 8601），NULL 表示不封榜 |
| `allow_hint` | bool | 否 | 是否允许使用提示，默认 true |
| `max_instances_per_user` | int | 否 | 每用户最大并发实例数 |
| `submission_rate_limit` | object | 否 | 提交频率限制 |
| `awd_config` | object | 否 | AWD 模式专属配置（仅 mode=awd/awd_plus 时有效） |

实现备注（2026-03-10）：

- 上述 `contests.config.awd_config` 仍属于目标数据设计，当前后端代码尚未把 AWD 轮次参数持久化到 `contests.config`。
- 现阶段实际生效的是全局配置 `contest.awd.scheduler_interval`、`contest.awd.scheduler_batch_size`、`contest.awd.round_interval`、`contest.awd.round_lock_ttl`，用于后台自动轮次推进。

### 11.4 cheat_reports.evidence — 作弊证据详情

```json
{
  "type": "flag_sharing",
  "submissions": [
    {
      "user_id": 101,
      "username": "alice",
      "submit_ip": "192.168.1.50",
      "submitted_at": "2026-03-15T14:30:00Z",
      "flag_fingerprint": "sha256:3b3e... (仅指纹/脱敏，禁止落库明文)"
    },
    {
      "user_id": 202,
      "username": "bob",
      "submit_ip": "192.168.1.50",
      "submitted_at": "2026-03-15T14:30:05Z",
      "flag_fingerprint": "sha256:3b3e... (仅指纹/脱敏，禁止落库明文)"
    }
  ],
  "similarity_score": 0.98,
  "detection_rule": "same_ip_same_flag_within_60s",
  "auto_detected": true
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `type` | string | 是 | 与 cheat_reports.type 一致 |
| `submissions` | array | 是 | 涉及的提交记录快照 |
| `similarity_score` | float | 否 | 相似度评分（0~1），similarity 类型时使用 |
| `detection_rule` | string | 是 | 触发的检测规则名称 |
| `auto_detected` | bool | 是 | 是否为系统自动检测（false 表示人工举报） |

---

## 12. Redis 缓存键设计

> 键名统一前缀 `ctf:`，使用冒号 `:` 分隔层级。所有 TTL 均通过配置文件注入，以下为推荐默认值。

### 12.1 用户与认证

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:auth:session:{session_id}` | STRING (session json) | 7d | 服务端 session 存储；登出或强制下线时删除 |
| `ctf:login_fail:{username}` | STRING (int 计数) | 30min | 登录失败计数，达到阈值后触发账户锁定 |
| `ctf:user:profile:{user_id}` | HASH | 30min | 用户基本信息缓存（减少 DB 查询） |
| `ctf:user:roles:{user_id}` | SET | 30min | 用户角色列表缓存 |

### 12.2 靶场与靶机

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:challenge:detail:{challenge_id}` | HASH | 1h | 靶机详情缓存（名称、描述、难度、分值等） |
| `ctf:challenge:list:active` | ZSET (score=sort_order, member=challenge_id) | 10min | 已上线靶机列表缓存，前台首页/列表页使用 |
| `ctf:challenge:solve_count:{challenge_id}` | STRING (int) | 无过期 | 靶机解出人数计数器（原子递增，定期与 DB 校准） |
| `ctf:challenge:flag:{challenge_id}` | HASH (field: hash, salt) | 1h | 静态 Flag 哈希+盐值缓存（验证时比对哈希，禁止缓存明文） |
| `ctf:image:status:{image_id}` | STRING | 5min | 镜像构建状态缓存（轮询构建进度时使用） |

### 12.3 容器实例

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:instance:user:{user_id}` | HASH (field=challenge_id, value=instance_id) | 与实例过期时间对齐 | 用户当前运行实例映射，快速判断是否已有实例 |
| `ctf:instance:count:{user_id}` | STRING (int) | 无过期 | 用户并发实例计数器（原子操作，创建+1 销毁-1） |
| `ctf:instance:nonce:{instance_id}` | STRING | 与实例过期时间对齐 | 实例 nonce 缓存（如不落库时使用）；动态 Flag 验证通过重算完成，禁止缓存/存储明文 Flag |
| `ctf:instance:expire_queue` | ZSET (score=expire_timestamp, member=instance_id) | 无过期 | 实例过期队列，定时任务扫描清理 |

### 12.4 提交与限流

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:submit:rate:{user_id}:{challenge_id}` | STRING (int 计数) | 60s（滑动窗口） | Flag 提交频率限制，防止暴力猜测 |
| `ctf:submit:solved:{user_id}` | SET (member=challenge_id) | 无过期 | 用户已解出的靶机集合，快速判断是否重复提交 |
| `ctf:submit:lock:{user_id}:{challenge_id}` | STRING ("1") | 5s | 提交分布式锁，防止同一用户对同一题并发提交 |
| `ctf:cheat:flag:{contest_id}:{challenge_id}` | HASH (field=flag_hash) | 至竞赛结束 | 作弊检测索引：记录某 Flag 哈希的首次提交元信息（避免存明文 Flag） |

### 12.5 竞赛与排行榜

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:contest:detail:{contest_id}` | HASH | 10min | 竞赛详情缓存 |
| `ctf:contest:challenges:{contest_id}` | LIST (JSON 序列化) | 5min | 竞赛题目列表缓存 |
| `ctf:rank:global` | ZSET (score=composite_score, member=user_id) | 无过期 | 全站排行榜（实时更新） |
| `ctf:rank:contest:{contest_id}:user` | ZSET (score=composite_score, member=user_id) | 无过期 | 竞赛个人排行榜 |
| `ctf:rank:contest:{contest_id}:team` | ZSET (score=composite_score, member=team_id) | 无过期 | 竞赛队伍排行榜 |
| `ctf:rank:contest:{contest_id}:frozen` | STRING (JSON 快照) | 至竞赛结束 | 封榜后的排行榜快照 |
| `ctf:awd:round:lock:{contest_id}:{round_number}` | STRING ("1") | 30s | AWD 轮次推进幂等锁，防止多实例重复创建/推进同一轮 |

> **排行榜同分排序（composite_score 编码规则）：**
>
> Redis ZSET 的 score 为 float64，仅存 `total_score` 无法处理同分排序（CTF 规则：同分时最后一次正确提交时间更早的排名靠前）。
> 采用复合编码将得分和时间戳压缩到一个 float64 中：
>
> ```
> composite_score = total_score * 1e10 + (MAX_TIMESTAMP - last_solve_timestamp)
> ```
>
> - `total_score`：用户/队伍总得分（整数部分）
> - `MAX_TIMESTAMP`：一个足够大的固定值（如 `9999999999`，即 2286 年）
> - `last_solve_timestamp`：最后一次正确提交的 Unix 秒级时间戳
> - 效果：`ZREVRANGE` 降序排列时，得分高的在前；同分时 `MAX_TS - last_solve_ts` 更大（即提交更早）的在前
> - 精度：float64 有效精度约 15-16 位，`total_score` 不超过 99999（5 位）+ 时间戳 10 位 = 15 位，精度足够

### 12.6 AWD 实时状态

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:awd:{contest_id}:current_round` | STRING (int) | 无过期 | 当前轮次编号（轮次切换时更新） |
| `ctf:awd:{contest_id}:round:{round_id}:flags` | HASH (field=team_id:challenge_id, value=flag) | 至轮次结束 | 每轮每队每服务的动态 Flag（轮次结束后清理） |
| `ctf:awd:{contest_id}:service_status` | HASH (field=team_id:challenge_id, value=status) | 无过期 | 各队服务实时状态（健康检查更新） |
| `ctf:awd:{contest_id}:scoreboard` | STRING (JSON) | 10s | AWD 实时计分板缓存（高频刷新场景） |

### 12.7 通知与杂项

| 键名模式 | 数据结构 | TTL | 用途 |
|----------|----------|-----|------|
| `ctf:notify:unread:{user_id}` | STRING (int 计数) | 无过期 | 用户未读通知数（原子递增，读取时递减） |
| `ctf:notify:broadcast:{contest_id}` | LIST (JSON) | 至竞赛结束 | 竞赛广播通知队列（WebSocket 推送消费） |
| `ctf:captcha:{session_id}` | STRING (验证码值) | 5min | 图形验证码缓存 |
| `ctf:config:global` | HASH | 10min | 全局系统配置缓存（如最大实例数、默认 TTL 等） |

---

## 13. 数据归档策略

### 13.1 归档原则

- 归档不等于删除：数据从主表迁移到归档表（`*_archive`），保留完整记录
- 归档操作在低峰期执行（凌晨 2:00-5:00），通过定时任务触发
- 归档前先写入归档表，确认成功后再从主表删除（两阶段操作，失败可重试）
- 归档表与主表结构完全一致，额外增加 `archived_at TIMESTAMPTZ` 字段

### 13.2 各表归档规则

| 表名 | 归档条件 | 保留周期 | 归档表 | 说明 |
|------|----------|----------|--------|------|
| `audit_logs` | `created_at < NOW() - INTERVAL '90 days'` | 主表保留 90 天 | `audit_logs_archive` | 审计日志量最大，优先归档 |
| `submissions` | `created_at < NOW() - INTERVAL '180 days'` AND 不属于活跃竞赛 | 主表保留 180 天 | `submissions_archive` | 归档前确保关联竞赛已结束 |
| `instances` | `status IN ('stopped', 'expired', 'error')` AND `updated_at < NOW() - INTERVAL '30 days'` | 已终止实例保留 30 天 | `instances_archive` | 运行中实例永不归档 |
| `awd_attack_logs` | 关联竞赛状态为 `archived` | 竞赛归档后 30 天 | `awd_attack_logs_archive` | 随竞赛生命周期归档 |
| `awd_team_services` | 关联竞赛状态为 `archived` | 竞赛归档后 30 天 | `awd_team_services_archive` | 随竞赛生命周期归档 |
| `awd_traffic_events` | 关联竞赛状态为 `archived` | 竞赛归档后 30 天 | `awd_traffic_events_archive` | 随竞赛生命周期归档 |
| `notifications` | `created_at < NOW() - INTERVAL '90 days'` AND `is_read = TRUE` | 已读通知保留 90 天 | `notifications_archive` | 未读通知不归档 |
| `hint_unlocks` | `created_at < NOW() - INTERVAL '365 days'` | 保留 1 年 | `hint_unlocks_archive` | 低频数据，年度归档即可 |

### 13.3 不归档的表

以下表数据量可控或具有长期引用价值，不做归档：

| 表名 | 原因 |
|------|------|
| `users` | 用户主数据，软删除即可，不归档 |
| `roles` / `user_roles` | 数据量极小（角色固定 3 种），无需归档 |
| `challenges` / `images` | 靶机和镜像为核心资产，软删除管理 |
| `tags` / `challenge_tags` | 数据量小，标签体系需长期保留 |
| `topologies` / `topology_nodes` | 拓扑为教学资产，软删除管理 |
| `writeups` / `hints` | 知识沉淀内容，长期保留 |
| `contests` | 竞赛主记录保留，仅状态流转到 `archived` |
| `contest_challenges` | 随竞赛保留 |
| `teams` / `team_members` | 随竞赛保留 |
| `contest_registrations` | 数据量可控，随竞赛保留 |
| `contest_announcements` | 数据量小，随竞赛保留 |
| `user_scores` | 每用户一条，数据量 = 用户数，无需归档 |
| `skill_dimensions` / `user_skill_profiles` | 能力画像需长期追踪 |
| `learning_paths` | 教学资产，软删除管理 |

### 13.4 归档表 DDL 示例

以 `audit_logs` 为例，其他归档表结构类似：

```sql
CREATE TABLE audit_logs_archive (
    -- 与主表完全一致的字段
    id              BIGINT          NOT NULL,
    user_id         BIGINT          DEFAULT NULL,
    action          VARCHAR(32)     NOT NULL,
    resource_type   VARCHAR(64)     NOT NULL,
    resource_id     BIGINT          DEFAULT NULL,
    detail          JSONB           NOT NULL DEFAULT '{}',
    ip_address      VARCHAR(45)     NOT NULL,
    user_agent      VARCHAR(512)    DEFAULT NULL,
    created_at      TIMESTAMPTZ     NOT NULL,
    -- 归档专属字段
    archived_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- 归档表按月分区（PostgreSQL 声明式分区），便于按月清理超龄数据
-- 示例：按 created_at 范围分区
-- CREATE TABLE audit_logs_archive_2026_01 PARTITION OF audit_logs_archive
--     FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- 归档表索引（按需建立，比主表精简）
CREATE INDEX idx_archive_audit_created ON audit_logs_archive(created_at);
CREATE INDEX idx_archive_audit_user ON audit_logs_archive(user_id, created_at);
```

### 13.5 Redis 缓存清理策略

| 清理场景 | 策略 | 说明 |
|----------|------|------|
| 竞赛结束 | 主动清理 | 竞赛状态流转到 `ended` 时，异步清理该竞赛相关的所有 Redis 键（排行榜、题目缓存、AWD 状态等） |
| 实例过期 | 定时扫描 | 每分钟扫描 `ctf:instance:expire_queue`，清理过期实例对应的所有缓存键 |
| 用户封禁/删除 | 事件驱动 | 用户状态变更时，清理该用户的 token、profile、roles 等缓存 |
| 靶机状态变更 | 事件驱动 | 靶机上线/下线/修改时，清理靶机详情缓存和列表缓存 |
| 全局配置变更 | 手动触发 | 管理员修改全局配置后，提供"刷新缓存"按钮清理 `ctf:config:global` |
