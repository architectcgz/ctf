import type { RouteRecordRaw } from 'vue-router'

export const errorRoutes: RouteRecordRaw[] = [
{
  path: '/401',
  name: 'Unauthorized',
  component: () => import('@/pages/errors/UnauthorizedRoutePage.vue'),
  meta: { title: '未认证' },
},
{
  path: '/403',
  name: 'Forbidden',
  component: () => import('@/pages/errors/ForbiddenRoutePage.vue'),
  meta: { title: '无权限' },
},
{
  name: 'NotFound',
  path: '/404',
  component: () => import('@/pages/errors/NotFoundRoutePage.vue'),
  meta: { title: '页面不存在' },
},
{
  path: '/429',
  name: 'TooManyRequests',
  component: () => import('@/pages/errors/TooManyRequestsRoutePage.vue'),
  meta: { title: '请求过多' },
},
{
  path: '/500',
  name: 'InternalServerError',
  component: () => import('@/pages/errors/InternalServerErrorRoutePage.vue'),
  meta: { title: '系统错误' },
},
{
  path: '/502',
  name: 'BadGateway',
  component: () => import('@/pages/errors/BadGatewayRoutePage.vue'),
  meta: { title: '网关异常' },
},
{
  path: '/503',
  name: 'ServiceUnavailable',
  component: () => import('@/pages/errors/ServiceUnavailableRoutePage.vue'),
  meta: { title: '服务不可用' },
},
{
  path: '/504',
  name: 'GatewayTimeout',
  component: () => import('@/pages/errors/GatewayTimeoutRoutePage.vue'),
  meta: { title: '服务超时' },
},
]
