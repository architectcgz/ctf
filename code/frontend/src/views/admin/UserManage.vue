<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-[#e6edf3]">用户管理</h1>
      <button class="rounded-lg bg-[#0891b2] px-4 py-2 text-sm font-medium text-white transition-colors duration-150 hover:bg-[#06b6d4]">
        创建用户
      </button>
    </div>

    <div class="flex gap-3">
      <input
        type="text"
        placeholder="搜索用户..."
        class="rounded-lg border border-[#30363d] bg-[#0f1117] px-3 py-2 text-sm text-[#e6edf3] placeholder:text-[#6e7681] focus:border-[#0891b2] focus:outline-none focus:ring-1 focus:ring-[#0891b2]/50"
      />
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-4 border-[#30363d] border-t-[#0891b2]"></div>
    </div>

    <div v-else class="overflow-hidden rounded-lg border border-[#30363d]">
      <table class="w-full">
        <thead class="bg-[#161b22]">
          <tr>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">学号</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">姓名</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">班级</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">角色</th>
            <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-[#8b949e]">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="user in mockUsers"
            :key="user.id"
            class="border-b border-[#21262d] transition-colors duration-100 hover:bg-[#1c2128]"
          >
            <td class="px-4 py-3 text-sm text-[#e6edf3]">{{ user.studentId }}</td>
            <td class="px-4 py-3 text-sm text-[#e6edf3]">{{ user.name }}</td>
            <td class="px-4 py-3 text-sm text-[#8b949e]">{{ user.className }}</td>
            <td class="px-4 py-3 text-sm text-[#8b949e]">{{ user.role }}</td>
            <td class="px-4 py-3 text-sm">
              <button class="text-[#0891b2] hover:text-[#06b6d4]">编辑</button>
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
