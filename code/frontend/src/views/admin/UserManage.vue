<script setup lang="ts">
import { onMounted, useTemplateRef } from 'vue'

import AdminUserFormDialog from '@/components/admin/user/AdminUserFormDialog.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import { useAdminUsers } from '@/composables/useAdminUsers'

const {
  list,
  total,
  page,
  pageSize,
  loading,
  refresh,
  changePage,
  keyword,
  roleFilter,
  statusFilter,
  dialogOpen,
  dialogMode,
  saving,
  formDraft,
  importResult,
  openCreateDialog,
  openEditDialog,
  closeDialog,
  saveUser,
  removeUser,
  importUserFile,
} = useAdminUsers()

const importInput = useTemplateRef<HTMLInputElement>('importInput')

onMounted(() => {
  void refresh()
})

function triggerImport() {
  importInput.value?.click()
}

async function handleImportChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }
  try {
    await importUserFile(file)
  } finally {
    input.value = ''
  }
}

async function handleDelete(userId: string) {
  const user = list.value.find((item) => item.id === userId)
  if (!user || !window.confirm(`确定删除用户 ${user.username} 吗？`)) {
    return
  }
  await removeUser(user)
}

function handleDialogOpenChange(value: boolean) {
  if (!value) {
    closeDialog()
  }
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Admin"
      title="用户管理"
      description="当前页已接入真实的管理员用户接口，支持列表、筛选、创建、编辑、删除和 CSV 批量导入。"
    >
      <div class="flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="rounded-xl border border-border px-4 py-2 text-sm font-medium text-slate-100 transition hover:border-primary"
          @click="refresh"
        >
          刷新列表
        </button>
        <button
          type="button"
          class="rounded-xl border border-border px-4 py-2 text-sm font-medium text-slate-100 transition hover:border-primary"
          @click="triggerImport"
        >
          批量导入
        </button>
        <button
          type="button"
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
          @click="openCreateDialog"
        >
          创建用户
        </button>
      </div>
    </PageHeader>

    <input
      ref="importInput"
      type="file"
      accept=".csv,text/csv"
      class="hidden"
      @change="handleImportChange"
    />

    <SectionCard
      title="筛选与导入"
      subtitle="列表筛选和 CSV 导入都直接走主线后端接口，不再保留说明型降级卡片。"
    >
      <div class="grid gap-4 xl:grid-cols-[1.1fr_0.9fr]">
        <div class="grid gap-4 md:grid-cols-3">
          <label class="space-y-2">
            <span class="text-sm text-slate-300">关键词</span>
            <input
              v-model="keyword"
              type="text"
              class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
              placeholder="用户名 / 邮箱 / 班级"
            />
          </label>

          <label class="space-y-2">
            <span class="text-sm text-slate-300">角色</span>
            <select
              v-model="roleFilter"
              class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
            >
              <option value="all">全部角色</option>
              <option value="student">student</option>
              <option value="teacher">teacher</option>
              <option value="admin">admin</option>
            </select>
          </label>

          <label class="space-y-2">
            <span class="text-sm text-slate-300">状态</span>
            <select
              v-model="statusFilter"
              class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
            >
              <option value="all">全部状态</option>
              <option value="active">active</option>
              <option value="inactive">inactive</option>
              <option value="locked">locked</option>
              <option value="banned">banned</option>
            </select>
          </label>
        </div>

        <div class="rounded-2xl border border-border bg-surface-alt/60 p-5">
          <p class="text-sm font-medium text-slate-100">CSV 导入格式</p>
          <p class="mt-2 text-sm leading-6 text-slate-400">
            按列顺序上传：`username,password,email,class_name,role,status`。首行可带表头；已存在用户名会执行更新。
          </p>
          <div
            v-if="importResult"
            class="mt-4 rounded-xl border border-border bg-surface px-4 py-4 text-sm text-slate-300"
          >
            <p>
              创建 {{ importResult.created }}，更新 {{ importResult.updated }}，失败
              {{ importResult.failed }}
            </p>
            <ul v-if="importResult.errors?.length" class="mt-3 space-y-2 text-rose-300">
              <li
                v-for="item in importResult.errors.slice(0, 5)"
                :key="`${item.row}-${item.message}`"
              >
                第 {{ item.row }} 行：{{ item.message }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </SectionCard>

    <SectionCard
      title="用户列表"
      subtitle="所有数据都来自 `/admin/users`，编辑只暴露后端已支持的字段。"
    >
      <div v-if="loading && list.length === 0" class="flex justify-center py-10">
        <AppLoading>正在同步用户列表...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="list.length === 0"
        title="暂无用户"
        description="当前筛选条件下没有匹配用户。你可以调整筛选，或者直接创建新用户。"
        icon="UsersRound"
      >
        <template #action>
          <button
            type="button"
            class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
            @click="openCreateDialog"
          >
            创建第一个用户
          </button>
        </template>
      </AppEmpty>

      <div v-else class="space-y-5">
        <div class="overflow-hidden rounded-2xl border border-border">
          <table class="min-w-full divide-y divide-border">
            <thead class="bg-surface-alt/70">
              <tr class="text-left text-xs font-semibold uppercase tracking-[0.2em] text-slate-400">
                <th class="px-4 py-3">用户</th>
                <th class="px-4 py-3">角色</th>
                <th class="px-4 py-3">状态</th>
                <th class="px-4 py-3">班级</th>
                <th class="px-4 py-3">创建时间</th>
                <th class="px-4 py-3">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border bg-surface/70">
              <tr v-for="user in list" :key="user.id" class="transition hover:bg-surface-alt/60">
                <td class="px-4 py-4 align-top">
                  <p class="font-medium text-slate-100">{{ user.username }}</p>
                  <p class="mt-1 text-sm text-slate-400">{{ user.email || '未填写邮箱' }}</p>
                </td>
                <td class="px-4 py-4 align-top text-sm text-slate-300">
                  {{ user.roles.join(', ') }}
                </td>
                <td class="px-4 py-4 align-top">
                  <span
                    class="rounded-full bg-surface-alt px-3 py-1 text-xs font-semibold text-slate-200"
                  >
                    {{ user.status }}
                  </span>
                </td>
                <td class="px-4 py-4 align-top text-sm text-slate-300">
                  {{ user.class_name || '未分配' }}
                </td>
                <td class="px-4 py-4 align-top text-sm text-slate-300">
                  {{ new Date(user.created_at).toLocaleString('zh-CN') }}
                </td>
                <td class="px-4 py-4 align-top">
                  <div class="flex gap-2">
                    <button
                      type="button"
                      class="rounded-xl border border-border px-3 py-1.5 text-sm text-slate-100 transition hover:border-primary"
                      @click="openEditDialog(user)"
                    >
                      编辑
                    </button>
                    <button
                      type="button"
                      class="rounded-xl border border-rose-500/30 px-3 py-1.5 text-sm text-rose-300 transition hover:bg-rose-500/10"
                      @click="handleDelete(user.id)"
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          class="flex flex-col gap-3 text-sm text-slate-400 sm:flex-row sm:items-center sm:justify-between"
        >
          <span>共 {{ total }} 个用户</span>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="rounded-xl border border-border px-3 py-1.5 text-slate-200 transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-40"
              :disabled="page <= 1"
              @click="changePage(page - 1)"
            >
              上一页
            </button>
            <span>{{ page }} / {{ Math.max(1, Math.ceil(total / pageSize)) }}</span>
            <button
              type="button"
              class="rounded-xl border border-border px-3 py-1.5 text-slate-200 transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-40"
              :disabled="page >= Math.max(1, Math.ceil(total / pageSize))"
              @click="changePage(page + 1)"
            >
              下一页
            </button>
          </div>
        </div>
      </div>
    </SectionCard>

    <AdminUserFormDialog
      :open="dialogOpen"
      :mode="dialogMode"
      :draft="formDraft"
      :saving="saving"
      @update:open="handleDialogOpenChange"
      @save="saveUser"
    />
  </div>
</template>
