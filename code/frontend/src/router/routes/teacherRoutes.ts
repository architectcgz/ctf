import type { RouteRecordRaw } from 'vue-router'

import { redirectWithQuery } from './route-helpers'

export const teacherRoutes: RouteRecordRaw[] = [
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
]
