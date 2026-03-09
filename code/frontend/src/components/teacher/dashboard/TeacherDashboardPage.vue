<script setup lang="ts">
import type { TeacherClassItem, TeacherStudentItem } from '@/api/contracts'
import PageHeader from '@/components/common/PageHeader.vue'

const props = defineProps<{
  classes: TeacherClassItem[]
  students: TeacherStudentItem[]
  selectedClassName: string
  selectedClass: TeacherClassItem | null
  error: string | null
}>()

const emit = defineEmits<{
  retry: []
  openClassManagement: []
  openReportExport: []
}>()
</script>

<template>
  <div class="space-y-6">
    <PageHeader eyebrow="Teacher Flight Deck" title="教学介入台" description="查看班级状态。">
      <ElButton plain @click="emit('openClassManagement')">班级管理</ElButton>
      <ElButton type="primary" @click="emit('openReportExport')">导出报告</ElButton>
    </PageHeader>

    <section>
      <div class="rounded-[30px] border border-cyan-500/20 bg-[linear-gradient(145deg,rgba(8,47,73,0.82),rgba(15,23,42,0.94))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]">
        <div class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-100/75">
          <span>Intervention Deck</span>
          <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">{{ selectedClassName || '未选择班级' }}</span>
        </div>
        <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">
          {{ selectedClassName ? `${selectedClassName} 的教学概览` : '先选择一个班级' }}
        </h2>
        <p class="mt-3 text-sm leading-7 text-cyan-50/80">
          {{
            selectedClassName
              ? '查看当前班级人数、学生数量和教学入口。'
              : '选择班级后查看当前班级概览。'
          }}
        </p>

        <div class="mt-6 grid gap-3 md:grid-cols-3">
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">班级人数</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ selectedClass?.student_count || students.length }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前班级纳入视图的人数</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">管理班级</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ classes.length }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前教师可访问的班级总数</div>
          </div>
          <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
            <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">已加载学生</div>
            <div class="mt-2 text-2xl font-semibold text-white">{{ students.length }}</div>
            <div class="mt-2 text-sm text-cyan-50/70">当前班级已同步到页面的学生数</div>
          </div>
        </div>
      </div>
    </section>

    <div v-if="error" class="rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm text-red-600">
      {{ error }}
      <button type="button" class="ml-3 font-medium underline" @click="emit('retry')">重试</button>
    </div>
  </div>
</template>
