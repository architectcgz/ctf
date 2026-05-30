import type { RouteRecordRaw } from 'vue-router'

export const utilityRoutes: RouteRecordRaw[] = [
{
  path: '/:pathMatch(.*)*',
  redirect: '/404',
},
{
  path: '/ui-lab',
  name: 'UILab',
  component: () => import('@/pages/utility/UILabRoutePage.vue'),
  meta: { requiresAuth: true, roles: ['admin'], title: 'UI 设计实验室' },
},
]
