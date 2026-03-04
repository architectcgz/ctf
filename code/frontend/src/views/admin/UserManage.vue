<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[var(--color-text-primary)]">用户管理</h1>
      <button class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors duration-150 hover:bg-[#06b6d4]">
        创建用户
      </button>
    </div>

    <div class="flex gap-3">
      <input
        type="text"
        placeholder="搜索用户..."
        class="rounded-lg border border-[var(--color-border-default)] bg-[var(--color-bg-base)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none focus:ring-1 focus:ring-[#0891b2]/50"
      />
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[var(--color-border-default)] border-t-[var(--color-primary)]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[var(--color-border-default)]">
      <table class="w-full">
        <thead class="bg-[var(--color-bg-surface)]">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">学号</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">姓名</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">班级</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">角色</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[var(--color-text-secondary)]">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="user in mockUsers"
            :key="user.id"
            class="border-b border-[var(--color-border-subtle)] transition-colors duration-100 hover:bg-[var(--color-bg-elevated)]"
          >
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ user.studentId }}</td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-primary)]">{{ user.name }}</td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{{ user.className }}</td>
            <td class="px-4 py-3 text-sm text-[var(--color-text-secondary)]">{{ user.role }}</td>
            <td class="px-4 py-3 text-sm">
              <button class="text-[var(--color-primary)] hover:text-[#06b6d4]">编辑</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const loading = ref(false)
const mockUsers = ref([
  { id: '1', studentId: '2022001', name: '张三', className: '信安2201', role: '学员' },
  { id: '2', studentId: '2022002', name: '李四', className: '信安2201', role: '学员' }
])
</script>
