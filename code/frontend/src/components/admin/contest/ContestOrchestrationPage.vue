<script setup lang="ts">
import { computed } from 'vue'
import { CalendarClock, Flag, RefreshCw, ShieldCheck, Trophy, UserPlus } from 'lucide-vue-next'

import type { ContestDetailData, ContestStatus } from '@/api/contracts'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import AdminContestTable from '@/components/admin/contest/AdminContestTable.vue'

type StatusFilter = 'all' | Extract<ContestStatus, 'draft' | 'registering' | 'running' | 'frozen' | 'ended'>

const props = defineProps<{
  list: ContestDetailData[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  statusFilter: StatusFilter
}>()

const emit = defineEmits<{
  refresh: []
  openCreateDialog: []
  updateStatusFilter: [value: StatusFilter]
  openEditDialog: [contest: ContestDetailData]
  changePage: [page: number]
}>()

const registeringCount = computed(() => props.list.filter((item) => item.status === 'registering').length)
const runningCount = computed(() => props.list.filter((item) => item.status === 'running').length)
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Contest Orchestration"
      title="赛事编排台"
      description="这页不再只是竞赛列表，而是把状态窗口、接入边界和赛事编排动作收进同一个管理员工作台。"
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
          class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
          @click="emit('openCreateDialog')"
        >
          <UserPlus class="h-4 w-4" />
          创建竞赛
        </button>
      </div>
    </PageHeader>

    <section class="grid gap-4 xl:grid-cols-[1.06fr_0.94fr]">
      <AppCard
        variant="hero"
        accent="warning"
        eyebrow="Contest Timeline"
        title="当前赛事编排视角"
        subtitle="这里主要看赛事窗口和状态流转。创建、编辑和状态筛选都接真实接口，但删除能力仍然保持关闭，不再留假按钮。"
      >
        <template #header>
          <span
            class="rounded-full border px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em]"
            style="border-color: rgba(210,153,34,0.24); background-color: rgba(210,153,34,0.12); color: var(--color-warning);"
          >
            真实接口
          </span>
        </template>

        <div class="grid gap-3 md:grid-cols-3">
          <AppCard variant="metric" accent="warning" eyebrow="当前页赛事" :title="String(list.length)" subtitle="当前筛选结果内的本页赛事数。" />
          <AppCard variant="metric" accent="warning" eyebrow="报名中" :title="String(registeringCount)" subtitle="便于快速判断当前公开报名窗口。" />
          <AppCard variant="metric" accent="warning" eyebrow="进行中" :title="String(runningCount)" subtitle="当前正处于比赛中的场次数量。" />
        </div>
      </AppCard>

      <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
        <AppCard variant="metric" accent="warning" eyebrow="赛事总量" :title="String(total)" subtitle="当前筛选条件下的赛事总数。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-amber-500/20 bg-amber-500/10 text-amber-300">
              <Trophy class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard variant="metric" accent="primary" eyebrow="接入边界" title="显式" subtitle="删除接口未提供，所以页面继续隐藏删除能力。">
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary">
              <ShieldCheck class="h-5 w-5" />
            </div>
          </template>
        </AppCard>

        <AppCard
          variant="metric"
          accent="primary"
          eyebrow="状态筛选"
          :title="statusFilter === 'all' ? '全部' : statusFilter"
          subtitle="用于快速切到某个赛事阶段做编排调整。"
        >
          <template #header>
            <div class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary">
              <CalendarClock class="h-5 w-5" />
            </div>
          </template>
        </AppCard>
      </div>
    </section>

    <section class="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
      <div class="space-y-6">
        <SectionCard title="状态窗口" subtitle="状态筛选和接入边界合在一起，先收敛当前编排范围。">
          <label class="space-y-2">
            <span class="text-sm text-slate-300">状态筛选</span>
            <select
              :value="statusFilter"
              class="w-full rounded-xl border border-border bg-surface px-3 py-3 text-sm text-slate-100 outline-none transition focus:border-primary"
              @change="emit('updateStatusFilter', ($event.target as HTMLSelectElement).value as StatusFilter)"
            >
              <option value="all">全部状态</option>
              <option value="draft">草稿</option>
              <option value="registering">报名中</option>
              <option value="running">进行中</option>
              <option value="frozen">已冻结</option>
              <option value="ended">已结束</option>
            </select>
          </label>

          <div class="mt-4 grid gap-3">
            <AppCard variant="action" accent="success" eyebrow="已接入" subtitle="竞赛列表、创建、编辑都走真实接口。">
              <template #default />
            </AppCard>
            <AppCard variant="action" accent="warning" eyebrow="受后端约束" subtitle="状态流转、时间字段可编辑范围与后端规则保持一致。">
              <template #default />
            </AppCard>
            <AppCard variant="action" accent="neutral" eyebrow="暂未暴露" subtitle="删除接口主线未提供，页面不再展示假删除能力。">
              <template #default />
            </AppCard>
          </div>
        </SectionCard>
      </div>

      <SectionCard title="赛事列表" subtitle="列表保留真实编辑能力，但页面语义切到“赛事编排”。">
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
              @click="emit('openCreateDialog')"
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
          @edit="emit('openEditDialog', $event)"
          @change-page="emit('changePage', $event)"
        />
      </SectionCard>
    </section>
  </div>
</template>
