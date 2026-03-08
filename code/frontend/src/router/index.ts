import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { setupRouterGuards } from './guards'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: { title: '注册' },
  },
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { requiresAuth: true, title: '仪表盘', icon: 'LayoutDashboard' },
      },
      {
        path: 'challenges',
        name: 'Challenges',
        component: () => import('@/views/challenges/ChallengeList.vue'),
        meta: { requiresAuth: true, roles: ['student', 'teacher'], title: '靶场', icon: 'Swords' },
      },
      {
        path: 'challenges/:id',
        name: 'ChallengeDetail',
        component: () => import('@/views/challenges/ChallengeDetail.vue'),
        meta: { requiresAuth: true, roles: ['student', 'teacher'], title: '靶场详情' },
      },
      {
        path: 'contests',
        name: 'Contests',
        component: () => import('@/views/contests/ContestList.vue'),
        meta: { requiresAuth: true, title: '竞赛', icon: 'Trophy' },
      },
      {
        path: 'contests/:id',
        name: 'ContestDetail',
        component: () => import('@/views/contests/ContestDetail.vue'),
        meta: { requiresAuth: true, title: '竞赛详情' },
      },
      {
        path: 'scoreboard',
        name: 'Scoreboard',
        component: () => import('@/views/scoreboard/ScoreboardView.vue'),
        meta: { requiresAuth: true, title: '排行榜', icon: 'BarChart3' },
      },
      {
        path: 'instances',
        name: 'Instances',
        component: () => import('@/views/instances/InstanceList.vue'),
        meta: { requiresAuth: true, roles: ['student'], title: '我的实例', icon: 'Server' },
      },
      {
        path: 'skill-profile',
        name: 'SkillProfile',
        component: () => import('@/views/profile/SkillProfile.vue'),
        meta: { requiresAuth: true, roles: ['student'], title: '能力画像', icon: 'Radar' },
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/profile/UserProfile.vue'),
        meta: { requiresAuth: true, title: '个人资料', icon: 'User' },
      },
      {
        path: 'settings/security',
        name: 'SecuritySettings',
        component: () => import('@/views/profile/SecuritySettings.vue'),
        meta: { requiresAuth: true, title: '安全设置', icon: 'Settings' },
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('@/views/notifications/NotificationList.vue'),
        meta: { requiresAuth: true, title: '通知', icon: 'Bell' },
      },

      // Teacher
      {
        path: 'teacher/dashboard',
        name: 'TeacherDashboard',
        component: () => import('@/views/teacher/TeacherDashboard.vue'),
        meta: { requiresAuth: true, roles: ['teacher', 'admin'], title: '教学概览', icon: 'GraduationCap' },
      },
      {
        path: 'teacher/classes',
        name: 'ClassManagement',
        component: () => import('@/views/teacher/ClassManagement.vue'),
        meta: { requiresAuth: true, roles: ['teacher', 'admin'], title: '班级管理', icon: 'Users' },
      },
      {
        path: 'teacher/reports',
        name: 'ReportExport',
        component: () => import('@/views/teacher/ReportExport.vue'),
        meta: { requiresAuth: true, roles: ['teacher', 'admin'], title: '报告导出', icon: 'FileDown' },
      },

      // Admin
      {
        path: 'admin/dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/AdminDashboard.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '系统概览', icon: 'Shield' },
      },
      {
        path: 'admin/challenges',
        name: 'ChallengeManage',
        component: () => import('@/views/admin/ChallengeManage.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '靶场管理', icon: 'Settings' },
      },
      {
        path: 'admin/challenges/:id',
        name: 'AdminChallengeDetail',
        component: () => import('@/views/admin/ChallengeDetail.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '靶场详情' },
      },
      {
        path: 'admin/contests',
        name: 'ContestManage',
        component: () => import('@/views/admin/ContestManage.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '竞赛管理', icon: 'Trophy' },
      },
      {
        path: 'admin/users',
        name: 'UserManage',
        component: () => import('@/views/admin/UserManage.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '用户管理', icon: 'Users' },
      },
      {
        path: 'admin/images',
        name: 'ImageManage',
        component: () => import('@/views/admin/ImageManage.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '镜像管理', icon: 'Layers' },
      },
      {
        path: 'admin/cheat',
        name: 'CheatDetection',
        component: () => import('@/views/admin/CheatDetection.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '作弊检测', icon: 'ScanEye' },
      },
      {
        path: 'admin/audit',
        name: 'AuditLog',
        component: () => import('@/views/admin/AuditLog.vue'),
        meta: { requiresAuth: true, roles: ['admin'], title: '审计日志', icon: 'ClipboardList' },
      },
    ],
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/errors/ForbiddenView.vue'),
    meta: { title: '无权限' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/errors/NotFoundView.vue'),
    meta: { title: '页面不存在' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

setupRouterGuards(router)

export default router
export { routes }

