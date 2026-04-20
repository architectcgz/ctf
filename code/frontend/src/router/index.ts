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
        path: 'academy/classes/:className/trend',
        name: 'TeacherClassTrend',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '班级趋势',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className/trend',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}/trend`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'academy/classes/:className/review',
        name: 'TeacherClassReview',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '教学复盘',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className/review',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}/review`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'academy/classes/:className/insights',
        name: 'TeacherClassInsights',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '学生洞察',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className/insights',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}/insights`,
          query: to.query,
          hash: to.hash,
        }),
      },
      {
        path: 'academy/classes/:className/intervention',
        name: 'TeacherClassIntervention',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '介入建议',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/classes/:className/intervention',
        redirect: (to) => ({
          path: `/academy/classes/${encodeURIComponent(String(to.params.className || ''))}/intervention`,
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
        path: 'academy/awd-reviews',
        name: 'TeacherAWDReviewIndex',
        component: () => import('@/views/teacher/TeacherAWDReviewIndex.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: 'AWD复盘',
          icon: 'ScanEye',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/awd-reviews',
        redirect: redirectWithQuery('/academy/awd-reviews'),
      },
      {
        path: 'academy/awd-reviews/:contestId',
        name: 'TeacherAWDReviewDetail',
        component: () => import('@/views/teacher/TeacherAWDReviewDetail.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: 'AWD复盘详情',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'teacher/awd-reviews/:contestId',
        redirect: (to) => ({
          path: `/academy/awd-reviews/${encodeURIComponent(String(to.params.contestId || ''))}`,
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

      // Platform Governance
      {
        path: 'platform/overview',
        name: 'PlatformOverview',
        component: () => import('@/views/platform/PlatformOverview.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '系统概览',
          icon: 'Shield',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes',
        name: 'PlatformClassManagement',
        component: () => import('@/views/platform/ClassManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '班级管理',
          icon: 'Users',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/students',
        name: 'PlatformStudentManagement',
        component: () => import('@/views/platform/StudentManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '学生管理',
          icon: 'GraduationCap',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className',
        name: 'PlatformClassStudents',
        component: () => import('@/views/teacher/TeacherClassStudents.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '班级学生',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className/trend',
        name: 'PlatformClassTrend',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '班级趋势',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className/review',
        name: 'PlatformClassReview',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '教学复盘',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className/insights',
        name: 'PlatformClassInsights',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '学生洞察',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className/intervention',
        name: 'PlatformClassIntervention',
        component: () => import('@/views/teacher/TeacherClassWorkspaceSection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '介入建议',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className/students/:studentId',
        name: 'PlatformStudentAnalysis',
        component: () => import('@/views/teacher/TeacherStudentAnalysis.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '学员分析',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/classes/:className/students/:studentId/review-archive',
        name: 'PlatformStudentReviewArchive',
        component: () => import('@/views/teacher/TeacherStudentReviewArchive.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '学生复盘归档',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/awd-reviews',
        name: 'PlatformAwdReviewIndex',
        component: () => import('@/views/platform/AWDReviewIndex.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: 'AWD复盘',
          icon: 'ScanEye',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/awd-reviews/:contestId',
        name: 'PlatformAwdReviewDetail',
        component: () => import('@/views/teacher/TeacherAWDReviewDetail.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: 'AWD复盘详情',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/instances',
        name: 'PlatformInstanceManagement',
        component: () => import('@/views/platform/InstanceManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '实例管理',
          icon: 'Server',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges',
        name: 'ChallengeManage',
        component: () => import('@/views/platform/ChallengeManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '题目管理',
          icon: 'Settings',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/package-format',
        name: 'PlatformChallengePackageFormat',
        component: () => import('@/views/platform/ChallengePackageFormat.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '题目包示例',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/imports',
        name: 'PlatformChallengeImportManage',
        component: () => import('@/views/platform/ChallengeImportManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '导入资源包',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/imports/:importId',
        name: 'PlatformChallengeImportPreview',
        component: () => import('@/views/platform/ChallengeImportPreview.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '导入预览',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/:id',
        name: 'PlatformChallengeDetail',
        component: () => import('@/views/platform/ChallengeDetail.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '题目详情',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/:id/topology',
        name: 'PlatformChallengeTopologyStudio',
        component: () => import('@/views/platform/ChallengeTopologyStudio.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '拓扑编排',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/environment-templates',
        name: 'PlatformEnvironmentTemplateLibrary',
        component: () => import('@/views/platform/EnvironmentTemplateLibrary.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '环境模板',
          icon: 'Server',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/awd-service-templates',
        name: 'PlatformAwdServiceTemplateLibrary',
        component: () => import('@/views/platform/AWDServiceTemplateLibrary.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: 'AWD 服务模板',
          icon: 'Shield',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/:id/writeup',
        name: 'PlatformChallengeWriteup',
        component: () => import('@/views/platform/ChallengeWriteup.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '题解管理',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/challenges/:id/writeup/view',
        name: 'PlatformChallengeWriteupView',
        component: () => import('@/views/platform/ChallengeWriteupView.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '查看题解',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contests',
        name: 'ContestManage',
        component: () => import('@/views/platform/ContestManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '竞赛目录',
          icon: 'Trophy',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contests/:id/edit',
        name: 'ContestEdit',
        component: () => import('@/views/platform/ContestEdit.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '竞赛工作室',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contests/:id/manage',
        name: 'ContestOperations',
        component: () => import('@/views/platform/ContestOperations.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '运维指挥中心',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contest-ops/environment',
        redirect: { name: 'PlatformContestOpsEnvironment' },
      },
      {
        path: 'platform/contest-ops/contests',
        name: 'PlatformContestOpsEnvironment',
        component: () => import('@/views/platform/ContestOperationsHub.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '环境管理',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contest-ops/traffic',
        name: 'PlatformContestOpsTraffic',
        component: () => import('@/views/platform/ContestOperationsHub.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '流量监控',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contest-ops/projector',
        name: 'PlatformContestOpsProjector',
        component: () => import('@/views/platform/ContestOperationsHub.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '大屏投射',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/contest-ops/scoreboard',
        name: 'PlatformContestOpsScoreboard',
        component: () => import('@/views/platform/ContestOperationsHub.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '排行榜',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/users',
        name: 'UserManage',
        component: () => import('@/views/platform/UserManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '用户管理',
          icon: 'Users',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/images',
        name: 'ImageManage',
        component: () => import('@/views/platform/ImageManage.vue'),
        meta: {
          requiresAuth: true,
          roles: ['teacher', 'admin'],
          title: '镜像管理',
          icon: 'Layers',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/integrity',
        name: 'CheatDetection',
        component: () => import('@/views/platform/CheatDetection.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '作弊检测',
          icon: 'ScanEye',
          contentLayout: 'bleed',
        },
      },
      {
        path: 'platform/audit',
        name: 'AuditLog',
        component: () => import('@/views/platform/AuditLog.vue'),
        meta: {
          requiresAuth: true,
          roles: ['admin'],
          title: '审计日志',
          icon: 'ClipboardList',
          contentLayout: 'bleed',
        },
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
  {
    path: '/ui-lab',
    name: 'UILab',
    component: () => import('@/views/UILab.vue'),
    meta: { title: 'UI 设计实验室' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

setupRouterGuards(router)

export default router
export { routes }
