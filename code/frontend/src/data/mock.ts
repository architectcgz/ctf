import type { UserRole } from '@/utils/constants'

export const currentUser = {
  name: '陈星河',
  role: 'student' as UserRole,
  className: '网安 2201',
  rank: 42,
  solved: 42,
  totalChallenges: 156,
}

export const dashboardStats = [
  { label: '已解题数', value: '42 / 156', hint: '本周 +6', tone: 'cyan' },
  { label: '当前排名', value: '#42', hint: '较昨日上升 5 位', tone: 'violet' },
  { label: '活跃实例', value: '2 / 3', hint: '1 个实例即将过期', tone: 'amber' },
  { label: '进行中竞赛', value: '1', hint: '春季选拔赛进行中', tone: 'emerald' },
]

export const skillScores = [
  { name: 'Web 安全', value: 78, color: '#06b6d4' },
  { name: '密码学', value: 62, color: '#a78bfa' },
  { name: 'Pwn', value: 45, color: '#ef4444' },
  { name: '逆向工程', value: 40, color: '#fb923c' },
  { name: 'Misc', value: 55, color: '#22c55e' },
  { name: '取证分析', value: 25, color: '#3b82f6' },
]

export const activityFeed = [
  { title: '解出 SQL 注入基础', meta: 'Web · 简单 · 8 分钟前', tone: 'success' },
  { title: '启动 栈溢出入门 实例', meta: 'Pwn · 中等 · 13 分钟前', tone: 'primary' },
  { title: '解锁提示 RSA 基础 Hint 2', meta: 'Crypto · 简单 · 36 分钟前', tone: 'warning' },
  { title: '提交 Flag 失败', meta: 'XSS 反射型 · 1 小时前', tone: 'danger' },
]

export const recommendations = [
  {
    title: '流量包取证训练',
    category: 'Forensics',
    level: '中等',
    reason: '补齐你在取证分析上的短板',
  },
  { title: 'ELF 入门逆向', category: 'Reverse', level: '简单', reason: '适合作为逆向工程热身题' },
  { title: 'RCE 链路梳理', category: 'Web', level: '困难', reason: '提升漏洞利用链构建能力' },
]

export const challengeCards = [
  {
    id: 1,
    title: 'SQL 注入基础',
    category: 'Web',
    level: '简单',
    solved: 128,
    points: 100,
    summary: '通过登录注入获取管理员数据表中的敏感信息。',
    solvedByMe: true,
  },
  {
    id: 2,
    title: 'XSS 反射型',
    category: 'Web',
    level: '中等',
    solved: 95,
    points: 200,
    summary: '定位输入点并构造反射型脚本完成 Cookie 窃取。',
    solvedByMe: false,
  },
  {
    id: 3,
    title: '栈溢出入门',
    category: 'Pwn',
    level: '中等',
    solved: 44,
    points: 300,
    summary: '利用函数返回地址覆盖拿到 shell。',
    solvedByMe: false,
  },
  {
    id: 4,
    title: 'RSA 基础',
    category: 'Crypto',
    level: '简单',
    solved: 157,
    points: 150,
    summary: '识别错误密钥生成逻辑并恢复明文。',
    solvedByMe: true,
  },
  {
    id: 5,
    title: '流量包取证训练',
    category: 'Forensics',
    level: '困难',
    solved: 19,
    points: 450,
    summary: '从大体量抓包中抽丝剥茧定位失陷主机。',
    solvedByMe: false,
  },
  {
    id: 6,
    title: 'PE 壳识别',
    category: 'Reverse',
    level: '中等',
    solved: 31,
    points: 260,
    summary: '识别打包壳并恢复关键逻辑路径。',
    solvedByMe: false,
  },
]

export const contestCards = [
  {
    id: 1,
    title: '春季选拔赛',
    status: '进行中',
    time: '剩余 01:42:18',
    teams: 38,
    solved: 17,
    accent: 'primary',
  },
  {
    id: 2,
    title: '校内周赛 #12',
    status: '即将开始',
    time: '今晚 19:30',
    teams: 24,
    solved: 0,
    accent: 'accent',
  },
  {
    id: 3,
    title: '新生训练营结营赛',
    status: '已结束',
    time: '2 天前结束',
    teams: 42,
    solved: 23,
    accent: 'muted',
  },
]

export const contestScoreboard = [
  { rank: 1, team: 'Overflow', solved: 9, score: 2840, delta: '+120' },
  { rank: 2, team: 'ZeroDay', solved: 8, score: 2710, delta: '+90' },
  { rank: 3, team: 'ByteStorm', solved: 8, score: 2600, delta: '+30' },
  { rank: 4, team: 'NightShift', solved: 7, score: 2370, delta: '-12' },
  { rank: 5, team: 'CyanTeam', solved: 7, score: 2250, delta: '+70' },
]

export const instances = [
  {
    id: 1,
    title: 'SQL 注入基础',
    category: 'Web',
    level: '简单',
    status: 'running',
    address: '10.10.1.42:8080',
    timeLeft: '01:23:45',
    note: '终端票据已签发，可在浏览器内直接进入 Web Terminal',
  },
  {
    id: 2,
    title: '栈溢出入门',
    category: 'Pwn',
    level: '中等',
    status: 'warning',
    address: '10.10.1.43:9999',
    timeLeft: '00:04:32',
    note: '建议立即延时，防止 shell 会话被回收',
  },
  {
    id: 3,
    title: '流量包取证训练',
    category: 'Forensics',
    level: '困难',
    status: 'queue',
    queuePosition: 3,
    estimate: '~2 分钟',
    progress: 46,
    note: '当前高峰时段启用排队系统，实例创建完成后自动刷新',
  },
]

export const notifications = [
  { id: 1, type: '竞赛公告', title: '春季赛新增 3 道 Web 题目', time: '5 分钟前', unread: true },
  {
    id: 2,
    type: '系统通知',
    title: '你的靶机“栈溢出入门”将在 5 分钟后过期',
    time: '1 小时前',
    unread: true,
  },
  {
    id: 3,
    type: '解题通知',
    title: '恭喜解出“SQL 注入基础”，获得 100 分',
    time: '2 小时前',
    unread: false,
  },
  { id: 4, type: '排队更新', title: '“流量包取证训练”已前进至第 3 位', time: '刚刚', unread: true },
]

export const adminStats = [
  { label: '在线实例', value: '186', hint: '较昨日 +12' },
  { label: '容器健康率', value: '98.4%', hint: '过去 24h' },
  { label: '今日提交数', value: '1,284', hint: '错误率 8.7%' },
  { label: '告警事件', value: '7', hint: '2 条待处理' },
]

export const profileFields = [
  { label: '学号', value: '2022012345' },
  { label: '姓名', value: '陈星河' },
  { label: '班级', value: '网络空间安全 2201' },
  { label: '邮箱', value: 'xinghe@campus.edu' },
]

export function toneClass(tone: string): string | undefined {
  return {
    cyan: 'text-cyan-300 bg-cyan-500/10 border-cyan-500/20',
    violet: 'text-violet-300 bg-violet-500/10 border-violet-500/20',
    amber: 'text-amber-300 bg-amber-500/10 border-amber-500/20',
    emerald: 'text-emerald-300 bg-emerald-500/10 border-emerald-500/20',
    success: 'text-emerald-300',
    primary: 'text-cyan-300',
    warning: 'text-amber-300',
    danger: 'text-red-300',
    muted: 'text-slate-400 bg-slate-500/10 border-slate-500/20',
    accent: 'text-violet-300 bg-violet-500/10 border-violet-500/20',
  }[tone]
}
