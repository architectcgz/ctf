import type { RouteRecordRaw } from 'vue-router'

export const teacherRoutes: RouteRecordRaw[] = [
{
  path: 'academy/overview',
  name: 'TeacherDashboard',
  component: () => import('@/pages/teacher/TeacherDashboardRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '教学概览',
    icon: 'GraduationCap',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes',
  name: 'ClassManagement',
  component: () => import('@/pages/teacher/ClassManagementRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '班级管理',
    icon: 'Users',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/students',
  name: 'TeacherStudentManagement',
  component: () => import('@/pages/teacher/StudentManagementRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '学生管理',
    icon: 'GraduationCap',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className',
  name: 'TeacherClassStudents',
  component: () => import('@/pages/teacher/TeacherClassStudentsRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '班级学生',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className/trend',
  name: 'TeacherClassTrend',
  component: () => import('@/pages/teacher/TeacherClassWorkspaceSectionRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '班级趋势',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className/review',
  name: 'TeacherClassReview',
  component: () => import('@/pages/teacher/TeacherClassWorkspaceSectionRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '教学复盘',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className/insights',
  name: 'TeacherClassInsights',
  component: () => import('@/pages/teacher/TeacherClassWorkspaceSectionRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '学生洞察',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className/intervention',
  name: 'TeacherClassIntervention',
  component: () => import('@/pages/teacher/TeacherClassWorkspaceSectionRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '介入建议',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className/students/:studentId',
  name: 'TeacherStudentAnalysis',
  component: () => import('@/pages/teacher/TeacherStudentAnalysisRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '学员分析',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/classes/:className/students/:studentId/review-archive',
  name: 'TeacherStudentReviewArchive',
  component: () => import('@/pages/review-archive/StudentReviewArchiveRoutePage.vue'),
  props: (route) => ({
    className: String(route.params.className || ''),
    studentId: String(route.params.studentId || ''),
  }),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '学生复盘归档',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/awd-reviews',
  name: 'TeacherAWDReviewIndex',
  component: () => import('@/pages/awd-review/TeacherAwdReviewIndexRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: 'AWD复盘',
    icon: 'ScanEye',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/awd-reviews/:contestId',
  name: 'TeacherAWDReviewDetail',
  component: () => import('@/pages/awd-review/TeacherAwdReviewDetailRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: 'AWD复盘详情',
    contentLayout: 'bleed',
  },
},
{
  path: 'academy/instances',
  name: 'TeacherInstanceManagement',
  component: () => import('@/pages/teacher/InstanceManagementRoutePage.vue'),
  meta: {
    requiresAuth: true,
    roles: ['teacher', 'admin'],
    title: '实例管理',
    icon: 'Server',
    contentLayout: 'bleed',
  },
},
]
