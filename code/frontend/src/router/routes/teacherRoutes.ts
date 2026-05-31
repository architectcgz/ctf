import type { RouteRecordRaw } from 'vue-router'

import { redirectWithQuery } from './route-helpers'

function encodeRouteParam(value: unknown): string {
  return encodeURIComponent(String(value || ''))
}

function redirectWithDynamicAcademyPath(
  buildPath: (route: Parameters<NonNullable<RouteRecordRaw['redirect']>>[0]) => string
): NonNullable<RouteRecordRaw['redirect']> {
  return (to) => ({
    path: buildPath(to),
    query: to.query,
    hash: to.hash,
  })
}

const teacherLegacyRedirectDefinitions = [
  {
    legacyPath: 'teacher/dashboard',
    redirect: redirectWithQuery('/academy/overview'),
  },
  {
    legacyPath: 'teacher/classes',
    redirect: redirectWithQuery('/academy/classes'),
  },
  {
    legacyPath: 'teacher/students',
    redirect: redirectWithQuery('/academy/students'),
  },
  {
    legacyPath: 'teacher/classes/:className',
    redirect: redirectWithDynamicAcademyPath(
      (to) => `/academy/classes/${encodeRouteParam(to.params.className)}`
    ),
  },
  {
    legacyPath: 'teacher/classes/:className/trend',
    redirect: redirectWithDynamicAcademyPath(
      (to) => `/academy/classes/${encodeRouteParam(to.params.className)}/trend`
    ),
  },
  {
    legacyPath: 'teacher/classes/:className/review',
    redirect: redirectWithDynamicAcademyPath(
      (to) => `/academy/classes/${encodeRouteParam(to.params.className)}/review`
    ),
  },
  {
    legacyPath: 'teacher/classes/:className/insights',
    redirect: redirectWithDynamicAcademyPath(
      (to) => `/academy/classes/${encodeRouteParam(to.params.className)}/insights`
    ),
  },
  {
    legacyPath: 'teacher/classes/:className/intervention',
    redirect: redirectWithDynamicAcademyPath(
      (to) => `/academy/classes/${encodeRouteParam(to.params.className)}/intervention`
    ),
  },
  {
    legacyPath: 'teacher/classes/:className/students/:studentId',
    redirect: redirectWithDynamicAcademyPath(
      (to) =>
        `/academy/classes/${encodeRouteParam(to.params.className)}/students/${encodeRouteParam(to.params.studentId)}`
    ),
  },
  {
    legacyPath: 'teacher/classes/:className/students/:studentId/review-archive',
    redirect: redirectWithDynamicAcademyPath(
      (to) =>
        `/academy/classes/${encodeRouteParam(to.params.className)}/students/${encodeRouteParam(to.params.studentId)}/review-archive`
    ),
  },
  {
    legacyPath: 'teacher/awd-reviews',
    redirect: redirectWithQuery('/academy/awd-reviews'),
  },
  {
    legacyPath: 'teacher/awd-reviews/:contestId',
    redirect: redirectWithDynamicAcademyPath(
      (to) => `/academy/awd-reviews/${encodeRouteParam(to.params.contestId)}`
    ),
  },
  {
    legacyPath: 'teacher/instances',
    redirect: redirectWithQuery('/academy/instances'),
  },
] as const

export const teacherLegacyRedirectAllowlist = teacherLegacyRedirectDefinitions.map(
  ({ legacyPath }) => legacyPath
)

const teacherLegacyRedirectRoutes: RouteRecordRaw[] = teacherLegacyRedirectDefinitions.map(
  ({ legacyPath, redirect }) => ({
    path: legacyPath,
    redirect,
  })
)

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
...teacherLegacyRedirectRoutes,
]
