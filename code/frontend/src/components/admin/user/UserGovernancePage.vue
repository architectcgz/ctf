<script setup lang="ts">
import { computed, useTemplateRef } from 'vue'
import { FileUp, RefreshCw, ShieldCheck, UserPlus, UsersRound, UserRoundCheck } from 'lucide-vue-next'

import type { AdminUserImportData, AdminUserListItem, UserStatus } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import type { UserRole } from '@/utils/constants'

type UserFilterRole = UserRole | 'all'
type UserFilterStatus = UserStatus | 'all'

const props = defineProps<{
  list: AdminUserListItem[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  keyword: string
  roleFilter: UserFilterRole
  statusFilter: UserFilterStatus
  importResult: AdminUserImportData | null
}>()

const emit = defineEmits<{
  refresh: []
  updateKeyword: [value: string]
  updateRoleFilter: [value: UserFilterRole]
  updateStatusFilter: [value: UserFilterStatus]
  openCreateDialog: []
  openEditDialog: [user: AdminUserListItem]
  deleteUser: [userId: string]
  changePage: [page: number]
  importFile: [file: File]
}>()

const importInput = useTemplateRef<HTMLInputElement>('importInput')
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const activeCount = computed(() => props.list.filter((item) => item.status === 'active').length)
const teacherCount = computed(() => props.list.filter((item) => item.roles.includes('teacher')).length)

function triggerImport(): void {
  importInput.value?.click()
}

async function handleImportChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  try {
    emit('importFile', file)
  } finally {
    input.value = ''
  }
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="User Governance"
      title="用户治理台"
      description="这里不再是通用表格页，而是围绕筛选、导入、状态治理和用户编排单独设计的管理员工作台。"
    >
      <div class="flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-slate-100 transition hover:border-primary"
          @click="emit('refresh')"
        >
          <RefreshCw class="h-4 w-4" />
          刷新列表
        </button>
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-slate-100 transition hover:border-primary"
          @click="triggerImport"
        >
          <FileUp class="h-4 w-4" />
          批量导入
        </button>
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
          @click="emit('openCreateDialog')"
        >
          <UserPlus class="h-4 w-4" />
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

    <section class="grid gap-4 xl:grid-cols-[1.06fr_0.94fr]">
      <div class="overflow-hidden rounded-[30px] border border-emerald-500/20 bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.16),transparent_42%),linear-gradient(145deg,rgba(2,6,23,0.98),rgba(15,23,42,0.92))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.24em] text-emerald-100/80">
          <span>Governance Deck</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">实时列表</span>
        </div>
        <h2 class="mt-4 text-3xl font-semibold tracking-tight text-white">当前治理视角</h2>
        <p class="mt-3 max-w-2xl text-sm leading-7 text-slate-200/78">
          先收敛筛选条件，再决定是治理单个账号还是走批量导入。导入结果和列表状态都聚合在同一页，不再分散成说明型卡片。
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-emerald-100/60">当前页用户</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ list.length }}</div>
            <div class="mt-2 text-sm text-slate-200/70">当前筛选结果内的本页样本数</div>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-emerald-100/60">活跃账号</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ activeCount }}</div>
            <div class="mt-2 text-sm text-slate-200/70">当前页处于 active 状态的用户数</div>
          </div>
          <div class="rounded-[22px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-emerald-100/60">教师角色</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ teacherCount }}</div>
            <div class="mt-2 text-sm text-slate-200/70">用于快速判断教学侧用户分布</div>
          </div>
        </div>
      </div>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">用户总量</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ total }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <UsersRound class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">当前筛选条件下的用户总数。</div>
        </article>

        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">导入回执</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">{{ importResult ? `${importResult.created}/${importResult.updated}` : '--' }}</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <FileUp class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">创建数 / 更新数。失败行会在左下方导入回执内展示。</div>
        </article>

        <article class="rounded-[24px] border border-border bg-surface/88 px-5 py-5 shadow-[0_18px_40px_var(--color-shadow-soft)]">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="text-[11px] font-semibold uppercase tracking-[0.2em] text-text-muted">治理状态</div>
              <div class="mt-2 text-2xl font-semibold text-text-primary">稳定</div>
            </div>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/12 text-primary">
              <ShieldCheck class="h-5 w-5" />
            </div>
          </div>
          <div class="mt-3 text-sm leading-6 text-text-secondary">创建、编辑、删除与导入都已经切到真实接口。</div>
        </article>
      </div>
    </section>

    <section class="grid gap-6 xl:grid-cols-[0.94fr_1.06fr]">
      <div class="space-y-6">
        <SectionCard title="筛选与导入" subtitle="先收敛治理视角，再做单个用户或批量动作。">
          <div class="grid gap-4">
            <label class="space-y-2">
              <span class="text-sm text-slate-300">关键词</span>
              <input
                :value="keyword"
                type="text"
                class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
                placeholder="用户名 / 邮箱 / 班级"
                @input="emit('updateKeyword', ($event.target as HTMLInputElement).value)"
              />
            </label>

            <div class="grid gap-4 md:grid-cols-2">
              <label class="space-y-2">
                <span class="text-sm text-slate-300">角色</span>
                <select
                  :value="roleFilter"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
                  @change="emit('updateRoleFilter', ($event.target as HTMLSelectElement).value as UserFilterRole)"
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
                  :value="statusFilter"
                  class="w-full rounded-xl border border-border bg-surface px-4 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
                  @change="emit('updateStatusFilter', ($event.target as HTMLSelectElement).value as UserFilterStatus)"
                >
                  <option value="all">全部状态</option>
                  <option value="active">active</option>
                  <option value="inactive">inactive</option>
                  <option value="locked">locked</option>
                  <option value="banned">banned</option>
                </select>
              </label>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="导入回执" subtitle="CSV 导入结果直接留在治理页，不再跳到别处查看。">
          <div class="rounded-2xl border border-border bg-surface-alt/60 p-5">
            <p class="text-sm font-medium text-slate-100">CSV 格式</p>
            <p class="mt-2 text-sm leading-6 text-slate-400">
              按列顺序上传：`username,password,email,class_name,role,status`。首行可带表头；已存在用户名会执行更新。
            </p>
          </div>

          <div
            v-if="importResult"
            class="mt-4 rounded-2xl border border-border bg-surface px-4 py-4 text-sm text-slate-300"
          >
            <p>创建 {{ importResult.created }}，更新 {{ importResult.updated }}，失败 {{ importResult.failed }}</p>
            <ul v-if="importResult.errors?.length" class="mt-3 space-y-2 text-rose-300">
              <li
                v-for="item in importResult.errors.slice(0, 5)"
                :key="`${item.row}-${item.message}`"
              >
                第 {{ item.row }} 行：{{ item.message }}
              </li>
            </ul>
          </div>
          <div v-else class="mt-4 rounded-2xl border border-dashed border-border px-4 py-6 text-sm text-text-secondary">
            还没有导入记录，执行一次 CSV 导入后会在这里看到回执。
          </div>
        </SectionCard>
      </div>

      <SectionCard title="用户编排" subtitle="当前页保留创建、编辑、删除和分页；布局改成治理视角下的用户卡片清单。">
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
              @click="emit('openCreateDialog')"
            >
              创建第一个用户
            </button>
          </template>
        </AppEmpty>

        <div v-else class="space-y-4">
          <article
            v-for="user in list"
            :key="user.id"
            class="rounded-[24px] border border-border bg-[linear-gradient(180deg,rgba(15,23,42,0.9),rgba(8,15,32,0.76))] px-5 py-5"
          >
            <div class="flex flex-wrap items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="font-semibold text-slate-100">{{ user.username }}</p>
                  <span class="rounded-full bg-surface-alt px-3 py-1 text-xs font-semibold text-slate-200">{{ user.status }}</span>
                </div>
                <p class="mt-2 text-sm text-slate-400">{{ user.email || '未填写邮箱' }}</p>
              </div>
              <div class="text-right text-sm text-slate-400">
                <div>{{ user.class_name || '未分配班级' }}</div>
                <div class="mt-1">{{ new Date(user.created_at).toLocaleString('zh-CN') }}</div>
              </div>
            </div>

            <div class="mt-4 flex flex-wrap items-center justify-between gap-3">
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="role in user.roles"
                  :key="`${user.id}-${role}`"
                  class="rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-medium text-primary"
                >
                  <UserRoundCheck class="mr-1 inline h-3.5 w-3.5" />
                  {{ role }}
                </span>
              </div>
              <div class="flex gap-2">
                <button
                  type="button"
                  class="rounded-xl border border-border px-3 py-1.5 text-sm text-slate-100 transition hover:border-primary"
                  @click="emit('openEditDialog', user)"
                >
                  编辑
                </button>
                <button
                  type="button"
                  class="rounded-xl border border-rose-500/30 px-3 py-1.5 text-sm text-rose-300 transition hover:bg-rose-500/10"
                  @click="emit('deleteUser', user.id)"
                >
                  删除
                </button>
              </div>
            </div>
          </article>

          <div class="flex flex-col gap-3 text-sm text-slate-400 sm:flex-row sm:items-center sm:justify-between">
            <span>共 {{ total }} 个用户</span>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="rounded-xl border border-border px-3 py-1.5 text-slate-200 transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="page <= 1"
                @click="emit('changePage', page - 1)"
              >
                上一页
              </button>
              <span>{{ page }} / {{ totalPages }}</span>
              <button
                type="button"
                class="rounded-xl border border-border px-3 py-1.5 text-slate-200 transition hover:border-primary disabled:cursor-not-allowed disabled:opacity-40"
                :disabled="page >= totalPages"
                @click="emit('changePage', page + 1)"
              >
                下一页
              </button>
            </div>
          </div>
        </div>
      </SectionCard>
    </section>
  </div>
</template>
