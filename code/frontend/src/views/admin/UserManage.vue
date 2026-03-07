<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Admin"
      title="用户管理"
      description="用户列表、导入、编辑与状态管理的页面框架已保留，但当前后端主线还没有提供 /admin/users 路由。这里明确展示为降级态，避免继续使用假数据误导联调。"
    >
      <div
        class="inline-flex items-center gap-2 rounded-full border border-amber-500/30 bg-amber-500/10 px-3 py-1 text-xs font-medium text-amber-200"
      >
        <span class="h-2 w-2 rounded-full bg-amber-300" />
        后端接口待补齐
      </div>
    </PageHeader>

    <div class="grid gap-4 lg:grid-cols-[1.4fr,1fr]">
      <SectionCard
        title="当前状态"
        subtitle="移除了静态 mock 用户表。只有在后端补齐管理员用户路由后，才会恢复真实的列表、搜索和编辑交互。"
      >
        <AppEmpty
          icon="UsersRound"
          title="暂时无法展示用户列表"
          description="当前主线仅提供管理员 dashboard、审计日志、镜像、靶场和竞赛管理接口，尚未提供 /api/v1/admin/users 相关能力。"
        >
          <template #action>
            <div class="flex flex-wrap items-center justify-center gap-2 text-xs text-text-muted">
              <span class="rounded-full border border-border px-3 py-1">列表查询</span>
              <span class="rounded-full border border-border px-3 py-1">批量导入</span>
              <span class="rounded-full border border-border px-3 py-1">状态编辑</span>
              <span class="rounded-full border border-border px-3 py-1">角色调整</span>
            </div>
          </template>
        </AppEmpty>
      </SectionCard>

      <SectionCard
        title="后续接入清单"
        subtitle="等后端补齐接口后，页面可以按下面的顺序直接恢复成真实管理页。"
      >
        <ol class="space-y-3 text-sm text-text-secondary">
          <li class="rounded-xl border border-border bg-elevated px-4 py-3">
            <div class="font-medium text-text-primary">1. 用户列表与分页</div>
            <div class="mt-1">
              接入 `GET /admin/users`，替换当前降级卡片，恢复搜索、筛选和分页。
            </div>
          </li>
          <li class="rounded-xl border border-border bg-elevated px-4 py-3">
            <div class="font-medium text-text-primary">2. 创建 / 编辑用户</div>
            <div class="mt-1">
              接入 `POST /admin/users` 与 `PUT /admin/users/:id`，补齐用户表单与角色配置。
            </div>
          </li>
          <li class="rounded-xl border border-border bg-elevated px-4 py-3">
            <div class="font-medium text-text-primary">3. 删除与批量导入</div>
            <div class="mt-1">接入删除和导入接口，补齐批量导入反馈、失败行提示和审计记录联动。</div>
          </li>
        </ol>
      </SectionCard>
    </div>

    <SectionCard title="接口缺口" subtitle="这是当前页面明确依赖但后端主线尚未提供的接口集合。">
      <div class="grid gap-3 md:grid-cols-2">
        <div
          v-for="endpoint in missingEndpoints"
          :key="endpoint.path"
          class="rounded-xl border border-border bg-elevated px-4 py-3"
        >
          <div class="font-mono text-xs text-cyan-300">
            {{ endpoint.method }} {{ endpoint.path }}
          </div>
          <div class="mt-2 text-sm text-text-secondary">{{ endpoint.description }}</div>
        </div>
      </div>
    </SectionCard>
  </div>
</template>

<script setup lang="ts">
import AppEmpty from '@/components/common/AppEmpty.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

interface MissingEndpoint {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  path: string
  description: string
}

const missingEndpoints: MissingEndpoint[] = [
  {
    method: 'GET',
    path: '/api/v1/admin/users',
    description: '管理员查询用户列表、搜索和分页。',
  },
  {
    method: 'POST',
    path: '/api/v1/admin/users',
    description: '管理员创建用户并分配初始角色、班级信息。',
  },
  {
    method: 'PUT',
    path: '/api/v1/admin/users/:id',
    description: '管理员更新用户状态、角色和资料。',
  },
  {
    method: 'DELETE',
    path: '/api/v1/admin/users/:id',
    description: '管理员删除或停用用户。',
  },
  {
    method: 'POST',
    path: '/api/v1/admin/users/import',
    description: '管理员通过文件批量导入用户。',
  },
]
</script>
