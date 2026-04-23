import type { RouteRecordRaw } from 'vue-router'

export const errorRoutes: RouteRecordRaw[] = [
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
]
