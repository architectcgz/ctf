<script setup lang="ts">
import { ChevronRight, Users } from 'lucide-vue-next'

import type { TeacherClassItem } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import MetricCard from '@/components/common/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  loading: boolean
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openDashboard: []
  openReportExport: []
  openClass: [className: string]
}>()
</script>

<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Class Directory"
      title="班级管理"
    >
      <ElButton plain @click="emit('openDashboard')">教学概览</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section class="grid gap-4 md:grid-cols-2">
      <div class="grid gap-3 md:grid-cols-2">
        <MetricCard label="班级数量" :value="classes.length" hint="当前可管理班级总数" accent="primary" />
        <MetricCard
          label="学生总量"
          :value="classes.reduce((sum, item) => sum + (item.student_count || 0), 0)"
          hint="各班级学生数汇总"
          accent="success"
        />
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>

    <SectionCard title="班级列表" subtitle="查看当前可管理的班级。">
      <div v-if="loading" class="space-y-3">
        <div v-for="index in 5" :key="index" class="h-14 animate-pulse rounded-2xl bg-[var(--color-bg-base)]" />
      </div>

      <AppEmpty
        v-else-if="classes.length === 0"
        icon="Users"
        title="暂无班级"
        description="当前教师账号下还没有可访问的班级。"
      />

      <div v-else>
        <ElTable
          :data="classes"
          row-key="name"
          class="teacher-class-table"
          empty-text="暂无班级"
        >
          <ElTableColumn prop="name" label="班级名称" min-width="260">
            <template #default="{ row }">
              <div class="py-1">
                <div class="font-semibold text-text-primary">{{ row.name }}</div>
                <div class="mt-1 text-sm text-text-secondary">查看班级学生名单。</div>
              </div>
            </template>
          </ElTableColumn>

          <ElTableColumn prop="student_count" label="学生数" width="140" align="center">
            <template #default="{ row }">
              <span class="text-sm font-medium text-text-primary">{{ row.student_count || 0 }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn label="操作" width="160" align="right">
            <template #default="{ row }">
              <ElButton type="primary" plain @click="emit('openClass', row.name)">
                进入班级
                <ChevronRight class="ml-1 h-4 w-4" />
              </ElButton>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
    </SectionCard>
  </div>
</template>

<style scoped>
:deep(.teacher-class-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-expanded-cell-bg-color: transparent;
  --el-table-header-bg-color: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  --el-table-border-color: var(--color-border-default);
  --el-table-row-hover-bg-color: color-mix(in srgb, var(--color-primary) 8%, var(--color-bg-surface));
  --el-table-text-color: var(--color-text-primary);
  --el-table-header-text-color: var(--color-text-secondary);
}

:deep(.teacher-class-table th.el-table__cell) {
  background: color-mix(in srgb, var(--color-bg-surface) 84%, var(--color-bg-base));
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

:deep(.teacher-class-table td.el-table__cell),
:deep(.teacher-class-table th.el-table__cell) {
  border-bottom-color: var(--color-border-default);
}

:deep(.teacher-class-table .el-table__inner-wrapper::before) {
  display: none;
}
</style>
