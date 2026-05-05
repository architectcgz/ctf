import type { RouteRecordRaw } from 'vue-router'

import { redirectWithQuery } from './route-helpers'

export const studentRoutes: RouteRecordRaw[] = [
{
  path: 'student/dashboard',
  name: 'Dashboard',
  component: () => import('@/views/dashboard/DashboardView.vue'),
  meta: {
    requiresAuth: true,
    title: '仪表盘',
    icon: 'LayoutDashboard',
    contentLayout: 'bleed',
  },
},
{
  path: 'dashboard',
  redirect: redirectWithQuery('/student/dashboard'),
},
{
  path: 'challenges',
  name: 'Challenges',
  component: () => import('@/views/challenges/ChallengeList.vue'),
  meta: {
    requiresAuth: true,
    roles: ['student', 'teacher'],
    title: '题目',
    icon: 'Swords',
    contentLayout: 'bleed',
  },
},
{
  path: 'challenges/:id',
  name: 'ChallengeDetail',
  component: () => import('@/views/challenges/ChallengeDetail.vue'),
  meta: {
    requiresAuth: true,
    roles: ['student', 'teacher'],
    title: '题目详情',
    contentLayout: 'bleed',
  },
},
{
  path: 'contests',
  name: 'Contests',
  component: () => import('@/views/contests/ContestList.vue'),
  meta: { requiresAuth: true, title: '竞赛', icon: 'Trophy', contentLayout: 'bleed' },
},
{
  path: 'contests/:id',
  name: 'ContestDetail',
  component: () => import('@/views/contests/ContestDetail.vue'),
  meta: { requiresAuth: true, title: '竞赛详情', contentLayout: 'bleed' },
},
{
  path: 'contests/:id/awd/defense/:serviceId',
  name: 'ContestAWDDefenseWorkbench',
  component: () => import('@/views/contests/ContestAWDDefenseWorkbench.vue'),
  meta: { requiresAuth: true, title: '防守内容', contentLayout: 'bleed' },
},
{
  path: 'scoreboard',
  name: 'Scoreboard',
  component: () => import('@/views/scoreboard/ScoreboardView.vue'),
  meta: { requiresAuth: true, title: '排行榜', icon: 'BarChart3', contentLayout: 'bleed' },
},
{
  path: 'scoreboard/:contestId',
  name: 'ScoreboardDetail',
  component: () => import('@/views/scoreboard/ScoreboardDetail.vue'),
  meta: { requiresAuth: true, title: '排行详情', contentLayout: 'bleed' },
},
{
  path: 'student/instances',
  name: 'Instances',
  component: () => import('@/views/instances/InstanceList.vue'),
  meta: {
    requiresAuth: true,
    roles: ['student'],
    title: '我的实例',
    icon: 'Server',
    contentLayout: 'bleed',
  },
},
{
  path: 'instances',
  redirect: redirectWithQuery('/student/instances'),
},
{
  path: 'student/skill-profile',
  name: 'SkillProfile',
  component: () => import('@/views/profile/SkillProfile.vue'),
  meta: {
    requiresAuth: true,
    roles: ['student'],
    title: '能力画像',
    icon: 'Radar',
    contentLayout: 'bleed',
  },
},
{
  path: 'skill-profile',
  redirect: redirectWithQuery('/student/skill-profile'),
},
{
  path: 'profile',
  name: 'Profile',
  component: () => import('@/views/profile/UserProfile.vue'),
  meta: { requiresAuth: true, title: '个人资料', icon: 'User', contentLayout: 'bleed' },
},
{
  path: 'student/profile',
  redirect: redirectWithQuery('/profile'),
},
{
  path: 'settings/security',
  name: 'SecuritySettings',
  component: () => import('@/views/profile/SecuritySettings.vue'),
  meta: { requiresAuth: true, title: '安全设置', icon: 'Settings', contentLayout: 'bleed' },
},
{
  path: 'student/settings/security',
  redirect: redirectWithQuery('/settings/security'),
},
{
  path: 'notifications',
  name: 'Notifications',
  component: () => import('@/views/notifications/NotificationList.vue'),
  meta: { requiresAuth: true, title: '通知', icon: 'Bell', contentLayout: 'bleed' },
},
{
  path: 'notifications/:id',
  name: 'NotificationDetail',
  component: () => import('@/views/notifications/NotificationDetail.vue'),
  meta: { requiresAuth: true, title: '通知详情', contentLayout: 'bleed' },
},
]
