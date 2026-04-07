import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import { setupRouterGuards } from './guards'

function redirectWithQuery(path: string): NonNullable<RouteRecordRaw['redirect']> {
  return (to) => ({
    path,
    query: to.query,
    hash: to.hash,
  })
}

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
    redirect: '/student/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'student/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/DashboardView.vue'),
        meta: { requiresAuth: true, title: '仪表盘', icon: 'LayoutDashboard', contentLayout: 'bleed' },
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
          title: '靶场',
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
          title: '靶场详情',
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
        path: 'scoreboard',
        name: 'Scoreboard',
        component: () => import('@/views/scoreboard/ScoreboardView.vue'),
        meta: { requiresAuth: true, title: '排行榜', icon: 'BarChart3', contentLayout: 'bleed' },
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

      // Teaching Operations
      {
        path: 'academy/overview',
        name: 'TeacherDashboard',
        component: () => import('@/views/teacher/TeacherDashboard.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '教学概览',
          icon: 'GraduationCap',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/dashboard',
        redirect: redirectWithQuery('/academy/overview'),
      },
      {
        path: 'academy/classes',
        name: 'ClassManagement',
        component: () => import('@/views/teacher/ClassManagement.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '班级管理',
          icon: 'Users',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes',
        redirect: redirectWithQuery('/academy/classes'),
      },
      {
        path: 'academy/students',
        name: 'TeacherStudentManagement',
        component: () => import('@/views/teacher/TeacherStudentManagement.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '学生管理',
          icon: 'GraduationCap',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/students',
        redirect: redirectWithQuery('/academy/students'),
      },
      {
        path: 'academy/classes/:className',
        name: 'TeacherClassStudents',
        component: () => import('@/views/teacher/TeacherClassStudents.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '班级学生',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'academy/classes/:className/students/:studentId',
        name: 'TeacherStudentAnalysis',
        component: () => import('@/views/teacher/TeacherStudentAnalysis.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '学员分析',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className/students/:studentId',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}/students/${encodeURIComponent(String(to.params.studentId || ''))}`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'academy/classes/:className/students/:studentId/review-archive',
        name: 'TeacherStudentReviewArchive',
        component: () => import('@/views/teacher/TeacherStudentReviewArchive.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '学生复盘归档',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className/students/:studentId/review-archive',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}/students/${encodeURIComponent(String(to.params.studentId || ''))}/review-archive`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'academy/instances',
        name: 'TeacherInstanceManagement',
        component: () => import('@/views/teacher/InstanceManagement.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '实例管理',
          icon: 'Server',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/instances',
        redirect: redirectWithQuery('/academy/instances'),
      },
      {
        path: 'academy/reports',
        name: 'ReportExport',
        component: () => import('@/views/teacher/ReportExport.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '报告导出',
          icon: 'FileDown',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/reports',
        redirect: redirectWithQuery('/academy/reports'),
      },

      // Platform Governance
      {
        path: 'admin/dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/AdminDashboard.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '系统概览',
          icon: 'Shield',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/overview',
        redirect: redirectWithQuery('/admin/dashboard'),
      },
      {
        path: 'platform/challenges',
        name: 'ChallengeManage',
        component: () => import('@/views/admin/ChallengeManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '靶场管理',
          icon: 'Settings',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/package-format',
        name: 'AdminChallengePackageFormat',
        component: () => import('@/views/admin/ChallengePackageFormat.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '题目包示例',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'admin/challenges',
        redirect: redirectWithQuery('/platform/challenges'),
      },
      {
        path: 'admin/challenges/package-format',
        redirect: redirectWithQuery('/platform/challenges/package-format'),
      },
      {
        path: 'platform/challenges/:id',
        name: 'AdminChallengeDetail',
        component: () => import('@/views/admin/ChallengeDetail.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '靶场详情',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'admin/challenges/:id',
        redirect: (to) => ({
          path: `/platform/challenges/${encodeURIComponent(String(to.params.id || ''))}`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'platform/challenges/:id/topology',
        name: 'AdminChallengeTopologyStudio',
        component: () => import('@/views/admin/ChallengeTopologyStudio.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '拓扑编排',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'admin/challenges/:id/topology',
        redirect: (to) => ({
          path: `/platform/challenges/${encodeURIComponent(String(to.params.id || ''))}/topology`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'platform/environment-templates',
        name: 'AdminEnvironmentTemplateLibrary',
        component: () => import('@/views/admin/EnvironmentTemplateLibrary.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '环境模板',
          icon: 'Server',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'admin/environment-templates',
        redirect: redirectWithQuery('/platform/environment-templates'),
      },
      {
        path: 'platform/challenges/:id/writeup',
        name: 'AdminChallengeWriteup',
        component: () => import('@/views/admin/ChallengeWriteup.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '题解管理',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'admin/challenges/:id/writeup',
        redirect: (to) => ({
          path: `/platform/challenges/${encodeURIComponent(String(to.params.id || ''))}/writeup`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'admin/contests',
        name: 'ContestManage',
        component: () => import('@/views/admin/ContestManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '竞赛管理',
          icon: 'Trophy',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contests',
        redirect: redirectWithQuery('/admin/contests'),
      },
      {
        path: 'admin/users',
        name: 'UserManage',
        component: () => import('@/views/admin/UserManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '用户管理',
          icon: 'Users',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/users',
        redirect: redirectWithQuery('/admin/users'),
      },
      {
        path: 'platform/images',
        name: 'ImageManage',
        component: () => import('@/views/admin/ImageManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '镜像管理',
          icon: 'Layers',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'admin/images',
        redirect: redirectWithQuery('/platform/images'),
      },
      {
        path: 'admin/integrity',
        name: 'CheatDetection',
        component: () => import('@/views/admin/CheatDetection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '作弊检测',
          icon: 'ScanEye',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/integrity',
        redirect: redirectWithQuery('/admin/integrity'),
      },
      {
        path: 'admin/audit',
        name: 'AuditLog',
        component: () => import('@/views/admin/AuditLog.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '审计日志',
          icon: 'ClipboardList',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/audit',
        redirect: redirectWithQuery('/admin/audit'),
      },
    ],
  },
  {
    path: '/401',
    name: 'Unauthorized',
    component: () => import('@/views/errors/UnauthorizedView.vue'),
    meta: { title: '未认证' },
  },
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/errors/ForbiddenView.vue'),
    meta: { title: '无权限' },
  },
  {
    name: 'NotFound',
    path: '/404',
    component: () => import('@/views/errors/NotFoundView.vue'),
    meta: { title: '页面不存在' },
  },
  {
    path: '/429',
    name: 'TooManyRequests',
    component: () => import('@/views/errors/TooManyRequestsView.vue'),
    meta: { title: '请求过多' },
  },
  {
    path: '/500',
    name: 'InternalServerError',
    component: () => import('@/views/errors/InternalServerErrorView.vue'),
    meta: { title: '系统错误' },
  },
  {
    path: '/502',
    name: 'BadGateway',
    component: () => import('@/views/errors/BadGatewayView.vue'),
    meta: { title: '网关异常' },
  },
  {
    path: '/503',
    name: 'ServiceUnavailable',
    component: () => import('@/views/errors/ServiceUnavailableView.vue'),
    meta: { title: '服务不可用' },
  },
  {
    path: '/504',
    name: 'GatewayTimeout',
    component: () => import('@/views/errors/GatewayTimeoutView.vue'),
    meta: { title: '服务超时' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

setupRouterGuards(router)

export default router
export { routes }
