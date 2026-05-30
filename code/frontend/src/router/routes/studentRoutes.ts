import type { RouteRecordRaw } from 'vue-router'

import { redirectWithQuery } from './route-helpers'

export const studentRoutes: RouteRecordRaw[] = [
  {
    path: 'student/dashboard',
    name: 'Dashboard',
    component: () => import('@/pages/dashboard/DashboardRoutePage.vue'),
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
    component: () => import('@/pages/challenges/ChallengeListRoutePage.vue'),
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
    component: () => import('@/pages/challenges/ChallengeDetailRoutePage.vue'),
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
    component: () => import('@/pages/contests/ContestListRoutePage.vue'),
    meta: { requiresAuth: true, title: '竞赛', icon: 'Trophy', contentLayout: 'bleed' },
  },
  {
    path: 'contests/:id',
    name: 'ContestDetail',
    component: () => import('@/pages/contests/ContestDetailRoutePage.vue'),
    meta: { requiresAuth: true, title: '竞赛详情', contentLayout: 'bleed' },
  },
  {
    path: 'scoreboard',
    name: 'Scoreboard',
    component: () => import('@/pages/scoreboard/ScoreboardViewRoutePage.vue'),
    meta: { requiresAuth: true, title: '排行榜', icon: 'BarChart3', contentLayout: 'bleed' },
  },
  {
    path: 'scoreboard/:contestId',
    name: 'ScoreboardDetail',
    component: () => import('@/pages/scoreboard/ScoreboardDetailRoutePage.vue'),
    props: (route) => ({
      contestId: String(route.params.contestId || ''),
    }),
    meta: { requiresAuth: true, title: '排行详情', contentLayout: 'bleed' },
  },
  {
    path: 'student/instances',
    name: 'Instances',
    component: () => import('@/pages/instances/InstanceListRoutePage.vue'),
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
    component: () => import('@/pages/profile/SkillProfileRoutePage.vue'),
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
    component: () => import('@/pages/profile/UserProfileRoutePage.vue'),
    meta: { requiresAuth: true, title: '个人资料', icon: 'User', contentLayout: 'bleed' },
  },
  {
    path: 'student/profile',
    redirect: redirectWithQuery('/profile'),
  },
  {
    path: 'settings/security',
    name: 'SecuritySettings',
    component: () => import('@/pages/profile/SecuritySettingsRoutePage.vue'),
    meta: { requiresAuth: true, title: '安全设置', icon: 'Settings', contentLayout: 'bleed' },
  },
  {
    path: 'student/settings/security',
    redirect: redirectWithQuery('/settings/security'),
  },
  {
    path: 'notifications',
    name: 'Notifications',
    component: () => import('@/pages/notifications/NotificationListRoutePage.vue'),
    meta: { requiresAuth: true, title: '通知', icon: 'Bell', contentLayout: 'bleed' },
  },
  {
    path: 'notifications/:id',
    name: 'NotificationDetail',
    component: () => import('@/pages/notifications/NotificationDetailRoutePage.vue'),
    props: (route) => ({
      id: String(route.params.id || ''),
    }),
    meta: { requiresAuth: true, title: '通知详情', contentLayout: 'bleed' },
  },
]
