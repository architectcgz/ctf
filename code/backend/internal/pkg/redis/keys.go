package redis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ============================================================
// 全局命名空间前缀
// ============================================================

const Namespace = "ctf"

// withNS 拼接全局命名空间前缀
func withNS(key string) string {
	return Namespace + ":" + key
}

// ============================================================
// 用户与认证模块
// ============================================================

const (
	// keyTokenPrefix 用户 Refresh Token 存储，登出时删除实现强制下线
	keyTokenPrefix = "token:"
	// keyLoginFailPrefix 登录失败计数，达到阈值后触发账户锁定
	keyLoginFailPrefix = "login_fail:"
	// keyUserProfilePrefix 用户基本信息缓存
	keyUserProfilePrefix = "user:profile:"
	// keyUserRolesPrefix 用户角色列表缓存
	keyUserRolesPrefix = "user:roles:"
)

// TokenKey 用户 Refresh Token
// 数据结构: STRING (JWT refresh token hash) | TTL: 7d
func TokenKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyTokenPrefix, userID))
}

// LoginFailKey 登录失败计数
// 数据结构: STRING (int) | TTL: 30min
func LoginFailKey(username string) string {
	return withNS(fmt.Sprintf("%s%s", keyLoginFailPrefix, normalizedUsernameSegment(username)))
}

// UserProfileKey 用户基本信息缓存
// 数据结构: HASH | TTL: 30min
func UserProfileKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyUserProfilePrefix, userID))
}

// UserRolesKey 用户角色列表缓存
// 数据结构: SET | TTL: 30min
func UserRolesKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyUserRolesPrefix, userID))
}

// ============================================================
// 靶场与靶机模块
// ============================================================

const (
	// keyChallengeDetailPrefix 靶机详情缓存
	keyChallengeDetailPrefix = "challenge:detail:"
	// keyChallengeListActive 已上线靶机列表缓存
	keyChallengeListActive = "challenge:list:active"
	// keyChallengeSolveCountPrefix 靶机解出人数计数器
	keyChallengeSolveCountPrefix = "challenge:solve_count:"
	// keyChallengeFlagPrefix 静态 Flag 缓存
	keyChallengeFlagPrefix = "challenge:flag:"
	// keyImageStatusPrefix 镜像构建状态缓存
	keyImageStatusPrefix = "image:status:"
)

// ChallengeDetailKey 靶机详情缓存
// 数据结构: HASH | TTL: 1h
func ChallengeDetailKey(challengeID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyChallengeDetailPrefix, challengeID))
}

// ChallengeListActiveKey 已上线靶机列表缓存
// 数据结构: ZSET (score=sort_order, member=challenge_id) | TTL: 10min
func ChallengeListActiveKey() string {
	return withNS(keyChallengeListActive)
}

// ChallengeSolveCountKey 靶机解出人数计数器
// 数据结构: STRING (int) | TTL: 无过期（定期与 DB 校准）
func ChallengeSolveCountKey(challengeID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyChallengeSolveCountPrefix, challengeID))
}

// ChallengeFlagKey 静态 Flag 缓存
// 数据结构: STRING | TTL: 1h
func ChallengeFlagKey(challengeID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyChallengeFlagPrefix, challengeID))
}

// ImageStatusKey 镜像构建状态缓存
// 数据结构: STRING | TTL: 5min
func ImageStatusKey(imageID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyImageStatusPrefix, imageID))
}

// ============================================================
// 容器实例模块
// ============================================================

const (
	// keyInstanceUserPrefix 用户当前运行实例映射
	keyInstanceUserPrefix = "instance:user:"
	// keyInstanceCountPrefix 用户并发实例计数器
	keyInstanceCountPrefix = "instance:count:"
	// keyInstanceFlagPrefix 动态 Flag 实例级缓存
	keyInstanceFlagPrefix = "instance:flag:"
	// keyInstanceExpireQueue 实例过期队列
	keyInstanceExpireQueue = "instance:expire_queue"
)

// InstanceUserKey 用户当前运行实例映射
// 数据结构: HASH (field=challenge_id, value=instance_id) | TTL: 与实例过期时间对齐
func InstanceUserKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyInstanceUserPrefix, userID))
}

// InstanceCountKey 用户并发实例计数器
// 数据结构: STRING (int) | TTL: 无过期（创建+1 销毁-1）
func InstanceCountKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyInstanceCountPrefix, userID))
}

// InstanceFlagKey 动态 Flag 实例级缓存
// 数据结构: STRING | TTL: 与实例过期时间对齐
func InstanceFlagKey(instanceID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyInstanceFlagPrefix, instanceID))
}

// InstanceExpireQueueKey 实例过期队列
// 数据结构: ZSET (score=expire_timestamp, member=instance_id) | TTL: 无过期
func InstanceExpireQueueKey() string {
	return withNS(keyInstanceExpireQueue)
}

// ============================================================
// 提交与限流模块
// ============================================================

const (
	// keySubmitRatePrefix Flag 提交频率限制，防止暴力猜测
	keySubmitRatePrefix = "submit:rate:"
	// keySubmitSolvedPrefix 用户已解出的靶机集合
	keySubmitSolvedPrefix = "submit:solved:"
	// keySubmitLockPrefix 提交分布式锁，防止并发提交
	keySubmitLockPrefix = "submit:lock:"
)

// SubmitRateKey Flag 提交频率限制
// 数据结构: STRING (int) | TTL: 60s（滑动窗口）
func SubmitRateKey(userID, challengeID int64) string {
	return withNS(fmt.Sprintf("%s%d:%d", keySubmitRatePrefix, userID, challengeID))
}

// SubmitSolvedKey 用户已解出的靶机集合
// 数据结构: SET (member=challenge_id) | TTL: 无过期
func SubmitSolvedKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keySubmitSolvedPrefix, userID))
}

// SubmitLockKey 提交分布式锁
// 数据结构: STRING ("1") | TTL: 5s
func SubmitLockKey(userID, challengeID int64) string {
	return withNS(fmt.Sprintf("%s%d:%d", keySubmitLockPrefix, userID, challengeID))
}

// ============================================================
// 竞赛与排行榜模块
// ============================================================

const (
	// keyContestDetailPrefix 竞赛详情缓存
	keyContestDetailPrefix = "contest:detail:"
	// keyContestChallengesPrefix 竞赛题目列表缓存
	keyContestChallengesPrefix = "contest:challenges:"
	// keyRankGlobal 全站排行榜
	keyRankGlobal = "rank:global"
	// keyRankContestUserPrefix 竞赛个人排行榜
	keyRankContestUserPrefix = "rank:contest:%d:user"
	// keyRankContestTeamPrefix 竞赛队伍排行榜
	keyRankContestTeamPrefix = "rank:contest:%d:team"
	// keyRankContestFrozenPrefix 封榜后的排行榜快照
	keyRankContestFrozenPrefix = "rank:contest:%d:frozen"
	// keyContestFreezeFlagPrefix 竞赛排行榜冻结标记
	keyContestFreezeFlagPrefix = "contest:freeze_flag:"
	// keyAWDRoundLockPrefix AWD 轮次推进分布式锁
	keyAWDRoundLockPrefix = "awd:round:lock:"
)

// ContestDetailKey 竞赛详情缓存
// 数据结构: HASH | TTL: 10min
func ContestDetailKey(contestID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyContestDetailPrefix, contestID))
}

// ContestChallengesKey 竞赛题目列表缓存
// 数据结构: LIST (JSON 序列化) | TTL: 5min
func ContestChallengesKey(contestID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyContestChallengesPrefix, contestID))
}

// RankGlobalKey 全站排行榜
// 数据结构: ZSET (score=total_score, member=user_id) | TTL: 无过期
func RankGlobalKey() string {
	return withNS(keyRankGlobal)
}

// RankContestUserKey 竞赛个人排行榜
// 数据结构: ZSET (score=total_score, member=user_id) | TTL: 无过期
func RankContestUserKey(contestID int64) string {
	return withNS(fmt.Sprintf(keyRankContestUserPrefix, contestID))
}

// RankContestTeamKey 竞赛队伍排行榜
// 数据结构: ZSET (score=total_score, member=team_id) | TTL: 无过期
func RankContestTeamKey(contestID int64) string {
	return withNS(fmt.Sprintf(keyRankContestTeamPrefix, contestID))
}

// RankContestFrozenKey 封榜后的排行榜快照
// 数据结构: ZSET (score=total_score, member=team_id) | TTL: 至竞赛结束
func RankContestFrozenKey(contestID int64) string {
	return withNS(fmt.Sprintf(keyRankContestFrozenPrefix, contestID))
}

// ContestFreezeFlagKey 竞赛排行榜冻结标记
// 数据结构: STRING ("1") | TTL: 至竞赛结束
func ContestFreezeFlagKey(contestID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyContestFreezeFlagPrefix, contestID))
}

// AWDRoundLockKey AWD 轮次推进锁
// 数据结构: STRING ("1") | TTL: 30s（默认）
func AWDRoundLockKey(contestID int64, roundNumber int) string {
	return withNS(fmt.Sprintf("%s%d:%d", keyAWDRoundLockPrefix, contestID, roundNumber))
}

// ============================================================
// AWD 实时状态模块
// ============================================================

const (
	// keyAWDCurrentRoundFmt 当前轮次编号
	keyAWDCurrentRoundFmt = "awd:%d:current_round"
	// keyAWDRoundFlagsFmt 每轮每队的动态 Flag
	keyAWDRoundFlagsFmt = "awd:%d:round:%d:flags"
	// keyAWDServiceStatusFmt 各队服务实时状态
	keyAWDServiceStatusFmt = "awd:%d:service_status"
	// keyAWDScoreboardFmt AWD 实时计分板缓存
	keyAWDScoreboardFmt = "awd:%d:scoreboard"
)

// AWDCurrentRoundKey 当前轮次编号
// 数据结构: STRING (int) | TTL: 无过期
func AWDCurrentRoundKey(contestID int64) string {
	return withNS(fmt.Sprintf(keyAWDCurrentRoundFmt, contestID))
}

// AWDRoundFlagsKey 每轮每队的动态 Flag
// 数据结构: HASH (field=team_id, value=flag) | TTL: 至轮次结束
func AWDRoundFlagsKey(contestID int64, roundID int64) string {
	return withNS(fmt.Sprintf(keyAWDRoundFlagsFmt, contestID, roundID))
}

// AWDRoundFlagField 每轮 Flag 哈希字段
// 结构: {team_id}:{challenge_id}
func AWDRoundFlagField(teamID, challengeID int64) string {
	return fmt.Sprintf("%d:%d", teamID, challengeID)
}

// AWDServiceStatusKey 各队服务实时状态
// 数据结构: HASH (field=team_id:challenge_id, value=status) | TTL: 无过期
func AWDServiceStatusKey(contestID int64) string {
	return withNS(fmt.Sprintf(keyAWDServiceStatusFmt, contestID))
}

// AWDScoreboardKey AWD 实时计分板缓存
// 数据结构: STRING (JSON) | TTL: 10s
func AWDScoreboardKey(contestID int64) string {
	return withNS(fmt.Sprintf(keyAWDScoreboardFmt, contestID))
}

// ============================================================
// 能力评估与推荐模块
// ============================================================

const (
	// keyRecommendationUserPrefix 用户推荐靶场缓存
	keyRecommendationUserPrefix = "recommendation:user:"
)

// RecommendationKey 用户推荐靶场缓存
// 数据结构: STRING (JSON) | TTL: 1h
func RecommendationKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyRecommendationUserPrefix, userID))
}

// ============================================================
// 通知与杂项模块
// ============================================================

const (
	// keyNotifyUnreadPrefix 用户未读通知数
	keyNotifyUnreadPrefix = "notify:unread:"
	// keyNotifyBroadcastPrefix 竞赛广播通知队列
	keyNotifyBroadcastPrefix = "notify:broadcast:"
	// keyCaptchaPrefix 图形验证码缓存
	keyCaptchaPrefix = "captcha:"
	// keyConfigGlobal 全局系统配置缓存
	keyConfigGlobal = "config:global"
)

// NotifyUnreadKey 用户未读通知数
// 数据结构: STRING (int) | TTL: 无过期（原子递增，读取时递减）
func NotifyUnreadKey(userID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyNotifyUnreadPrefix, userID))
}

// NotifyBroadcastKey 竞赛广播通知队列
// 数据结构: LIST (JSON) | TTL: 至竞赛结束
func NotifyBroadcastKey(contestID int64) string {
	return withNS(fmt.Sprintf("%s%d", keyNotifyBroadcastPrefix, contestID))
}

// CaptchaKey 图形验证码缓存
// 数据结构: STRING (验证码值) | TTL: 5min
func CaptchaKey(sessionID string) string {
	return withNS(fmt.Sprintf("%s%s", keyCaptchaPrefix, sessionID))
}

// ConfigGlobalKey 全局系统配置缓存
// 数据结构: HASH | TTL: 10min
func ConfigGlobalKey() string {
	return withNS(keyConfigGlobal)
}

func normalizedUsernameSegment(username string) string {
	normalized := strings.ToLower(strings.TrimSpace(username))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}
