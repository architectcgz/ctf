<script setup lang="ts">
import { onMounted } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import AdminContestFormDialog from '@/components/admin/contest/AdminContestFormDialog.vue'
import AdminContestTable from '@/components/admin/contest/AdminContestTable.vue'
import { useAdminContests } from '@/composables/useAdminContests'

const {
  list,
  total,
  page,
  pageSize,
  loading,
  refresh,
  changePage,
  statusFilter,
  dialogOpen,
  mode,
  saving,
  formDraft,
  fieldLocks,
  statusOptions,
  openCreateDialog,
  openEditDialog,
  closeDialog,
  saveContest,
} = useAdminContests()

onMounted(() => {
  void refresh()
})

function handleDialogOpenChange(value: boolean) {
  if (!value) {
    closeDialog()
  }
}
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Admin Console"
      title="竞赛管理"
      description="当前页已接入真实的 /admin/contests 接口，支持列表、状态筛选、创建与编辑。删除入口保持关闭，避免继续暴露不存在的后端行为。"
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
          class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
          @click="openCreateDialog"
        >
          创建竞赛
        </button>
      </div>
    </PageHeader>

    <SectionCard
      title="接入边界"
      subtitle="这页只保留主线后端已经提供的能力；未提供的行为会显式说明，不再保留占位按钮。"
    >
      <div class="grid gap-3 md:grid-cols-3">
        <div class="rounded-2xl border border-emerald-500/25 bg-emerald-500/8 p-4">
          <p class="text-xs font-semibold uppercase tracking-[0.2em] text-emerald-300">已接入</p>
          <p class="mt-2 text-sm text-slate-200">竞赛列表、创建、编辑都走真实接口。</p>
        </div>
        <div class="rounded-2xl border border-amber-500/25 bg-amber-500/8 p-4">
          <p class="text-xs font-semibold uppercase tracking-[0.2em] text-amber-300">受后端约束</p>
          <p class="mt-2 text-sm text-slate-200">状态流转、时间字段可编辑范围与后端规则保持一致。</p>
        </div>
        <div class="rounded-2xl border border-slate-500/25 bg-slate-500/8 p-4">
          <p class="text-xs font-semibold uppercase tracking-[0.2em] text-slate-300">暂未暴露</p>
          <p class="mt-2 text-sm text-slate-200">删除接口主线未提供，页面不再展示假删除能力。</p>
        </div>
      </div>
    </SectionCard>

    <SectionCard
      title="竞赛列表"
      subtitle="支持按状态回看当前竞赛窗口，并在编辑时遵守后端的状态与时间流转限制。"
    >
      <template #header>
        <div class="flex flex-wrap items-center gap-3">
          <label class="text-sm text-slate-400" for="contest-status-filter">状态筛选</label>
          <select
            id="contest-status-filter"
            v-model="statusFilter"
            class="rounded-xl border border-border bg-surface px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-primary"
          >
            <option value="all">全部状态</option>
            <option value="draft">草稿</option>
            <option value="registering">报名中</option>
            <option value="running">进行中</option>
            <option value="frozen">已冻结</option>
            <option value="ended">已结束</option>
          </select>
        </div>
      </template>

      <div v-if="loading && list.length === 0" class="flex justify-center py-10">
        <AppLoading>正在同步竞赛列表...</AppLoading>
      </div>

      <AppEmpty
        v-else-if="list.length === 0"
        title="暂无竞赛"
        description="当前筛选条件下没有竞赛数据。你可以直接创建新竞赛，或者切换状态查看其他竞赛。"
        icon="Flag"
      >
        <template #action>
          <button
            type="button"
            class="rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
            @click="openCreateDialog"
          >
            创建第一场竞赛
          </button>
        </template>
      </AppEmpty>

      <AdminContestTable
        v-else
        :contests="list"
        :page="page"
        :page-size="pageSize"
        :total="total"
        @edit="openEditDialog"
        @change-page="changePage"
      />
    </SectionCard>

    <AdminContestFormDialog
      :open="dialogOpen"
      :mode="mode"
      :draft="formDraft"
      :saving="saving"
      :status-options="statusOptions"
      :field-locks="fieldLocks"
      @update:open="handleDialogOpenChange"
      @save="saveContest"
    />
  </div>
</template>
